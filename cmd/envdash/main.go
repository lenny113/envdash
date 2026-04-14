package main

import (
	currencyclient "assignment-2/internal/client/currency"
	aqclient "assignment-2/internal/client/openaq"
	weatherclient "assignment-2/internal/client/openmeteo"
	countryclient "assignment-2/internal/client/restcountries"
	handler "assignment-2/internal/handlers"
	store "assignment-2/internal/store"
	utils "assignment-2/internal/utils"
	"context"
	"strings"
	"time"

	"log"
	"net/http"
	"os"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"
)

func main() {

	startedAt := time.Now() // starts the timer
	ctx := context.Background()

	credFile := os.Getenv("FIREBASE_CREDENTIALS_FILE")

	if !fileExists(credFile) {
		log.Panic("Firebase credentials file not found or not mounted")
	}

	openAQAPIKey := os.Getenv("OPENAQ_API_KEY")
	if strings.TrimSpace(openAQAPIKey) == "" {
		log.Panic("OPENAQ_API_KEY is not set")
	}

	client, err := firestore.NewClient(ctx, "cachemea2",
		option.WithCredentialsFile(credFile),
	)
	if err != nil {
		log.Fatal("Failed to initialize Firestore:", err)
	}
	defer client.Close()
	st := store.NewFirestoreStore(client)
	h := handler.NewFirestoreHandler(st)

	httpClient := utils.NewHttpClient()

	countryClient := countryclient.NewRestCountriesClient(httpClient)
	weatherClient := weatherclient.NewWeatherClient(httpClient)
	currencyClient := currencyclient.NewCurrencyClient(httpClient)

	aqClient := aqclient.NewOpenAQClient(httpClient, openAQAPIKey)

	statusHandler := handler.NewStatusHandler(
		countryClient,
		weatherClient,
		aqClient,
		currencyClient,
		nil, // keep notification/webhook plumbing as skeletons for now
		startedAt,
	)

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
	/*
		restCountriesHTTPClient := utils.NewHttpClient()
		restCountriesClient := client.NewRestCountriesClient(restCountriesHTTPClient)
		h := handler.NewHandler(nil, restCountriesClient)
		openAQClient := aqclient.NewOpenAQClient(httpClient, openAQAPIKey)
	*/

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

	//public routes:
	router.HandleFunc(utils.AUTHENTICATION_PATH, h.RegisterAuth)
	router.HandleFunc(utils.AUTHENTICATION_PATH+"/", h.RegisterAuth)
	router.HandleFunc(utils.STATUS_PATH, statusHandler.GetStatus)
	router.HandleFunc(utils.STATUS_PATH+"/", statusHandler.GetStatus)

	//Private routes (api check in middelware)
	privateRouter := http.NewServeMux()
	privateRouter.HandleFunc("/", handler.DefaultHandler)
	///Auth
	privateRouter.HandleFunc(utils.AUTHENTICATION_PATH, h.RegisterAuth)
	privateRouter.HandleFunc(utils.AUTHENTICATION_PATH+"/", h.RegisterAuth)
	privateRouter.HandleFunc(utils.AUTHENTICATION_PATH+"/{id}", h.DeleteAuth)
	privateRouter.HandleFunc(utils.AUTHENTICATION_PATH+"/{id}"+"/", h.DeleteAuth)
	///Notification
	privateRouter.HandleFunc(utils.NOTIFICATION_PATH, h.NotificationSpinner)
	privateRouter.HandleFunc(utils.NOTIFICATION_PATH+"/", h.NotificationSpinner)
	privateRouter.HandleFunc(utils.NOTIFICATION_PATH+"/{id}", h.NotificationSpinnerById)
	privateRouter.HandleFunc(utils.NOTIFICATION_PATH+"/{id}/", h.NotificationSpinnerById)

	///Registration
	privateRouter.HandleFunc(utils.REGISTRATION_PATH, h.RegistrationHandler)

	//Only for some of the routnes, not global
	router.Handle("/", h.AuthMiddleware(privateRouter))

	// Configure the HTTP server with the network address and
	// the router wrapped in logging middleware.
	router.HandleFunc(utils.REGISTRATION_PATH, h.RegistrationHandler)

	server := http.Server{
		Addr:    ":" + port,
		Handler: utils.Logging(router),
	}

	//Starting server
	log.Println("Starting server on port " + port)
	log.Fatal(server.ListenAndServe(), "Error starting server")
}

// fileExists checks if a file exists and is not a directory.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
