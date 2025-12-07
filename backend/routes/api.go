package routes

import (
	"backend/controller"
	"backend/middleware"
	"backend/models"
	"backend/service"
	"backend/websocket"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SetupRoutes configures all API routes
func SetupRoutes(router *gin.Engine, userController *controller.UserController, messageController *controller.MessageController, groupController *controller.GroupController, hub *websocket.Hub, userService *service.UserService, groupService *service.GroupService) {
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

		// Chat routes (protected)
		chat := api.Group("/chat")
		chat.Use(middleware.AuthMiddleware())
		{
			// Get list of online users only
			chat.GET("/online-users", func(c *gin.Context) {
				users := hub.GetOnlineUsers()
				c.JSON(200, gin.H{
					"online_users": users,
					"count":        len(users),
				})
			})

			// Get users with conversation history (like Slack DM sidebar)
			chat.GET("/users", func(c *gin.Context) {
				// Get current user ID from context (set by auth middleware)
				currentUserID, exists := c.Get("user_id")
				if !exists {
					c.JSON(401, gin.H{"error": "user not authenticated"})
					return
				}

				userID, ok := currentUserID.(uuid.UUID)
				if !ok {
					c.JSON(500, gin.H{"error": "invalid user ID"})
					return
				}

				// Get users the current user has had conversations with
				conversationUsers, err := userService.GetUsersWithConversation(userID)
				if err != nil {
					c.JSON(500, gin.H{"error": "failed to fetch users"})
					return
				}

				// Get online users from hub
				onlineUsers := hub.GetOnlineUsers()
				onlineMap := make(map[uuid.UUID]bool)
				for _, u := range onlineUsers {
					onlineMap[u.ID] = true
				}

				// Build response with status
				usersWithStatus := make([]models.UserWithStatus, len(conversationUsers))
				for i, user := range conversationUsers {
					usersWithStatus[i] = models.UserWithStatus{
						ID:       user.ID,
						Username: user.Username,
						Email:    user.Email,
						IsOnline: onlineMap[user.ID],
					}
				}

				c.JSON(200, gin.H{
					"users": usersWithStatus,
					"count": len(usersWithStatus),
				})
			})

			// Search users by username or email
			chat.GET("/users/search", func(c *gin.Context) {
				query := c.Query("q")
				if query == "" {
					c.JSON(400, gin.H{"error": "search query 'q' is required"})
					return
				}

				// Search users in database
				users, err := userService.SearchUsers(query)
				if err != nil {
					c.JSON(500, gin.H{"error": "failed to search users"})
					return
				}

				// Get online users from hub
				onlineUsers := hub.GetOnlineUsers()
				onlineMap := make(map[uuid.UUID]bool)
				for _, u := range onlineUsers {
					onlineMap[u.ID] = true
				}

				// Build response with status
				usersWithStatus := make([]models.UserWithStatus, len(users))
				for i, user := range users {
					usersWithStatus[i] = models.UserWithStatus{
						ID:       user.ID,
						Username: user.Username,
						Email:    user.Email,
						IsOnline: onlineMap[user.ID],
					}
				}

				c.JSON(200, gin.H{
					"users": usersWithStatus,
					"count": len(usersWithStatus),
				})
			})

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

			// Get group members
			groups.GET("/:id/members", groupController.GetGroupMembers)

			// Add members to a group
			groups.POST("/:id/members", groupController.AddMembers)

			// Get group messages
			groups.GET("/:id/messages", groupController.GetGroupMessages)
		}
	}
}
