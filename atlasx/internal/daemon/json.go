package daemon

import (
	"encoding/json"
	"errors"
	"net/http"
)

func (s Status) Render() string {
	payload, _ := json.MarshalIndent(s, "", "  ")
	return string(payload) + "\n"
}

func serveStatus(w http.ResponseWriter, _ *http.Request) {
	status, err := Bootstrap()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, status)
}

func writeJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, code int, err error) {
	writeJSON(w, code, map[string]string{
		"error": err.Error(),
	})
}

func decodeRequiredJSON(r *http.Request, target any) error {
	if r.Body == nil {
		return errors.New("request body is required")
	}
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}

func decodeOptionalJSON(r *http.Request, target any) error {
	if r.Body == nil || r.ContentLength == 0 {
		return nil
	}
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}
