package handlers

import (
	"assignment-2/internal/models"
	"assignment-2/internal/utils"
	"bytes"   //for sending the payload in the POST request to the webhook URL
	"context" //for handling context in the checkWhatNotificationsToSend function
	"encoding/json"
	"fmt"
	"net/http"
	"net/url" //for validating the URL in the notification registration
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
	defer r.Body.Close()

	var request models.RegisterWebhook

	//checks if json is valid and can be decoded into the struct, if not, it returns an error
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		utils.SetMessageForLogger(w, "Invalid JSON in notification registration")
		writeJSONError(w, http.StatusBadRequest, "Invalid JSON in notification registration")
		return
	}
	request.Country = strings.ToUpper(request.Country) //convert country to uppercase
	request.Event = strings.ToUpper(request.Event)     //convert event to uppercase to make it case insensitive

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

	//adding api key to the request struct, so we can store it under the assosiated user
	api := r.Header.Get("X-Api-Key")
	//adding time created:
	timeCreated := time.Now().Format("20060102 15:04")
	request.Time = timeCreated
	//Stores in Firestore
	notificationId, err := h.store.CreateNotification(r.Context(), request, api)
	if err != nil {
		fmt.Println("Error creating notification in Firestore: ", err)
		utils.SetMessageForLogger(w, "Error creating notification in Firestore")
		writeJSONError(w, http.StatusInternalServerError, "Error creating notification")
		return
	}
	//time created

	//After successful creation of notification, it returns a success message:
	var response models.RegisteredWebhookResponse
	response.Id = notificationId

	responseJSON, err := json.MarshalIndent(response, "", "   ")
	if err != nil {
		utils.SetMessageForLogger(w, "Error marshaling response JSON in notification registration")
		writeJSONError(w, http.StatusInternalServerError, "Error marshaling response JSON in notification registration")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(responseJSON)
}

