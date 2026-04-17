package handlers

import (
	model "assignment-2/internal/models"
	"assignment-2/internal/utils"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// This function directs a request to the corresponding handler based on REST method
func (h *Handler) RegistrationHandler(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodOptions:
		h.RegistrationOptionsHandler(w, r)

	case http.MethodPost:
		h.RegistrationPostHandler(w, r)

	case http.MethodGet:
		h.RegistrationGetHandler(w, r)

	case http.MethodPut:
		h.RegistrationPutHandler(w, r)

	case http.MethodDelete:
		h.RegistrationDeleteHandler(w, r)

	case http.MethodHead:
		h.RegistrationHeadHandler(w, r)

	case http.MethodPatch:
		h.RegistrationPatchHandler(w, r)
	default:
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// RegistrationPostHandler handles creation of a new registration. Flow:
// 1. authenticate API-key
// 2. Decode and normalize input
// 3. Validate
// 4. Persist to firestore
// 5. Return registration ID and timestamp to the user
func (h *Handler) RegistrationPostHandler(w http.ResponseWriter, r *http.Request) {

	//Creating an implementation of the registration struct
	//to be used when creating a registration
	var reg model.Registration

	//getting hashed apiKey from request header
	apiKey := GetAndHashAPIKey(r)

	//Decode incoming JSON. If any of the features are not true or false, this will fail
	//and this is why it is not added as a check in the "validateRegistration" function

	if err := json.NewDecoder(r.Body).Decode(&reg); err != nil {
		utils.SetMessageForLogger(w, "Error decoding registration")
		writeJSONError(w, http.StatusBadRequest, "Invalid request body, features must be either true or false")
		return
	}

	//Formatting countryname and/or isocode so that we can get either uppercase or lowercase letters
	//or a mix as input and still make a valid registration
	reg.Country, reg.IsoCode = formatCountryNameAndIso(reg.Country, reg.IsoCode)

	//Features are either true or false so we format the countryname and/or isocode and check them
	//with the "validateRegistration function". Currencodes will be checked if countryname and isocode is valid

	if err, logmessage := validateRegistration(&reg); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body: "+
			err.Error())
		utils.SetMessageForLogger(w, logmessage)
		return
	}

	//Adding timestamp as the registration is valid
	reg.LastChange = time.Now().Format("20060102 15:04")

	//Storing registration in firebase
	id, err := h.store.CreateRegistration(r.Context(), apiKey, reg)
	if err != nil {
		utils.SetMessageForLogger(w, "failed to store registration")
		writeJSONError(w, http.StatusInternalServerError, "failed to save registration")
		return
	}

	// Creating the response for the user which includes the id of the registration
	// and the timestamp of the creation
	response := map[string]string{
		"id":         id,
		"lastChange": reg.LastChange,
	}

	//logging the creation internally
	utils.SetMessageForLogger(w, "registration created "+id)

	w.Header().Set("Content-Type", "application/json") //Setting content type
	w.WriteHeader(http.StatusCreated)                  // Set HTTP status code to 201 CREATED

	//Writing response body for the user
	json.NewEncoder(w).Encode(response)
	h.CheckLifecycleNotifications(r.Context(), resolveIsoCode(reg.IsoCode, reg.Country), "REGISTER")

}

