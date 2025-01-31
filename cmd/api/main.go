package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"kick-points/internal/database"
	"kick-points/internal/server"

	"github.com/gorilla/websocket"
)

func gracefulShutdown(apiServer *http.Server, done chan bool) {
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Listen for the interrupt signal.
	<-ctx.Done()

	log.Println("shutting down gracefully, press Ctrl+C again to force")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := apiServer.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown with error: %v", err)
	}

	log.Println("Server exiting")

	// Notify the main goroutine that the shutdown is complete
	done <- true
}

func main() {

	server := server.NewServer()

	// Create a done channel to signal when the shutdown is complete
	done := make(chan bool, 1)

	// Run graceful shutdown in a separate goroutine
	go gracefulShutdown(server, done)

	go connectWebSocket()

	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		panic(fmt.Sprintf("http server error: %s", err))
	}

	// Wait for the graceful shutdown to complete
	<-done
	log.Println("Graceful shutdown complete.")
}

// Kick websocket client stuff

type Identity struct {
	Badges []Badge `json:"badges"`
}

type Badge struct {
	Type string `json:"type"`
}

type Sender struct {
	Slug     string   `json:"slug"`
	Identity Identity `json:"identity"`
}

type Data struct {
	Sender Sender `json:"sender"`
}

type Message struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data"`
}

func isSubscriber(tags []Badge) bool {
	for _, tag := range tags {
		if tag.Type == "subscriber" {
			return true
		}
	}
	return false
}

func connectWebSocket() {
	var url = os.Getenv("KICK_WEBSOCKET_URL")
	// var channelId = os.Getenv("KICK_CHANNEL_ID")

	headers := http.Header{}
	headers.Set("Origin", "https://kick.com")

	conn, _, err := websocket.DefaultDialer.Dial(url, headers)
	if err != nil {
		fmt.Println("Error connecting to WebSocket server:", err)
		return
	}
	defer conn.Close()

	channelIds := strings.Split(os.Getenv("KICK_CHANNEL_ID"), ",")

	for _, channelId := range channelIds {
		subscribeMessage := map[string]interface{}{
			"event": "pusher:subscribe",
			"data": map[string]string{
				"channel": fmt.Sprintf("chatrooms.%s.v2", channelId),
			},
		}

		message, err := json.Marshal(subscribeMessage)
		if err != nil {
			fmt.Println("Error marshaling subscribe message:", err)
			return
		}

		err = conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			fmt.Println("Error sending subscribe message:", err)
			return
		}

		fmt.Printf("Connected to channel: %s\n", channelId)
	}

	for {
		_, message, err := conn.ReadMessage()

		if err != nil {
			fmt.Println("Error reading message:", err)
			return
		}

		var msg Message
		err = json.Unmarshal(message, &msg)
		if err != nil {
			log.Printf("Error parsing JSON: %v", err)
			continue
		}

		if msg.Event == "App\\Events\\ChatMessageEvent" {
			var dataString string
			err = json.Unmarshal(msg.Data, &dataString)
			if err != nil {
				log.Printf("Error parsing data string: %v", err)
				continue
			}

			var data Data
			err = json.Unmarshal([]byte(dataString), &data)
			if err != nil {
				log.Printf("Error parsing data JSON: %v", err)
				continue
			}

			var amountToAdd int
			if isSubscriber(data.Sender.Identity.Badges) {
				amountToAdd, err = strconv.Atoi(os.Getenv("POINTS_PER_MESSAGE_SUBSCRIBER"))

				if err != nil {
					log.Printf("Error parsing string to number: %v", err)
					continue
				}
			} else {
				amountToAdd, err = strconv.Atoi(os.Getenv("POINTS_PER_MESSAGE"))

				if err != nil {
					log.Printf("Error parsing string to number: %v", err)
					continue
				}
			}

			if amountToAdd > 0 {
				database.New().AddPoints(data.Sender.Slug, amountToAdd)
			}
		}
	}
}
