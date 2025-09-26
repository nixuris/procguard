package gui

import (
	"crypto/rand"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"procguard/internal/auth"
	"procguard/internal/config"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

//go:embed dashboard.html
var dashboard []byte

//go:embed login.html
var loginPage []byte

//go:embed create_password.html
var createPasswordPage []byte

var GuiCmd = &cobra.Command{
	Use:   "gui",
	Short: "Run the web-based GUI",
	Run:   runGUI,
}

// --- Session Management ---
type session struct {
	expires time.Time
}

var (
	sessions      = make(map[string]session)
	sessionMutex  = &sync.Mutex{}
	sessionCookie = "procguard_session"
)

func runGUI(cmd *cobra.Command, args []string) {
	const defaultPort = "58141"
	addr := "127.0.0.1:" + defaultPort
	fmt.Println("Starting GUI on http://" + addr)
	StartWebServer(addr)
}

func StartWebServer(addr string) {
	r := http.NewServeMux()

	// Public endpoints for login/creation
	r.HandleFunc("/api/login", apiLogin)
	r.HandleFunc("/api/create_password", apiCreatePassword)

	// All other endpoints are protected by the auth middleware
	r.Handle("/", authMiddleware(http.HandlerFunc(serveDashboard)))
	r.Handle("/api/logout", authMiddleware(http.HandlerFunc(apiLogout)))
	r.Handle("/api/search", authMiddleware(http.HandlerFunc(apiSearch)))
	r.Handle("/api/block", authMiddleware(http.HandlerFunc(apiBlock)))
	r.Handle("/api/blocklist", authMiddleware(http.HandlerFunc(apiBlockList)))
	r.Handle("/api/unblock", authMiddleware(http.HandlerFunc(apiUnblock)))

	fmt.Println("GUI listening on http://" + addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		fmt.Fprintln(os.Stderr, "Error running server:", err)
		os.Exit(1)
	}
}

func serveDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(dashboard)
}

// --- API Handlers ---

func apiCreatePassword(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	hash, err := auth.HashPassword(creds.Password)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	cfg, _ := config.Load()
	cfg.PasswordHash = hash
	if err := cfg.Save(); err != nil {
		http.Error(w, "Failed to save config", http.StatusInternalServerError)
		return
	}

	createSession(w)
	w.WriteHeader(http.StatusOK)
}

func apiLogin(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	cfg, _ := config.Load()
	if !auth.CheckPasswordHash(creds.Password, cfg.PasswordHash) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	createSession(w)
	w.WriteHeader(http.StatusOK)
}

func apiLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookie)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	sessionMutex.Lock()
	delete(sessions, cookie.Value)
	sessionMutex.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:   sessionCookie,
		Value:  "",
		Path:   "/",
		MaxAge: -1, // Delete cookie
	})
	w.WriteHeader(http.StatusOK)
}

func apiSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	since := r.URL.Query().Get("since")
	until := r.URL.Query().Get("until")

	args := []string{"find", q}
	if since != "" {
		args = append(args, "--since", since)
	}
	if until != "" {
		args = append(args, "--until", until)
	}

	cmd, err := runProcGuardCommand(args...)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		http.Error(w, string(out), 500)
		return
	}

	// The find command now always produces JSON, either rich or simple.
	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

func apiBlock(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Names    []string `json:"names"`
		Password string   `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// The block command takes multiple arguments
	args := []string{"block", "add"}
	args = append(args, req.Names...)
	cmd, err := runProcGuardCommand(args...)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Pass the password to the CLI command via stdin
	stdin, _ := cmd.StdinPipe()
	go func() {
		defer stdin.Close()
		io.WriteString(stdin, req.Password+"\n")
	}()

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Assume any error from the command is an auth failure or other issue
		http.Error(w, string(output), http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func apiBlockList(w http.ResponseWriter, r *http.Request) {
	cmd, _ := runProcGuardCommand("block", "list", "--json")
	out, _ := cmd.Output()
	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

func apiUnblock(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Names    []string `json:"names"`
		Password string   `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	args := []string{"block", "rm"}
	args = append(args, req.Names...)
	cmd, err := runProcGuardCommand(args...)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Pass the password to the CLI command via stdin
	stdin, _ := cmd.StdinPipe()
	go func() {
		defer stdin.Close()
		io.WriteString(stdin, req.Password+"\n")
	}()

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Assume any error from the command is an auth failure or other issue
		http.Error(w, string(output), http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// --- Middleware & Helpers ---

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg, _ := config.Load()

		// If no password is set yet, serve the creation page.
		if cfg.PasswordHash == "" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write(createPasswordPage)
			return
		}

		cookie, err := r.Cookie(sessionCookie)
		if err != nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write(loginPage)
			return
		}

		sessionMutex.Lock()
		session, exists := sessions[cookie.Value]
		sessionMutex.Unlock()

		if !exists || time.Now().After(session.expires) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write(loginPage)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func createSession(w http.ResponseWriter) {
	b := make([]byte, 16)
	rand.Read(b)
	token := hex.EncodeToString(b)

	sessionMutex.Lock()
	sessions[token] = session{expires: time.Now().Add(12 * time.Hour)}
	sessionMutex.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Now().Add(12 * time.Hour),
	})
}