// validateRegistration performs FULL validation of a registration object.
// It assumes the object is complete (not partial like PATCH).
// This function enforces:
// - Required fields (country or isoCode)
// - Country <-> ISO consistency
// - Valid country/ISO existence
// - Valid currency codes
//
// returns: error for the user and a string for the logger
//
// NOTE: Input is assumed to be normalized (ISO uppercase, country formatted) per the formatCountryNameAndIso function.
// It is also assumed that the external API's providing the isocodes + countrynames and currencycodes keep the
// same query requirements.
func validateRegistration(reg *model.Registration) (error, string) {

	//Check if country name or isocode exists
	if reg.Country == "" && reg.IsoCode == "" {
		return errors.New("Missing required field. A country name and/or iso code is required."),
			"Missing required field in registration. A country name and/or iso code is required."
	}

	//Country and/or isocode exists so we create a map to use for validation
	cNameAndIsoMap, err := getCountryNameAndIsoMap()
	if err != nil {
		return err, "failed to get country and iso map"
	}

	//Countryname and isocode is provided so we validate them
	if reg.Country != "" && reg.IsoCode != "" {
		//Checking if countryname has a valid length
		if err, logMessage := checkCountryNameLength(reg.Country); err != nil {
			return err, logMessage
		}
		//Checking if ISOcode has a valid length
		if err, logMessage := checkIsoCodeLength(reg.IsoCode); err != nil {
			return err, logMessage
		}
		//Validating isocode
		name, ok := cNameAndIsoMap[reg.IsoCode]
		if !ok {
			return errors.New("invalid country iso code"), "invalid iso code provided"
		}
		//Validating countryname
		if !strings.EqualFold(name, reg.Country) {
			return errors.New("country name and iso code do not match"), "Country name and iso code do not match"
		}

	}

	//Validating only isocode as country is not provided
	if reg.Country == "" && reg.IsoCode != "" {
		if _, ok := cNameAndIsoMap[strings.ToUpper(reg.IsoCode)]; !ok {
			return errors.New("invalid country iso code"), "invalid iso code provided"
		}

	}

	//Validating only countryname as isocode is not provided
	if reg.IsoCode == "" && reg.Country != "" {
		valid := false
		for _, name := range cNameAndIsoMap {
			if strings.EqualFold(name, reg.Country) {
				valid = true
				break
			}
		}
		if !valid {
			return errors.New("invalid country name provided"), "invalid country name provided"
		}

	}

	//Validating currencies
	if err, logmessage := validateCurrencies(reg.Features.TargetCurrencies); err != nil {
		return err, logmessage
	}

	//Currencies are valid if we get here so we make them uppercase
	for i := range reg.Features.TargetCurrencies {
		reg.Features.TargetCurrencies[i] = strings.ToUpper(reg.Features.TargetCurrencies[i])
	}

	//No checks returned an error so the registration is valid
	return nil, "Valid registration provided"
}

// RegistrationGetHandler fetches an existing registration or registrations as long as the user has
// the api-key corresponding to the registration/registrations.

func (h *Handler) RegistrationGetHandler(w http.ResponseWriter, r *http.Request) {

	//Getting registration id from path
	id := getIDFromRegPath(r.URL.Path)
	//getting apikey from request header and hashing it
	apiKey := GetAndHashAPIKey(r)

	//No id provided, fetch all registrations
	if id == "" {
		registrations, err := h.store.GetAllRegistrations(r.Context(), apiKey)
		if err != nil {
			utils.SetMessageForLogger(w, err.Error())
		}
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		enc.Encode(registrations)
		return
	} else { //id provided so we try to fetch specified registration from firebase
		registration, err := h.store.GetRegistration(r.Context(), apiKey, id)
		if err != nil {
			utils.SetMessageForLogger(w, err.Error())
		}

		//to send notification that a reg of this country is fetched

		enc := json.NewEncoder(w)

		w.Header().Set("Content-Type", "application/json") // Set the content type of response
		w.WriteHeader(http.StatusOK)                       // Set HTTP status code to 200 OK

		//formatting
		enc.SetIndent("", "  ")
		//Encoding registration for user
		enc.Encode(registration)
	}

}

// RegistrationPutHandler handles the replacement of a registration with a new one. Flow:
// 1. authenticate API-key
// 2. Decode and normalize input
// 3. Validate
// 4. Persist to firestore
// 5. Return new registration to the user

