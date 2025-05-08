mod story_data;
mod game_state;
mod language_model;
mod storyteller;

use std::collections::HashMap;
use std::io::{self, Write};

use story_data::{Story, StoryNode, Branch, TextSource, Effect, StateCondition};
use storyteller::{Storyteller, GO_EAST, GO_WEST, TAKE_ITEM, LOOK, INVENTORY};

fn main() -> Result<(), String> {
    println!("Welcome to the Interactive Story AI!");

    let mut nodes = HashMap::new();

    nodes.insert("clearing".into(), StoryNode {
        id: "clearing".into(),
        text_source: TextSource::LMPrompt("You are in a quiet forest clearing. Sunlight filters through the leaves. Describe the atmosphere and what you see.".into()),
        branches: vec![
            Branch {
                id: "clearing_to_east_path".into(),
                required_input_pattern: Some(GO_EAST.clone()),
                required_state_conditions: vec![],
                target_node_id: "forest_path".into(),
                choice_text: Some("Go East into the forest path".into()),
            },
             Branch {
                id: "clearing_look_around".into(),
                required_input_pattern: Some(LOOK.clone()),
                required_state_conditions: vec![],
                target_node_id: "clearing_look".into(),
                choice_text: Some("Look around the clearing".into()),
            },
        ],
        entry_effects: vec![],
    });

    nodes.insert("clearing_look".into(), StoryNode {
        id: "clearing_look".into(),
        text_source: TextSource::Static("You look around the clearing. To the east, you see a dark path leading into thicker woods. There's nothing else of note here.".into()),
        branches: vec![
             Branch {
                id: "clearing_look_return".into(),
                required_input_pattern: None,
                required_state_conditions: vec![],
                target_node_id: "clearing".into(),
                choice_text: None,
            }
        ],
        entry_effects: vec![],
    });


    nodes.insert("forest_path".into(), StoryNode {
        id: "forest_path".into(),
        text_source: TextSource::Template("You are on a narrow forest path. Trees tower above you. You see a glint to the south. You have {item_count} items.".into()),
        branches: vec![
            Branch {
                id: "path_to_west_clearing".into(),
                required_input_pattern: Some(GO_WEST.clone()),
                required_state_conditions: vec![],
                target_node_id: "clearing".into(),
                choice_text: Some("Go West back to the clearing".into()),
            },
            Branch {
                id: "path_to_south_glint".into(),
                required_input_pattern: Some(TAKE_ITEM.clone()),
                required_state_conditions: vec![StateCondition::FlagIsTrue("glint_present".into())],
                target_node_id: "forest_path_item_taken".into(),
                choice_text: Some("Take the shiny object".into()),
            },
             Branch {
                id: "path_inventory".into(),
                required_input_pattern: Some(INVENTORY.clone()),
                required_state_conditions: vec![],
                target_node_id: "inventory_node".into(),
                choice_text: Some("Check your inventory".into()),
            },
        ],
        entry_effects: vec![Effect::SetFlag("glint_present".into(), true)],
    });

     nodes.insert("forest_path_item_taken".into(), StoryNode {
        id: "forest_path_item_taken".into(),
        text_source: TextSource::Static("You pick up the shiny object. It's a simple, smooth river stone. You pocket it.".into()),
        branches: vec![
             Branch {
                id: "forest_path_item_taken_return".into(),
                required_input_pattern: None,
                required_state_conditions: vec![],
                target_node_id: "forest_path".into(),
                choice_text: None,
            }
        ],
        entry_effects: vec![
            Effect::AddItem("Shiny Stone".into()),
            Effect::SetFlag("glint_present".into(), false)
        ],
    });

     nodes.insert("inventory_node".into(), StoryNode {
         id: "inventory_node".into(),
         text_source: TextSource::Template("Your inventory contains: {inventory_list}. You are currently at {current_location_id}.".into()),
         branches: vec![
              Branch {
                id: "inventory_return".into(),
                required_input_pattern: None,
                required_state_conditions: vec![],
                target_node_id: "forest_path".into(),
                choice_text: None,
             }
         ],
         entry_effects: vec![],
     });

    let story = Story {
        nodes,
        start_node_id: "clearing".into(),
    };

    let mut storyteller = Storyteller::new(story, "Hero".into())?;

    let initial_narrative = storyteller.start()?;
    println!("{}", initial_narrative);

    let game_over = false;

    while !game_over {
        println!("\nWhat do you do?");
        match storyteller.get_available_choices() {
            Ok(choices) => {
                if choices.is_empty() {
                    println!("(No specific options available, try a general action like 'look' or 'inventory')");
                } else {
                    for (i, choice) in choices.iter().enumerate() {
                        if let Some(text) = &choice.choice_text {
                            println!("{}) {}", i + 1, text);
                        }
                    }
                }
            }
            Err(e) => eprintln!("Error getting choices: {}", e),
        }


        print!("> ");
        io::stdout().flush().unwrap();

        let mut input = String::new();
        io::stdin().read_line(&mut input).map_err(|e| e.to_string())?;

        if input.trim().to_lowercase() == "quit" {
            println!("Exiting game.");
            break;
        }

        let narrative_result = storyteller.handle_input(&input);

        match narrative_result {
            Ok(narrative) => {
                 let final_narrative = if storyteller.current_node_id() == "inventory_node" {
                       let inventory_list = if storyteller.game_state().inventory.is_empty() {
                           "nothing".into()
                       } else {
                           storyteller.game_state().inventory.join(", ")
                       };
                       let current_location_id = storyteller.story().get_node(storyteller.current_node_id())
                           .and_then(|node| Some(node.id.as_str()))
                           .unwrap_or("unknown");

                       narrative.replace("{inventory_list}", &inventory_list)
                                  .replace("{current_location_id}", current_location_id)
                 } else {
                     narrative
                 };
                println!("{}", final_narrative);
            }
            Err(e) => {
                eprintln!("Error: {}", e);
            }
        }
    }

    Ok(())
}
