package tabs

import "testing"

func TestSearchTargetsMatchesTitleAndURLCaseInsensitively(t *testing.T) {
	targets := SearchTargets([]Target{
		{ID: "tab-1", Type: "page", Title: "Atlas Docs", URL: "https://openai.com/docs/atlas"},
		{ID: "tab-2", Type: "page", Title: "Blog", URL: "https://openai.com/blog"},
		{ID: "worker-1", Type: "service_worker", Title: "Atlas Docs", URL: "https://openai.com/docs/atlas"},
	}, "ATLAS")

	if len(targets) != 1 {
		t.Fatalf("unexpected targets: %+v", targets)
	}
	if targets[0].ID != "tab-1" {
		t.Fatalf("unexpected target: %+v", targets[0])
	}
}

func TestSearchTargetsReturnsEmptyForBlankQuery(t *testing.T) {
	targets := SearchTargets([]Target{
		{ID: "tab-1", Type: "page", Title: "Atlas Docs", URL: "https://openai.com/docs/atlas"},
	}, "   ")

	if len(targets) != 0 {
		t.Fatalf("unexpected targets: %+v", targets)
	}
}
