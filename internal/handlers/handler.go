package handlers

import (
	"assignment-2/internal/client"
	utils "assignment-2/internal/models"
	"assignment-2/internal/store"
	"encoding/json"
	"net/http"
)

type Handler struct {
	store               *store.Store // firestore
	restCountriesClient client.RestCountriesClient
}

/*
func NewHandler(s *store.Store, restCountriesClient client.RestCountriesClient) *Handler {
	return &Handler{
		store:               s,
		restCountriesClient: restCountriesClient,
	}
}
*/
func NewHandler(s *store.Store) *Handler {
	return &Handler{store: s}
}
func writeJSONError(w http.ResponseWriter, code int, errMsg string) {
	// Create an instance of the custom error struct
	response := utils.ErrorResponse{
		Code:    code,
		Message: errMsg,
	}

	// Marshal the struct into a JSON byte slice
	jsonBytes, err := json.Marshal(response)
	if err != nil {
		// If marshaling fails (rare), fall back to a plain text error
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Set the Content-Type header to application/json
	w.Header().Set("Content-Type", "application/json")

	// Set the HTTP status code
	w.WriteHeader(code)

	// Write the JSON response body
	w.Write(jsonBytes)
}
