package routes

import (
	"backend/controller"
	"backend/middleware"
	"backend/websocket"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all API routes
func SetupRoutes(router *gin.Engine, userController *controller.UserController, messageController *controller.MessageController, groupController *controller.GroupController, hub *websocket.Hub) {
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
			auth.PUT("/user", middleware.AuthMiddleware(), userController.UpdateUsername)
		}

		// Chat routes (protected)
		chat := api.Group("/chat")
		chat.Use(middleware.AuthMiddleware())
		{
			// Get list of online users only
			chat.GET("/online-users", userController.GetOnlineUsers)

			// Get users with conversation history (like Slack DM sidebar)
			chat.GET("/users", userController.GetUsersWithConversation)

			// Search users by username or email
			chat.GET("/users/search", userController.SearchUsers)

			// Get chat history with a specific user
			chat.GET("/history", messageController.GetChatHistory)

			// Get all messages for current user
			chat.GET("/messages", messageController.GetAllMessages)
		}

		// Group routes (protected)
		groups := api.Group("/groups")
		groups.Use(middleware.AuthMiddleware())
		{
			// Create a new group
			groups.POST("", groupController.CreateGroup)

			// Get all groups for current user
			groups.GET("", groupController.GetUserGroups)

			// Search user's groups (must be before /:id to avoid conflict)
			groups.GET("/search", groupController.SearchUserGroups)

			// Search available groups to join
			groups.GET("/search/available", groupController.SearchAvailableGroups)

			// Get group details
			groups.GET("/:id", groupController.GetGroup)

			// Update group name/description
			groups.PUT("/:id", groupController.UpdateGroup)

			// Get group members
			groups.GET("/:id/members", groupController.GetGroupMembers)

			// Add members to a group
			groups.POST("/:id/members", groupController.AddMembers)

			// Get group messages
			groups.GET("/:id/messages", groupController.GetGroupMessages)
		}
	}
}
