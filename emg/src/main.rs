use anyhow::Result;
use tch::Device;

mod model;
mod tokenizer;
mod dataset;
mod trainer;
mod util;

fn main() -> Result<()> {
    let device = util::cuda_if_available();
    println!("Using device: {:?}", device);

    // Initialize tokenizer
    let tokenizer = tokenizer::Tokenizer::new(1000)?;
    
    // Load dataset
    let dataset = dataset::Dataset::new("data/corpus.txt", &tokenizer)?;
    
    // Initialize model
    let config = model::ModelConfig {
        vocab_size: tokenizer.vocab_size(),
        n_layers: 2,
        n_heads: 4,
        d_model: 128,
        d_ff: 512,
        max_seq_len: 128,
    };
    let mut model = model::Transformer::new(config, device)?;
    
    // Train model
    let trainer = trainer::Trainer::new(0.001, 10, 32);
    trainer.train(&mut model, &dataset, device)?;
    
    // Generate text
    let prompt = "It is a truth";
    let output = model.generate(prompt, &tokenizer, 50, device)?;
    println!("Generated: {}", output);
    
    Ok(())
}