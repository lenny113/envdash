package handlers

import (
	"assignment-2/internal/models"
	"assignment-2/internal/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time" //for generating time of creation of notification
)

func (h *Handler) NotificationSpinner(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		h.postRequest(w, r)
	case "GET":
		h.allNotifications(w, r)
	default:
		http.Error(w, "method is not ok", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) NotificationSpinnerById(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		h.specificNotification(w, r)
	case "DELETE":
		h.deleteNotification(w, r)
	default:
		http.Error(w, "method is not ok", http.StatusMethodNotAllowed)
	}

}

func (h *Handler) postRequest(w http.ResponseWriter, r *http.Request) {

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
	request.Event = strings.ToUpper(request.Event) //convert event to uppercase to make it case insensitive

	//checks if the required fields are present and valid, if not, it returns an error with the specific missing fields
	err, errorMessage := validateNotification(request)
	if err != nil {
		http.Error(w, errorMessage, http.StatusBadRequest)
		return
	}

	//Stores in Firestore
	notificationId, err := h.store.CreateNotification(r.Context(), request)
	if err != nil {
		utils.SetMessageForLogger(w, "Error creating notification in Firestore")
		http.Error(w, "Error creating notification", http.StatusInternalServerError)
		return
	}
	//time created
	timeCreated := time.Now().Format("20060102 15:04")

	//After successful creation of notification, it returns a success message:
	var response models.RegisteredWebhookResponse
	response.Id = notificationId
	response.Country = request.Country
	response.Event = request.Event
	response.Time = timeCreated

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	responseJSON, err := json.MarshalIndent(response, "", "   ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(responseJSON)

}

func validateNotification(request models.RegisterWebhook) (error, string) {
	var errors []string

	//URL check:
	if request.Url == "" {
		errors = append(errors, "Missing URL in request body")
	}
	//TODO: add check for valid URL, maybe by using regex or the net/url package

	//country check
	if request.Country == "" {
		errors = append(errors, "Missing Country in request body")
	}
	//TODO: add check for valid country, maybe by checking if the country is in the list of countries we have in our database

	//event check
	if request.Event == "" {
		errors = append(errors, "Missing Event in request body")
	}

	//if there are any errors, return them as a single string,
	if len(errors) > 0 {
		return fmt.Errorf("validation failed"), strings.Join(errors, ", ")
	}

	//If there are no errors, check if the event is one of the supported events, if not, it returns an error with the valid events
	//This is last because if there are missing fields, it is not necessary to check if the event is valid,
	// and it is more efficient to check for missing fields first before checking for valid events

	//checks if event is one of the supported events
	find := false
	for _, validEvent := range utils.VALIDEVENTS {
		if strings.ToUpper(request.Event) == validEvent {
			find = true
			break
		}
	}
	if !find {
		return fmt.Errorf("validation failed"), "Invalid Event in request body, valid events are: " + strings.Join(utils.VALIDEVENTS, ", ")
	}

	//If there are no errors, return nil
	return nil, ""

}

func (h *Handler) allNotifications(w http.ResponseWriter, r *http.Request) {
	AllSaved, err := h.store.GetAllNotifications(r.Context())
	if err != nil {
		http.Error(w, "Error fetching notifications", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	responseJSON, err := json.MarshalIndent(AllSaved, "", "   ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(responseJSON)

}

func (h *Handler) specificNotification(w http.ResponseWriter, r *http.Request) {
	//get id from url path
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Missing id in URL path", http.StatusBadRequest)
		return
	}

	notification, err := h.store.GetSpecificNotification(r.Context(), id)
	if err != nil {
		//if not found, return 404 error
		http.Error(w, "Notification not found", http.StatusNotFound)
		return
	}

	//if found, return the notification as json
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	responseJSON, err := json.MarshalIndent(notification, "", "   ")
	if err != nil {
		//if there is an error marshaling the response, return 500 error
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//send the response back to the client
	w.Write(responseJSON)
}

func (h *Handler) deleteNotification(w http.ResponseWriter, r *http.Request) {
	//get id from url path
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Missing id in URL path", http.StatusBadRequest)
		return
	}
	err := h.store.DeleteNotification(r.Context(), id)
	if err != nil {
		//if not found, return 404 error
		http.Error(w, "Notification not found", http.StatusNotFound)
		return
	}
	//if deleted successfully, return a success message
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
	w.Write([]byte("Notification with id " + id + " successfully deleted"))
}
