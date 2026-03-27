package chrome

import "testing"

func TestBuildLaunchArgs(t *testing.T) {
	args := BuildLaunchArgs("/Applications/Google Chrome.app", "https://chatgpt.com/atlas?get-started", "/tmp/atlasx-profile", false)
	if len(args) == 0 || args[0] != "-na" {
		t.Fatalf("unexpected isolated launch args: %#v", args)
	}
	shared := BuildLaunchArgs("/Applications/Google Chrome.app", "https://chatgpt.com/atlas?get-started", "", true)
	if len(shared) == 0 || shared[0] != "-a" {
		t.Fatalf("unexpected shared launch args: %#v", shared)
	}
}
