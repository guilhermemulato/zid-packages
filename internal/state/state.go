package state

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"zid-packages/internal/secure"
)

type LicenseState struct {
	LastAttempt time.Time       `json:"last_attempt"`
	LastSuccess time.Time       `json:"last_success"`
	Licensed    map[string]bool `json:"licensed"`
	Signature   string          `json:"signature"`
}

const DefaultPath = "/var/db/zid-packages/state.json"

func Load(path string) (LicenseState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return LicenseState{Licensed: map[string]bool{}}, nil
		}
		return LicenseState{}, err
	}
	var st LicenseState
	if err := json.Unmarshal(data, &st); err != nil {
		return LicenseState{}, err
	}
	if st.Licensed == nil {
		st.Licensed = map[string]bool{}
	}
	if st.Signature != "" {
		ok, err := verifySignature(st)
		if err != nil {
			return LicenseState{}, err
		}
		if !ok {
			return LicenseState{}, errors.New("state signature invalid")
		}
	}
	return st, nil
}

func Save(path string, st LicenseState) error {
	if dir := filepath.Dir(path); dir != "" {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return err
		}
	}
	sig, err := signState(st)
	if err != nil {
		return err
	}
	st.Signature = sig
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func signState(st LicenseState) (string, error) {
	key, err := secure.DeriveKey()
	if err != nil {
		return "", err
	}
	st.Signature = ""
	payload, err := json.Marshal(st)
	if err != nil {
		return "", err
	}
	return secure.SignHex(key, payload), nil
}

func verifySignature(st LicenseState) (bool, error) {
	key, err := secure.DeriveKey()
	if err != nil {
		return false, err
	}
	sig := st.Signature
	st.Signature = ""
	payload, err := json.Marshal(st)
	if err != nil {
		return false, err
	}
	return secure.VerifyHex(key, payload, sig), nil
}
