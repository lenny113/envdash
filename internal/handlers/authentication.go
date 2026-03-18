package handlers

import (
	"crypto/md5"   //for generarting hash to create api key
	"encoding/hex" //for converting md5 hash to string
	"encoding/json"
	"fmt"
	"net/http"
	"time" //time of creating api key and for generating unique api key based partly on time hash
)

type Login struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type key struct {
	Key       string `json:"key"`
	CreatedAt string `json:"createdAt"`
}

func RegisterAuth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { //if not a POST request, return method not allowed
		http.Error(w, "Only POST method allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the JSON body of the request into a Login struct
	var login Login
	defer r.Body.Close()

	err := json.NewDecoder(r.Body).Decode(&login)
	if err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// Validate the input (check if name and email are not empty, and if email contains @)
	if login.Email == "" || login.Name == "" {
		http.Error(w, "Missing name or email", http.StatusBadRequest)
		return
	}

	//check if email contains @, if not, return bad request
	isAtValid := false
	for i := 0; i < len(login.Email); i++ {
		if login.Email[i] == '@' {
			isAtValid = true
			break
		}
	}
	if !isAtValid {
		http.Error(w, "Invalid email format, no @ found", http.StatusBadRequest)
		return
	}

	//creating api key
	createAPI, timeCreateApi := createAPIKey(login.Email)

	//formatting the response to the user, which includes the generated API key and the time of API creation
	key := key{
		Key:       createAPI,
		CreatedAt: timeCreateApi,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(key); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
	fmt.Println("API key generated and sent to user:", key.Key)

}

func createAPIKey(email string) (string, string) {

	timeCreateApi := time.Now().Format("20060102 15:04") //this is the format specified in the assignement for time
	fmt.Println("Time of API creation:", timeCreateApi)  //maybe logg this

	startOfUserAPI := "sk-envdash-"

	// Generate hash of email + current time using md5
	//this will be used as the api key for the user, and will be unique for each registration
	//unique even, even same email cant make same key, because of the time component
	hash := md5.Sum([]byte(email + time.Now().String()))
	hashString := hex.EncodeToString(hash[:])
	createAPI := startOfUserAPI + hashString

	// Display the character
	fmt.Println("createAPI:", createAPI)

	return createAPI, timeCreateApi
}