func (h *Handler) RegistrationPutHandler(w http.ResponseWriter, r *http.Request) {

	//Getting registration id from path
	id := getIDFromRegPath(r.URL.Path)
	//getting apikey from request header and hashing it
	apiKey := GetAndHashAPIKey(r)

	if id == "" {
		//Writing JSON error to user with status 400 BAD REQUEST
		writeJSONError(w, http.StatusBadRequest, "Missing id")
		return
	}

	var reg model.Registration

	//Decoding registration from user
	if err := json.NewDecoder(r.Body).Decode(&reg); err != nil {
		//Writing JSON error to user with status 400 BAD REQUEST
		writeJSONError(w, http.StatusBadRequest, "Invalid request body, please refer to the documentation")
		return
	}

	//Formatting countryname and isocode so it doese'nt fail the validation because of wrong casing
	reg.Country, reg.IsoCode = formatCountryNameAndIso(reg.Country, reg.IsoCode)

	//Validating the new registration
	if err, logmessage := validateRegistration(&reg); err != nil {
		utils.SetMessageForLogger(w, logmessage)
		writeJSONError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	//updating timestamp
	reg.LastChange = time.Now().Format("20060102 15:04")

	//Replacing the old registration with the new one provided by the user
	err := h.store.UpdateRegistration(r.Context(), apiKey, id, reg)
	if err != nil {
		utils.SetMessageForLogger(w, err.Error())
		//Writing JSON error to user with status 404 NOT FOUND
		writeJSONError(w, http.StatusNotFound, "Registration not found "+id)
		return
	}
	//Logging successful registration update
	utils.SetMessageForLogger(w, "registration updated "+id)

	w.Header().Set("Content-Type", "application/json") // Set the content type of response
	w.WriteHeader(http.StatusOK)                       // Set HTTP status code to 200 OK

	//Encoding the new registration for the user
	json.NewEncoder(w).Encode(reg)
	//Sending lifecycle notification for update of registration, we have to know the country of the registration to send the correct notifications
	h.CheckLifecycleNotifications(r.Context(), reg.IsoCode, "CHANGE")
}

// RegistrationDeleteHandler handles the deletion of a registration with a given id
func (h *Handler) RegistrationDeleteHandler(w http.ResponseWriter, r *http.Request) {
	//Getting id from url path
	id := getIDFromRegPath(r.URL.Path)
	if id == "" {
		//Writing JSON error to user with status 400 BAD REQUEST
		writeJSONError(w, http.StatusBadRequest, "Missing id")
		return
	}

	//fetching apikey from request header and hashing it
	apiKey := GetAndHashAPIKey(r)

	//For notification purposes we need to know the registration before deleting
	reg, err := h.store.GetRegistration(r.Context(), apiKey, id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "Registration not found")
		return
	}
	RegistrationIso := reg.IsoCode
	RegistrationCountry := reg.Country
	//Deleting specified registration from firestore if found under provided apikey
	err = h.store.DeleteRegistration(r.Context(), apiKey, id)
	if err != nil {
		utils.SetMessageForLogger(w, err.Error())
		//Writing JSON error to user with status 404 NOT FOUND
		writeJSONError(w, http.StatusNotFound, "Registration not found")
		return
	}
	//Logging successful deletion to logfile
	utils.SetMessageForLogger(w, "Registration with id "+id+" deleted")

	w.Header().Set("Content-Type", "application/json")                      //Setting content type
	w.WriteHeader(http.StatusNoContent)                                     // Set HTTP status code to 204 NO CONTENT
	w.Write([]byte("Registration with id " + id + " successfully deleted")) //Writing result to user

	//Sending lifecycle notification for deletion of registration, we have to know the country of the registration to send the correct notifications
	h.CheckLifecycleNotifications(r.Context(), resolveIsoCode(RegistrationIso, RegistrationCountry), "DELETE")
}

// RegistrationHeadHandler handles gets the head of the response when querying for one or all registrations
// belonging to a user

func (h *Handler) RegistrationHeadHandler(w http.ResponseWriter, r *http.Request) {
	//getting registration id from URL path
	id := getIDFromRegPath(r.URL.Path)
	//Getting apikey from request header and hashing it, then store it in the apiKey variable
	apiKey := GetAndHashAPIKey(r)

	//No id provided, fetch all registrations
	if id == "" {
		_, err := h.store.GetAllRegistrations(r.Context(), apiKey)
		if err != nil {
			utils.SetMessageForLogger(w, "Failed to get all registrations")
			writeJSONError(w, http.StatusInternalServerError, "Failed to get all registrations")
			return
		}
		w.Header().Set("Content-Type", "application/json") //Set content type
		w.WriteHeader(http.StatusOK)                       // Set HTTP status code to 200 OK
		return

	} else { //id provided so we try to fetch specified registration from firebase
		_, err := h.store.GetRegistration(r.Context(), apiKey, id)
		if err != nil {
			utils.SetMessageForLogger(w, "Failed to get registration "+id)
			//Writing JSON error if registration not found with HTTP status code "404 NOT FOUND"
			writeJSONError(w, http.StatusNotFound, "Failed to get registration "+id)
			return
		}
		w.Header().Set("Content-Type", "application/json") //setting content type
		w.WriteHeader(http.StatusOK)                       // Set HTTP status code to 200 OK

	}
}

