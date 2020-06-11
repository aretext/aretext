use std::io::{Error, ErrorKind, Result};

#[derive(Clone, Copy)]
enum State {
    Valid,
    Invalid,
    AwaitingOneByte,
    AwaitingTwoBytesA,
    AwaitingTwoBytesB,
    AwaitingTwoBytesC,
    AwaitingThreeBytesA,
    AwaitingThreeBytesB,
    AwaitingThreeBytesC,
}

/// Streaming validator for UTF-8 text.
pub struct Utf8Validator {
    processed_count: usize,
    state: State,
}

impl Utf8Validator {
    pub fn new() -> Self {
        Utf8Validator {
            processed_count: 0,
            state: State::Valid,
        }
    }

    /// Check that bytes are valid UTF-8.
    /// Returns an io::Error with kind set to InvalidData otherwise.
    pub fn validate(&mut self, bytes: &[u8]) -> Result<()> {
        // Fast path for ASCII text
        if let State::Valid = self.state {
            if Self::is_ascii(bytes) {
                self.processed_count += bytes.len();
                return Ok(());
            }
        }

        // Slow path for non-ASCII
        for b in bytes.iter() {
            self.process_byte(b)?;
            self.processed_count += 1;
        }
        Ok(())
    }

    /// Check that the bytestream ends in a valid state.
    /// Call this when there are no more bytes to process.
    pub fn validate_end(&self) -> Result<()> {
        match self.state {
            State::Valid => Ok(()),
            _ => {
                let msg = format!("Expected continuation byte at end of stream");
                Err(Error::new(ErrorKind::InvalidData, msg))
            }
        }
    }

    fn is_ascii(bytes: &[u8]) -> bool {
        bytes.iter().all(|b| (b >> 7) == 0)
    }

    fn process_byte(&mut self, b: &u8) -> Result<()> {
        // See http://bjoern.hoehrmann.de/utf-8/decoder/dfa/
        self.state = match (self.state, b) {
            (State::Valid, 0x00..=0x7f) => State::Valid,
            (State::Valid, 0xc2..=0xdf) => State::AwaitingOneByte,
            (State::Valid, 0xe1..=0xec) | (State::Valid, 0xee..=0xef) => State::AwaitingTwoBytesA,
            (State::Valid, 0xe0) => State::AwaitingTwoBytesB,
            (State::Valid, 0xed) => State::AwaitingTwoBytesC,
            (State::Valid, 0xf0) => State::AwaitingThreeBytesA,
            (State::Valid, 0xf1..=0xf3) => State::AwaitingThreeBytesB,
            (State::Valid, 0xf4) => State::AwaitingThreeBytesC,
            (State::AwaitingOneByte, 0x80..=0xbf) => State::Valid,
            (State::AwaitingTwoBytesA, 0x80..=0xbf)
            | (State::AwaitingTwoBytesB, 0xa0..=0xbf)
            | (State::AwaitingTwoBytesC, 0x80..=0x9f) => State::AwaitingOneByte,
            (State::AwaitingThreeBytesA, 0x90..=0xbf)
            | (State::AwaitingThreeBytesB, 0x80..=0xbf)
            | (State::AwaitingThreeBytesC, 0x80..=0xbf) => State::AwaitingTwoBytesA,
            _ => State::Invalid,
        };

        match self.state {
            State::Invalid => {
                let msg = format!("Invalid byte at position {}", self.processed_count);
                Err(Error::new(ErrorKind::InvalidData, msg))
            }
            _ => Ok(()),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use proptest::prelude::*;
    use std::str;

    #[test]
    fn it_validates_single_byte_chars() {
        let mut v = Utf8Validator::new();
        let b = "abcdefghijklmnopqrstuvwxyz";
        v.validate(b.as_bytes())
            .and_then(|_| v.validate_end())
            .expect("Input is valid");
    }

    #[test]
    fn it_validates_multi_byte_chars() {
        let mut v = Utf8Validator::new();
        let b = "丂丄丅丆丏";
        v.validate(b.as_bytes())
            .and_then(|_| v.validate_end())
            .expect("Input is valid");
    }

    #[test]
    fn it_validates_multi_byte_char_split_between_reads() {
        let mut v = Utf8Validator::new();
        let s = "¢ह€한";
        for b in s.as_bytes() {
            v.validate(&[*b]).expect("Input is valid");
        }
        v.validate_end().expect("Input is valid");
    }

    #[test]
    fn it_rejects_invalid_start_byte() {
        let mut v = Utf8Validator::new();
        let b = vec![0b11111111];
        assert!(v.validate(&b).and_then(|_| v.validate_end()).is_err());
    }

    #[test]
    fn it_rejects_too_many_continuation_chars() {
        let mut v = Utf8Validator::new();
        let b = vec![0b11000000, 0b10000000, 0b10000000];
        assert!(v.validate(&b).and_then(|_| v.validate_end()).is_err());
    }

    #[test]
    fn it_rejects_missing_continuation_chars() {
        let mut v = Utf8Validator::new();
        let b = vec![0b11100000, 0b10000000, 0b00000000];
        assert!(v.validate(&b).and_then(|_| v.validate_end()).is_err());
    }

    #[test]
    fn it_rejects_missing_continuation_chars_at_end() {
        let mut v = Utf8Validator::new();
        let b = vec![0b11110000, 0b10000000];
        assert!(v.validate(&b).and_then(|_| v.validate_end()).is_err());
    }

    #[test]
    fn it_rejects_overlong_sequences() {
        let mut v = Utf8Validator::new();
        let b = vec![0b11000000, 0b10000000]; // could be encoded in one byte, so reject
        assert!(v.validate(&b).and_then(|_| v.validate_end()).is_err());
    }

    #[test]
    fn it_rejects_too_large_codepoints() {
        let mut v = Utf8Validator::new();
        let b = vec![0b11110111, 0b10111111, 0b10111111, 0b10111111];
        assert!(v.validate(&b).and_then(|_| v.validate_end()).is_err());
    }

    proptest! {
        #[test]
        fn it_matches_stdlib_str_behavior_no_splits(b: Vec<u8>) {
            let mut v = Utf8Validator::new();
            let expect_valid = str::from_utf8(&b).is_ok();
            let result = v.validate(&b).and_then(|_| v.validate_end());
            assert_eq!(result.is_ok(), expect_valid);
        }

        #[test]
        fn it_matches_stdlib_str_behavior_with_splits(b: Vec<u8>) {
            let mut v = Utf8Validator::new();
            let expect_valid = str::from_utf8(&b).is_ok();
            let mut valid = true;
            for byte in b.iter() {
                valid &= v.validate(&[*byte]).is_ok();
            }
            valid &= v.validate_end().is_ok();
            assert_eq!(valid, expect_valid);
        }
    }
}
