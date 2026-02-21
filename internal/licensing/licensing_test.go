package licensing

import "testing"

func TestParseLicenseMap(t *testing.T) {
	t.Run("bools", func(t *testing.T) {
		raw := []byte(`{"zid-proxy":true,"zid-geolocation":false}`)
		got, err := parseLicenseMap(raw)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if got["zid-proxy"] != true || got["zid-geolocation"] != false {
			t.Fatalf("unexpected map: %#v", got)
		}
	})

	t.Run("strings", func(t *testing.T) {
		raw := []byte(`{"zid-proxy":"true","zid-geolocation":"false"}`)
		got, err := parseLicenseMap(raw)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if got["zid-proxy"] != true || got["zid-geolocation"] != false {
			t.Fatalf("unexpected map: %#v", got)
		}
	})

	t.Run("numbers", func(t *testing.T) {
		raw := []byte(`{"zid-proxy":1,"zid-geolocation":0}`)
		got, err := parseLicenseMap(raw)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if got["zid-proxy"] != true || got["zid-geolocation"] != false {
			t.Fatalf("unexpected map: %#v", got)
		}
	})

	t.Run("invalid", func(t *testing.T) {
		raw := []byte(`{"zid-proxy":"maybe"}`)
		if _, err := parseLicenseMap(raw); err == nil {
			t.Fatalf("expected error")
		}
	})
}

func TestLicensedByKey_OrchestratorAlias(t *testing.T) {
	cases := []struct {
		name string
		out  map[string]bool
		key  string
		want bool
	}{
		{
			name: "direct key",
			out:  map[string]bool{"zid-orchestrator": true},
			key:  "zid-orchestrator",
			want: true,
		},
		{
			name: "alias key",
			out:  map[string]bool{"zid-orchestration": true},
			key:  "zid-orchestrator",
			want: true,
		},
		{
			name: "missing key",
			out:  map[string]bool{},
			key:  "zid-orchestrator",
			want: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := licensedByKey(tc.out, tc.key); got != tc.want {
				t.Fatalf("licensedByKey(%v, %q)=%v; want %v", tc.out, tc.key, got, tc.want)
			}
		})
	}
}
