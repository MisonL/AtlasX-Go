package daemon

import (
	"fmt"
	"net/http"
)

func serveTabActivateWindow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, r.Method, http.MethodPost)
		return
	}

	var request windowIDRequest
	if err := decodeRequiredJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if request.WindowID <= 0 {
		writeError(w, http.StatusBadRequest, fmt.Errorf("window_id must be positive"))
		return
	}

	paths, err := discoverPaths()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	client, err := newTabsClient(paths)
	if err != nil {
		writeTabsClientUnavailable(w, err)
		return
	}

	result, err := client.ActivateWindow(request.WindowID)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}
