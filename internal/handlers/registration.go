package handlers

import (
	model "assignment-2/internal/models"
	"assignment-2/internal/utils"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
)

func (h *Handler) RegistrationHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {

	case http.MethodPost:
		h.RegistrationPostHandler(w, r)

	case http.MethodGet:
		h.RegistrationGetHandler(w, r)

	case http.MethodPut:
		h.UpdateRegistration(w, r)

	case http.MethodDelete:
		h.DeleteRegistration(w, r)

	case http.MethodHead:
		h.RegistrationHeadHandler(w, r)

	//case http.MethodPatch:
	//	h.RegistrationPatchHandler(w, r)

	default:
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) RegistrationPostHandler(w http.ResponseWriter, r *http.Request) {

	//Creating an implementation of the registration struct
	//to be used when creating a registration
	var reg model.Registration

	//Decode incoming JSON. If any of the features are not true or false, this will fail
	//and this is why it is not added as a check in the "validateRegistration" function

	if err := json.NewDecoder(r.Body).Decode(&reg); err != nil {
		utils.SetMessageForLogger(w, "Error decoding registration")
		writeJSONError(w, http.StatusBadRequest, "Invalid request body, please refer to the documentation")
		//TO DO: legge til lenke til dokumentasjon
		return
	}

	if err, logmessage := validateRegistration(reg); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body, please refer to the documentation\n"+
			err.Error())
		//TO DO: legge til lenke til dokumentasjon
		utils.SetMessageForLogger(w, logmessage)
		return
	}

	//Adding timestamp as the registration is valid
	reg.LastChange = time.Now().Format("20060102 15:04")

	//Storing registration in firebase
	id, err := h.store.CreateRegistration(r.Context(), reg)
	if err != nil {
		utils.SetMessageForLogger(w, "failed to store registration")
		writeJSONError(w, http.StatusInternalServerError, "failed to save registration")
		return
	}

	// addin the last
	response := map[string]string{
		"id":         id,
		"lastChange": reg.LastChange,
	}

	utils.SetMessageForLogger(w, "registration created "+id)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)

}

/*validateRegistration validates the registration based on these criteria:
- Country or Isocode exists
- If the country or Isocode is valid
- If both country and Isocode is provided, they match
- If the country matches the provided neighbouring currencies
if the registration fails to satisfy any of these criteria, it's invalid
*/

func validateRegistration(reg model.Registration) (error, string) {

	//Check if country name or isocode exists
	if reg.Country == "" && reg.IsoCode == "" {

		return errors.New("Missing required field. A country name or iso code is required."),
			"Missing required field in registration. A country name or iso code is required."
	}

	//Check if country name and isocode matches
	if reg.Country != "" && reg.IsoCode != "" {

	}

	//Check if countryname and/or isocode exists

	return nil, "Valid registration provided"
}

//Valideringide
//===========
//må legge med api-key som kan være "dev-key" for øyeblikket
//POST
//Vi får en registration
//Sjekker om valid
//Hvis valid så legger vi den i databasen

// GET
func (h *Handler) RegistrationGetHandler(w http.ResponseWriter, r *http.Request) {

	id := getIDFromPath(r.URL.Path)

	//No id provided, fetch all registrations
	if id == "" {
		registrations, err := h.store.GetAllRegistrations(r.Context())
		if err != nil {
			utils.SetMessageForLogger(w, err.Error())
		}
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		enc.Encode(registrations)
		return
	} else { //id provided so we try to fetch specified registration from firebase
		registration, err := h.store.GetRegistration(r.Context(), id)
		if err != nil {
			utils.SetMessageForLogger(w, err.Error())
		}

		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		enc.Encode(registration)
	}

}

// PUT
// add check so you cant replace with an identical registration?
func (h *Handler) UpdateRegistration(w http.ResponseWriter, r *http.Request) {
	id := getIDFromPath(r.URL.Path)

	if id == "" {
		writeJSONError(w, http.StatusBadRequest, "Missing id")
		return
	}

	var reg model.Registration

	if err := json.NewDecoder(r.Body).Decode(&reg); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body, please refer to the documentation")
		return
	}

	if err, logmessage := validateRegistration(reg); err != nil {
		utils.SetMessageForLogger(w, logmessage)
		writeJSONError(w, http.StatusBadRequest, "Invalid request body, please refer to the documentation")
		return
	}

	reg.LastChange = time.Now().Format("20060102 15:04")

	err := h.store.UpdateRegistration(r.Context(), id, reg)
	if err != nil {
		utils.SetMessageForLogger(w, err.Error())
		writeJSONError(w, http.StatusNotFound, "Registration not found "+id)
		return
	}
	utils.SetMessageForLogger(w, "registration updated "+id)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(reg)
}

// DELETE
func (h *Handler) DeleteRegistration(w http.ResponseWriter, r *http.Request) {
	id := getIDFromPath(r.URL.Path)
	if id == "" {
		writeJSONError(w, http.StatusBadRequest, "Missing id")
		return
	}
	err := h.store.DeleteRegistration(r.Context(), id)
	if err != nil {
		utils.SetMessageForLogger(w, err.Error())
		writeJSONError(w, http.StatusNotFound, "Registration not found")
		return
	}
	utils.SetMessageForLogger(w, "Registration with id "+id+" deleted")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
	w.Write([]byte("Registration with id " + id + " successfully deleted"))
}

// HEAD (for later)
func (h *Handler) RegistrationHeadHandler(w http.ResponseWriter, r *http.Request) {
	id := getIDFromPath(r.URL.Path)

	//No id provided, fetch all registrations
	if id == "" {
		_, err := h.store.GetAllRegistrations(r.Context())
		if err != nil {
			utils.SetMessageForLogger(w, "Failed to get all registrations")
			writeJSONError(w, http.StatusInternalServerError, "Failed to get all registrations")
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		return

	} else { //id provided so we try to fetch specified registration from firebase
		_, err := h.store.GetRegistration(r.Context(), id)
		if err != nil {
			utils.SetMessageForLogger(w, "Failed to get registration "+id)
			writeJSONError(w, http.StatusNotFound, "Failed to get registration "+id)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
	}
}

//PATCH (for later)
//func (h *Handler) RegistrationPatchHandler(w http.ResponseWriter, r *http.Request) {}

// Extract id from path
func getIDFromPath(path string) string {
	id := strings.TrimPrefix(path, utils.REGISTRATION_PATH)
	return strings.Trim(id, "/")
}
