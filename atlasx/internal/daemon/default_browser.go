package daemon

import (
	"fmt"
	"net/http"

	"atlasx/internal/defaultbrowser"
)

var readDaemonDefaultBrowserStatus = defaultbrowser.ReadStatus

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
