package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set(HeaderContentType, ContentTypeJSON)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, destination any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 2<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decodeSingleJSON(decoder, destination)
}

func decodeSingleJSON(decoder *json.Decoder, destination any) error {
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return errMultipleJSONValues
		}
		return err
	}
	return nil
}
