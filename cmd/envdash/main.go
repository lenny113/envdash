package main

import (
	"assignment-2/internal/client"
	handler "assignment-2/internal/handlers"
	utils "assignment-2/internal/utils"
	"log"
	"net/http"
	"os"
)

func main() {

	/*
		ctx := context.Background()
		//client, err := firestore.NewClient(ctx, "<PROJECT_ID>", ADD PROJECT ID HERTE
		//	option.WithCredentialsFile(<CREDENTIALS_FILE>)) ADD CREDENTIALS FILE HERE
		if err != nil {
			log.Fatal("Failed to initialize Firestore:", err)
		}
		defer client.Close()

		st := store.NewFirestoreStore(client)
	*/

	restCountriesHTTPClient := utils.NewHttpClient()
	restCountriesClient := client.NewRestCountriesClient(restCountriesHTTPClient)
	h := handler.NewHandler(nil, restCountriesClient)

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
