package controllers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestRegister(t *testing.T) {
	r := gin.Default()
	r.POST("/register", func(c *gin.Context) {
		// Use mock DB for testing
		var mockDB *gorm.DB
		Register(c, mockDB)
	})

	reqBody := `{"email": "testuser@example.com", "password": "password123"}`
	req, _ := http.NewRequest("POST", "/register", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"message": "User registered successfully"}`, w.Body.String())
}