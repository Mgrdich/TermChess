"""
Monte Carlo Tree Search (MCTS) for AlphaZero-style Chess

This module implements MCTS with neural network guidance, following the AlphaZero
algorithm. Unlike traditional MCTS which uses random rollouts, this version uses
a neural network to both guide the search (policy) and evaluate positions (value).

Key Concepts:
-------------
1. UCB (Upper Confidence Bound): Balances exploitation (moves with high value)
   and exploration (moves that haven't been tried much).

2. PUCT (Predictor + UCT): A variant of UCB that uses neural network policy
   priors to guide exploration toward promising moves.

3. No rollouts: Instead of playing random games to the end, we use the neural
   network's value prediction directly.

Algorithm Flow:
---------------
For each simulation:
1. SELECT: Start at root, use UCB formula to pick children until we reach a leaf
2. EXPAND: At the leaf, query the neural network and create children for all legal moves
3. EVALUATE: Use the neural network's value output (no rollout needed)
4. BACKUP: Propagate the value back up the tree, negating at each level

After all simulations, return visit counts as move probabilities.
"""

import math
from typing import Dict, List, Optional, Tuple

import chess
import numpy as np
import torch
import torch.nn.functional as F

from board_encoder import encode_board_tensor, get_device, NUM_HISTORY_POSITIONS
from model import ChessNet, POLICY_OUTPUT_SIZE


class MCTSNode:
    """
    A node in the Monte Carlo search tree.

    Each node represents a chess position and stores statistics used for UCB
    selection. The tree is built incrementally during search.

    Attributes:
        state: The chess board position at this node
        parent: Parent node (None for root)
        parent_move: The move that led from parent to this node
        children: Dict mapping moves to child nodes
        visit_count (N): Number of times this node has been visited
        total_value (W): Sum of all backed-up values through this node
        prior_probability (P): Neural network's prior probability for this move
    """

    def __init__(
        self,
        state: chess.Board,
        parent: Optional["MCTSNode"] = None,
        parent_move: Optional[chess.Move] = None,
        prior: float = 0.0
    ):
        """
        Initialize an MCTS node.

        Args:
            state: Chess board position at this node
            parent: Parent node in the tree (None for root)
            parent_move: Move that led from parent to this node
            prior: Neural network's prior probability for this move
        """
        self.state = state
        self.parent = parent
        self.parent_move = parent_move
        self.prior_probability = prior

        # Statistics for UCB calculation
        self.visit_count = 0  # N(s, a)
        self.total_value = 0.0  # W(s, a)

        # Children: move -> MCTSNode
        # None means unexpanded, empty dict means terminal (no legal moves)
        self.children: Optional[Dict[chess.Move, "MCTSNode"]] = None

    @property
    def q_value(self) -> float:
        """
        Q(s, a) = W(s, a) / N(s, a)

        The average value of this node. Returns 0 if not visited yet.
        This represents our current estimate of how good this position is.
        """
        if self.visit_count == 0:
            return 0.0
        return self.total_value / self.visit_count

    def is_expanded(self) -> bool:
        """Check if this node has been expanded (children created)."""
        return self.children is not None

    def is_terminal(self) -> bool:
        """
        Check if this is a terminal position (game over).

        A terminal position is one where the game has ended:
        - Checkmate
        - Stalemate
        - Draw by insufficient material, repetition, 50-move rule, etc.
        """
        return self.state.is_game_over()

    def get_terminal_value(self) -> float:
        """
        Get the value of a terminal position from the perspective of the side to move.

        Returns:
            +1.0 if the side to move has won (shouldn't happen - they can't move)
            -1.0 if the side to move has lost (they are checkmated)
            0.0 for draws (stalemate, insufficient material, etc.)

        Understanding checkmate in python-chess:
        - When a position is checkmate, board.turn returns the CHECKMATED side
        - board.outcome().winner returns the WINNING side (the one who delivered mate)
        - So if board.turn == BLACK and outcome.winner == WHITE, black is mated

        Example: After 1.f3 e5 2.g4 Qh4#
        - board.turn == WHITE (white would move next, but can't)
        - outcome.winner == BLACK (black delivered the checkmate)
        - Value from white's perspective: -1.0 (white lost)
        """
        result = self.state.outcome()
        if result is None:
            return 0.0  # Game not over (shouldn't happen if is_terminal() is True)

        if result.winner is None:
            # Draw (stalemate, insufficient material, 50-move rule, repetition)
            return 0.0

        # In checkmate: the side to move is the one who is mated
        # The winner is the opposite side
        # From the side-to-move's perspective, they lost
        if result.winner != self.state.turn:
            # The opponent (not side to move) won
            # This means the side to move is checkmated -> they lost
            return -1.0
        else:
            # The side to move won - this is unusual in standard chess
            # (can't happen in checkmate, but included for completeness)
            return 1.0


