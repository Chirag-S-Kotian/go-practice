use std::io;

fn main() {
    let mut task = String::new();
    println!("Hello, Enter your task!");
    io::stdin()
        .read_line(&mut task)
        .expect("Failed to read line");
    insert_task(task);
    view_task();
}

fn insert_task(task: String) {
    let mut task = task;
    task.push_str(" inserted");
    println!("Task inserted: {}", task);
}

// fn update_task(task: String) {
//     let mut task = task;
//     task.push_str(" updated");
//     println!("Task updated: {}", task);
// }

// fn delete_task(task: String) {
//     let mut task = task;
//     task.push_str(" deleted");
//     println!("Task deleted: {}", task);
// }

fn view_task() {
    println!("Task viewed");
}