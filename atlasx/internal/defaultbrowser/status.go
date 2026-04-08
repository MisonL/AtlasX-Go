package defaultbrowser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

const launchServicesDomain = "com.apple.LaunchServices/com.apple.launchservices.secure"

var readLaunchServicesJSON = exportLaunchServicesJSON

type Status struct {
	Source        string `json:"source"`
	HTTPBundleID  string `json:"http_bundle_id"`
	HTTPRole      string `json:"http_role"`
	HTTPKnown     bool   `json:"http_known"`
	HTTPSBundleID string `json:"https_bundle_id"`
	HTTPSRole     string `json:"https_role"`
	HTTPSKnown    bool   `json:"https_known"`
	Consistent    bool   `json:"consistent"`
}

type launchServicesHandlers struct {
	LSHandlers []launchServicesHandler `json:"LSHandlers"`
}

type launchServicesHandler struct {
	URLScheme string `json:"LSHandlerURLScheme"`

	RoleAll    string `json:"LSHandlerRoleAll"`
	RoleViewer string `json:"LSHandlerRoleViewer"`
	RoleEditor string `json:"LSHandlerRoleEditor"`
	RoleShell  string `json:"LSHandlerRoleShell"`
}

func ReadStatus() (Status, error) {
	payload, err := readLaunchServicesJSON()
	if err != nil {
		return Status{}, err
	}

	return parseStatus(payload)
}

func parseStatus(payload []byte) (Status, error) {
	var handlers launchServicesHandlers
	if err := json.Unmarshal(payload, &handlers); err != nil {
		return Status{}, fmt.Errorf("decode launchservices handlers: %w", err)
	}

	status := Status{
		Source:        launchServicesDomain,
		HTTPBundleID:  "unknown",
		HTTPRole:      "unknown",
		HTTPSBundleID: "unknown",
		HTTPSRole:     "unknown",
	}
	for _, handler := range handlers.LSHandlers {
		scheme := strings.ToLower(strings.TrimSpace(handler.URLScheme))
		role, bundleID := resolveRole(handler)
		if role == "" || bundleID == "" {
			continue
		}

		switch scheme {
		case "http":
			status.HTTPBundleID = bundleID
			status.HTTPRole = role
			status.HTTPKnown = true
		case "https":
			status.HTTPSBundleID = bundleID
			status.HTTPSRole = role
			status.HTTPSKnown = true
		}
	}

	status.Consistent = status.HTTPKnown && status.HTTPSKnown && status.HTTPBundleID == status.HTTPSBundleID
	return status, nil
}

func resolveRole(handler launchServicesHandler) (string, string) {
	switch {
	case strings.TrimSpace(handler.RoleAll) != "":
		return "all", strings.TrimSpace(handler.RoleAll)
	case strings.TrimSpace(handler.RoleViewer) != "":
		return "viewer", strings.TrimSpace(handler.RoleViewer)
	case strings.TrimSpace(handler.RoleEditor) != "":
		return "editor", strings.TrimSpace(handler.RoleEditor)
	case strings.TrimSpace(handler.RoleShell) != "":
		return "shell", strings.TrimSpace(handler.RoleShell)
	default:
		return "", ""
	}
}

func exportLaunchServicesJSON() ([]byte, error) {
	plistPayload, err := commandOutput("/usr/bin/defaults", "export", launchServicesDomain, "-")
	if err != nil {
		return nil, fmt.Errorf("export launchservices defaults: %w", err)
	}

	cmd := exec.Command("/usr/bin/plutil", "-convert", "json", "-o", "-", "-")
	cmd.Stdin = bytes.NewReader(plistPayload)
	jsonPayload, err := cmd.CombinedOutput()
	if err != nil {
		return nil, commandError("convert launchservices plist to json", err, jsonPayload)
	}
	return jsonPayload, nil
}

func commandOutput(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, commandError(strings.Join(append([]string{name}, args...), " "), err, output)
	}
	return output, nil
}

func commandError(action string, err error, output []byte) error {
	trimmed := strings.TrimSpace(string(output))
	if trimmed == "" {
		return fmt.Errorf("%s: %w", action, err)
	}
	return fmt.Errorf("%s: %w: %s", action, err, trimmed)
}

func (s Status) Render() string {
	return strings.Join([]string{
		fmt.Sprintf("source=%s", s.Source),
		fmt.Sprintf("http_bundle_id=%s", s.HTTPBundleID),
		fmt.Sprintf("http_role=%s", s.HTTPRole),
		fmt.Sprintf("http_known=%t", s.HTTPKnown),
		fmt.Sprintf("https_bundle_id=%s", s.HTTPSBundleID),
		fmt.Sprintf("https_role=%s", s.HTTPSRole),
		fmt.Sprintf("https_known=%t", s.HTTPSKnown),
		fmt.Sprintf("consistent=%t", s.Consistent),
	}, "\n") + "\n"
}
