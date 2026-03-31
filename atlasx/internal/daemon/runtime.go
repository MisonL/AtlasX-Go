package daemon

import (
	"errors"
	"fmt"
	"net/http"

	"atlasx/internal/managedruntime"
)

type runtimeStageRequest struct {
	BundlePath string `json:"bundle_path"`
	Version    string `json:"version"`
	Channel    string `json:"channel"`
}

type runtimePlanCreateRequest struct {
	Version     string `json:"version"`
	Channel     string `json:"channel"`
	URL         string `json:"url"`
	SHA256      string `json:"sha256"`
	ArchivePath string `json:"archive_path"`
	BundlePath  string `json:"bundle_path"`
}

type runtimeVerifyResponse struct {
	managedruntime.VerifyReport
	Error string `json:"error,omitempty"`
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

func serveRuntimeVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method %s is not allowed", r.Method))
		return
	}

	paths, err := discoverPaths()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	report, err := managedruntime.Verify(paths)
	response := runtimeVerifyResponse{VerifyReport: report}
	if err != nil {
		response.Error = err.Error()
		writeJSON(w, runtimeVerifyStatusCode(err), response)
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func serveRuntimePlan(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		serveRuntimePlanStatus(w)
	case http.MethodPost:
		serveRuntimePlanCreate(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method %s is not allowed", r.Method))
	}
}

func serveRuntimePlanStatus(w http.ResponseWriter) {
	paths, err := discoverPaths()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	status, err := managedruntime.InstallPlanInfo(paths)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, status)
}

func serveRuntimePlanCreate(w http.ResponseWriter, r *http.Request) {
	var request runtimePlanCreateRequest
	if err := decodeRequiredJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	paths, err := discoverPaths()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	plan, err := managedruntime.NewInstallPlan(managedruntime.InstallPlanOptions{
		Version:          request.Version,
		Channel:          request.Channel,
		SourceURL:        request.URL,
		ExpectedSHA256:   request.SHA256,
		ArchivePath:      request.ArchivePath,
		StagedBundlePath: request.BundlePath,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := managedruntime.SaveInstallPlan(paths, plan); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	status, err := managedruntime.InstallPlanInfo(paths)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, status)
}

func serveRuntimePlanClear(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method %s is not allowed", r.Method))
		return
	}

	paths, err := discoverPaths()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	if err := managedruntime.ClearInstallPlan(paths); err != nil {
		switch {
		case errors.Is(err, managedruntime.ErrInstallPlanNotFound):
			writeError(w, http.StatusConflict, err)
		default:
			writeError(w, http.StatusInternalServerError, err)
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"cleared_install_plan": paths.RuntimeInstallPlanFile,
	})
}

func runtimeVerifyStatusCode(err error) int {
	switch {
	case errors.Is(err, managedruntime.ErrRuntimeManifestNotFound),
		errors.Is(err, managedruntime.ErrRuntimeManifestSHA256Missing),
		errors.Is(err, managedruntime.ErrRuntimeBundleNotFound),
		errors.Is(err, managedruntime.ErrRuntimeBinaryNotFound),
		errors.Is(err, managedruntime.ErrRuntimeBinaryNotExecutable),
		errors.Is(err, managedruntime.ErrRuntimeSHA256Mismatch):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
