package main

import (
	"fmt"
	"os"
	"slices"
)

var tasks []string

func main() {
	for {
		fmt.Println("\nEnter your choice")
		fmt.Println("1. Insert task")
		fmt.Println("2. Delete task")
		fmt.Println("3. View tasks")
		fmt.Println("4. Exit")
		var choice int
		fmt.Scan(&choice)
		if choice == 1 {
			fmt.Println("Enter the task")
			var task string
			fmt.Scan(&task)
			insert_task(task)
		} else if choice == 2 {
			fmt.Println("Enter the index")
			var index int
			fmt.Scan(&index)
			delete_task(index)
		} else if choice == 3 {
			view_tasks()
		} else {
			fmt.Println("Invalid choice")
			os.Exit(0)
		}
	}
}

func insert_task(task string) {
	if task == "" {
		fmt.Println("Task cannot be empty")
		return
	}
	tasks = append(tasks, task)
	fmt.Println("Task inserted")
}

func view_tasks() {
	if len(tasks) == 0 {
		fmt.Println("No tasks")
		return
	}
	fmt.Println("Tasks:")
	for i, task := range tasks {
		fmt.Println(i, task)
	}
}

func delete_task(index int) {
	if index < 0 || index >= len(tasks) {
		fmt.Println("Invalid index")
		return
	}
	tasks = slices.Delete(tasks, index, index+1)
	fmt.Println("Task deleted")
}
