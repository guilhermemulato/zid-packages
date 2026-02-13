package packages

import "testing"

func TestMatchesPath_Suffix(t *testing.T) {
	stack := []string{"pfsense", "installedpackages", "zidaccess", "config", "enable"}

	if !matchesPath(stack, []string{"installedpackages", "zidaccess", "config", "enable"}) {
		t.Fatalf("matchesPath() should match contiguous suffix")
	}
	if !matchesPath(stack, []string{"enable"}) {
		t.Fatalf("matchesPath() should match leaf element")
	}
	if matchesPath(stack, []string{"pfsense", "installedpackages", "zidaccess", "config", "enable", "extra"}) {
		t.Fatalf("matchesPath() should not match when path is longer than stack")
	}
	if matchesPath(stack, nil) {
		t.Fatalf("matchesPath() should not match empty path")
	}
	if matchesPath(stack, []string{}) {
		t.Fatalf("matchesPath() should not match empty path")
	}
}

func TestMatchesPathLoose_Subsequence(t *testing.T) {
	stack := []string{"pfsense", "installedpackages", "zidaccess", "config", "enable"}
	if !matchesPathLoose(stack, []string{"installedpackages", "zidaccess", "enable"}) {
		t.Fatalf("matchesPathLoose() should match subsequence")
	}
	if matchesPathLoose(stack, nil) {
		t.Fatalf("matchesPathLoose() should not match empty path")
	}
	if matchesPathLoose(stack, []string{}) {
		t.Fatalf("matchesPathLoose() should not match empty path")
	}
}
