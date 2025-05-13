use std::io::{self, Write};

fn main() {
    let mut tasks: Vec<String> = Vec::new();

    loop {
        println!("\n--- Task Manager ---");
        println!("1. Insert task");
        println!("2. View tasks");
        println!("3. Update task");
        println!("4. Delete task");
        println!("5. Exit");
        print!("Enter your choice: ");
        io::stdout().flush().unwrap();

        let mut choice = String::new();
        io::stdin().read_line(&mut choice).unwrap();

        match choice.trim() {
            "1" => insert_task(&mut tasks),
            "2" => view_tasks(&tasks),
            "3" => update_task(&mut tasks),
            "4" => delete_task(&mut tasks),
            "5" => {
                println!("Exiting...");
                break;
            }
            _ => println!("Invalid choice, try again."),
        }
    }
}

fn insert_task(tasks: &mut Vec<String>) {
    print!("Enter the new task: ");
    io::stdout().flush().unwrap();

    let mut task = String::new();
    io::stdin().read_line(&mut task).unwrap();
    let task = task.trim().to_string();

    if task.is_empty() {
        println!("Task cannot be empty.");
        return;
    }

    tasks.push(task);
    println!("Task inserted.");
}

fn view_tasks(tasks: &Vec<String>) {
    if tasks.is_empty() {
        println!("No tasks to show.");
        return;
    }

    println!("\nYour tasks:");
    for (i, task) in tasks.iter().enumerate() {
        println!("{}: {}", i + 1, task);
    }
}

fn update_task(tasks: &mut Vec<String>) {
    view_tasks(tasks);

    if tasks.is_empty() {
        return;
    }

    print!("Enter task number to update: ");
    io::stdout().flush().unwrap();

    let mut index_input = String::new();
    io::stdin().read_line(&mut index_input).unwrap();

    if let Ok(index) = index_input.trim().parse::<usize>() {
        if index == 0 || index > tasks.len() {
            println!("Invalid task number.");
            return;
        }

        print!("Enter new description: ");
        io::stdout().flush().unwrap();

        let mut new_task = String::new();
        io::stdin().read_line(&mut new_task).unwrap();

        tasks[index - 1] = new_task.trim().to_string();
        println!("Task updated.");
    } else {
        println!("Please enter a valid number.");
    }
}

fn delete_task(tasks: &mut Vec<String>) {
    view_tasks(tasks);

    if tasks.is_empty() {
        return;
    }

    print!("Enter task number to delete: ");
    io::stdout().flush().unwrap();

    let mut index_input = String::new();
    io::stdin().read_line(&mut index_input).unwrap();

    if let Ok(index) = index_input.trim().parse::<usize>() {
        if index == 0 || index > tasks.len() {
            println!("Invalid task number.");
            return;
        }

        tasks.remove(index - 1);
        println!("Task deleted.");
    } else {
        println!("Please enter a valid number.");
    }
}