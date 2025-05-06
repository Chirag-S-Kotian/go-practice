// models/todo.go
package models

import "gorm.io/gorm"

type Todo struct {
	gorm.Model
	Title     string
	Completed bool
	UserID    uint
}