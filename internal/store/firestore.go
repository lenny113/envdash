package store

import (
	model "assignment-2/utils/models"
	"context"
	"reflect"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

// FireStore wraps the Firestore client and provides methods
// for interacting with data in firestore
type FireStore struct {
	client *firestore.Client
}

// NewFirestoreStore initializes a new firestore instance with the
// provided Firestore client
func NewFirestoreStore(client *firestore.Client) *FireStore {
	return &FireStore{client: client}
}

// CreateRegistration stores a new registration under an API key.
// A document id is generated automatically by firestore and assigned to the registratrion
//
// Returns:
// - string: generated document ID
// - error:  if the operation fails
func (f *FireStore) CreateRegistration(ctx context.Context, apiKey string, reg model.Registration) (string, error) {

	// Navigate to user_registrations subcollection
	col := f.client.Collection("registrations").Doc(apiKey).Collection("user_registrations")

	//generate new doc reference per auto-ID
	doc := col.NewDoc()

	//assign generated id to the document
	reg.ID = doc.ID

	//store registration in firestore
	_, err := doc.Set(ctx, reg)
	if err != nil {
		return "", err
	}

	return doc.ID, nil
}

// GetRegistration retrieves a single registration by ID
//
// Returns:
// - *model.Registration: the requested registration
// - error: if the document does not exist or decoding fails
func (f *FireStore) GetRegistration(ctx context.Context, apiKey string, id string) (*model.Registration, error) {

	doc, err := f.client.Collection("registrations").
		Doc(apiKey).
		Collection("user_registrations").
		Doc(id).
		Get(ctx)

	if err != nil {
		return nil, err
	}

	var reg model.Registration
	if err := doc.DataTo(&reg); err != nil {
		return nil, err
	}

	// Include the document ID
	reg.ID = doc.Ref.ID

	return &reg, nil
}

// GetAllRegistrations retrieves all registrations associated with an API key.
//
// Returns:
// - []model.Registration: list of registrations
// - error: if iteration or decoding fails
func (f *FireStore) GetAllRegistrations(ctx context.Context, apiKey string) ([]model.Registration, error) {
	iter := f.client.Collection("registrations").
		Doc(apiKey).
		Collection("user_registrations").
		Documents(ctx)

	defer iter.Stop()

	var registrations []model.Registration

	// Iterate through all documents in collection
	for {
		doc, err := iter.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}
			return nil, err
		}

		var reg model.Registration
		doc.DataTo(&reg)

		// Include the document ID
		reg.ID = doc.Ref.ID

		registrations = append(registrations, reg)
	}

	return registrations, nil
}

// UpdateRegistration replaces an existing registration entirely.
//
// NOTE: This performs a full overwrite of the document.
func (f *FireStore) UpdateRegistration(ctx context.Context, apiKey string, id string, reg model.Registration) error {
	docRef := f.client.Collection("registrations").
		Doc(apiKey).
		Collection("user_registrations").
		Doc(id)
	_, err := docRef.Set(ctx, reg)
	if err != nil {
		return err
	}
	// Overwrites the entire document
	_, err = docRef.Set(ctx, reg)
	return err
}

// DeleteRegistration removes a registration by ID.
//
// Returns error if the document does not exist or deletion fails.
func (f *FireStore) DeleteRegistration(ctx context.Context, apiKey string, id string) error {

	docRef := f.client.Collection("registrations").Doc(apiKey).Collection("user_registrations").Doc(id)

	// check if exists
	_, err := docRef.Get(ctx)
	if err != nil {
		return err
	}

	_, err = docRef.Delete(ctx)
	return err
}

// TweakRegistration performs a partial update (PATCH) on a registration.
//
// Only fields provided in the patch object are updated.
// Uses Firestore's Update() to modify specific fields instead of overwriting.
//
// Supports nested updates for the "features" object using reflection.
func (f *FireStore) TweakRegistration(
	ctx context.Context,
	apiKey string,
	id string,
	patch model.RegistrationPatch,
) error {

	docRef := f.client.Collection("registrations").
		Doc(apiKey).
		Collection("user_registrations").
		Doc(id)

	var updates []firestore.Update

	// Handle top-level fields
	if patch.Country != nil {
		updates = append(updates, firestore.Update{
			Path:  "country",
			Value: *patch.Country,
		})
	}

	if patch.IsoCode != nil {
		updates = append(updates, firestore.Update{
			Path:  "isoCode",
			Value: *patch.IsoCode,
		})
	}

	// Handle nested "features" fields dynamically using reflection

	if patch.Features != nil {
		v := reflect.ValueOf(*patch.Features)
		t := reflect.TypeOf(*patch.Features)

		for i := 0; i < v.NumField(); i++ {
			fieldValue := v.Field(i)
			fieldType := t.Field(i)

			// Skip nil pointers safely
			if fieldValue.Kind() == reflect.Ptr && fieldValue.IsNil() {
				continue
			}

			// Extract JSON tag to match Firestore field naming
			jsonTag := fieldType.Tag.Get("json")
			if jsonTag == "" {
				continue
			}
			jsonTag = strings.Split(jsonTag, ",")[0]

			// Extract actual value (handle pointer vs non-pointer)
			var value interface{}
			if fieldValue.Kind() == reflect.Ptr {
				value = fieldValue.Elem().Interface()
			} else {
				value = fieldValue.Interface()
			}
			// Append nested field update (e.g., "features.temperature")
			updates = append(updates, firestore.Update{
				Path:  "features." + jsonTag,
				Value: value,
			})
		}
	}

	// Update timestamp
	updates = append(updates, firestore.Update{
		Path:  "lastChange",
		Value: time.Now().Format("20060102 15:04"),
	})

	// Execute partial update
	_, err := docRef.Update(ctx, updates)
	return err
}

// APIKeyExists checks whether a given API key (hashed) exists in Firestore.
//
// Returns:
// - true if the key exists
// - false if not found or an error occurs
//
// NOTE: Errors are treated as "not found" for simplicity.
func (f *FireStore) APIKeyExists(ctx context.Context, keyHash string) bool {
	_, err := f.client.Collection("all_api_keys").Doc(keyHash).Get(ctx)
	if err != nil {
		return false
	}
	return true

}
