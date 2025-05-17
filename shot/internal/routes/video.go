package routes

import (
	"github.com/gin-gonic/gin"
	"shot/internal/controllers"
	"shot/internal/middleware"
)

func VideoRoutes(r *gin.Engine) {
	v := r.Group("/videos")
	v.Use(middleware.AuthMiddleware()) // 🔒 protect routes
	{
		v.POST("/upload", controllers.UploadVideo)
	}
}