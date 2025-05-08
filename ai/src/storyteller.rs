use regex::Regex;

use lazy_static::lazy_static;

use crate::story_data::{Story, StoryNode, Branch, StateCondition, TextSource, Effect};
use crate::game_state::GameState;
use crate::language_model::LanguageModel;

pub struct Storyteller {
    story: Story,
    game_state: GameState,
    language_model: LanguageModel,
}

impl Storyteller {
    pub fn new(story: Story, player_name: String) -> Result<Self, String> {
        let start_node_id = story.start_node_id.clone();
        let game_state = GameState::new(start_node_id, player_name);
        let language_model = LanguageModel::new()?;

        Ok(Storyteller {
            story,
            game_state,
            language_model,
        })
    }

    pub fn start(&mut self) -> Result<String, String> {
        self.apply_current_node_entry_effects()?;
        self.generate_current_node_text()
    }

    pub fn handle_input(&mut self, input: &str) -> Result<String, String> {
        let processed_input = input.trim().to_lowercase();

        let current_node = self.story.get_node(&self.game_state.current_node_id)
            .ok_or_else(|| format!("Error: Current node '{}' not found!", self.game_state.current_node_id))?;

        let chosen_branch = self.find_matching_branch(&processed_input, current_node)?;

        match chosen_branch {
            Some(branch) => {
                println!("(Debug: Following branch: {})", branch.id);
                self.game_state.current_node_id = branch.target_node_id.clone();

                self.apply_current_node_entry_effects()?;

                self.generate_current_node_text()
            }
            None => {
                self.handle_invalid_input(&processed_input)
            }
        }
    }

    fn find_matching_branch<'a>(&self, input: &str, current_node: &'a StoryNode) -> Result<Option<&'a Branch>, String> {
        let mut fallback_branch: Option<&'a Branch> = None;

        for branch in &current_node.branches {
            match &branch.required_input_pattern {
                Some(regex) => {
                    // Check for specific patterns first
                    if regex.is_match(input) {
                        let conditions_met = self.check_state_conditions(&branch.required_state_conditions)?;
                        if conditions_met {
                            // Found a matching and valid specific branch
                            return Ok(Some(branch));
                        }
                    }
                }
                None => {
                    // If we encounter a branch with None pattern, store it as a potential fallback
                    // We assume there's at most one such fallback branch per node for simplicity
                    if fallback_branch.is_none() {
                         let conditions_met = self.check_state_conditions(&branch.required_state_conditions)?;
                         if conditions_met {
                            fallback_branch = Some(branch);
                         }
                    } else {
                         // Handle case where there's more than one fallback branch if necessary
                         // For now, we just take the first one found
                    }
                }
            }
        }

        // If no specific pattern matched, return the fallback branch if found
        Ok(fallback_branch)
    }

    fn check_state_conditions(&self, conditions: &[StateCondition]) -> Result<bool, String> {
        for condition in conditions {
            match condition {
                StateCondition::HasItem(item) => {
                    if !self.game_state.inventory.contains(item) {
                        println!("(Debug: Condition failed: Requires item '{}')", item);
                        return Ok(false);
                    }
                }
                StateCondition::FlagIsTrue(flag) => {
                    if self.game_state.flags.get(flag) != Some(&true) {
                         println!("(Debug: Condition failed: Requires flag '{}' to be true)", flag);
                        return Ok(false);
                    }
                }
            }
        }
        Ok(true)
    }

    fn apply_current_node_entry_effects(&mut self) -> Result<(), String> {
        let current_node = self.story.get_node(&self.game_state.current_node_id)
             .ok_or_else(|| format!("Error: Current node '{}' not found for effects!", self.game_state.current_node_id))?;

        println!("(Debug: Applying entry effects for node '{}')", current_node.id);

        for effect in &current_node.entry_effects {
            match effect {
                Effect::SetFlag(flag, value) => {
                    self.game_state.flags.insert(flag.clone(), *value);
                    println!("(Debug: Set flag '{}' to {})", flag, value);
                }
                Effect::AddItem(item) => {
                    if !self.game_state.inventory.contains(item) {
                         self.game_state.inventory.push(item.clone());
                         println!("You acquired: {}!", item);
                    } else {
                        println!("You already have the {}!", item);
                    }
                }
                Effect::RemoveItem(item) => {
                    let old_len = self.game_state.inventory.len();
                    self.game_state.inventory.retain(|i| i != item);
                    if self.game_state.inventory.len() < old_len {
                         println!("You used/lost: {}!", item);
                    } else {
                         println!("You don't have the {} to lose!", item);
                    }
                }
            }
        }
        Ok(())
    }

    fn generate_current_node_text(&self) -> Result<String, String> {
         let current_node = self.story.get_node(&self.game_state.current_node_id)
             .ok_or_else(|| format!("Error: Current node '{}' not found for text generation!", self.game_state.current_node_id))?;

        println!("(Debug: Generating text for node '{}')", current_node.id);

        match &current_node.text_source {
            TextSource::Static(text) => Ok(text.clone()),
            TextSource::Template(template) => {
                let mut text = template.clone();
                text = text.replace("{player_name}", &self.game_state.player_name);
                text = text.replace("{item_count}", &self.game_state.inventory.len().to_string());
                Ok(text)
            }
            TextSource::LMPrompt(prompt) => {
                 self.language_model.generate_text(prompt, &self.game_state)
                     .map_err(|e| format!("LM Generation Error: {}", e))
            }
        }
    }

    fn handle_invalid_input(&self, _input: &str) -> Result<String, String> {
        Ok(format!("You can't do that here. Try something else or look around."))
    }

     pub fn get_available_choices(&self) -> Result<Vec<&Branch>, String> {
          let current_node = self.story.get_node(&self.game_state.current_node_id)
             .ok_or_else(|| format!("Error: Current node '{}' not found for choices!", self.game_state.current_node_id))?;

          let mut choices = Vec::new();
          for branch in &current_node.branches {
               if branch.choice_text.is_some() {
                    if self.check_state_conditions(&branch.required_state_conditions)? {
                         choices.push(branch);
                    }
               }
          }
          Ok(choices)
     }

     pub fn current_node_id(&self) -> &str {
         &self.game_state.current_node_id
     }

     pub fn game_state(&self) -> &GameState {
         &self.game_state
     }

      pub fn story(&self) -> &Story {
          &self.story
      }
}

lazy_static! {
    pub static ref GO_NORTH: Regex = Regex::new(r"^(go|move) north$").expect("Invalid regex");
    pub static ref GO_EAST: Regex = Regex::new(r"^(go|move) east$").expect("Invalid regex");
    pub static ref GO_SOUTH: Regex = Regex::new(r"^(go|move) south$").expect("Invalid regex");
    pub static ref GO_WEST: Regex = Regex::new(r"^(go|move) west$").expect("Invalid regex");
    pub static ref TAKE_ITEM: Regex = Regex::new(r"^(take|get) (.+)$").expect("Invalid regex");
    pub static ref LOOK: Regex = Regex::new(r"^look( around)?$").expect("Invalid regex");
    pub static ref INVENTORY: Regex = Regex::new(r"^inventory$").expect("Invalid regex");
    pub static ref TALK_TO: Regex = Regex::new(r"^talk to (.+)$").expect("Invalid regex");
}