package controllers

import (
	"net/http"
	"shot/internal/models"
	"github.com/gin-gonic/gin"
)

func UploadVideo(c *gin.Context) {
	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		VideoURL    string `json:"video_url"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid data"})
		return
	}

	// Get user_id from JWT middleware (coming next)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	err := models.CreateVideo(req.Title, req.Description, req.VideoURL, userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not upload video"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "video uploaded"})
}