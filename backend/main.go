package main

import (
	"backend/config"
	"backend/controller"
	"backend/db"
	"backend/routes"
	"backend/service"
	"backend/websocket"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize database
	database := config.InitDB()

	// Initialize repository layer
	userRepo := db.NewUserRepository(database)

	// Initialize service layer
	userService := service.NewUserService(userRepo)

	// Initialize controller layer
	userController := controller.NewUserController(userService)

	// Initialize WebSocket hub
	hub := websocket.NewHub()
	go hub.Run()

	// Initialize Gin router
	router := gin.Default()

	// Setup routes
	routes.SetupRoutes(router, userController)

	// WebSocket endpoint
	router.GET("/ws", websocket.HandleWebSocket(hub))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":         "ok",
			"active_clients": hub.GetActiveClients(),
		})
	})

	// Start server
	log.Println("Server starting on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
