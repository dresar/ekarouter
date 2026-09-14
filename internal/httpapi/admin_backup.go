package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type BackupItem struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	SizeBytes int64  `json:"size_bytes"`
	CreatedAt string `json:"created_at"`
}

func (a *AdminHandler) CreateBackup(w http.ResponseWriter, r *http.Request) {
	var body struct {
		DestPath string `json:"dest_path"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	destPath := body.DestPath
	if destPath == "" {
		destDir := "data"
		if a.cfg != nil && a.cfg.DatabasePath != "" {
			destDir = filepath.Dir(a.cfg.DatabasePath)
		}
		destPath = filepath.Join(destDir, fmt.Sprintf("backup-%d.db", time.Now().UnixNano()))
	}

	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"failed to create directory: %v"}`, err), http.StatusInternalServerError)
		return
	}

	escaped := strings.ReplaceAll(destPath, "'", "''")
	query := fmt.Sprintf("VACUUM INTO '%s'", escaped)
	if _, err := a.db.ExecContext(r.Context(), query); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"sqlite vacuum into backup failed: %v"}`, err), http.StatusInternalServerError)
		return
	}

	info, _ := os.Stat(destPath)
	var sz int64
	if info != nil {
		sz = info.Size()
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":     "created",
		"path":       destPath,
		"size_bytes": sz,
		"created_at": time.Now().UTC().Format(time.RFC3339),
	})
}

func (a *AdminHandler) ListBackups(w http.ResponseWriter, r *http.Request) {
	destDir := "data"
	if a.cfg != nil && a.cfg.DatabasePath != "" {
		destDir = filepath.Dir(a.cfg.DatabasePath)
	}

	entries, err := os.ReadDir(destDir)
	if err != nil {
		if os.IsNotExist(err) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"backups": []BackupItem{}})
			return
		}
		http.Error(w, `{"error":"failed to list backup directory"}`, http.StatusInternalServerError)
		return
	}

	mainDbName := ""
	if a.cfg != nil && a.cfg.DatabasePath != "" {
		mainDbName = filepath.Base(a.cfg.DatabasePath)
	}

	var backups []BackupItem
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".db") && entry.Name() != mainDbName {
			info, err := entry.Info()
			if err == nil {
				backups = append(backups, BackupItem{
					Name:      entry.Name(),
					Path:      filepath.Join(destDir, entry.Name()),
					SizeBytes: info.Size(),
					CreatedAt: info.ModTime().UTC().Format(time.RFC3339),
				})
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"backups": backups,
	})
}
