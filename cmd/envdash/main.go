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

	credPath := os.Getenv("FIREBASE_CREDENTIALS_FILE")
	if credPath == "" {
		//For docker deployment
		credPath = "/run/secrets/Firebase"
	}

	if !fileExists(credPath) {
		log.Fatal("Firebase credentials file not found at: " + credPath)
	}

	openAQAPIKey := os.Getenv("OPENAQ_API_KEY")
	if strings.TrimSpace(openAQAPIKey) == "" {
		log.Fatal("OPENAQ_API_KEY is not set")
	}

	client, err := firestore.NewClient(ctx, "cachemea2",
		option.WithCredentialsFile(credPath),
	)
	if err != nil {
		log.Fatal("Failed to initialize Firestore:", err)
	}

	defer client.Close()

	httpClient := utils.NewHttpClient()

	countryClient := countryclient.NewRestCountriesClient(httpClient)
	weatherClient := weatherclient.NewWeatherClient(httpClient)
	currencyClient := currencyclient.NewCurrencyClient(httpClient)
	aqClient := aqclient.NewOpenAQClient(httpClient, openAQAPIKey)

	cache := store.InitializeCache(
		countryClient,
		weatherClient,
		currencyClient,
		aqClient,
	)

	defer client.Close()
	st := store.NewFirestoreStore(client)
	h := handler.NewFirestoreHandler(st, cache)

	statusHandler := handler.NewStatusHandler(
		countryClient,
		weatherClient,
		aqClient,
		currencyClient,
		st,
		startedAt,
	)

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

	//Public routes:
	///Auth
	router.HandleFunc(utils.AUTHENTICATION_PATH, h.Auth)
	router.HandleFunc(utils.AUTHENTICATION_PATH+"/", h.Auth)
	router.HandleFunc(utils.AUTHENTICATION_PATH+"/{id}", h.Auth)
	router.HandleFunc(utils.AUTHENTICATION_PATH+"/{id}"+"/", h.Auth)
	///Status
	router.HandleFunc(utils.STATUS_PATH, statusHandler.GetStatus)
	router.HandleFunc(utils.STATUS_PATH+"/", statusHandler.GetStatus)

	//Private routes (api check in middelware)
	privateRouter := http.NewServeMux()
	privateRouter.HandleFunc("/", handler.DefaultHandler)

	//Dashboards
	privateRouter.HandleFunc(utils.DASHBOARD_PATH, h.DashboardHandler)
	privateRouter.HandleFunc(utils.DASHBOARD_PATH+"/", h.DashboardHandler)
	privateRouter.HandleFunc(utils.DASHBOARD_PATH+"/{id}", h.DashboardHandler)

	///Notification
	privateRouter.HandleFunc(utils.NOTIFICATION_PATH, h.NotificationSpinner)
	privateRouter.HandleFunc(utils.NOTIFICATION_PATH+"/", h.NotificationSpinner)
	privateRouter.HandleFunc(utils.NOTIFICATION_PATH+"/{id}", h.NotificationSpinnerById)
	privateRouter.HandleFunc(utils.NOTIFICATION_PATH+"/{id}/", h.NotificationSpinnerById)

	///Registration
	privateRouter.HandleFunc(utils.REGISTRATION_PATH, h.RegistrationHandler)
	privateRouter.HandleFunc(utils.REGISTRATION_PATH+"/", h.RegistrationHandler)

	//Only for some of the routnes, not global
	router.Handle("/", h.AuthMiddleware(privateRouter))

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

// fileExists checks if a file exists and is not a directory.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
