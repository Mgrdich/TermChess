//! A minimal single-line text input (replaces `bubbles/textinput`).

/// A single-line text buffer with a cursor and character limit.
#[derive(Clone, Debug)]
pub struct TextField {
    value: String,
    cursor: usize,
    pub placeholder: String,
    pub char_limit: usize,
}

impl TextField {
    /// Creates a text field with the given placeholder and character limit.
    pub fn new(placeholder: &str, char_limit: usize) -> Self {
        TextField {
            value: String::new(),
            cursor: 0,
            placeholder: placeholder.to_string(),
            char_limit,
        }
    }

    /// Returns the current value.
    pub fn value(&self) -> &str {
        &self.value
    }

    /// Replaces the value and moves the cursor to the end.
    pub fn set_value(&mut self, v: &str) {
        self.value = v.to_string();
        self.cursor = self.value.chars().count();
    }

    /// Inserts a character at the cursor, honoring the char limit.
    pub fn insert(&mut self, c: char) {
        if self.char_limit > 0 && self.value.chars().count() >= self.char_limit {
            return;
        }
        let byte_idx = self
            .value
            .char_indices()
            .nth(self.cursor)
            .map(|(i, _)| i)
            .unwrap_or(self.value.len());
        self.value.insert(byte_idx, c);
        self.cursor += 1;
    }

    /// Deletes the character before the cursor.
    pub fn backspace(&mut self) {
        if self.cursor == 0 {
            return;
        }
        let remove_idx = self
            .value
            .char_indices()
            .nth(self.cursor - 1)
            .map(|(i, _)| i);
        if let Some(i) = remove_idx {
            self.value.remove(i);
            self.cursor -= 1;
        }
    }

    /// Moves the cursor one character left.
    pub fn left(&mut self) {
        if self.cursor > 0 {
            self.cursor -= 1;
        }
    }

    /// Moves the cursor one character right.
    pub fn right(&mut self) {
        if self.cursor < self.value.chars().count() {
            self.cursor += 1;
        }
    }

    /// Renders the field as a display string (value or placeholder, with a cursor block).
    pub fn display(&self) -> String {
        if self.value.is_empty() {
            format!("> {}", self.placeholder)
        } else {
            format!("> {}_", self.value)
        }
    }
}
