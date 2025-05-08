use regex::Regex;
use std::collections::HashMap;

#[derive(Debug, Clone)]
pub enum StateCondition {
    HasItem(String),
    FlagIsTrue(String),
}

#[derive(Debug, Clone)]
pub struct Branch {
    pub id: String,
    pub required_input_pattern: Option<Regex>,
    pub required_state_conditions: Vec<StateCondition>,
    pub target_node_id: String,
    pub choice_text: Option<String>,
}

#[derive(Debug, Clone)]
pub enum TextSource {
    Static(String),
    Template(String),
    LMPrompt(String),
}

#[derive(Debug, Clone)]
pub enum Effect {
    SetFlag(String, bool),
    AddItem(String),
    RemoveItem(String),
}

#[derive(Debug, Clone)]
pub struct StoryNode {
    pub id: String,
    pub text_source: TextSource,
    pub branches: Vec<Branch>,
    pub entry_effects: Vec<Effect>,
}

#[derive(Debug)]
pub struct Story {
    pub nodes: HashMap<String, StoryNode>,
    pub start_node_id: String,
}

impl Story {
    pub fn get_node(&self, node_id: &str) -> Option<&StoryNode> {
        self.nodes.get(node_id)
    }
}