// RegistrationOptionsHandler provides the user with allowed REST methods for the registrations endpoint
func (h *Handler) RegistrationOptionsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")                //set content type
	w.Header().Set("Allow", "GET, POST, PUT, PATCH, OPTIONS, DELETE") //Set allow to all the available
	//REST methods
	w.WriteHeader(http.StatusOK) //set HTTP status code to 200 OK
}

// RegistrationPatchHandler handles changes to an existing registration. Flow:
// 1. authenticate API-key
// 2. Decode and normalize input
// 3. Validate input (changes to registration)
// 4. Retrieve existing registration from firestore
// 5. Decode and normalize existing registration
// 6. Combine existing registration with provided changes
// 7. validate the new registration (registration combined with changes)
// 8. Persist to firestore

func (h *Handler) RegistrationPatchHandler(w http.ResponseWriter, r *http.Request) {

	//Get registration id from url path
	id := getIDFromRegPath(r.URL.Path)
	//Getting API-key from request header, hashing it and storing it in the apiKey variable
	apiKey := GetAndHashAPIKey(r)

	if id == "" {
		//Writing JSON error if id is empty with HTTP status code "400 BAD REQUEST"
		writeJSONError(w, http.StatusBadRequest, "Missing id")
		return
	}

	var patch model.RegistrationPatch

	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		//Writing JSON error if wrong datatype provided in patch data "400 BAD REQUEST"
		writeJSONError(w, http.StatusBadRequest, "Invalid body")
		return
	}

	//Validating patch data
	if err, log := validatePatch(&patch); err != nil {
		utils.SetMessageForLogger(w, log)
		//Writing JSON error if patch invalid with http status code "400 BAD REQUEST"
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	//Now that the patch data is valid we check if the update from the patch
	//is valid combined with the original registration

	//fetch the original registration from firestore
	originalReg, err := h.store.GetRegistration(r.Context(), apiKey, id)
	if err != nil {
		utils.SetMessageForLogger(w, err.Error())
	}

	//combine original registration with updates
	updated := applyPatch(*originalReg, patch)

	//validating the new registration
	if err, log := validateRegistration(&updated); err != nil {
		utils.SetMessageForLogger(w, log)
		//Writing JSON error if registration is invalid with http status code "400 BAD REQUEST"
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	//Change the registration in firestore
	err = h.store.TweakRegistration(r.Context(), apiKey, id, patch)
	if err != nil {
		//Writing JSON error if update fails with http status code "500 INTERNAL SERVER ERROR"
		writeJSONError(w, http.StatusInternalServerError, "Failed to update registration "+id)
		return
	}

	//logging successful patch to logfile
	utils.SetMessageForLogger(w, "patched "+id)

	w.WriteHeader(http.StatusNoContent) //set HTTP status code to 204 "NO CONTENT"
	h.CheckLifecycleNotifications(r.Context(), resolveIsoCode(updated.IsoCode, updated.Country), "CHANGE")
}

// Extract id from path
func getIDFromRegPath(path string) string {
	//Trimming prefix so that we are left with the id provided by the user
	id := strings.TrimPrefix(path, utils.REGISTRATION_PATH)
	//Removing eventual remaining frontslashes  and returning registration id
	return strings.Trim(id, "/")
}

// checkIsoCodeLength checks if the isocode provided is the appropriate length and returns an
// error and a string for the logger
func checkIsoCodeLength(isoCode string) (error, string) {
	if len(isoCode) < utils.ISOCODE_LENGTH { //isocode too short
		return errors.New("iso code is too short"), "User provided iso code is too short"
	} else if len(isoCode) > utils.ISOCODE_LENGTH { //isocode too long
		return errors.New("iso code is too long"), "User provided iso code is too long"
	}
	// isocode is 2 characters long
	return nil, "User provided a correct iso code"

}

// checkCountryNameLength checks if the country name provided is the appropriate length and returns an
// error and a string for the logger
func checkCountryNameLength(countryName string) (error, string) {
	if len(countryName) > utils.LONGEST_COUNTRYNAME { //Countryname too long
		return errors.New("country name is too long"), "User provided country name is too long"
	} else if len(countryName) < utils.SHORTEST_COUNTRYNAME { //Countryname too short
		return errors.New("country name is too short"), "User provided country name is too short"
	}
	//Countryname has valid length
	return nil, "User provided a correct country name"
}

// getNamesMakeMap gets the name and isocode of all countries from the countriesapi
// and puts them in a map that will be used to  validate if country names and
// iso codes correspond in the registrations
func getNamesMakeMap() (map[string]string, error) {
	//fetching countrynames and isocodes from external api
	resp, err := http.Get(utils.COUNTRY_AND_ISO_URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var countries []model.CountryNameAndISO

	//decoding response from external api
	if err := json.NewDecoder(resp.Body).Decode(&countries); err != nil {
		return nil, err
	}
	countryMap := make(map[string]string)

	//creating map from response
	for _, c := range countries {
		countryMap[c.CCA2] = c.Name.Common
	}
	//returning created map
	return countryMap, nil
}

// Map with countryname and isocode to ber used for validating registrations
var cNameAndIsoMap map[string]string

// bool to check if the map is loaded into memory
var cMapLoaded bool

// getCountryNameAndIsoMap returns a cached map of ISO codes to country names.
//
// On first call, it fetches the data from an external API (via getNamesMakeMap),
// stores it in memory, and marks it as loaded.
// Subsequent calls return the cached map to avoid repeated network requests.
//
// This improves performance and reduces dependency on external services.
//
// Returns:
// - map[string]string: ISO code -> country name mapping
// - error: if fetching or building the map fails
//
// NOTE: This implementation is not concurrency-safe.

func getCountryNameAndIsoMap() (map[string]string, error) {

	if cMapLoaded {
		return cNameAndIsoMap, nil
	}

	m, err := getNamesMakeMap()
	if err != nil {
		return nil, err
	}
	cNameAndIsoMap = m
	cMapLoaded = true
	return cNameAndIsoMap, nil
}

// This function has the purpouse of making the countryname and the isocode the correct format
func formatCountryNameAndIso(countryName string, isoCode string) (string, string) {

	//If countryname is provided we change it to the correct format
	if countryName != "" {
		//Make the whole string lowercase
		countryName = strings.ToLower(countryName)
		//Make the first letter uppercase and add the rest of the string to create the final countryname version
		countryName = strings.ToUpper(countryName[0:1]) + countryName[1:]
	}

	//If Isocode is provided we make it uppercase
	if isoCode != "" {
		isoCode = strings.ToUpper(isoCode)
	}

	//Return the correct format
	return countryName, isoCode
}

// FetchCurrencyMap gets currencies from an external api and puts it in a map and returns the map and an error
func FetchCurrencyMap() (map[string]struct{}, error) {
	//fetching currency codes from external api
	resp, err := http.Get(utils.CURRENCY_URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data model.CurrencyAPIResponse

	//Decode JSON response into struct
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	//Create a set  of currency codes for O(1) validation lookups
	currencySet := make(map[string]struct{})

	//Extract currency codes from response
	for code := range data.Rates {
		currencySet[code] = struct{}{}
	}

	return currencySet, nil
}

// currencyMap holds the cached set of currency codes
var currencyMap map[string]struct{}

// currencyLoaded indicates whether the currency map has been initialized
var currencyLoaded bool

func GetCurrencyMap() (map[string]struct{}, error) {
	if currencyLoaded {
		return currencyMap, nil
	}

	m, err := FetchCurrencyMap()
	if err != nil {
		return nil, err
	}

	currencyMap = m
	currencyLoaded = true

	return currencyMap, nil
}

// GetCurrencyMap returns a cached map of currency codes.
// If the map has not been loaded yet, it fetches the data once
// and stores it in memory for future use.
//
// This avoids repeated external API calls and improves performance.
//
// NOTE: This implementation is not concurrency-safe.
func validateCurrencies(codes []string) (error, string) {
	//If too many currencies are provided we return an error
	if len(codes) > utils.TARGETCURRENCIES_MAX_LENGTH {
		return errors.New("too many currencies provided"), "too many currencies provided"
	}

	//trying to load the currencymap
	currencyMap, err := GetCurrencyMap()
	if err != nil {
		return errors.New("failed to load currency data"), "failed to load currency data"
	}

	//Iterating through the currencies and if a currencycode has the wrong length, and approptiate error will
	//be returned and if not, the currencycode is looked for in the currencymap to check if it exists
	for _, code := range codes {
		if len(code) != utils.CURRENCYCODE_LENGTH {
			if len(code) < utils.CURRENCYCODE_LENGTH {
				return fmt.Errorf("Currencycode '%s' is too short, currencycode has to be 3 letters", code),
					fmt.Sprintf("Currency code '%s' is too short", code)
			}
			if len(code) > utils.CURRENCYCODE_LENGTH {
				return fmt.Errorf("Currencycode '%s' is too long, currencycode has to be 3 letters", code),
					fmt.Sprintf("Currency code '%s' is too long", code)
			}

		}
		code = strings.ToUpper(code)
		if _, ok := currencyMap[code]; !ok {
			return fmt.Errorf("invalid currency: %s", code), "invalid currency code provided"
		}
	}

	return nil, "valid currencies provided"
}

// ValidatePatch makes sure the patch data is valid (when not combined with unpatched data
// "validateRegistration" will be used to validate the "New" registration)
func validatePatch(patch *model.RegistrationPatch) (error, string) {

	// Validate country / iso if provided
	if patch.Country != nil || patch.IsoCode != nil {

		cNameAndIsoMap, err := getCountryNameAndIsoMap()
		if err != nil {
			return err, "failed to load country map"
		}

		// Format values before validation
		country := ""
		iso := ""

		if patch.Country != nil {
			country = *patch.Country
		}
		if patch.IsoCode != nil {
			iso = *patch.IsoCode
		}

		country, iso = formatCountryNameAndIso(country, iso)

		if patch.Country != nil {
			*patch.Country = country
		}

		if patch.IsoCode != nil {
			*patch.IsoCode = iso
		}

		// If both isocode and country provided must match
		if country != "" && iso != "" {
			name, ok := cNameAndIsoMap[iso]
			if !ok {
				return errors.New("invalid iso code"), "invalid iso code"
			}
			if !strings.EqualFold(name, country) {
				return errors.New("country and iso do not match"), "country iso mismatch"
			}
		}

		// Only iso
		if iso != "" && country == "" {
			if _, ok := cNameAndIsoMap[iso]; !ok {
				return errors.New("invalid iso code"), "invalid iso code"
			}
		}

		// Only country
		if country != "" && iso == "" {
			valid := false
			for _, name := range cNameAndIsoMap {
				if strings.EqualFold(name, country) {
					valid = true
					break
				}
			}
			if !valid {
				return errors.New("invalid country"), "invalid country"
			}
		}
	}

	// Validate features
	if patch.Features != nil {

		// Only validate currencies if provided
		if patch.Features.TargetCurrencies != nil {
			if err, log := validateCurrencies(*patch.Features.TargetCurrencies); err != nil {
				return err, log
			}
		}
	}

	return nil, "valid patch"
}

// applyPatch combines a registration with provided patch data and returns the new registration
func applyPatch(existing model.Registration, patch model.RegistrationPatch) model.Registration {

	updated := existing

	// Top-level fields
	if patch.Country != nil {
		updated.Country = *patch.Country
	}

	if patch.IsoCode != nil {
		updated.IsoCode = *patch.IsoCode
	}

	// Nested features
	if patch.Features != nil {

		if patch.Features.Temperature != nil {
			updated.Features.Temperature = *patch.Features.Temperature
		}

		if patch.Features.Precipitation != nil {
			updated.Features.Precipitation = *patch.Features.Precipitation
		}

		if patch.Features.AirQuality != nil {
			updated.Features.AirQuality = *patch.Features.AirQuality
		}

		if patch.Features.Capital != nil {
			updated.Features.Capital = *patch.Features.Capital
		}

		if patch.Features.Coordinates != nil {
			updated.Features.Coordinates = *patch.Features.Coordinates
		}

		if patch.Features.Population != nil {
			updated.Features.Population = *patch.Features.Population
		}

		if patch.Features.Area != nil {
			updated.Features.Area = *patch.Features.Area
		}

		if patch.Features.TargetCurrencies != nil {
			updated.Features.TargetCurrencies = *patch.Features.TargetCurrencies
		}
	}

	return updated
}
