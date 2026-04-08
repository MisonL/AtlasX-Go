package daemon

import (
	"fmt"
	"net/http"
	"strings"
)

func serveTabMoveToWindow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method %s is not allowed", r.Method))
		return
	}

	var request windowMoveRequest
	if err := decodeRequiredJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if strings.TrimSpace(request.ID) == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("id is required"))
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
		writeError(w, http.StatusConflict, err)
		return
	}

	result, err := client.MoveToWindow(request.ID, request.WindowID)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}
