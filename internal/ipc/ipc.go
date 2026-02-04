package ipc

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"zid-packages/internal/licensing"
	"zid-packages/internal/logx"
	"zid-packages/internal/packages"
	"zid-packages/internal/secure"
)

const (
	SocketPath    = "/var/run/zid-packages.sock"
	nonceTTL      = 5 * time.Minute
	maxTimeSkew   = 5 * time.Minute
	opCheck       = "CHECK"
	defaultReason = "unavailable"
)

type Request struct {
	Op      string `json:"op"`
	Package string `json:"package"`
	TS      int64  `json:"ts"`
	Nonce   string `json:"nonce"`
	Sig     string `json:"sig"`
}

type Response struct {
	OK         bool   `json:"ok"`
	Licensed   bool   `json:"licensed"`
	Mode       string `json:"mode"`
	ValidUntil int64  `json:"valid_until"`
	Reason     string `json:"reason"`
	TS         int64  `json:"ts"`
	Sig        string `json:"sig"`
}

type Server struct {
	listener net.Listener
	mu       sync.Mutex
	nonces   map[string]time.Time
	closed   bool
	logger   *logx.Logger
	debug    bool
	logKeys  bool
}

func NewServer(logger *logx.Logger) *Server {
	return &Server{
		nonces:  map[string]time.Time{},
		logger:  logger,
		debug:   envTrue("ZID_PACKAGES_IPC_DEBUG"),
		logKeys: envTrue("ZID_PACKAGES_IPC_LOG_KEYS"),
	}
}

func (s *Server) Start() error {
	if s.listener != nil {
		return errors.New("ipc already started")
	}
	_ = os.Remove(SocketPath)
	ln, err := net.Listen("unix", SocketPath)
	if err != nil {
		return err
	}
	if err := os.Chmod(SocketPath, 0600); err != nil {
		_ = ln.Close()
		return err
	}
	os.Chown(SocketPath, 0, 0)

	s.listener = ln
	s.logInfo("ipc listen: " + SocketPath)
	go s.acceptLoop()
	return nil
}

func (s *Server) Stop() error {
	s.mu.Lock()
	s.closed = true
	ln := s.listener
	s.listener = nil
	s.mu.Unlock()
	if ln != nil {
		return ln.Close()
	}
	return nil
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.isClosed() {
				return
			}
			continue
		}
		go s.handleConn(conn)
	}
}

func (s *Server) isClosed() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.closed
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()
	dec := json.NewDecoder(conn)
	var req Request
	if err := dec.Decode(&req); err != nil {
		s.logInfo("ipc request invalid: " + err.Error())
		return
	}
	if s.debug {
		s.logInfo("ipc request op=" + req.Op + " pkg=" + req.Package + " ts=" + strconv.FormatInt(req.TS, 10) + " nonce=" + req.Nonce + " sig=" + req.Sig)
	}
	resp := s.handleRequest(req)
	enc := json.NewEncoder(conn)
	_ = enc.Encode(resp)
}

func (s *Server) handleRequest(req Request) Response {
	now := time.Now().UTC()
	if req.Op != opCheck || req.Package == "" || req.Nonce == "" || req.TS == 0 {
		return s.respond(req, Response{OK: false, Licensed: false, Mode: licensing.ModeNeverOK, Reason: "invalid_request", TS: now.Unix()})
	}
	if err := packages.ValidateKey(req.Package); err != nil {
		return s.respond(req, Response{OK: false, Licensed: false, Mode: licensing.ModeNeverOK, Reason: "unknown_package", TS: now.Unix()})
	}
	if skew := now.Sub(time.Unix(req.TS, 0)); skew < -maxTimeSkew || skew > maxTimeSkew {
		return s.respond(req, Response{OK: false, Licensed: false, Mode: licensing.ModeNeverOK, Reason: "invalid_ts", TS: now.Unix()})
	}
	if !s.acceptNonce(req.Nonce, now) {
		return s.respond(req, Response{OK: false, Licensed: false, Mode: licensing.ModeNeverOK, Reason: "replay", TS: now.Unix()})
	}

	key, err := secure.DeriveKey()
	if err != nil {
		return s.respond(req, Response{OK: false, Licensed: false, Mode: licensing.ModeNeverOK, Reason: defaultReason, TS: now.Unix()})
	}
	if s.logKeys {
		s.logInfo("ipc key hex=" + hex.EncodeToString(key))
	}

	unsigned := req
	unsigned.Sig = ""
	payload, err := json.Marshal(unsigned)
	if err != nil || !secure.VerifyHex(key, payload, req.Sig) {
		if s.debug && err == nil {
			s.logInfo("ipc bad_sig expected=" + secure.SignHex(key, payload))
		}
		return s.respond(req, Response{OK: false, Licensed: false, Mode: licensing.ModeNeverOK, Reason: "bad_sig", TS: now.Unix()})
	}

	st, err := licensing.LoadState()
	mode := licensing.ModeNeverOK
	validUntil := int64(0)
	licensed := false
	reason := "no_state"
	if err == nil {
		m, vu := licensing.Evaluate(st, now)
		mode = m
		validUntil = unixOrZero(vu)
		reason = mode
		if mode == licensing.ModeOK || mode == licensing.ModeOfflineGrace {
			licensed = st.Licensed[req.Package]
		}
	}

	resp := Response{
		OK:         true,
		Licensed:   licensed,
		Mode:       mode,
		ValidUntil: validUntil,
		Reason:     reason,
		TS:         now.Unix(),
	}
	return s.respond(req, resp)
}

func (s *Server) signResponse(resp Response) Response {
	key, err := secure.DeriveKey()
	if err != nil {
		resp.Sig = ""
		return resp
	}
	unsigned := resp
	unsigned.Sig = ""
	payload, err := json.Marshal(unsigned)
	if err != nil {
		resp.Sig = ""
		return resp
	}
	resp.Sig = secure.SignHex(key, payload)
	return resp
}

func (s *Server) acceptNonce(nonce string, now time.Time) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for k, ts := range s.nonces {
		if now.Sub(ts) > nonceTTL {
			delete(s.nonces, k)
		}
	}
	if _, exists := s.nonces[nonce]; exists {
		return false
	}
	s.nonces[nonce] = now
	return true
}

func unixOrZero(t time.Time) int64 {
	if t.IsZero() {
		return 0
	}
	return t.Unix()
}

func (s *Server) respond(req Request, resp Response) Response {
	signed := s.signResponse(resp)
	s.logAttempt(req, signed)
	return signed
}

func (s *Server) logAttempt(req Request, resp Response) {
	if s.logger == nil || !s.debug {
		return
	}
	msg := "ipc response pkg=" + req.Package + " ok=" + strconv.FormatBool(resp.OK) + " licensed=" + strconv.FormatBool(resp.Licensed) + " mode=" + resp.Mode + " reason=" + resp.Reason
	s.logInfo(msg)
}

func (s *Server) logInfo(msg string) {
	if s.logger == nil {
		return
	}
	s.logger.Info(msg)
}

func envTrue(name string) bool {
	val := strings.TrimSpace(strings.ToLower(os.Getenv(name)))
	return val == "1" || val == "true" || val == "yes"
}
