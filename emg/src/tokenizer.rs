use anyhow::Result;
use std::collections::HashMap;

#[derive(Clone)]
pub struct Tokenizer {
    vocab: HashMap<String, u32>,
    reverse_vocab: HashMap<u32, String>,
    vocab_size: usize,
}

impl Tokenizer {
    pub fn new(max_vocab: usize) -> Result<Self> {
        let mut vocab = HashMap::new();
        let mut reverse_vocab = HashMap::new();
        vocab.insert("<PAD>".to_string(), 0);
        reverse_vocab.insert(0, "<PAD>".to_string());
        vocab.insert("<UNK>".to_string(), 1);
        reverse_vocab.insert(1, "<UNK>".to_string());
        Ok(Tokenizer {
            vocab,
            reverse_vocab,
            vocab_size: 2,
        })
    }

    pub fn build_vocab(&mut self, text: &str, max_vocab: usize) -> Result<()> {
        let words: Vec<&str> = text.split_whitespace().collect();
        for word in words {
            if !self.vocab.contains_key(word) && self.vocab_size < max_vocab {
                self.vocab.insert(word.to_string(), self.vocab_size as u32);
                self.reverse_vocab.insert(self.vocab_size as u32, word.to_string());
                self.vocab_size += 1;
            }
        }
        Ok(())
    }

    pub fn encode(&self, text: &str) -> Result<Vec<u32>> {
        let words: Vec<&str> = text.split_whitespace().collect();
        let mut tokens = Vec::new();
        for word in words {
            tokens.push(*self.vocab.get(word).unwrap_or(&1));
        }
        Ok(tokens)
    }

    pub fn decode(&self, tokens: &[u32]) -> Result<String> {
        let words: Vec<String> = tokens
            .iter()
            .map(|&t| self.reverse_vocab.get(&t).unwrap_or(&"<UNK>".to_string()).clone())
            .collect();
        Ok(words.join(" "))
    }

    pub fn vocab_size(&self) -> usize {
        self.vocab_size
    }
}