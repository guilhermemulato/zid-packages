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
