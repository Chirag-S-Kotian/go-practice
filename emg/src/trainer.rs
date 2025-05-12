use anyhow::Result;
use tch::{nn, Device, Tensor};

pub struct Trainer {
    learning_rate: f64,
    epochs: usize,
    batch_size: usize,
}

impl Trainer {
    pub fn new(learning_rate: f64, epochs: usize, batch_size: usize) -> Self {
        Trainer {
            learning_rate,
            epochs,
            batch_size,
        }
    }

    pub fn train(&self, model: &mut crate::model::Transformer, dataset: &crate::dataset::Dataset, device: Device) -> Result<()> {
        let optimizer = nn::Adam::default()
            .build(model.var_store(), self.learning_rate)?;
        
        for epoch in 0..self.epochs {
            let (inputs, targets) = dataset.get_batch(self.batch_size)?;
            let inputs = inputs.to_device(device);
            let targets = targets.to_device(device);
            
            let logits = model.forward(&inputs)?;
            let loss = logits
                .cross_entropy_for_logits(&targets)?;
            
            optimizer.zero_grad();
            loss.backward();
            optimizer.step();
            
            println!("Epoch {}: Loss = {}", epoch + 1, f64::from(loss));
        }
        
        Ok(())
    }
}