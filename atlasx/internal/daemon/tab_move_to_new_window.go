package daemon

import (
	"fmt"
	"net/http"
	"strings"
)

func serveTabMoveToNewWindow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, r.Method, http.MethodPost)
		return
	}

	var request targetIDRequest
	if err := decodeRequiredJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	request.ID = strings.TrimSpace(request.ID)
	if request.ID == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("id is required"))
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

	result, err := client.MoveToNewWindow(request.ID)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}
