package daemon

import (
	"fmt"
	"net/http"
	"strings"
)

func serveTabOpenDevToolsPanelWindowIntoWindow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method %s is not allowed", r.Method))
		return
	}

	var request tabOpenDevToolsPanelWindowIntoWindowRequest
	if err := decodeRequiredJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if request.SourceWindowID <= 0 {
		writeError(w, http.StatusBadRequest, fmt.Errorf("source_window_id must be positive"))
		return
	}
	if strings.TrimSpace(request.Panel) == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("panel is required"))
		return
	}
	if request.TargetWindowID <= 0 {
		writeError(w, http.StatusBadRequest, fmt.Errorf("target_window_id must be positive"))
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

	result, err := client.OpenDevToolsPanelWindowIntoWindow(request.SourceWindowID, request.Panel, request.TargetWindowID)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}
