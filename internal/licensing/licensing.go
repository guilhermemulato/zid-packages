package licensing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"zid-packages/internal/logx"
	"zid-packages/internal/packages"
	"zid-packages/internal/secure"
	"zid-packages/internal/state"
)

const (
	ModeOK           = "OK"
	ModeOfflineGrace = "OFFLINE_GRACE"
	ModeExpired      = "EXPIRED"
	ModeNeverOK      = "NEVER_OK"
)

const (
	licenseURL    = "https://webhook.c01.soulsolucoes.com.br/webhook/bf26a31e-11f4-4dfd-8659-94ce045b3323/soul/licensing"
	licenseHeader = "x-auth-n8n"
	licenseToken  = "58ff7159c6d562c4d665de1d4d9a60f9546a0fcec885a15239f5bf5d25a48c80"
)

type licenseRequest struct {
	Hostname string `json:"hostname"`
	ID       string `json:"id"`
}

func Sync(logger *logx.Logger) error {
	now := time.Now().UTC()
	st, err := state.Load(state.DefaultPath)
	if err != nil {
		logger.Error("falha ao carregar state: " + err.Error())
		st = state.LicenseState{Licensed: map[string]bool{}}
	}

	st.LastAttempt = now

	payload, err := buildRequest()
	if err != nil {
		_ = state.Save(state.DefaultPath, st)
		return err
	}

	body, err := json.Marshal(payload)
	if err != nil {
		_ = state.Save(state.DefaultPath, st)
		return err
	}

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest(http.MethodPost, licenseURL, bytes.NewReader(body))
	if err != nil {
		_ = state.Save(state.DefaultPath, st)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(licenseHeader, licenseToken)

	resp, err := client.Do(req)
	if err != nil {
		_ = state.Save(state.DefaultPath, st)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		_ = state.Save(state.DefaultPath, st)
		return fmt.Errorf("licensing http status %d", resp.StatusCode)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		_ = state.Save(state.DefaultPath, st)
		return err
	}
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		if err := state.Save(state.DefaultPath, st); err != nil {
			return err
		}
		logger.Info("licensing sync sem resposta, mantendo estado anterior")
		return nil
	}

	out, err := parseLicenseMap(raw)
	if err != nil {
		_ = state.Save(state.DefaultPath, st)
		return err
	}

	licensed := map[string]bool{}
	for _, pkg := range packages.All() {
		licensed[pkg.Key] = licensedByKey(out, pkg.Key)
	}

	st.LastSuccess = now
	st.Licensed = licensed

	if err := state.Save(state.DefaultPath, st); err != nil {
		return err
	}

	logger.Info("licenciamento atualizado")
	return nil
}

func parseLicenseMap(raw []byte) (map[string]bool, error) {
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}

	out := make(map[string]bool, len(payload))
	for key, value := range payload {
		if len(value) == 0 {
			return nil, fmt.Errorf("licensing invalid value for %s", key)
		}
		var asBool bool
		if err := json.Unmarshal(value, &asBool); err == nil {
			out[key] = asBool
			continue
		}
		var asString string
		if err := json.Unmarshal(value, &asString); err == nil {
			parsed, err := strconv.ParseBool(asString)
			if err != nil {
				return nil, fmt.Errorf("licensing invalid bool string for %s", key)
			}
			out[key] = parsed
			continue
		}
		var asNumber float64
		if err := json.Unmarshal(value, &asNumber); err == nil {
			if asNumber == 0 || asNumber == 1 {
				out[key] = asNumber == 1
				continue
			}
		}
		return nil, fmt.Errorf("licensing unsupported value for %s", key)
	}
	return out, nil
}

func licensedByKey(out map[string]bool, key string) bool {
	if v, ok := out[key]; ok {
		return v
	}
	// Compatibilidade de nomenclatura: pacote novo usa chave "zid-orchestrator"
	// no zid-packages, mas alguns clientes/webhook usam "zid-orchestration".
	if key == "zid-orchestrator" {
		return out["zid-orchestration"]
	}
	return false
}

func buildRequest() (licenseRequest, error) {
	host, err := secure.ShortHostname()
	if err != nil {
		return licenseRequest{}, err
	}
	uid, err := secure.UniqueID()
	if err != nil {
		return licenseRequest{}, err
	}
	return licenseRequest{Hostname: host, ID: uid}, nil
}

func LoadState() (state.LicenseState, error) {
	return state.Load(state.DefaultPath)
}

func Evaluate(st state.LicenseState, now time.Time) (string, time.Time) {
	if st.LastSuccess.IsZero() {
		return ModeNeverOK, time.Time{}
	}
	validUntil := st.LastSuccess.Add(7 * 24 * time.Hour)
	if now.After(validUntil) {
		return ModeExpired, validUntil
	}
	if st.LastAttempt.After(st.LastSuccess) {
		return ModeOfflineGrace, validUntil
	}
	return ModeOK, validUntil
}
