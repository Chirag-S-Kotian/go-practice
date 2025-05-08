use std::collections::HashMap;

pub struct GameState {
    pub current_node_id: String,
    pub inventory: Vec<String>,
    pub flags: HashMap<String, bool>,
    pub player_name: String,
}

impl GameState {
    pub fn new(start_node_id: String, player_name: String) -> Self {
        GameState {
            current_node_id: start_node_id,
            inventory: Vec::new(),
            flags: HashMap::new(),
            player_name,
        }
    }
}