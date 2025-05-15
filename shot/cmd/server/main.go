package main

import (
	"shot/internal/models"
	"shot/internal/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	models.InitDB()
	r :=gin.Default()
	routes.AuthRoutes(r)

	r.GET("/", func(c *gin.Context){
		c.JSON(200, gin.H{"message": "Youtube  API is running"})
	})

	r.Run()
}
 