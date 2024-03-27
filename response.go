package awesomemy

import (
	"encoding/json"
	"net/http"
)

func Render(w http.ResponseWriter, status int, data any) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func RenderNotFound(w http.ResponseWriter) {
	Render(w, http.StatusNotFound, map[string]string{
		"message": "The resource you are looking for could not be found.",
	})
}
