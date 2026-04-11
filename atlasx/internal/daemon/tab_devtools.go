package daemon

import (
	"fmt"
	"net/http"
)

func serveTabDevTools(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, r.Method, http.MethodGet)
		return
	}

	targetID := r.URL.Query().Get("id")
	if targetID == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("missing query parameter id"))
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

	target, err := client.DevTools(targetID)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}

	writeJSON(w, http.StatusOK, target)
}

func serveTabDevToolsPanel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, r.Method, http.MethodGet)
		return
	}

	targetID := r.URL.Query().Get("id")
	if targetID == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("missing query parameter id"))
		return
	}
	panel := r.URL.Query().Get("panel")
	if panel == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("missing query parameter panel"))
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

	target, err := client.DevToolsPanel(targetID, panel)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}

	writeJSON(w, http.StatusOK, target)
}
