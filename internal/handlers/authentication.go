package handlers

import (
	model "assignment-2/internal/models"
	"assignment-2/internal/utils"
	"crypto/md5" //for generarting hash to create api key
	"crypto/rand"
	"encoding/hex" //for converting md5 hash to string
	"encoding/json"
	"net/http"
	"net/mail" //for email check (private emails also accepted)
	"time"     //time of creating api key and for generating unique api key based partly on time hash
)

type Login struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Key struct {
	ApiKey    string `json:"key"`
	CreatedAt string `json:"createdAt"`
}

func (h *Handler) Auth(w http.ResponseWriter, r *http.Request) {
	PathValue := r.PathValue("id")
	if PathValue == "" {
		h.RegisterAuth(w, r)
		return
	} else {
		h.DeleteAuth(w, r, PathValue)
	}
}

/*
This function is the handler for the /register endpoint, it is responsible for handling the registration of users and generating API keys for them.
The function first checks if the request method is POST, if not, it returns a method not allowed error.
Then it parses the JSON body of the request into a Login struct, and validates the input (check if name and email are not empty, and if email contains @).
If the input is valid, it generates an API key for the user by calling the createAPIKey function, and checks if the generated API key is already in use by calling the isAPIKeyUsed function.
If the API key is unique, it encodes the key struct to JSON and sends it back to the user as a response.

@param w - the http.ResponseWriter used to send the response back to the client
@param r - the http.Request containing the details of the incoming request, including the JSON body with the user's name and email

@see createAPIKey - the function responsible for generating a unique API key for the user
@see isAPIKeyUsed - the function responsible for checking if the generated API key is already in use and not empty or incomplete
*/
func (h *Handler) RegisterAuth(w http.ResponseWriter, r *http.Request) {
	//used in handling connection to Firestore
	ctx := r.Context()

	if r.Method != http.MethodPost { //if not a POST request, return method not allowed
		writeJSONError(w, http.StatusMethodNotAllowed, "Only POST method allowed")
		utils.SetMessageForLogger(w, "Method not allowed")
		return
	}

	// Parse the JSON body of the request into a Login struct
	var login Login
	defer r.Body.Close()

	err := json.NewDecoder(r.Body).Decode(&login)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON format")
		utils.SetMessageForLogger(w, "Invalid JSON format")
		return
	}

	// Validate the input (check if name and email are not empty, and if email contains @)
	if login.Email == "" || login.Name == "" {
		writeJSONError(w, http.StatusBadRequest, "Missing name or email")
		utils.SetMessageForLogger(w, "AUTH_REGISTER_FAIL: missing name/email")
		return
	}

	//check if email is correctly formated (RFC 5322), if not, return bad request
	if !isValidEmail(login.Email) {
		writeJSONError(w, http.StatusBadRequest, "Invalid email format")
		utils.SetMessageForLogger(w, "AUTH_REGISTER_FAIL: invalid email format")
		return
	}

	//getting how manny api keys the user alreaddy have to confirm is they can make more
	howMannyApiKeys, err := h.store.CountApiPerUser(ctx, login.Email)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to reach Firestore")
		utils.SetMessageForLogger(w, "AUTH_REGISTER_FAIL: failed to reach Firestore")
		return
	}
	if (howMannyApiKeys) > utils.MAXAPIKEYS-1 {
		writeJSONError(w, http.StatusTooManyRequests, "Too many api keys registered to this user, try deleting one")
		utils.SetMessageForLogger(w, "AUTH_REGISTER_FAIL: too many api keys registered to this user")
		return
	}

	// Generate API key for the user
	var createAPI, timeCreateApi string
	var ok bool

	//checks for duplicate API keys
	for i := 0; i < utils.MAXATTEMPTSFORKEYGENERATION; i++ {
		createAPI, timeCreateApi = createAPIKey(login.Email)

		if !h.store.ApiKeyExists(ctx, createAPI) {
			ok = true
			break
		}
	}

	//this would be very unlikely to happen, but if we have tried to generate a unique API key 10 times and failed, we return an error to avoid an infinite loop
	//this is to ensure that we dont have duplicate API keys, which would be a security risk and cause issues with the functionality of the API
	if !ok {
		utils.SetMessageForLogger(w, "AUTH_REGISTER_FAIL: api key generation exhausted")
		writeJSONError(w, http.StatusLoopDetected, "Failed to generate a unique API key")
		return
	}

	// Store the API key in Firestore
	reg := model.Authentication{
		Name:      login.Name,
		Email:     login.Email,
		ApiKey:    createAPI,
		CreatedAt: timeCreateApi,
	}

	err = h.store.CreateApiStorage(ctx, reg)
	if err != nil {
		utils.SetMessageForLogger(w, "Firestore write failed")
		writeJSONError(w, http.StatusInternalServerError, "Failed to save api keys in Firestore")
		return
	}

	//formatting the response to the user, which includes the generated API key and the time of API creation
	keyResponse := Key{
		ApiKey:    createAPI,
		CreatedAt: timeCreateApi,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(keyResponse); err != nil {
		utils.SetMessageForLogger(w, "Failed to encode response: "+err.Error())
		writeJSONError(w, http.StatusInternalServerError, "Failed to encode response")
		return
	}

}

