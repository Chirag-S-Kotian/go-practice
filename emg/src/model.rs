use anyhow::Result;
use tch::{nn, Device, Tensor, Kind};

pub struct ModelConfig {
    pub vocab_size: usize,
    pub n_layers: usize,
    pub n_heads: usize,
    pub d_model: usize,
    pub d_ff: usize,
    pub max_seq_len: usize,
}

pub struct Transformer {
    vs: nn::VarStore,
    embedding: nn::Embedding,
    pos_embedding: nn::Embedding,
    layers: Vec<nn::TransformerDecoderLayer>,
    final_layer: nn::Linear,
    config: ModelConfig,
}

impl Transformer {
    pub fn new(config: ModelConfig, device: Device) -> Result<Self> {
        let vs = nn::VarStore::new(device);
        let p = &vs.root();
        
        let embedding = nn::embedding(
            p / "embedding",
            config.vocab_size as i64,
            config.d_model as i64,
            Default::default(),
        );
        
        let pos_embedding = nn::embedding(
            p / "pos_embedding",
            config.max_seq_len as i64,
            config.d_model as i64,
            Default::default(),
        );
        
        let mut layers = Vec::new();
        for i in 0..config.n_layers {
            let layer = nn::transformer_decoder_layer(
                p / format!("layer{}", i),
                config.d_model as i64,
                config.n_heads as i64,
                config.d_ff as i64,
                0.1,
                Default::default(),
            );
            layers.push(layer);
        }
        
        let final_layer = nn::linear(
            p / "final_layer",
            config.d_model as i64,
            config.vocab_size as i64,
            Default::default(),
        );
        
        Ok(Transformer {
            vs,
            embedding,
            pos_embedding,
            layers,
            final_layer,
            config,
        })
    }

    pub fn forward(&self, input: &Tensor) -> Result<Tensor> {
        let seq_len = input.size()[1];
        let input_emb = input.apply(&self.embedding)?;
        let pos_ids = Tensor::arange(seq_len, (Kind::Int64, input.device()));
        let pos_emb = pos_ids.apply(&self.pos_embedding)?;
        let mut x = input_emb + pos_emb;
        
        for layer in &self.layers {
            x = x.apply_t(layer, false)?;
        }
        
        let output = x.apply(&self.final_layer)?;
        Ok(output)
    }

    pub fn generate(&self, prompt: &str, tokenizer: &crate::tokenizer::Tokenizer, max_len: i64, device: Device) -> Result<String> {
        let mut input = tokenizer.encode(prompt)?;
        let mut output = Vec::new();
        
        for _ in 0..max_len {
            let input_tensor = Tensor::of_slice(&input).to_device(device);
            let logits = self.forward(&input_tensor.unsqueeze(0))?.squeeze(0);
            let next_token = logits
                .slice(1, -1, None, 1)
                .argmax(None, false)
                .int64_value(&[]);
            output.push(next_token as u32);
            input.push(next_token as u32);
            if input.len() > self.config.max_seq_len {
                input.drain(0..1);
            }
        }
        
        Ok(tokenizer.decode(&output)?)
    }

    pub fn var_store(&self) -> &nn::VarStore {
        &self.vs
    }
}