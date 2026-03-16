package handlers

import (
	"assignment-2/utils"
	"net/http"
)

func DefaultHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	_, err := w.Write([]byte("404 page not found"))
	if err != nil {
		utils.SetMessageForLogger(w, "Error writing response from default handler")
	}
	utils.SetMessageForLogger(w, "Successfull response from default handler")
}
