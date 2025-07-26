package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/browser"
	"github.com/pterm/pterm"
	"github.com/spf13/viper"
)

type authResult struct {
	token  string
	tenant string
	err    error
}

type callbackRequest struct {
	Token   string `json:"token"`
	Tenant  string `json:"tenant"`
	Success bool   `json:"success"`
}

// startLocalServer starts a local HTTP server to handle the OAuth callback
func startLocalServer(ctx context.Context, port int) (string, error) {
	url := viper.GetString("url")
	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    fmt.Sprintf("localhost:%d", port),
		Handler: mux,
	}

	// Open browser to login URL
	loginURL := fmt.Sprintf("%s/auth/cli?port=%d", url, port)
	pterm.Println("If you are not automatically redirected, you can manually open the following URL in your browser:")
	pterm.Println(pterm.Cyan(loginURL))
	browser.OpenURL(loginURL)

	// Channel to communicate when auth is complete
	authComplete := make(chan authResult, 1)

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight OPTIONS request
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var req callbackRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			authComplete <- authResult{"", "", fmt.Errorf("invalid request body: %v", err)}
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if !req.Success || req.Token == "" {
			authComplete <- authResult{"", "", fmt.Errorf("authentication failed")}
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Return success
		w.WriteHeader(http.StatusOK)

		// Signal auth is complete with token
		authComplete <- authResult{req.Token, req.Tenant, nil}
	})

	// Start server in goroutine
	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			authComplete <- authResult{"", "", fmt.Errorf("server error: %v", err)}
		}
	}()

	// Wait for either auth completion or timeout
	select {
	case result := <-authComplete:
		server.Shutdown(ctx)
		return result.token, result.err
	case <-time.After(20 * time.Minute):
		server.Shutdown(ctx)
		return "", fmt.Errorf("authentication timed out")
	case <-ctx.Done():
		server.Shutdown(ctx)
		return "", ctx.Err()
	}
}
