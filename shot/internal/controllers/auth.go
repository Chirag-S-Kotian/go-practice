package controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"shot/models"
	"shot/utils"
)

func Register(c *gin.Context,db *gorm.DB){
	var user models.user
	if err := c.shouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error":err.Error()})
		return
	}
	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to give hash password"})
	}
	user.Password
}