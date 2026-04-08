package daemon

import (
	"fmt"
	"net/http"

	"atlasx/internal/permissions"
)

var loadDaemonPermissionsStatus = permissions.LoadStatus

func servePermissions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method %s is not allowed", r.Method))
		return
	}

	writeJSON(w, http.StatusOK, loadDaemonPermissionsStatus())
}
