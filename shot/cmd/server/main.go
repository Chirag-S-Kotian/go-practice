package main

import (
	"github.com/gin-gonic/gin"
	"shot/internal/models"
	"shot/internal/routes"
)

func main() {
	models.InitDB()

	r := gin.Default()
	routes.AuthRoutes(r)
	routes.VideoRoutes(r)

	r.Run()
}