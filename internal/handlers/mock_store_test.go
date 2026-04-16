package handlers

import (
	"assignment-2/internal/models"
	"context"

	"cloud.google.com/go/firestore"
)

// MockStore is a test double for the Store interface used by Handler.
// Each field is a function so individual tests can override only what they need.
type MockStore struct {
	CreateNotificationFn         func(ctx context.Context, n models.RegisterWebhook, apiKey string) (string, error)
	GetAllNotificationsForUserFn func(ctx context.Context, apiKey string) ([]models.AllRegisteredWebhook, error)
	GetSpecificNotificationFn    func(ctx context.Context, id string) (models.AllRegisteredWebhook, *firestore.DocumentRef, error)
	DeleteNotificationFn         func(ctx context.Context, id string, apiKey string) error
	GetAllNotificationsFn        func(ctx context.Context) ([]models.AllRegisteredWebhook, error)
	GetRegistrationFn            func(ctx context.Context, apiKey string, id string) (*models.Registration, error)
	CreateRegistrationFn         func(ctx context.Context, apiKey string, reg models.Registration) (string, error)
	GetAllRegistrationsFn        func(ctx context.Context, apiKey string) ([]models.Registration, error)
	UpdateRegistrationFn         func(ctx context.Context, apiKey string, id string, reg models.Registration) error
	DeleteRegistrationFn         func(ctx context.Context, apiKey string, id string) error
	TweakRegistrationFn          func(ctx context.Context, apiKey string, id string, patch models.RegistrationPatch) error
	ApiKeyExistsFn               func(ctx context.Context, apiKey string) bool
	CreateApiStorageFn           func(ctx context.Context, reg models.Authentication) error
	FindUserWithApiKeyFn         func(ctx context.Context, apiKey string) (string, error)
	CountApiPerUserFn            func(ctx context.Context, email string) (int, error)
	DeleteAPIkeyFn               func(ctx context.Context, apiKeyToDelete string, requestApiKey string) error
	DB_StatusFn                  func(ctx context.Context) bool
	CountFirestoreFn             func(ctx context.Context, collection string) (int, error)
}

func (m *MockStore) CreateNotification(ctx context.Context, n models.RegisterWebhook, apiKey string) (string, error) {
	return m.CreateNotificationFn(ctx, n, apiKey)
}
func (m *MockStore) GetAllNotificationsForUser(ctx context.Context, apiKey string) ([]models.AllRegisteredWebhook, error) {
	return m.GetAllNotificationsForUserFn(ctx, apiKey)
}
func (m *MockStore) GetSpecificNotification(ctx context.Context, id string) (models.AllRegisteredWebhook, *firestore.DocumentRef, error) {
	return m.GetSpecificNotificationFn(ctx, id)
}
func (m *MockStore) DeleteNotification(ctx context.Context, id string, apiKey string) error {
	return m.DeleteNotificationFn(ctx, id, apiKey)
}
func (m *MockStore) GetAllNotifications(ctx context.Context) ([]models.AllRegisteredWebhook, error) {
	return m.GetAllNotificationsFn(ctx)
}
func (m *MockStore) GetRegistration(ctx context.Context, apiKey string, id string) (*models.Registration, error) {
	return m.GetRegistrationFn(ctx, apiKey, id)
}
func (m *MockStore) CreateRegistration(ctx context.Context, apiKey string, reg models.Registration) (string, error) {
	return m.CreateRegistrationFn(ctx, apiKey, reg)
}
func (m *MockStore) GetAllRegistrations(ctx context.Context, apiKey string) ([]models.Registration, error) {
	return m.GetAllRegistrationsFn(ctx, apiKey)
}
func (m *MockStore) UpdateRegistration(ctx context.Context, apiKey string, id string, reg models.Registration) error {
	return m.UpdateRegistrationFn(ctx, apiKey, id, reg)
}
func (m *MockStore) DeleteRegistration(ctx context.Context, apiKey string, id string) error {
	return m.DeleteRegistrationFn(ctx, apiKey, id)
}
func (m *MockStore) TweakRegistration(ctx context.Context, apiKey string, id string, patch models.RegistrationPatch) error {
	return m.TweakRegistrationFn(ctx, apiKey, id, patch)
}
func (m *MockStore) ApiKeyExists(ctx context.Context, apiKey string) bool {
	return m.ApiKeyExistsFn(ctx, apiKey)
}
func (m *MockStore) CreateApiStorage(ctx context.Context, reg models.Authentication) error {
	return m.CreateApiStorageFn(ctx, reg)
}
func (m *MockStore) FindUserWithApiKey(ctx context.Context, apiKey string) (string, error) {
	return m.FindUserWithApiKeyFn(ctx, apiKey)
}
func (m *MockStore) CountApiPerUser(ctx context.Context, email string) (int, error) {
	return m.CountApiPerUserFn(ctx, email)
}
func (m *MockStore) DeleteAPIkey(ctx context.Context, apiKeyToDelete string, requestApiKey string) error {
	return m.DeleteAPIkeyFn(ctx, apiKeyToDelete, requestApiKey)
}
func (m *MockStore) DB_Status(ctx context.Context) bool {
	return m.DB_StatusFn(ctx)
}
func (m *MockStore) CountFirestore(ctx context.Context, collection string) (int, error) {
	return m.CountFirestoreFn(ctx, collection)
}

// newMockHandler creates a Handler backed by the given MockStore.
func newMockHandler(store *MockStore) *Handler {
	return &Handler{store: store}
}
