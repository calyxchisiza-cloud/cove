package cove

import (
	_ "embed"
	"encoding/json"
	"net"
	"net/http"
	"time"
)

//go:embed index.html
var indexHTML []byte

type runRequest struct {
	Path           string   `json:"path"`
	Args           []string `json:"args"`
	Dir            string   `json:"dir"`
	Env            []string `json:"env"`
	TimeoutSeconds float64  `json:"timeout_seconds"`
	MaxOutput      int      `json:"max_output"`
	Memory         uint64   `json:"memory"`
}

// Handler serves the management page and runs one command at a time.
func Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(indexHTML)
	})
	mux.HandleFunc("POST /v1/run", func(w http.ResponseWriter, r *http.Request) {
		var body runRequest
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&body); err != nil {
			http.Error(w, ErrInvalid.Error(), http.StatusBadRequest)
			return
		}
		ctx := r.Context()
		result, err := Run(ctx, Request{
			Path:      body.Path,
			Args:      body.Args,
			Dir:       body.Dir,
			Env:       body.Env,
			Timeout:   time.Duration(body.TimeoutSeconds * float64(time.Second)),
			MaxOutput: body.MaxOutput,
			Memory:    body.Memory,
		})
		if err != nil {
			code := http.StatusBadRequest
			if err != ErrInvalid && err != ErrLimited {
				code = http.StatusInternalServerError
			}
			http.Error(w, err.Error(), code)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(result)
	})
	return mux
}

// Serve listens on a loopback address.
func Serve(addr string) (*http.Server, error) {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, ErrInvalid
	}
	ip := net.ParseIP(host)
	if host != "localhost" && (ip == nil || !ip.IsLoopback()) {
		return nil, ErrInvalid
	}
	srv := &http.Server{Addr: addr, Handler: Handler(), ReadHeaderTimeout: 5 * time.Second}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	go func() { _ = srv.Serve(ln) }()
	return srv, nil
}
