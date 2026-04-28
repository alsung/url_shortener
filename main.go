// main.go
//
// Entry point: wires together the router, handlers, and server
//
// This file should stay thin - its only job is dependency wiring.
// Business logic lives in handlers/, server logic lives in server/.

package main

import (
	"log"
	"os"

	"github.com/alsung/url-shortener/handlers"
	"github.com/alsung/url-shortener/server"
)

func main() {
	// Read the port from environment, default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	// Wire up routes
	//
	// Routes are matched in registration order - first match wins.
	// Notice the order matters:
	// 	- /health must be registered before /:code or it would never match
	//	  (because /:code would match "health" as the short code first)
	// 	- /stats/:code must be registered before /:code for the same reason

	router := server.NewRouter()

	router.GET("/health", handlers.Health)
	router.POST("/shorten", handlers.Shorten)
	router.GET("/stats/:code", handlers.Stats)
	router.GET("/:code", handlers.Redirect)

	// Start the server
	//
	// server.New takes our router's HandlerFunc - a function that takes a
	// Request and returns a Response. The server handles all TCP/connection
	// logic; the router handles all dispatch logic. Clean separation.

	srv := server.New(addr, router.Handler())

	log.Printf("Starting URL shortener on %s", addr)
	log.Printf("Routers registered:")
	log.Printf("  GET	/health")
	log.Printf("  POST	/shorten")
	log.Printf("  GET	/stats/:code")
	log.Printf("  GET	/:code  (redirect)")

	// ListenAndServe blocks forever (until the process is killed)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
