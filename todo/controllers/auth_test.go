package controllers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"todo/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite" // Using SQLite for testing
	"gorm.io/gorm"
)

func TestRegister(t *testing.T) {
	// Initialize Gin router for testing
	r := gin.Default()

	// Create an in-memory SQLite database for testing
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})

	// Migrate the schema (create the User table)
	db.AutoMigrate(&models.User{})

	// Register the route
	r.POST("/register", func(c *gin.Context) {
		Register(c, db) // Pass the mock DB
	})

	// Prepare the request body
	reqBody := `{"email": "testuser@example.com", "password": "password123"}`
	req, _ := http.NewRequest("POST", "/register", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	// Create a response recorder
	w := httptest.NewRecorder()

	// Send the request to the route
	r.ServeHTTP(w, req)

	// Assert the response status and body
	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"message": "User registered successfully"}`, w.Body.String())
}