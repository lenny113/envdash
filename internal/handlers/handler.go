package handlers

import (
	client "assignment-2/internal/client/restcountries"
	utils "assignment-2/internal/models"
	"assignment-2/internal/store"
	"context"
	"encoding/json"
	"net/http"

	models "assignment-2/internal/models"
	"cloud.google.com/go/firestore"
)

// StoreInterface allows both *store.FireStore and *store.MockStore to be used interchangeably.
type StoreInterface interface {
	CreateRegistration(ctx context.Context, apiKey string, reg models.Registration) (string, error)
	GetRegistration(ctx context.Context, apiKey string, id string) (*models.Registration, error)
	GetAllRegistrations(ctx context.Context, apiKey string) ([]models.Registration, error)
	UpdateRegistration(ctx context.Context, apiKey string, id string, reg models.Registration) error
	DeleteRegistration(ctx context.Context, apiKey string, id string) error
	TweakRegistration(ctx context.Context, apiKey string, id string, patch models.RegistrationPatch) error
	ApiKeyExists(ctx context.Context, keyHash string) bool
	CreateApiStorage(ctx context.Context, reg models.Authentication) error
	FindUserWithApiKey(ctx context.Context, apiKey string) (string, error)
	CountApiPerUser(ctx context.Context, email string) (int, error)
	DeleteAPIkey(ctx context.Context, apiKeyToDelete string, requestApiKey string) error

	// Notifications
	CreateNotification(ctx context.Context, n models.RegisterWebhook, apiKey string) (string, error)
	GetAllNotificationsForUser(ctx context.Context, apiKey string) ([]models.AllRegisteredWebhook, error)
	GetSpecificNotification(ctx context.Context, id string) (models.AllRegisteredWebhook, *firestore.DocumentRef, error)
	GetAllNotifications(ctx context.Context) ([]models.AllRegisteredWebhook, error)
	DeleteNotification(ctx context.Context, id string, apiKey string) error

	// Status
	DB_Status(ctx context.Context) bool
	CountFirestore(ctx context.Context, collection string) (int, error)
}

// CacheInterface allows both *store.Cache and mockCache to be used interchangeably.
type CacheInterface interface {
	RequestFromCache(req store.CacheExternalRequest) (*store.CacheResponse, error)
}

type Handler struct {
	store               StoreInterface
	restCountriesClient client.RestCountriesClient
	cache               CacheInterface
}

func NewHandler(s StoreInterface, restCountriesClient client.RestCountriesClient) *Handler {
	return &Handler{
		store:               s,
		restCountriesClient: restCountriesClient,
	}
}

func NewFirestoreHandler(s *store.FireStore, cache CacheInterface) *Handler {
	return &Handler{
		store: s,
		cache: cache,
	}
}

func writeJSONError(w http.ResponseWriter, code int, errMsg string) {
	response := utils.ErrorResponse{
		Code:    code,
		Message: errMsg,
	}

	jsonBytes, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(jsonBytes)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
