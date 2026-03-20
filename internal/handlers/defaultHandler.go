package handlers

import (
	"assignment-2/internal/utils"
	"net/http"
)

// Defaulthandler answers all queries to non-specified routes
func DefaultHandler(w http.ResponseWriter, r *http.Request) {
	writeJSONError(w, http.StatusNotFound, "endpoint does not exist")
	utils.SetMessageForLogger(w, "Unknown route: "+r.URL.Path)
}
