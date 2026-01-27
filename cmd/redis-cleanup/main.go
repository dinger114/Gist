package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"gist/backend/internal/db"
)

// This utility cleans up user data from Redis storage
// Use this when you get stuck on login page with empty database
func main() {
	redisURL := os.Getenv("UPSTASH_URL")
	redisToken := os.Getenv("UPSTASH_TOKEN")

	if redisURL == "" {
		log.Fatal("UPSTASH_URL environment variable is required")
	}

	// Create Redis storage
	storage, err := db.NewStorage(redisURL, redisToken)
	if err != nil {
		log.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	ctx := context.Background()

	// User setting keys to clean up
	userKeys := []string{
		"user.username",
		"user.nickname",
		"user.email",
		"user.password_hash",
		"user.jwt_secret",
	}

	fmt.Println("Cleaning up user data from Redis...")

	for _, key := range userKeys {
		fmt.Printf("Deleting key: %s\n", key)
		err := storage.DeleteSetting(ctx, key)
		if err != nil {
			fmt.Printf("Warning: Failed to delete %s: %v\n", key, err)
		} else {
			fmt.Printf("Successfully deleted: %s\n", key)
		}
	}

	fmt.Println("\nUser data cleanup complete!")
	fmt.Println("You should now be able to register a new user.")
}