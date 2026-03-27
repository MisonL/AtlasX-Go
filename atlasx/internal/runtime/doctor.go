package runtime

import (
	"fmt"
	goruntime "runtime"

	"atlasx/internal/diagnostics"
	"atlasx/internal/platform/chrome"
)

type Report struct {
	GOOS         string
	GOARCH       string
	ChromeFound  bool
	ChromeApp    string
	DefaultData  string
	PrimaryURL   string
	SharedModeOK bool
}

func Detect() Report {
	doctorReport, err := diagnostics.Generate()
	report := Report{
		GOOS:         goruntime.GOOS,
		GOARCH:       goruntime.GOARCH,
		DefaultData:  chrome.DefaultUserDataDir(),
		PrimaryURL:   "https://chatgpt.com/atlas?get-started",
		SharedModeOK: true,
	}
	if err == nil && doctorReport.Chrome.BinaryPath != "" {
		report.ChromeFound = true
		report.ChromeApp = doctorReport.Chrome.BinaryPath
	}
	return report
}

func (r Report) String() string {
	return fmt.Sprintf(
		"goos=%s\ngoarch=%s\nchrome_found=%t\nchrome_app=%s\ndefault_user_data_dir=%s\nprimary_url=%s\nshared_mode_ok=%t",
		r.GOOS,
		r.GOARCH,
		r.ChromeFound,
		r.ChromeApp,
		r.DefaultData,
		r.PrimaryURL,
		r.SharedModeOK,
	)
}
