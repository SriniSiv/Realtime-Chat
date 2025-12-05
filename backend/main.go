package main

import (
	"backend/config"
	"backend/controller"
	"backend/db"
	"backend/routes"
	"backend/service"
	"backend/websocket"
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize database
	database := config.InitDB()

	// Initialize repository layer
	userRepo := db.NewUserRepository(database)
	messageRepo := db.NewMessageRepository(database)

	// Initialize service layer
	userService := service.NewUserService(userRepo)
	messageService := service.NewMessageService(messageRepo, userRepo)

	// Initialize controller layer
	userController := controller.NewUserController(userService)
	messageController := controller.NewMessageController(messageService)

	// Initialize WebSocket hub with message service for persistence
	hub := websocket.NewHub(messageService)
	go hub.Run()

	// Initialize Gin router
	router := gin.Default()

	// Configure CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://127.0.0.1:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Setup routes (now includes WebSocket)
	routes.SetupRoutes(router, userController, messageController, hub)

	// Start server
	log.Println("Server starting on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
