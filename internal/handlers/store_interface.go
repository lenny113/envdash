// internal/handlers/store_interface.go
package handlers

import (
	"assignment-2/internal/models"
	"context"

	"cloud.google.com/go/firestore"
)

// StoreInterface
// *store.FireStore oppfyller dette automatisk – ingen endring i store-pakken.
type StoreInterface interface {
	// Registratios
	CreateRegistration(ctx context.Context, apiKey string, reg models.Registration) (string, error)
	GetRegistration(ctx context.Context, apiKey string, id string) (*models.Registration, error)
	GetAllRegistrations(ctx context.Context, apiKey string) ([]models.Registration, error)
	UpdateRegistration(ctx context.Context, apiKey string, id string, reg models.Registration) error
	DeleteRegistration(ctx context.Context, apiKey string, id string) error
	TweakRegistration(ctx context.Context, apiKey string, id string, patch models.RegistrationPatch) error

	// Notifications
	CreateNotification(ctx context.Context, n models.RegisterWebhook, apiKey string) (string, error)
	GetAllNotificationsForUser(ctx context.Context, apiKey string) ([]models.AllRegisteredWebhook, error)
	GetSpecificNotification(ctx context.Context, id string) (models.AllRegisteredWebhook, *firestore.DocumentRef, error)
	GetAllNotifications(ctx context.Context) ([]models.AllRegisteredWebhook, error)
	DeleteNotification(ctx context.Context, id string, apiKey string) error

	// Authentification
	ApiKeyExists(ctx context.Context, apiKey string) bool
	CreateApiStorage(ctx context.Context, reg models.Authentication) error
	FindUserWithApiKey(ctx context.Context, apiKey string) (string, error)
	CountApiPerUser(ctx context.Context, email string) (int, error)
	DeleteAPIkey(ctx context.Context, apiKeyToDelete string, requestApiKey string) error

	// Status
	DB_Status(ctx context.Context) bool
	CountFirestore(ctx context.Context, collection string) (int, error)
}
