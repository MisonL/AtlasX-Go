package daemon

import (
	"fmt"
	"net/http"

	"atlasx/internal/logs"
)

func serveLogsStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method %s is not allowed", r.Method))
		return
	}

	limit, err := parseOptionalLimit(r.URL.Query().Get("limit"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	paths, err := discoverPaths()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	status, err := logs.LoadStatus(paths, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, status)
}
