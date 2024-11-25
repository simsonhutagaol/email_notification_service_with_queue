package utils

import (
	"encoding/json"
	"net/http"
)

func JsonResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
	// json.Marshal(data)
}

//data interface{} artinya bisa menerima tipe data apa saja seperti string, int, struct, map, slice, dan lain-lain.
