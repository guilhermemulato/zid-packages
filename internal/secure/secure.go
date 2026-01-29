package secure

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"strings"

	"golang.org/x/crypto/hkdf"
)

const masterSecret = "zid-packages-master-secret-2026-01"

func ShortHostname() (string, error) {
	host, err := os.Hostname()
	if err != nil {
		return "", err
	}
	host = strings.TrimSpace(host)
	if host == "" {
		return "", errors.New("hostname vazio")
	}
	if idx := strings.IndexByte(host, '.'); idx > 0 {
		host = host[:idx]
	}
	return host, nil
}

func UniqueID() (string, error) {
	raw, err := os.ReadFile("/var/db/uniqueid")
	if err != nil {
		return "", err
	}
	uid := strings.TrimSpace(string(raw))
	if uid == "" {
		return "", errors.New("uniqueid vazio")
	}
	return uid, nil
}

func DeriveKey() ([]byte, error) {
	host, err := ShortHostname()
	if err != nil {
		return nil, err
	}
	uid, err := UniqueID()
	if err != nil {
		return nil, err
	}
	salt := []byte(uid + ":" + host)
	info := []byte("zid-packages-hkdf")

	reader := hkdf.New(sha256.New, []byte(masterSecret), salt, info)
	key := make([]byte, 32)
	if _, err := io.ReadFull(reader, key); err != nil {
		return nil, err
	}
	return key, nil
}

func SignHex(key []byte, payload []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

func VerifyHex(key []byte, payload []byte, sigHex string) bool {
	expected := SignHex(key, payload)
	return hmac.Equal([]byte(expected), []byte(strings.ToLower(sigHex)))
}
