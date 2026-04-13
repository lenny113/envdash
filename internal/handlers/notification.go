package handlers

import (
	"assignment-2/internal/models"
	"assignment-2/internal/utils"
	"bytes"   //for sending the payload in the POST request to the webhook URL
	"context" //for handling context in the checkWhatNotificationsToSend function
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time" //for generating time of creation of notification
)

func (h *Handler) NotificationSpinner(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		h.registerNewNotification(w, r)
	case "GET":
		h.allNotifications(w, r)
	default:
		writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		utils.SetMessageForLogger(w, "Method not allowed in NotificationSpinner: "+r.Method)
	}
}

func (h *Handler) NotificationSpinnerById(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		h.specificNotification(w, r)
	case "DELETE":
		h.deleteNotification(w, r)
	default:
		writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		utils.SetMessageForLogger(w, "Method not allowed in NotificationSpinnerById: "+r.Method)
	}

}

func (h *Handler) registerNewNotification(w http.ResponseWriter, r *http.Request) {

	if r.Body == nil {
		utils.SetMessageForLogger(w, "Missing request body in notification registration")
		writeJSONError(w, http.StatusBadRequest, "Missing request body in notification registration")
		return
	}

	var request models.RegisterWebhook

	//checks if json is valid and can be decoded into the struct, if not, it returns an error
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		utils.SetMessageForLogger(w, "Invalid JSON in notification registration")
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON in notification registration")
		return
	}
	request.Event = strings.ToUpper(request.Event) //convert event to uppercase to make it case insensitive

	//checks if the required fields are present and valid, if not, it returns an error with the specific missing fields
	err, errorMessage := validateNotification(request)
	if err != nil {
		utils.SetMessageForLogger(w, "Invalid request body in notification registration: "+errorMessage)
		writeJSONError(w, http.StatusBadRequest, "Invalid request body in notification registration: "+errorMessage)
		return
	}

	//Only allows threashold field if event is threashold
	if request.Event != "THRESHOLD" && request.ThresholdNotification != nil {
		utils.SetMessageForLogger(w, "Threshold is only allowed when event is THRESHOLD")
		writeJSONError(w, http.StatusBadRequest, "Threshold is only allowed when event is THRESHOLD")
		return
	}

	//checks if notifcation is threashold:
	if request.Event == "THRESHOLD" {
		//check if threashold body is present, if not, return an error
		if request.ThresholdNotification == nil {
			utils.SetMessageForLogger(w, "Missing threshold")
			writeJSONError(w, http.StatusBadRequest, "Missing threshold")
			return
		}

		//reed the threashold body
		//DELETE?var threshold models.ThresholdNotification

		//validate threashold body for valid values
		err, errorMessage := validateThreshold(*request.ThresholdNotification)
		if err != nil {
			utils.SetMessageForLogger(w, "Invalid threshold in notification registration: "+errorMessage)
			writeJSONError(w, http.StatusBadRequest, "Invalid threshold in notification registration: "+errorMessage)
			return
		}
		request.ThresholdNotification.Field = strings.ToUpper(request.ThresholdNotification.Field) //convert threashold field to uppercase to make it case insensitive
	}

	//Stores in Firestore
	notificationId, err := h.store.CreateNotification(r.Context(), request)
	if err != nil {
		utils.SetMessageForLogger(w, "Error creating notification in Firestore")
		writeJSONError(w, http.StatusInternalServerError, "Error creating notification")
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
		utils.SetMessageForLogger(w, "Error marshaling response JSON in notification registration")
		writeJSONError(w, http.StatusInternalServerError, "Error marshaling response JSON in notification registration")
		return
	}

	w.Write(responseJSON)

	utils.SetMessageForLogger(w, "Notification registered with id "+response.Id)

}

func validateNotification(request models.RegisterWebhook) (error, string) {
	var errors []string

	if request.Url == "" {
		errors = append(errors, "Missing URL")
	} else if !strings.HasPrefix(request.Url, "http://") && !strings.HasPrefix(request.Url, "https://") {
		errors = append(errors, "Invalid URL")
	}

	//TODO: add check for valid country, maybe by checking if the country is in the list of countries we have in our database

	//event check
	if request.Event == "" {
		errors = append(errors, "Missing Event in request body, valid events are:"+strings.Join(utils.VALIDEVENTS, ", "))
	} else {
		find := false
		for _, validEvent := range utils.VALIDEVENTS {
			if strings.ToUpper(request.Event) == validEvent {
				find = true
				break
			}
		}
		if !find {
			errors = append(errors, "Invalid Event in request body, valid events are: "+strings.Join(utils.VALIDEVENTS, ", "))
		}
	}

	//if there are any errors, return them as a single string,
	if len(errors) > 0 {
		return fmt.Errorf("validation failed"), strings.Join(errors, ", ")
	}

	//If there are no errors, check if the event is one of the supported events, if not, it returns an error with the valid events
	//This is last because if there are missing fields, it is not necessary to check if the event is valid,
	// and it is more efficient to check for missing fields first before checking for valid events

	//checks if event is one of the supported events

	//If there are no errors, return nil
	return nil, ""

}

