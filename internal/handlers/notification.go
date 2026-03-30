package handlers

import (
	"fmt"
	"net/http"
)

func (h *Handler) NotificationSpinner(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		fmt.Println("METHOD POST")
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
