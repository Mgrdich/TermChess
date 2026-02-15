# Functional Specification: Custom RL Agent

- **Roadmap Item:** Custom RL Agent - A reinforcement-learning-trained bot as the top-tier AI opponent
- **Status:** Draft
- **Author:** Mgrdich

---

## 1. Overview and Rationale (The "Why")

**Problem:** The current Hard bot uses minimax with position evaluation, which provides a ceiling on difficulty. Experienced players need a stronger, more challenging opponent that plays in a more "human-like" strategic style.

**Solution:** Train a custom reinforcement learning chess agent using AlphaZero-style self-play, providing three difficulty tiers (1500, 2000, 2200 ELO) that offer progressively challenging gameplay.

**Secondary Goal:** This serves as a hands-on deep learning project, with training runnable on Mac ARM (Apple Silicon) using PyTorch MPS.

**Success Metrics:**
- RL bots provide meaningfully different challenge levels
- Users can select and play against RL bots in both standard play and Bot vs Bot mode
- Training pipeline runs successfully on Mac ARM

---

## 2. Functional Requirements (The "What")

### 2.1 RL Bot Selection

- **As a** user, **I want to** select an RL bot difficulty, **so that** I can play against a challenging AI trained via deep learning.
  - **Acceptance Criteria:**
    - [ ] Three RL bots appear in the bot selection list: "RL Intermediate (1500)", "RL Advanced (2000)", "RL Master (2200)"
    - [ ] RL bots are listed alongside existing bots (Easy, Medium, Hard) in a simple list
    - [ ] Selecting an RL bot starts a game against that bot

### 2.2 Info/Description Page

- **As a** user, **I want to** view more information about RL bot difficulties, **so that** I understand what each level means.
  - **Acceptance Criteria:**
    - [ ] A shortcut/option (e.g., "Press 'i' for info") is visible when viewing bot selection
    - [ ] The info page displays ELO ratings and a brief description of each RL bot tier
    - [ ] User can close the info page and return to bot selection

### 2.3 RL Bot Gameplay

- **As a** user, **I want** the RL bot to display quirky thinking messages, **so that** gameplay feels engaging.
  - **Acceptance Criteria:**
    - [ ] RL bots display humorous/quirky messages during move calculation
    - [ ] Messages are themed for RL bots (e.g., "Neural pathways firing...", "Consulting the matrix...", "Adjusting weights...")

### 2.4 Bot vs Bot Mode

- **As a** user, **I want to** pit RL bots against other bots, **so that** I can watch and compare their play styles.
  - **Acceptance Criteria:**
    - [ ] RL bots are available in Bot vs Bot mode selection
    - [ ] Any combination works (e.g., RL Master vs Hard, RL Intermediate vs RL Advanced)

### 2.5 Error Handling

- **As a** user, **I want** clear feedback if an RL model is unavailable, **so that** I understand why I can't start the game.
  - **Acceptance Criteria:**
    - [ ] RL bot options are always visible in the selection list
    - [ ] If the RL model file is missing or corrupted, an error message is displayed when user attempts to start (e.g., "RL model not found. Please reinstall or check model files.")
    - [ ] User cannot proceed until the issue is resolved

### 2.6 Training Pipeline (Developer-Facing)

- **As a** developer, **I want** a training pipeline that runs on Mac ARM, **so that** I can train and iterate on RL models locally.
  - **Acceptance Criteria:**
    - [ ] Training uses PyTorch with MPS (Metal) acceleration on Mac ARM
    - [ ] AlphaZero-style self-play training loop is implemented
    - [ ] Checkpoints are saved at intervals to produce different strength models (1500, 2000, 2200 ELO)
    - [ ] Training script is documented with instructions for running on Mac ARM
    - [ ] Trained model files are portable and device-agnostic

### 2.7 Model Inference (Runtime)

- **As a** user, **I want** RL bots to work on any supported platform, **so that** I can play regardless of my device.
  - **Acceptance Criteria:**
    - [ ] Inference auto-detects available backend (MPS, CUDA, or CPU)
    - [ ] No platform-specific configuration required from user
    - [ ] Model runs on macOS, Linux, and Windows

---

## 3. Scope and Boundaries

### In-Scope

- Three RL bot difficulties (1500, 2000, 2200 ELO) integrated into existing bot selection
- Info page accessible via shortcut explaining RL bot tiers
- Quirky thinking messages for RL bots
- RL bots available in Bot vs Bot mode
- Error handling for missing/corrupted models (visible options, error on selection)
- PyTorch MPS training pipeline for Mac ARM
- AlphaZero-style self-play training implementation
- Cross-platform inference (auto-detects MPS/CUDA/CPU)

### Out-of-Scope

- UCI Engine Integration (separate roadmap item)
- Cloud/distributed training infrastructure
- Pre-trained Leela Chess Zero integration (training from scratch for learning purposes)
- Mobile or non-Mac training support
- All completed Phase 1-5 features