func (h *Handler) allNotifications(w http.ResponseWriter, r *http.Request) {
	AllSaved, err := h.store.GetAllNotifications(r.Context())
	if err != nil {
		utils.SetMessageForLogger(w, "Error fetching notifications")
		writeJSONError(w, http.StatusInternalServerError, "Error fetching notifications")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	responseJSON, err := json.MarshalIndent(AllSaved, "", "   ")
	if err != nil {
		utils.SetMessageForLogger(w, "Error marshaling response JSON in fetching notifications")
		writeJSONError(w, http.StatusInternalServerError, "Error marshaling response JSON in fetching notifications")
		return
	}
	w.Write(responseJSON)
	utils.SetMessageForLogger(w, "Fetched all notifications, count: "+fmt.Sprint(len(AllSaved)))

}

func (h *Handler) specificNotification(w http.ResponseWriter, r *http.Request) {
	//get id from url path
	id := r.PathValue("id")
	if id == "" {
		utils.SetMessageForLogger(w, "Missing id in URL path")
		writeJSONError(w, http.StatusBadRequest, "Missing id in URL path")
		return
	}

	notification, err := h.store.GetSpecificNotification(r.Context(), id)
	if err != nil {
		//if not found, return 404 error
		utils.SetMessageForLogger(w, "Notification not found")
		writeJSONError(w, http.StatusNotFound, "Notification not found")
		return
	}

	//if found, return the notification as json
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	responseJSON, err := json.MarshalIndent(notification, "", "   ")
	if err != nil {
		//if there is an error marshaling the response, return 500 error
		utils.SetMessageForLogger(w, "Error marshaling response JSON in fetching specific notification")
		writeJSONError(w, http.StatusInternalServerError, "Error marshaling response JSON in fetching specific notification")
		return
	}

	//send the response back to the client
	w.Write(responseJSON)
}

func (h *Handler) deleteNotification(w http.ResponseWriter, r *http.Request) {
	//get id from url path
	id := r.PathValue("id")
	if id == "" {
		utils.SetMessageForLogger(w, "Missing id in URL path")
		writeJSONError(w, http.StatusBadRequest, "Missing id in URL path")
		return
	}
	err := h.store.DeleteNotification(r.Context(), id)
	if err != nil {
		//if not found, return 404 error
		utils.SetMessageForLogger(w, "Notification not found")
		writeJSONError(w, http.StatusNotFound, "Notification not found")
		return
	}

	//if deleted successfully, return a success message
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)

	utils.SetMessageForLogger(w, "Notification with id "+id+" deleted")
}

func validateThreshold(thresholdStruct models.ThresholdNotification) (error, string) {
	var errors []string

	//check if threshold body is present
	if thresholdStruct.Field == "" {
		errors = append(errors, "Missing threshold body in request")
	}
	find := false

	//if threshold body is present, check if the fields are valid
	for _, validField := range utils.VALIDTHRESHOLDS {
		if strings.ToUpper(thresholdStruct.Field) == validField {
			find = true
			break
		}
	}
	if !find {
		errors = append(errors, "Invalid threshold type in request body, valid threshold types are: "+strings.Join(utils.VALIDTHRESHOLDS, ", "))
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation failed"), strings.Join(errors, ", ")
	}
	return nil, ""

}

func sendingWebhook(key string, notification models.RegisterWebhook) error {

	//POST METHOD
	//Containting key,country, event, time
	//If threashold:field, operator,threshold value, and registerd value

	payload := map[string]interface{}{
		"id":      key,
		"country": notification.Country,
		"event":   notification.Event,
		"time":    time.Now().Format("20060102 15:04"),
	}
	if notification.ThresholdNotification != nil {
		payload["threshold"] = map[string]interface{}{
			"field":    notification.ThresholdNotification.Field,
			"operator": notification.ThresholdNotification.Operator,
			"value":    notification.ThresholdNotification.Value,
		}
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		//fmt.Errorf("Error marshaling webhook payload", err)
		return err
	}

	//send the POST request to the webhook URL
	request, err := http.NewRequest(http.MethodPost, notification.Url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		//utils.SetMessageForLogger(w, "Error creating POST request for webhook")
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{}

	resp, err := client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	//if no 200 status code from the webhook url, it was not successful, return an error
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		//utils.SetMessageForLogger(w, fmt.Sprintf("Error sending webhook, received status code %d", resp.StatusCode))
		return fmt.Errorf("Error sending webhook, received status code %d", resp.StatusCode)
	}
	return nil

}

func (h *Handler) CheckWhatNotificationsToSend(ctx context.Context, country string, event string) {
	//This function will check if there are any notifications that should be sent based on the event and country of the notification, and if so, it will call the sendingWebhook function to send the notification to the registered webhook URL
	//This function will be called whenever there is a change in the data, and it will check if there are any notifications that should be sent based on the event and country of the notification, and if so, it will call the sendingWebhook function to send the notification to the registered webhook URL
	var allNotifications []models.AllRegisteredWebhook
	//get all notifications from the database, and check if there are any that should be sent based on the event and country of the notification, and if so, it will call the sendingWebhook function to send the notification to the registered webhook URL
	allNotifications, err := h.store.GetAllNotifications(ctx)
	if err != nil {
		utils.SetMessageForLogger(nil, "Error fetching notifications from database")
		return
	}
	if len(allNotifications) == 0 {
		utils.SetMessageForLogger(nil, "No notifications found in database")
		return
	}

	for _, notification := range allNotifications {
		countryMatch := notification.Country == "" || notification.Country == country
		eventMatch := notification.Event == event

		if countryMatch && eventMatch {
			err := sendingWebhook(notification.Id, notification.RegisterWebhook)
			if err != nil {
				//log.Println("WEBHOOK_SEND_FAIL:", err)
				return
			}
		}
	}
}

func (h *Handler) GetRegWithOnlyIdForNotification(ctx context.Context, id string, event string) {
	//This function will be called right before a registration is deleted

	//first it gets what country this registration is for
	registration, err := h.store.GetRegistration(ctx, id)
	if err != nil {
		//utils.SetMessageForLogger(w, "Error fetching registration from database", err)
		return
	}
	h.CheckWhatNotificationsToSend(ctx, registration.IsoCode, event)
}

//Delete
//CHANGE
//INVOKE
//REGISTER
