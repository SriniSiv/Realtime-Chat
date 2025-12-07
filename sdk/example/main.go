package main

import (
	"fmt"
	"os"
	"os/signal"

	"sdk"
)

func main() {
	// Initialize client
	client, err := sdk.NewClient(sdk.Config{
		APIKey:   "your-jwt-token",
		Endpoint: "ws://localhost:8080/api/realtime-chat/ws",
	})
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Register message handler (Observer Pattern)
	client.OnMessage(func(msg sdk.Message) {
		fmt.Printf("[%s] %s: %s\n", msg.Type, msg.From, msg.Content)
	})

	// Connect to server
	if err := client.Connect(); err != nil {
		fmt.Println("Connection failed:", err)
		return
	}
	fmt.Println("Connected!")

	// Send message
	client.SendMessage("user-id", "Hello, world!")

	// Wait for interrupt
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	<-sigCh

	// Graceful shutdown
	client.Close()
	fmt.Println("Disconnected!")
}