class MCTS:
    """
    Monte Carlo Tree Search with neural network guidance.

    This implements the AlphaZero-style MCTS which uses a neural network to:
    1. Provide prior probabilities P(s,a) for moves (from policy head)
    2. Evaluate positions without rollouts (from value head)

    The search balances exploitation (high Q values) with exploration
    (high prior, low visit count) using the PUCT formula.

    Dirichlet noise is added to the root node's priors to ensure exploration,
    following the AlphaZero paper.

    Attributes:
        model: ChessNet neural network for position evaluation
        c_puct: Exploration constant (higher = more exploration)
        num_simulations: Number of MCTS simulations per search
        device: Torch device for neural network inference
        dirichlet_alpha: Alpha parameter for Dirichlet noise at root
        dirichlet_epsilon: Weight of Dirichlet noise vs network prior
    """

    def __init__(
        self,
        model: ChessNet,
        c_puct: float = 1.5,
        num_simulations: int = 100,
        device: Optional[torch.device] = None,
        dirichlet_alpha: float = 0.3,
        dirichlet_epsilon: float = 0.5
    ):
        """
        Initialize MCTS.

        Args:
            model: ChessNet instance for position evaluation
            c_puct: Exploration constant for UCB formula (default: 1.5)
            num_simulations: Number of simulations per search (default: 100)
            device: Torch device (default: auto-detect)
            dirichlet_alpha: Alpha for Dirichlet noise at root (default: 0.3)
            dirichlet_epsilon: Noise weight at root (default: 0.5)
        """
        self.model = model
        self.c_puct = c_puct
        self.num_simulations = num_simulations
        self.device = device if device is not None else get_device()
        self.dirichlet_alpha = dirichlet_alpha
        self.dirichlet_epsilon = dirichlet_epsilon

        # Ensure model is on the correct device and in eval mode
        self.model = self.model.to(self.device)
        self.model.eval()

    def _move_to_policy_index(self, move: chess.Move) -> int:
        """
        Convert a chess move to a policy index (0-4095).

        The policy is represented as a 64x64 matrix flattened to 4096 values,
        where index = from_square * 64 + to_square.

        This is a simple encoding that handles all moves including promotions
        (though promotion type is not encoded - we treat all promotions the same).

        Args:
            move: A chess.Move object

        Returns:
            Integer index in [0, 4095]
        """
        return move.from_square * 64 + move.to_square

    def _get_legal_move_mask(self, board: chess.Board) -> torch.Tensor:
        """
        Create a mask of legal moves for the current position.

        Args:
            board: Chess board position

        Returns:
            Tensor of shape [4096] with 1.0 for legal moves, 0.0 for illegal
        """
        mask = torch.zeros(POLICY_OUTPUT_SIZE, device=self.device)
        for move in board.legal_moves:
            idx = self._move_to_policy_index(move)
            mask[idx] = 1.0
        return mask

    def _evaluate_position(
        self, board: chess.Board,
        history: Optional[List[chess.Board]] = None
    ) -> Tuple[Dict[chess.Move, float], float]:
        """
        Use the neural network to evaluate a position.

        Returns both the policy (move priors) and value (position evaluation).
        The policy is masked to only include legal moves and normalized to sum to 1.

        Args:
            board: Chess board position to evaluate
            history: Optional list of previous board states for encoding

        Returns:
            Tuple of:
            - Dict mapping legal moves to their prior probabilities (sum to 1)
            - Value in [-1, 1] from the perspective of the side to move
        """
        # Encode the board with history and add batch dimension
        state_tensor = encode_board_tensor(board, self.device, history=history)
        state_tensor = state_tensor.unsqueeze(0)  # [1, C, 8, 8]

        # Run neural network inference (no gradient needed)
        with torch.no_grad():
            policy_logits, value = self.model(state_tensor)

        # policy_logits: [1, 4096], value: [1, 1]
        policy_logits = policy_logits.squeeze(0)  # [4096]
        value = value.squeeze().item()  # scalar

        # Mask illegal moves by setting their logits to -inf
        legal_mask = self._get_legal_move_mask(board)
        masked_logits = policy_logits.clone()
        masked_logits[legal_mask == 0] = float('-inf')

        # Apply softmax to get probabilities (only legal moves will have non-zero prob)
        probs = F.softmax(masked_logits, dim=0)

        # Convert to dictionary of move -> probability
        move_priors: Dict[chess.Move, float] = {}
        for move in board.legal_moves:
            idx = self._move_to_policy_index(move)
            move_priors[move] = probs[idx].item()

        return move_priors, value

    def _ucb_score(self, parent: MCTSNode, child: MCTSNode) -> float:
        """
        Calculate the UCB (Upper Confidence Bound) score for a child node.

        Formula (PUCT - Predictor Upper Confidence Tree):
            UCB(s, a) = Q(s, a) + c_puct * P(s, a) * sqrt(N(parent)) / (1 + N(child))

        Where:
            - Q(s, a) = average value of this action (exploitation term)
            - c_puct = exploration constant
            - P(s, a) = prior probability from neural network
            - N = visit count

        The exploration term is large when:
            - Prior probability is high (NN thinks this move is good)
            - Parent has been visited many times
            - This child hasn't been visited much

        Args:
            parent: Parent node
            child: Child node to score

        Returns:
            UCB score (higher = should explore this child)
        """
        # Exploitation term: average value of this child
        q_value = child.q_value

        # Exploration term: encourages visiting less-explored nodes
        # with high prior probability
        exploration = (
            self.c_puct
            * child.prior_probability
            * math.sqrt(parent.visit_count)
            / (1 + child.visit_count)
        )

        return q_value + exploration

    def _select_child(self, node: MCTSNode) -> MCTSNode:
        """
        Select the child with the highest UCB score.

        This is the "selection" phase of MCTS. We pick the child that
        best balances exploitation (high Q) and exploration (high prior,
        low visit count).

        Args:
            node: Parent node (must be expanded)

        Returns:
            Child node with highest UCB score
        """
        assert node.children is not None, "Node must be expanded"
        assert len(node.children) > 0, "Node must have children"

        best_score = float('-inf')
        best_child = None

        for move, child in node.children.items():
            score = self._ucb_score(node, child)
            if score > best_score:
                best_score = score
                best_child = child

        assert best_child is not None
        return best_child

    def _get_node_history(self, node: MCTSNode) -> List[chess.Board]:
        """
        Reconstruct the board history for a node by walking parent pointers
        and prepending the root's external history.

        Returns:
            List of board states leading up to (but not including) this node's position.
        """
        # Collect states from root down to this node's parent
        path_states: List[chess.Board] = []
        current = node.parent
        while current is not None:
            path_states.append(current.state)
            current = current.parent
        path_states.reverse()  # root first, then down to parent

        # Prepend external history from before the search
        full_history = list(self._root_history) + path_states

        # Only need the last NUM_HISTORY_POSITIONS
        return full_history[-NUM_HISTORY_POSITIONS:] if full_history else []

    def _expand(self, node: MCTSNode) -> float:
        """
        Expand a leaf node by creating children for all legal moves.

        This is the "expansion" phase of MCTS. We query the neural network
        to get move priors and the position value.

        Args:
            node: Leaf node to expand

        Returns:
            Value of this position from the neural network
        """
        # Get neural network evaluation with history context
        history = self._get_node_history(node)
        move_priors, value = self._evaluate_position(node.state, history=history)

        # Create children for all legal moves
        node.children = {}
        for move, prior in move_priors.items():
            # Make the move to get the new state
            new_state = node.state.copy()
            new_state.push(move)

            # Create child node
            child = MCTSNode(
                state=new_state,
                parent=node,
                parent_move=move,
                prior=prior
            )
            node.children[move] = child

        return value

    def _backpropagate(self, node: MCTSNode, value: float) -> None:
        """
        Backpropagate the value from a leaf node up to the root.

        This is the "backup" phase of MCTS. We update visit counts and
        total values for all nodes from the leaf to the root.

        Important: The value is negated at each level because chess is
        a zero-sum game. A position that's good for white (+1) is bad
        for black (-1), and vice versa.

        Args:
            node: Starting node (the leaf that was just expanded/evaluated)
            value: Value to propagate (from perspective of node's side to move)
        """
        current = node
        current_value = value

        while current is not None:
            current.visit_count += 1
            current.total_value += current_value

            # Negate value for parent (opponent's perspective)
            current_value = -current_value
            current = current.parent

    def _run_simulation(self, root: MCTSNode) -> None:
        """
        Run a single MCTS simulation.

        A simulation consists of:
        1. SELECT: Walk down the tree using UCB until reaching a leaf
        2. EXPAND: Create children for the leaf (if not terminal)
        3. EVALUATE: Get value from neural network (or terminal value)
        4. BACKUP: Propagate value up to root

        Args:
            root: Root node to start the simulation from
        """
        node = root

        # === SELECTION PHASE ===
        # Walk down the tree using UCB until we reach a leaf
        while node.is_expanded() and not node.is_terminal():
            if len(node.children) == 0:
                # No legal moves (shouldn't happen if not terminal, but be safe)
                break
            node = self._select_child(node)

        # === EXPANSION AND EVALUATION PHASE ===
        if node.is_terminal():
            # Terminal node: get value from game result
            value = node.get_terminal_value()
        else:
            # Non-terminal leaf: expand and get value from neural network
            value = self._expand(node)

        # === BACKUP PHASE ===
        # Propagate the value back up the tree
        self._backpropagate(node, value)

    def _add_dirichlet_noise(self, root: MCTSNode) -> None:
        """
        Add Dirichlet noise to the root node's children priors.

        This ensures exploration at the root, preventing the search from
        collapsing to always picking the same moves. Following AlphaZero:
            P(s, a) = (1 - epsilon) * P(s, a) + epsilon * Dir(alpha)

        Args:
            root: Root node (must be expanded with children)
        """
        if root.children is None or len(root.children) == 0:
            return

        moves = list(root.children.keys())
        noise = np.random.dirichlet([self.dirichlet_alpha] * len(moves))

        for i, move in enumerate(moves):
            child = root.children[move]
            child.prior_probability = (
                (1 - self.dirichlet_epsilon) * child.prior_probability
                + self.dirichlet_epsilon * noise[i]
            )

    def search(
        self, board: chess.Board,
        history: Optional[List[chess.Board]] = None
    ) -> Dict[chess.Move, int]:
        """
        Perform MCTS search from the given position.

        Runs `num_simulations` simulations, then returns the visit count
        for each legal move. Higher visit count = better move.

        Dirichlet noise is added to the root node's priors to ensure
        sufficient exploration during self-play training.

        Args:
            board: Chess board position to search from
            history: Optional list of previous board states for encoding context

        Returns:
            Dict mapping each legal move to its visit count
        """
        # Store history for use during tree traversal
        self._root_history = history or []

        # Create root node
        root = MCTSNode(state=board.copy())

        # Expand root (needed to initialize children)
        if not root.is_terminal():
            self._expand(root)
            # Add Dirichlet noise to root for exploration
            self._add_dirichlet_noise(root)

        # Run simulations
        for _ in range(self.num_simulations):
            self._run_simulation(root)

        # Return visit counts for each move
        if root.children is None or len(root.children) == 0:
            return {}

        return {move: child.visit_count for move, child in root.children.items()}

    def get_action_probabilities(
        self,
        board: chess.Board,
        temperature: float = 1.0,
        history: Optional[List[chess.Board]] = None
    ) -> Tuple[List[chess.Move], np.ndarray]:
        """
        Get move probabilities after MCTS search.

        This is the main method used during self-play training. It returns
        moves and their probabilities based on visit counts.

        Temperature controls exploration:
        - temp=0: Deterministic, always picks the most visited move
        - temp=1: Proportional to visit counts (more exploration)
        - temp>1: Even more exploration (flatter distribution)

        Args:
            board: Chess board position to search from
            temperature: Temperature for probability distribution (default: 1.0)
            history: Optional list of previous board states

        Returns:
            Tuple of:
            - List of legal moves
            - numpy array of probabilities (same order as moves, sums to 1)
        """
        # Run MCTS search
        visit_counts = self.search(board, history=history)

        if len(visit_counts) == 0:
            # No legal moves (game over)
            return [], np.array([])

        # Extract moves and their visit counts
        moves = list(visit_counts.keys())
        counts = np.array([visit_counts[m] for m in moves], dtype=np.float64)

        if temperature == 0:
            # Deterministic: pick the best move(s)
            # In case of ties, we put all probability on the tied moves
            best_count = counts.max()
            probs = (counts == best_count).astype(np.float64)
            probs /= probs.sum()
        else:
            # Apply temperature
            # P(a) = N(a)^(1/temp) / sum(N(b)^(1/temp))
            # For numerical stability, we work in log space when temp != 1
            if temperature == 1.0:
                probs = counts / counts.sum()
            else:
                # N^(1/temp) = exp(log(N) / temp)
                # Add small epsilon to avoid log(0)
                log_counts = np.log(counts + 1e-10)
                scaled_log_counts = log_counts / temperature
                # Subtract max for numerical stability
                scaled_log_counts -= scaled_log_counts.max()
                probs = np.exp(scaled_log_counts)
                probs /= probs.sum()

        return moves, probs

    def select_move(
        self,
        board: chess.Board,
        temperature: float = 1.0,
        history: Optional[List[chess.Board]] = None
    ) -> chess.Move:
        """
        Select a move using MCTS search.

        This is a convenience method that performs search and returns a single
        move, either sampled from the probability distribution (temp > 0)
        or deterministically (temp = 0).

        Args:
            board: Chess board position
            temperature: Temperature for move selection

        Returns:
            Selected chess move

        Raises:
            ValueError: If no legal moves available
        """
        moves, probs = self.get_action_probabilities(board, temperature, history=history)

        if len(moves) == 0:
            raise ValueError("No legal moves available")

        if temperature == 0:
            # Deterministic: pick move with highest probability
            # (which will be the most visited move)
            best_idx = np.argmax(probs)
            return moves[best_idx]
        else:
            # Sample from the distribution
            idx = np.random.choice(len(moves), p=probs)
            return moves[idx]
