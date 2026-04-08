package daemon

import (
	"fmt"
	"net/http"

	"atlasx/internal/tabgroups"
)

func serveTabOrganize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method %s is not allowed", r.Method))
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

	targets, err := client.List()
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}

	groups := tabgroups.Suggest(targets)
	writeJSON(w, http.StatusOK, map[string]any{
		"returned": len(groups),
		"groups":   groups,
	})
}

func serveTabOrganizeGroupToWindow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method %s is not allowed", r.Method))
		return
	}

	var request struct {
		GroupID string `json:"group_id"`
	}
	if err := decodeRequiredJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if request.GroupID == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("group_id is required"))
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

	result, err := tabgroups.ApplyToNewWindow(client, request.GroupID)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func serveTabOrganizeToWindows(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method %s is not allowed", r.Method))
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

	result, err := tabgroups.ApplyAllToNewWindows(client)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func serveTabOrganizeGroupIntoWindow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method %s is not allowed", r.Method))
		return
	}

	var request struct {
		GroupID  string `json:"group_id"`
		WindowID int    `json:"window_id"`
	}
	if err := decodeRequiredJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if request.GroupID == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("group_id is required"))
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

	result, err := tabgroups.ApplyGroupToWindow(client, request.GroupID, request.WindowID)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}
