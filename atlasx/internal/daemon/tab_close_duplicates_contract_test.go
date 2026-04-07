package daemon

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"atlasx/internal/tabs"
)

func TestTabCloseDuplicatesEndpointContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	restoreDaemonHooks(t, &stubTabsClient{
		duplicateClose: tabs.CloseDuplicatesResult{
			Returned: 1,
			Groups: []tabs.DuplicateCloseGroup{
				{
					URL:             "https://openai.com/docs",
					KeptTargetID:    "tab-1",
					ClosedTargetIDs: []string{"tab-2"},
					Returned:        1,
				},
			},
			ClosedTargets: []string{"tab-2"},
		},
	})

	request := httptest.NewRequest(http.MethodPost, "/v1/tabs/close-duplicates", nil)
	recorder := httptest.NewRecorder()

	NewMux(Status{}).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d body=%s", recorder.Code, recorder.Body.String())
	}

	payload := decodeObjectResponse(t, recorder)
	assertMapKeys(t, payload, "returned", "groups", "closed_targets")
}