func validateNotification(request models.RegisterWebhook) (error, string) {
	var errors []string

	if request.Url == "" {
		errors = append(errors, "Missing URL")
	} else {
		_, err := url.ParseRequestURI(request.Url)
		if err != nil {
			errors = append(errors, "Invalid URL")
		}
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
	AllSaved, err := h.store.GetAllNotificationsForUser(r.Context(), r.Header.Get("X-API-Key"))
	if err != nil {
		utils.SetMessageForLogger(w, "Error fetching notifications")
		writeJSONError(w, http.StatusInternalServerError, "Error fetching notifications")
		return
	}
	if AllSaved == nil {
		AllSaved = []models.AllRegisteredWebhook{}
		//Spec says we should return with no message if empty
		//http.StatusOK
		//How do we now send this message?

		utils.SetMessageForLogger(w, "No stored notifications for user")
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

	notification, _, err := h.store.GetSpecificNotification(r.Context(), id)
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

	err := h.store.DeleteNotification(r.Context(), id, r.Header.Get("X-API-Key"))
	if err != nil {
		if err.Error() == "does not exist" {
			utils.SetMessageForLogger(w, "Notification not found")
			writeJSONError(w, http.StatusNotFound, "Notification not found")
			return
		} else if err.Error() == "No access" {
			utils.SetMessageForLogger(w, "No access to this notification")
			writeJSONError(w, http.StatusForbidden, "No privileges to delete this notification")
			return
		} else {
			utils.SetMessageForLogger(w, "Error Firestore, deleting notification: ")
			writeJSONError(w, http.StatusInternalServerError, "Error deleting notification")
			return
		}
	}

	//if deleted successfully, return a success message
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)

	utils.SetMessageForLogger(w, "Notification with id "+id+" deleted")
}

func validateThreshold(thresholdStruct models.ThresholdNotification) (error, string) {
	var errors []string

	// Validate Field
	if thresholdStruct.Field == "" {
		errors = append(errors, "Missing threshold field in request")
	} else {
		found := false
		for _, validField := range utils.VALIDTHRESHOLDS {
			if strings.ToUpper(thresholdStruct.Field) == validField {
				found = true
				break
			}
		}
		if !found {
			errors = append(errors, "Invalid threshold field, valid fields are: "+strings.Join(utils.VALIDTHRESHOLDS, ", "))
		}
	}

	// Validate Operator
	if thresholdStruct.Operator == "" {
		errors = append(errors, "Missing threshold operator, valid operators are: "+strings.Join(utils.VALIDOPERATORS, ", "))
	} else {
		found := false
		for _, op := range utils.VALIDOPERATORS {
			thresholdStruct.Operator = strings.TrimSpace(thresholdStruct.Operator)
			if thresholdStruct.Operator == op {
				found = true
				break
			}
		}
		if !found {
			errors = append(errors, "Invalid threshold operator, valid operators are: "+strings.Join(utils.VALIDOPERATORS, ", "))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation failed"), strings.Join(errors, ", ")
	}
	return nil, ""
}

func sendingLifeCycleWebhook(key string, notification models.RegisterWebhook) error {
	payload := map[string]interface{}{
		"id":      key,
		"country": notification.Country,
		"event":   notification.Event,
		"time":    time.Now().Format("20060102 15:04"),
	}
	return postWebhook(notification.Url, payload)
}

func (h *Handler) CheckLifecycleNotifications(ctx context.Context, country string, event string) {
	//This function will check if there are any notifications that should be sent based on the event and country of the notification, and if so, it will call the sendingLifeCycleWebhook function to send the notification to the registered webhook URL
	//This function will be called whenever there is a change in the data, and it will check if there are any notifications that should be sent based on the event and country of the notification, and if so, it will call the sendingLifeCycleWebhook function to send the notification to the registered webhook URL
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
	//to handle anny formating errors with country codes, we convert to uppercase
	country = strings.ToUpper(country)
	for _, notification := range allNotifications {
		countryMatch := notification.Country == "" || strings.EqualFold(notification.Country, country)
		eventMatch := notification.Event == event

		if countryMatch && eventMatch {
			err := sendingLifeCycleWebhook(notification.Id, notification.RegisterWebhook)
			if err != nil {
				utils.SetMessageForLogger(nil, "Error sending lifecycle webhook for id "+notification.Id)
				continue
			}
		}
	}
}

func (h *Handler) CheckThresholdNotifications(ctx context.Context, country string, measured map[string]float64) {
	allNotifications, err := h.store.GetAllNotifications(ctx)
	if err != nil {
		utils.SetMessageForLogger(nil, "Error fetching notifications from database")
		return
	}

	for _, notification := range allNotifications {
		countryMatch := notification.Country == "" || strings.EqualFold(notification.Country, country)
		eventMatch := notification.Event == "THRESHOLD"
		threshold := notification.ThresholdNotification
		if countryMatch && eventMatch && threshold != nil {
			//check if the measured value meets the threshold condition
			measuredValue, ok := measured[strings.ToUpper(threshold.Field)]
			if !ok {
				utils.SetMessageForLogger(nil, "Measured value for threshold field not found: "+threshold.Field)
				continue
			}
			conditionMet := false
			switch threshold.Operator {
			case ">":
				conditionMet = measuredValue > threshold.Value
			case "<":
				conditionMet = measuredValue < threshold.Value
			case ">=":
				conditionMet = measuredValue >= threshold.Value
			case "<=":
				conditionMet = measuredValue <= threshold.Value
			case "==":
				conditionMet = measuredValue == threshold.Value
			default:
				utils.SetMessageForLogger(nil, "Invalid operator in threshold notification: "+threshold.Operator)
				continue
			}

			details := models.ThresholdDetails{
				Field:          threshold.Field,
				Operator:       threshold.Operator,
				ThresholdValue: threshold.Value,
				MeasuredValue:  measuredValue,
			}
			if conditionMet {
				err := sendThresholdWebhook(notification.Id, notification.Country, notification.Url, details)
				if err != nil {
					utils.SetMessageForLogger(nil, "Error sending threshold webhook for id "+notification.Id+": "+err.Error())
					continue
				}
			}
		}
	}
}

func sendThresholdWebhook(id, country, url string, details models.ThresholdDetails) error {
	payload := map[string]interface{}{
		"id":      id,
		"country": country,
		"event":   "THRESHOLD",
		"time":    time.Now().Format("20060102 15:04"),
		"details": map[string]interface{}{
			"field":         details.Field,
			"operator":      details.Operator,
			"threshold":     details.ThresholdValue,
			"measuredValue": details.MeasuredValue,
		},
	}
	return postWebhook(url, payload)
}

func postWebhook(url string, payload map[string]interface{}) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}
	return nil
}

func (h *Handler) GetRegWithOnlyIdForNotification(ctx context.Context, apiKey string, id string, event string) {
	//This function will be called right before a registration is deleted

	//first it gets what country this registration is for
	//TODO FIX
	registration, err := h.store.GetRegistration(ctx, apiKey, id)
	if err != nil {
		//utils.SetMessageForLogger(w, "Error fetching registration from database", err)
		return
	}
	h.CheckLifecycleNotifications(ctx, registration.IsoCode, event)
}

func resolveIsoCode(isoCode string, country string) string {
	if isoCode != "" {
		return strings.ToUpper(isoCode)
	}
	if country == "" {
		//then we have ISO code
		return ""
	}
	// get the map with all names and iso codes
	cMap, err := getCountryNameAndIsoMap()
	if err != nil {
		return ""
	}
	for iso, name := range cMap {
		if strings.EqualFold(name, country) {
			return iso
		}
	}
	return ""
}
