package daemon

import (
	"fmt"
	"net/http"

	"atlasx/internal/managedruntime"
)

type runtimeStageRequest struct {
	BundlePath string `json:"bundle_path"`
	Version    string `json:"version"`
	Channel    string `json:"channel"`
}

func serveRuntimeStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method %s is not allowed", r.Method))
		return
	}

	paths, err := discoverPaths()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	report, err := managedruntime.Status(paths)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, report)
}

func serveRuntimeStage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method %s is not allowed", r.Method))
		return
	}

	var request runtimeStageRequest
	if err := decodeRequiredJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	paths, err := discoverPaths()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	report, err := managedruntime.StageLocal(paths, managedruntime.StageOptions{
		BundlePath: request.BundlePath,
		Version:    request.Version,
		Channel:    request.Channel,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusOK, report)
}

func serveRuntimeClear(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method %s is not allowed", r.Method))
		return
	}

	paths, err := discoverPaths()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	if err := managedruntime.Clear(paths); err != nil {
		switch {
		case err == managedruntime.ErrStagedRuntimeNotFound:
			writeError(w, http.StatusConflict, err)
		default:
			writeError(w, http.StatusInternalServerError, err)
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"cleared_runtime_root": paths.RuntimeRoot,
	})
}
