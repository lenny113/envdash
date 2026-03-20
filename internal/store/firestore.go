package store

import (
	model "assignment-2/internal/models"
	"context"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

type Store struct {
	client *firestore.Client
}

func NewFirestoreStore(client *firestore.Client) *Store {
	return &Store{client: client}
}

func (f *Store) CreateRegistration(ctx context.Context, reg model.Registration) (string, error) {

	doc := f.client.Collection("registrations").NewDoc()

	reg.ID = doc.ID

	_, err := doc.Set(ctx, reg)
	if err != nil {
		return "", err
	}

	return doc.ID, nil
}

func (f *Store) GetRegistration(ctx context.Context, id string) (*model.Registration, error) {
	doc, err := f.client.Collection("registrations").Doc(id).Get(ctx)
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

func (f *Store) GetAllRegistrations(ctx context.Context) ([]model.Registration, error) {
	iter := f.client.Collection("registrations").Documents(ctx)
	defer iter.Stop()

	var registrations []model.Registration

	for {
		doc, err := iter.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}
			return nil, err
		}

		var reg model.Registration
		if err := doc.DataTo(&reg); err != nil {
			return nil, err
		}

		reg.ID = doc.Ref.ID

		registrations = append(registrations, reg)
	}

	return registrations, nil
}

func (f *Store) UpdateRegistration(ctx context.Context, id string, reg model.Registration) error {

	docRef := f.client.Collection("registrations").Doc(id)

	// Check if exists
	_, err := docRef.Get(ctx)
	if err != nil {
		return err
	}

	reg.ID = id

	_, err = docRef.Set(ctx, reg) // replaces entire document
	if err != nil {
		return err
	}

	return nil
}

func (f *Store) DeleteRegistration(ctx context.Context, id string) error {

	docRef := f.client.Collection("registrations").Doc(id)

	// check if exists
	_, err := docRef.Get(ctx)
	if err != nil {
		return err
	}

	_, err = docRef.Delete(ctx)
	return err
}

// function to change a specific part/parts of a registration with use of the patch method
//func (f *Store) TweakRegistration(ctx context.Context) error {}
