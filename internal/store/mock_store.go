package store

import (
	model "assignment-2/internal/models"
	"context"
	"errors"

	"cloud.google.com/go/firestore"
)

type MockStore struct {
	data    map[string]model.Registration
	apiKeys map[string]bool
}

func NewMockStore() *MockStore {
	return &MockStore{
		data: make(map[string]model.Registration),
		apiKeys: map[string]bool{
			"ec654fac9599f62e79e2706abef23dfb7c07c08185aa86db4d8695f0b718d1b3": true, // hashed "valid"
			"test-key": true,
		},
	}
}

// -------------------- Registration --------------------

func (m *MockStore) CreateRegistration(ctx context.Context, apiKey string, reg model.Registration) (string, error) {
	id := "test-id"
	reg.ID = id
	m.data[id] = reg
	return id, nil
}

func (m *MockStore) GetRegistration(ctx context.Context, apiKey string, id string) (*model.Registration, error) {
	reg, ok := m.data[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return &reg, nil
}

func (m *MockStore) GetAllRegistrations(ctx context.Context, apiKey string) ([]model.Registration, error) {
	var regs []model.Registration
	for _, r := range m.data {
		regs = append(regs, r)
	}
	return regs, nil
}

func (m *MockStore) UpdateRegistration(ctx context.Context, apiKey string, id string, reg model.Registration) error {
	if _, ok := m.data[id]; !ok {
		return errors.New("not found")
	}
	reg.ID = id
	m.data[id] = reg
	return nil
}

func (m *MockStore) DeleteRegistration(ctx context.Context, apiKey string, id string) error {
	if _, ok := m.data[id]; !ok {
		return errors.New("not found")
	}
	delete(m.data, id)
	return nil
}

func (m *MockStore) TweakRegistration(ctx context.Context, apiKey string, id string, patch model.RegistrationPatch) error {
	reg, ok := m.data[id]
	if !ok {
		return errors.New("not found")
	}
	if patch.Country != nil {
		reg.Country = *patch.Country
	}
	m.data[id] = reg
	return nil
}

// -------------------- Notifications --------------------

func (m *MockStore) CreateNotification(ctx context.Context, n model.RegisterWebhook, apiKey string) (string, error) {
	return "mock-id", nil
}

func (m *MockStore) GetAllNotificationsForUser(ctx context.Context, apiKey string) ([]model.AllRegisteredWebhook, error) {
	return []model.AllRegisteredWebhook{}, nil
}

func (m *MockStore) GetSpecificNotification(ctx context.Context, id string) (model.AllRegisteredWebhook, *firestore.DocumentRef, error) {
	return model.AllRegisteredWebhook{}, nil, errors.New("not found")
}

func (m *MockStore) GetAllNotifications(ctx context.Context) ([]model.AllRegisteredWebhook, error) {
	return []model.AllRegisteredWebhook{}, nil
}

func (m *MockStore) DeleteNotification(ctx context.Context, id string, apiKey string) error {
	return nil
}

// -------------------- Auth --------------------

func (m *MockStore) ApiKeyExists(ctx context.Context, apiKey string) bool {
	return m.apiKeys[apiKey]
}

func (m *MockStore) CreateApiStorage(ctx context.Context, reg model.Authentication) error {
	return nil
}

func (m *MockStore) FindUserWithApiKey(ctx context.Context, apiKey string) (string, error) {
	if m.apiKeys[apiKey] {
		return "user", nil
	}
	return "", errors.New("not found")
}

func (m *MockStore) CountApiPerUser(ctx context.Context, email string) (int, error) {
	return 1, nil
}

func (m *MockStore) DeleteAPIkey(ctx context.Context, apiKeyToDelete string, requestApiKey string) error {
	delete(m.apiKeys, apiKeyToDelete)
	return nil
}

// -------------------- Status --------------------

func (m *MockStore) DB_Status(ctx context.Context) bool {
	return true
}

func (m *MockStore) CountFirestore(ctx context.Context, collection string) (int, error) {
	return len(m.data), nil
}

func ValidStore() *MockStore {
	return &MockStore{
		data: make(map[string]model.Registration),
		apiKeys: map[string]bool{
			"ec654fac9599f62e79e2706abef23dfb7c07c08185aa86db4d8695f0b718d1b3": true, // hashed "valid"
		},
	}
}
