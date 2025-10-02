package gui

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/exec"
	"procguard/internal/auth"
	"procguard/internal/blocklist"
	"procguard/internal/config"
	"procguard/internal/logsearch"

	"procguard/internal/webblocklist"
	"slices"
	"strings"
	"time"
)

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	s.mu.Lock()
	localIsAuthenticated := s.isAuthenticated
	s.mu.Unlock()

	if !localIsAuthenticated {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := w.Write(dashboardHTML); err != nil {
		s.logger.Printf("Error writing response: %v", err)
	}
}

func (s *Server) handleLoginTemplate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := w.Write(loginHTML); err != nil {
		s.logger.Printf("Error writing response: %v", err)
	}
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	s.isAuthenticated = false
	s.mu.Unlock()
	http.Redirect(w, r, "/login", http.StatusFound)
}

func (s *Server) handlePing(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleHasPassword(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.Load()
	if err != nil {
		http.Error(w, "Failed to load config", http.StatusInternalServerError)
		return
	}
	hasPassword := cfg.PasswordHash != ""
	if err := json.NewEncoder(w).Encode(map[string]bool{"hasPassword": hasPassword}); err != nil {
		s.logger.Printf("Error encoding response: %v", err)
	}
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	cfg, err := config.Load()
	if err != nil {
		http.Error(w, "Failed to load config", http.StatusInternalServerError)
		return
	}

	if auth.CheckPasswordHash(req.Password, cfg.PasswordHash) {
		s.mu.Lock()
		s.isAuthenticated = true
		s.mu.Unlock()
		if err := json.NewEncoder(w).Encode(map[string]bool{"success": true}); err != nil {
			s.logger.Printf("Error encoding response: %v", err)
		}
	} else {
		if err := json.NewEncoder(w).Encode(map[string]bool{"success": false}); err != nil {
			s.logger.Printf("Error encoding response: %v", err)
		}
	}
}

func (s *Server) handleSetPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	cfg, err := config.Load()
	if err != nil {
		http.Error(w, "Failed to load config", http.StatusInternalServerError)
		return
	}

	if cfg.PasswordHash != "" {
		http.Error(w, "Password already set", http.StatusForbidden)
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	cfg.PasswordHash = hash
	if err := cfg.Save(); err != nil {
			http.Error(w, "Failed to save password", http.StatusInternalServerError)
			return
		}
	
		s.mu.Lock()
		s.isAuthenticated = true
		s.mu.Unlock()
		if err := json.NewEncoder(w).Encode(map[string]bool{"success": true}); err != nil {
			s.logger.Printf("Error encoding response: %v", err)
		}
}

func (s *Server) apiSearch(w http.ResponseWriter, r *http.Request) {
	query := strings.ToLower(r.URL.Query().Get("q"))
	sinceStr := r.URL.Query().Get("since")
	untilStr := r.URL.Query().Get("until")

	results, err := logsearch.Search(query, sinceStr, untilStr)
	if err != nil {
		http.Error(w, "Failed to search logs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(results); err != nil {
		s.logger.Printf("Error encoding response: %v", err)
	}
}

func (s *Server) apiBlock(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Names []string `json:"names"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	list, err := blocklist.Load()
	if err != nil {
		http.Error(w, "Failed to load blocklist", http.StatusInternalServerError)
		return
	}

	for _, name := range req.Names {
		lowerName := strings.ToLower(name)
		if !slices.Contains(list, lowerName) {
			list = append(list, lowerName)
		}
	}

	if err := blocklist.Save(list); err != nil {
		http.Error(w, "Failed to save blocklist", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]bool{"ok": true}); err != nil {
		s.logger.Printf("Error encoding response: %v", err)
	}
}

func (s *Server) apiBlockList(w http.ResponseWriter, r *http.Request) {
	list, err := blocklist.Load()
	if err != nil {
		http.Error(w, "Failed to load blocklist", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(list); err != nil {
		s.logger.Printf("Error encoding response: %v", err)
	}
}

func (s *Server) apiUnblock(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Names []string `json:"names"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	list, err := blocklist.Load()
	if err != nil {
		http.Error(w, "Failed to load blocklist", http.StatusInternalServerError)
		return
	}

	for _, name := range req.Names {
		lowerName := strings.ToLower(name)
		list = slices.DeleteFunc(list, func(item string) bool {
			return item == lowerName
		})
	}

	if err := blocklist.Save(list); err != nil {
		http.Error(w, "Failed to save blocklist", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]bool{"ok": true}); err != nil {
		s.logger.Printf("Error encoding response: %v", err)
	}
}

func (s *Server) apiUninstall(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	cfg, err := config.Load()
	if err != nil {
		http.Error(w, "Failed to load config", http.StatusInternalServerError)
		return
	}

	if !auth.CheckPasswordHash(req.Password, cfg.PasswordHash) {
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}

	exePath, err := os.Executable()
	if err != nil {
		http.Error(w, "Failed to get executable path", http.StatusInternalServerError)
		return
	}

	cmd := exec.Command(exePath, "uninstall", "--force-no-prompt")
	if err := cmd.Start(); err != nil {
		http.Error(w, "Failed to start uninstall process", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]bool{"ok": true}); err != nil {
		s.logger.Printf("Error encoding response: %v", err)
	}

	go func() {
		time.Sleep(1 * time.Second)
		os.Exit(0)
	}()
}

func (s *Server) apiClearBlocklist(w http.ResponseWriter, r *http.Request) {
	if err := blocklist.Save([]string{}); err != nil {
		http.Error(w, "Failed to clear blocklist", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]bool{"ok": true}); err != nil {
		s.logger.Printf("Error encoding response: %v", err)
	}
}

func (s *Server) apiSaveBlocklist(w http.ResponseWriter, r *http.Request) {
	list, err := blocklist.Load()
	if err != nil {
		http.Error(w, "Failed to get blocklist", http.StatusInternalServerError)
		return
	}

	header := map[string]interface{}{
		"exported_at": time.Now().Format(time.RFC3339),
		"blocked":     list,
	}

	b, err := json.MarshalIndent(header, "", "  ")
	if err != nil {
		http.Error(w, "Failed to marshal blocklist", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename=procguard_blocklist.json")
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(b); err != nil {
		s.logger.Printf("Error writing response: %v", err)
	}
}

func (s *Server) apiLoadBlocklist(w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file from form", http.StatusBadRequest)
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			s.logger.Printf("Error closing file: %v", err)
		}
	}()

	content, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read uploaded file", http.StatusInternalServerError)
		return
	}

	var newEntries []string
	var savedList struct {
		Blocked []string `json:"blocked"`
	}

	err = json.Unmarshal(content, &newEntries)
	if err != nil {
		err2 := json.Unmarshal(content, &savedList)
		if err2 != nil {
			http.Error(w, "Invalid JSON format in uploaded file", http.StatusBadRequest)
			return
		}
		newEntries = savedList.Blocked
	}

	existingList, err := blocklist.Load()
	if err != nil {
		http.Error(w, "Failed to load existing blocklist", http.StatusInternalServerError)
		return
	}

	for _, entry := range newEntries {
		if !slices.Contains(existingList, entry) {
			existingList = append(existingList, entry)
		}
	}

	if err := blocklist.Save(existingList); err != nil {
		http.Error(w, "Failed to save merged blocklist", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]bool{"ok": true}); err != nil {
		s.logger.Printf("Error encoding response: %v", err)
	}
}

// --- Web Blocking Handlers ---

func (s *Server) apiWebBlockList(w http.ResponseWriter, r *http.Request) {
	list, err := webblocklist.Load()
	if err != nil {
		http.Error(w, "Failed to load web blocklist", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(list); err != nil {
		s.logger.Printf("Error encoding response: %v", err)
	}
}

