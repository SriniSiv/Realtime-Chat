package routes

import (
	"backend/controller"
	"backend/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all API routes
func SetupRoutes(router *gin.Engine, userController *controller.UserController) {
	// API v1 group
	api := router.Group("/api/realtime-chat")
	{
		// Auth routes (public)
		auth := api.Group("/auth")
		{
			auth.POST("/register", userController.Register)
			auth.POST("/login", userController.Login)
			auth.POST("/refresh", userController.RefreshToken)

			// Protected routes
			auth.GET("/me", middleware.AuthMiddleware(), userController.GetMe)
		}
	}
}
