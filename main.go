package main

import (
	handler "assignment-2/internal/handlers"
	utils "assignment-2/utils"
	"log"
	"net/http"
	"os"
)

func main() {

	// Extract PORT variable from the OS environment variables
	port := os.Getenv("PORT")

	// Override port with default port if not provided (e.g., local deployment)
	if port == "" {
		log.Println("\n$PORT has not been set. Default: 8080")
		port = "8080"
	}

	//Initializing requestlogger
	utils.InitLogger()

	//initializing router
	router := http.NewServeMux()

	router.HandleFunc("/", handler.DefaultHandler)
	router.HandleFunc(utils.REGISTRATION_PATH, h.RegistrationHandler)

	// Configure the HTTP server with the network address and
	// the router wrapped in logging middleware.
	server := http.Server{
		Addr:    ":" + port,
		Handler: utils.Logging(router),
	}

	//Starting server
	log.Println("Starting server on port " + port)
	log.Fatal(server.ListenAndServe(), "Error starting server")
}
