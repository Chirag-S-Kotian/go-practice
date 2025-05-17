package controllers

import (
	"fmt"
	"net/http"
	"shot/internal/utils"
	"time"

	"github.com/gin-gonic/gin"
)

func GenerateImagekitAuth(c *gin.Context) {
	token := utils.GenerateRandomString(16)
	expire := time.Now().Unix() + 240 // 4 mins

	privateKey := utils.GetEnv("IMAGEKIT_PRIVATE_KEY")
	expireStr := fmt.Sprintf("%d", expire)

	// ✅ Correct signature generation using HMAC-SHA1
	signature := utils.GenerateSignature(token, expireStr, privateKey)

	c.JSON(http.StatusOK, gin.H{
		"token":     token,
		"expire":    expire,
		"signature": signature,
		"publicKey": utils.GetEnv("IMAGEKIT_PUBLIC_KEY"),
		"folder":    utils.GetEnv("IMAGEKIT_FOLDER"),
	})
}