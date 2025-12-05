package routes

import (
	"backend/controller"
	"backend/middleware"
	"backend/websocket"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all API routes
func SetupRoutes(router *gin.Engine, userController *controller.UserController, hub *websocket.Hub) {
	// API v1 group
	api := router.Group("/api/realtime-chat")
	{
		// Health check endpoint
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status": "ok",
			})
		})

		// WebSocket endpoint
		api.GET("/ws", websocket.HandleWebSocket(hub))

		// Auth routes (public)
		auth := api.Group("/auth")
		{
			auth.POST("/register", userController.Register)
			auth.POST("/login", userController.Login)
			auth.POST("/refresh", userController.RefreshToken)

			// Protected routes
			auth.GET("/user-details", middleware.AuthMiddleware(), userController.GetUserDetails)
		}
	}
}