/*
This function generates an API key for the user based on their email and the current time.
The API key is generated by creating a hash of the email and the current time using md5, and then encoding it to a string.
The API key is prefixed with "sk-envdash-" to make it more recognizable as an API key and .
The function also returns the time of API creation, which is formatted according to the specifications in the assignment.
This allows us to keep track of when the API key was created, which can be useful for expiration or auditing purposes.

@param email - the email of the user for whom the API key is being generated
@return createAPI - the generated API key for the user
@return timeCreateApi - the time of API creation, formatted according to the specifications in the assignment
*/
func createAPIKey(email string) (string, string) {

	timeCreateApi := time.Now().Format("20060102 15:04") //this is the format specified in the assignement for time

	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		panic(err) // eller håndter bedre
	}

	// Generate hash of email + current time using md5
	//this will be used as the api key for the user, and will be unique for each registration
	//unique even, even same email cant make same key, because of the time component
	hash := md5.Sum([]byte(email + time.Now().String() + hex.EncodeToString(b)))
	hashString := hex.EncodeToString(hash[:])
	createAPI := utils.STARTOFUSERAPI + hashString

	return createAPI, timeCreateApi
}

/*
Checks if the email provided is valid, if so return true
*/
func isValidEmail(email string) bool {
	address, err := mail.ParseAddress(email)

	if err != nil {
		return false
	}

	return address != nil
}

func (h *Handler) DeleteAuth(w http.ResponseWriter, r *http.Request, ApiToDelete string) {
	//used in handling connection to Firestore
	ctx := r.Context()

	if r.Method != http.MethodDelete { //if not a DELETE request, return method not allowed
		writeJSONError(w, http.StatusMethodNotAllowed, "Only DELETE method allowed")
		utils.SetMessageForLogger(w, "Method not allowed")
		return
	}

	//Checks if api key exists middelware can not do this because of routing issues
	ApiUserAuth := r.Header.Get("X-Api-Key")
	if !h.store.ApiKeyExists(ctx, ApiUserAuth) {
		//extended message, two api keys here, so need to be carefull so the user understand
		writeJSONError(w, http.StatusForbidden, "Looks like your api key in header are wrong")
		utils.SetMessageForLogger(w, "api key in header are wrong")
		return
	}

	//we have to delete this key in firestore
	err := h.store.DeleteAPIkey(ctx, ApiToDelete, ApiUserAuth)
	if err != nil {

		if err.Error() == "api key not found" {
			writeJSONError(w, http.StatusNotFound, "Can not find API key you want to delete")

		} else if err.Error() == "unauthorized" {
			writeJSONError(w, http.StatusForbidden, "Not allowed to delete someone else's api key!")

		} else {
			writeJSONError(w, http.StatusInternalServerError, "Something went wrong when deleting api key")
		}

		messageForLogger := "Problem in firestore, while trying to delete apikey: " + err.Error()
		utils.SetMessageForLogger(w, messageForLogger)
		return
	}

	//if it goes through, it works
	utils.SetMessageForLogger(w, "DELETE_AUTH_SUCCESS")
	w.WriteHeader(http.StatusNoContent)

}
