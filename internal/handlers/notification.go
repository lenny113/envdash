package handlers

import (
	"assignment-2/internal/models"
	"assignment-2/internal/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func (h *Handler) NotificationSpinner(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		h.postRequest(w, r)
	case "GET":
		fmt.Println("METHOD GET")
	default:
		http.Error(w, "method is not ok", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) NotificationSpinnerById(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		fmt.Println("METHOD GET")
	case "DELETE":
		fmt.Println("METHOD DELETE")
	default:
		http.Error(w, "method is not ok", http.StatusMethodNotAllowed)
	}

}

func (h *Handler) postRequest(w http.ResponseWriter, r *http.Request) {
	fmt.Println("METHOD POST")

	if r.Body == nil {
		http.Error(w, "Please send a request body", http.StatusBadRequest)
		return
	}

	var request models.RegisterWebhook

	//checks if json is valid and can be decoded into the struct, if not, it returns an error
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}


	//Stores in Firestore
	err = h.store.CreateNotification(r.Context(), request)
	if err != nil {
		utils.SetMessageForLogger(w, "Error creating notification in Firestore")
		http.Error(w, "Error creating notification", http.StatusInternalServerError)
		return
	}

	//Prints the received notification to the console for testing purposes
	fmt.Printf("URL: %s, Country: %s, Event: %s\n",
		request.Url, request.Country, request.Event)
}
