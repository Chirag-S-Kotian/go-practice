// routes/routes.go
package routes

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"todo/controllers"
	"todo/models"
)

func SetupRoutes(router *gin.Engine) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dsn := os.Getenv("DB_DSN")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to initialize database, got error %v", err)
	}

	// Auto-migrate models
	db.AutoMigrate(&models.User{}, &models.Todo{})

	// Inject DB into controllers if needed
	router.POST("/register", func(c *gin.Context) { controllers.Register(c, db) })
	router.POST("/login", func(c *gin.Context) { controllers.Login(c, db) })
	router.GET("/todos", func(c *gin.Context) { controllers.GetTodos(c, db) })
	router.POST("/todos", func(c *gin.Context) { controllers.CreateTodo(c, db) })
}
