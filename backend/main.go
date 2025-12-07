package main

import (
	"backend/config"
	"backend/controller"
	"backend/db"
	"backend/kafka"
	"backend/models"
	"backend/redis"
	"backend/routes"
	"backend/service"
	"backend/websocket"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func main() {
	// Initialize environment configuration first
	config.InitializeEnv()
	env := config.EnvironmentData

	// Initialize database
	database := config.InitDB()

	// Auto-migrate new tables and update existing ones
	database.AutoMigrate(&models.ChatMessage{}, &models.Group{}, &models.GroupMember{})

	// Initialize repository layer
	userRepo := db.NewUserRepository(database)
	messageRepo := db.NewMessageRepository(database)
	groupRepo := db.NewGroupRepository(database)

	// Initialize service layer
	userService := service.NewUserService(userRepo)
	messageService := service.NewMessageService(messageRepo, userRepo)
	groupService := service.NewGroupService(groupRepo, userRepo)

	// Initialize controller layer
	userController := controller.NewUserController(userService)
	messageController := controller.NewMessageController(messageService)
	groupController := controller.NewGroupController(groupService)
	groupController.SetMessageService(messageService)

	// Initialize WebSocket hub with message service for persistence
	hub := websocket.NewHub(messageService)
	hub.SetGroupMemberProvider(groupService)
	go hub.Run()

	// Set online users provider for user service (for online status in API responses)
	userService.SetOnlineUsersProvider(hub)

	// Get instance ID from environment
	instanceID := env.InstanceID
	log.Printf("Instance ID: %s", instanceID)

	// Initialize Kafka for message persistence (if enabled)
	kafkaConfig := config.GetKafkaConfig()
	var kafkaProducer *kafka.Producer
	var kafkaConsumer *kafka.Consumer

	if kafkaConfig.Enabled {
		// Create Kafka producer for persistence
		producer, err := kafka.NewProducer(kafkaConfig.Brokers, kafkaConfig.Topic, instanceID)
		if err != nil {
			log.Printf("Failed to create Kafka producer: %v", err)
		} else {
			kafkaProducer = producer
			hub.SetKafkaProducer(producer, instanceID)
			log.Println("Kafka producer initialized (for persistence)")
		}

		// Create Kafka consumer for persistence (same consumer group for exactly-once)
		// This consumer saves messages to database
		consumer, err := kafka.NewConsumer(
			kafkaConfig.Brokers,
			kafkaConfig.Topic,
			"chat-persistence-group", // Same group across all instances
			func(msg *kafka.ChatMessage) {
				// Save message to database
				if msg.Type != "system" {
					var err error
					if msg.Type == "group" && msg.GroupID != uuid.Nil {
						_, err = messageService.SaveGroupMessage(msg.From, msg.GroupID, msg.Content)
					} else {
						_, err = messageService.SaveMessage(msg.From, msg.To, msg.Content, msg.Type)
					}
					if err != nil {
						log.Printf("Kafka consumer: Error saving message: %v", err)
					} else {
						log.Printf("Kafka consumer: Message saved to database (type=%s)", msg.Type)
					}
				}
			},
		)
		if err != nil {
			log.Printf("Failed to create Kafka consumer: %v", err)
		} else {
			kafkaConsumer = consumer
			consumer.Start()
			log.Println("Kafka persistence consumer started")
		}
	}

	// Initialize Redis Pub/Sub for real-time fanout (if enabled)
	redisConfig := config.GetRedisConfig()
	var redisPubSub *redis.PubSub

	if redisConfig.Enabled {
		pubsub, err := redis.NewPubSub(
			redisConfig.Addr,
			redisConfig.Password,
			redisConfig.DB,
			redisConfig.Channel,
			instanceID,
			hub.HandleRedisMessage,
		)
		if err != nil {
			log.Printf("Failed to create Redis Pub/Sub: %v", err)
		} else {
			redisPubSub = pubsub
			hub.SetRedisPubSub(pubsub)
			pubsub.Subscribe()
			log.Println("Redis Pub/Sub initialized (for real-time fanout)")
		}
	}

	// Set Gin mode from environment
	gin.SetMode(env.GinMode)

	// Initialize Gin router
	router := gin.Default()

	// Configure CORS from environment
	router.Use(cors.New(cors.Config{
		AllowOrigins:     env.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Setup routes (now includes WebSocket)
	routes.SetupRoutes(router, userController, messageController, groupController, hub)

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down...")

		// Close Redis Pub/Sub
		if redisPubSub != nil {
			if err := redisPubSub.Close(); err != nil {
				log.Printf("Error closing Redis Pub/Sub: %v", err)
			}
		}

		// Close Kafka connections
		if kafkaProducer != nil {
			if err := kafkaProducer.Close(); err != nil {
				log.Printf("Error closing Kafka producer: %v", err)
			}
		}
		if kafkaConsumer != nil {
			if err := kafkaConsumer.Close(); err != nil {
				log.Printf("Error closing Kafka consumer: %v", err)
			}
		}

		os.Exit(0)
	}()

	// Start server
	port := ":" + env.ServerPort
	log.Printf("Server starting on %s", port)
	if err := router.Run(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
