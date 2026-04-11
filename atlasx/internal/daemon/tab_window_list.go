package daemon

import "net/http"

func serveTabWindows(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, r.Method, http.MethodGet)
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

	windows, err := client.Windows()
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"returned": len(windows),
		"windows":  windows,
	})
}
