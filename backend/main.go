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

	// Setup routes (now includes WebSocket)
	routes.SetupRoutes(router, userController, hub)

	// Start server
	log.Println("Server starting on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
