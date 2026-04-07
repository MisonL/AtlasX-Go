package daemon

import (
	"fmt"
	"net/http"
)

func serveTabWindowBounds(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method %s is not allowed", r.Method))
		return
	}

	var request windowBoundsRequest
	if err := decodeRequiredJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if request.WindowID <= 0 {
		writeError(w, http.StatusBadRequest, fmt.Errorf("window_id must be positive"))
		return
	}
	if request.Width <= 0 {
		writeError(w, http.StatusBadRequest, fmt.Errorf("width must be positive"))
		return
	}
	if request.Height <= 0 {
		writeError(w, http.StatusBadRequest, fmt.Errorf("height must be positive"))
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

	result, err := client.SetWindowBounds(request.WindowID, request.Left, request.Top, request.Width, request.Height)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}
