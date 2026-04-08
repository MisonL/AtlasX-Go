package daemon

import (
	"fmt"
	"net/http"
	"strings"

	"atlasx/internal/defaultbrowser"
)

var readDaemonDefaultBrowserStatus = defaultbrowser.ReadStatus
var setDaemonDefaultBrowserBundleID = defaultbrowser.SetBundleID

func serveDefaultBrowserStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method %s is not allowed", r.Method))
		return
	}

	status, err := readDaemonDefaultBrowserStatus()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, status)
}

func serveDefaultBrowserSet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method %s is not allowed", r.Method))
		return
	}

	var request struct {
		BundleID string `json:"bundle_id"`
	}
	if err := decodeRequiredJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if strings.TrimSpace(request.BundleID) == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("bundle_id is required"))
		return
	}

	status, err := setDaemonDefaultBrowserBundleID(request.BundleID)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, status)
}
