package packages

import "testing"

func TestServiceFirewallCleanupAction(t *testing.T) {
	tests := []struct {
		key          string
		wantPath     string
		wantFunction string
		wantOK       bool
	}{
		{
			key:          "zid-proxy",
			wantPath:     "/usr/local/pkg/zid-proxy.inc",
			wantFunction: "zidproxy_service_poststop_hook",
			wantOK:       true,
		},
		{
			key:          "zid-geolocation",
			wantPath:     "/usr/local/pkg/zid-geolocation.inc",
			wantFunction: "zid_geolocation_clear_floating_rules",
			wantOK:       true,
		},
		{
			key:    "zid-logs",
			wantOK: false,
		},
	}

	for _, tc := range tests {
		gotPath, gotFunction, gotOK := serviceFirewallCleanupAction(tc.key)
		if gotOK != tc.wantOK {
			t.Fatalf("serviceFirewallCleanupAction(%q) ok=%v; want %v", tc.key, gotOK, tc.wantOK)
		}
		if gotPath != tc.wantPath {
			t.Fatalf("serviceFirewallCleanupAction(%q) path=%q; want %q", tc.key, gotPath, tc.wantPath)
		}
		if gotFunction != tc.wantFunction {
			t.Fatalf("serviceFirewallCleanupAction(%q) function=%q; want %q", tc.key, gotFunction, tc.wantFunction)
		}
	}
}

func TestServiceStartPostAction(t *testing.T) {
	tests := []struct {
		key          string
		wantPath     string
		wantFunction string
		wantOK       bool
	}{
		{
			key:          "zid-geolocation",
			wantPath:     "/usr/local/pkg/zid-geolocation.inc",
			wantFunction: "zid_geolocation_apply_async",
			wantOK:       true,
		},
		{
			key:    "zid-proxy",
			wantOK: false,
		},
	}

	for _, tc := range tests {
		gotPath, gotFunction, gotOK := serviceStartPostAction(tc.key)
		if gotOK != tc.wantOK {
			t.Fatalf("serviceStartPostAction(%q) ok=%v; want %v", tc.key, gotOK, tc.wantOK)
		}
		if gotPath != tc.wantPath {
			t.Fatalf("serviceStartPostAction(%q) path=%q; want %q", tc.key, gotPath, tc.wantPath)
		}
		if gotFunction != tc.wantFunction {
			t.Fatalf("serviceStartPostAction(%q) function=%q; want %q", tc.key, gotFunction, tc.wantFunction)
		}
	}
}

func TestServicePHPFunctionScript(t *testing.T) {
	got := servicePHPFunctionScript("/usr/local/pkg/zid-geolocation.inc", "zid_geolocation_clear_floating_rules")
	want := `require_once("/usr/local/pkg/zid-geolocation.inc"); if (function_exists("zid_geolocation_clear_floating_rules")) { zid_geolocation_clear_floating_rules(); }`
	if got != want {
		t.Fatalf("servicePHPFunctionScript()=%q; want %q", got, want)
	}
}

func TestEnableSnapshotZidAccess_HasLegacyKeys(t *testing.T) {
	snap := EnableSnapshot("zid-access")
	wantKeys := []string{
		"php:installedpackages/(zidaccess|zid-access|zid_access)/config/enable",
		"config:installedpackages/zidaccess/config/enable",
		"config:installedpackages/zid-access/config/enable",
		"config-loose:installedpackages/zid-access/config/enable",
	}
	for _, k := range wantKeys {
		if _, ok := snap[k]; !ok {
			t.Fatalf("EnableSnapshot(zid-access) missing key: %q", k)
		}
	}
}

func TestExtractNumericVersion(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{in: "0.2.3", want: "0.2.3"},
		{in: "zid-logs version dev", want: ""},
		{in: "pkg 1.2.3_4", want: "1.2.3"},
	}
	for _, tc := range tests {
		if got := extractNumericVersion(tc.in); got != tc.want {
			t.Fatalf("extractNumericVersion(%q)=%q; want %q", tc.in, got, tc.want)
		}
	}
}
