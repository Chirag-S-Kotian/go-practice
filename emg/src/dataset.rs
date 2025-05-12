use anyhow::Result;
use std::fs::File;
use std::io::{BufRead, BufReader};
use rand::seq::SliceRandom;
use rand::thread_rng;
use tch::Tensor;

pub struct Dataset {
    tokens: Vec<u32>,
    seq_len: usize,
}

impl Dataset {
    pub fn new(file_path: &str, tokenizer: &crate::tokenizer::Tokenizer) -> Result<Self> {
        let file = File::open(file_path)?;
        let reader = BufReader::new(file);
        let mut text = String::new();
        for line in reader.lines() {
            text.push_str(&line?);
            text.push(' ');
        }
        
        let mut new_tokenizer = tokenizer.clone();
        new_tokenizer.build_vocab(&text, 1000)?;
        let tokens = new_tokenizer.encode(&text)?;
        
        Ok(Dataset {
            tokens,
            seq_len: 128,
        })
    }

    pub fn get_batch(&self, batch_size: usize) -> Result<(Tensor, Tensor)> {
        let mut rng = thread_rng();
        let mut indices: Vec<usize> = (0..self.tokens.len() - self.seq_len).collect();
        indices.shuffle(&mut rng);
        
        let mut inputs = Vec::new();
        let mut targets = Vec::new();
        
        for i in indices.iter().take(batch_size) {
            let slice = &self.tokens[*i..*i + self.seq_len];
            inputs.extend_from_slice(slice);
            targets.extend_from_slice(&self.tokens[*i + 1..*i + self.seq_len + 1]);
        }
        
        let input_tensor = Tensor::of_slice(&inputs)
            .reshape(&[batch_size as i64, self.seq_len as i64]);
        let target_tensor = Tensor::of_slice(&targets)
            .reshape(&[batch_size as i64, self.seq_len as i64]);
        
        Ok((input_tensor, target_tensor))
    }
}