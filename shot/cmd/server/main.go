package main

import (
	"shot/internal/models"

	"github.com/gin-gonic/gin"
)

func main() {
	models.InitDB()
	r :=gin.Default()
	r.GET("/ping", func(c *gin.Context){
		c.JSON(200, gin.H{"message": "pong"})
	})

	r.Run()
}
 