package store

import (
	model "assignment-2/internal/models"
	"cloud.google.com/go/firestore"
	"context"
	"crypto/sha256" //hash api key for database
	"encoding/hex"  //for converting hash to string
	"fmt"
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

/*
You may want to use this function to check is API key gives acces.

Checks if api key exists
Takes api key, hashes it, checks database if it exists.
This will be used when authenticating incomming api requests, if it exists returns true

This method is part of the Store struct, which holds the Firestore client.

@see			-hashAPIKey() for hashing implementation

@param ctx 		-keeping track of firestore connection(timeout etc)
@apiKey			-the key you want to check
@return bool	-if api key exists:true, if not in Firestore:false
*/
func (f *Store) ApiKeyExists(ctx context.Context, apiKey string) bool {
	hashedApiKey := hashAPIKey(apiKey)
	_, err := f.client.
		Collection("all_api_keys").
		Doc(hashedApiKey).
		Get(ctx)

	if err != nil {
		//api key cant be found e.g it is unique
		return false
	}

	//This api key exists!
	return true
}

/*
Storess API
Apis are currently stored in two different ways:

	1: All apis stored in one collection ass documents
			-Data stored: "time of creation" and what email used
	2: All users (email, addresses) have nested collection storing each api key
		These are the same api keys, stored in different ways
		This is donne for effecient lookup (if we letssay have 1 million users this would still work)
			-Data stored: "time of creation" and "name of api key"

This function hashes api key so no clairtext api key is stored on server

This method is part of the Store struct, which holds the Firestore client.

@see			-hashAPIKey() for hashing implementation

@param ctx 		- keeping track of firestore connection(timeout etc)
@param reg 		- struct of all data that we want to store (api key gets hashed)
@return error 	- if anny errors cam when storing api key in firestore, if nil, the keys were stored!
*/
func (f *Store) CreateApiStorage(ctx context.Context, reg model.Authentication) error {
	//first hashes api key generated:
	hashedApiKey := hashAPIKey(reg.ApiKey)
	//setts api
	AllApi := f.client.Collection("all_api_keys").Doc(hashedApiKey)
	_, err := AllApi.Set(ctx, map[string]interface{}{
		"time of creation": reg.CreatedAt,
		"user":             reg.Email,
	})

	emailDoc := f.client.Collection("authentication_info").Doc(reg.Email)
	//creating nested api key structure
	EmailApiDoc := emailDoc.Collection("api_keys").Doc(hashedApiKey)

	_, err = EmailApiDoc.Set(ctx, map[string]interface{}{
		"time of creation": reg.CreatedAt,
		"name of api key":  reg.Name,
	})
	if err != nil {
		return err
	}

	return nil
}

/*
Counts how manny Api's the speccified user have (checks Firestore)
and return the number of Apis's and anny errors if appropirate
If error, return 0 apis registerd to this user.

This method is part of the Store struct, which holds the Firestore client.

@param ctx	-keeping track of firestore connection(timeout etc)
@param email-The email you want to check, email is the user all apis are registerd under
@return int	-Return if anny, how manny Api's this user have registerd in Firestore, 0 if error
return error-Returns anny error and dont complete the function
*/
func (h *Store) CountApiPerUser(ctx context.Context, email string) (int, error) {
	//getting info about spesific email
	EmailDoc := h.client.Collection("authentication_info").Doc(email)
	//seeing how manny api keys that user hve
	ApiKeyDoc := EmailDoc.Collection("api_keys")

	doc, err := ApiKeyDoc.Documents(ctx).GetAll()
	if err != nil {
		return 0, err
	}
	//returns length of
	return len(doc), nil
}

/*
Deletes Api stored in Firestore. Deletes both places where Api is stored (global storage and per user)
First extract what email(user) this is api is registerd to, then delete in global storage (All_api_Keys)
Then delete this exact api from user.

This function don't delete user from database if this is the last api. This is because we
want to keep our user stored. We may want to enhance the functionality, and want to link maybe some other
information about this user

This method is part of the Store struct, which holds the Firestore client.

@see			-hashAPIKey() for hashing implementation

@param ctx		-keeping track of firestore connection(timeout etc)
@param apiKey	-api key from the user
@return error	-returns error if something goes wrong, example: wrong format stored in Firestore
*/
func (f *Store) DeleteAPIkey(ctx context.Context, apiKey string) error {
	apiKeyHashed := hashAPIKey(apiKey)
	docRef := f.client.Collection("all_api_keys").Doc(apiKeyHashed)

	// check if exists
	docSnap, err := docRef.Get(ctx)
	if err != nil {
		return err
	}

	//Finds mail to this user that this api is registerd under
	data := docSnap.Data()

	userMail, ok := data["user"].(string)
	if !ok {
		//TODO: log this in logg file
		return fmt.Errorf("Cant get email, user field missing or not a string (Firestore)")
	}

	//delete api under "ALL_API_KEYS"
	_, err = docRef.Delete(ctx)
	if err != nil {
		return err
	}

	//now goes to right user, and deletes that API key:
	userDoc := f.client.Collection("authentication_info").Doc(userMail)

	nestedDocRef := userDoc.Collection("api_keys").Doc(apiKeyHashed)

	_, err = nestedDocRef.Delete(ctx)
	if err != nil {
		return err
	}

	return nil
}

/*
Hashes API key
This is done BEFORE being stored in database
Use this function when checking api key in database, since all are hashed
Uses sha 256 hash. Stores as string.

@param apiKeyUnhashed	-String (api key) you want hashed
@return string			-Returns sha 256 hashed string
*/
func hashAPIKey(apiKeyUnhashed string) string {
	apiKeyHashed := sha256.Sum256([]byte(apiKeyUnhashed))
	apiKeyHashedString := hex.EncodeToString(apiKeyHashed[:])
	return apiKeyHashedString
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
