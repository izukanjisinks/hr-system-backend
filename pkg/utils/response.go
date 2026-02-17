package utils

import (
	"encoding/json"
	"net/http"
)

func RespondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func RespondError(w http.ResponseWriter, status int, message string) {
	RespondJSON(w, status, map[string]interface{}{
		"error": map[string]string{
			"message": message,
		},
	})
}

func DecodeJson(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}
