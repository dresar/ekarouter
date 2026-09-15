package httpapi

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/dresar/ekarouter/internal/providers/storage"
	"github.com/go-chi/chi/v5"
)

func (a *AdminHandler) GetStorageConfig(w http.ResponseWriter, r *http.Request) {
	svc := storage.NewService(a.db, "")
	cfg, err := svc.GetConfig(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to get storage config"}`, http.StatusInternalServerError)
		return
	}

	masked := *cfg
	if masked.CloudinaryAPISecret != "" {
		if a.crypto != nil {
			if dec, err := a.crypto.Decrypt(masked.CloudinaryAPISecret); err == nil {
				_ = dec
			}
		}
		masked.CloudinaryAPISecret = "••••••••"
	}
	if masked.ImageKitPrivateKey != "" {
		if a.crypto != nil {
			if dec, err := a.crypto.Decrypt(masked.ImageKitPrivateKey); err == nil {
				_ = dec
			}
		}
		masked.ImageKitPrivateKey = "••••••••"
	}
	if masked.GitHubToken != "" {
		if a.crypto != nil {
			if dec, err := a.crypto.Decrypt(masked.GitHubToken); err == nil {
				_ = dec
			}
		}
		masked.GitHubToken = "••••••••"
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(masked)
}

func (a *AdminHandler) SaveStorageConfig(w http.ResponseWriter, r *http.Request) {
	var input storage.Config
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	svc := storage.NewService(a.db, "")
	current, _ := svc.GetConfig(r.Context())

	if input.CloudinaryAPISecret == "••••••••" || input.CloudinaryAPISecret == "" {
		if current != nil {
			input.CloudinaryAPISecret = current.CloudinaryAPISecret
		}
	} else if a.crypto != nil && input.CloudinaryAPISecret != "" {
		if enc, err := a.crypto.Encrypt(input.CloudinaryAPISecret); err == nil {
			input.CloudinaryAPISecret = enc
		}
	}

	if input.ImageKitPrivateKey == "••••••••" || input.ImageKitPrivateKey == "" {
		if current != nil {
			input.ImageKitPrivateKey = current.ImageKitPrivateKey
		}
	} else if a.crypto != nil && input.ImageKitPrivateKey != "" {
		if enc, err := a.crypto.Encrypt(input.ImageKitPrivateKey); err == nil {
			input.ImageKitPrivateKey = enc
		}
	}

	if input.GitHubToken == "••••••••" || input.GitHubToken == "" {
		if current != nil {
			input.GitHubToken = current.GitHubToken
		}
	} else if a.crypto != nil && input.GitHubToken != "" {
		if enc, err := a.crypto.Encrypt(input.GitHubToken); err == nil {
			input.GitHubToken = enc
		}
	}

	if err := svc.SaveConfig(r.Context(), &input); err != nil {
		http.Error(w, `{"error":"failed to save storage config"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":   "ok",
		"provider": input.Provider,
	})
}

func (a *AdminHandler) TestStorageConnection(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Provider string `json:"provider"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	svc := storage.NewService(a.db, "")
	cfg, err := svc.GetConfig(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to load storage config"}`, http.StatusInternalServerError)
		return
	}

	if a.crypto != nil {
		if cfg.CloudinaryAPISecret != "" {
			if dec, err := a.crypto.Decrypt(cfg.CloudinaryAPISecret); err == nil {
				cfg.CloudinaryAPISecret = dec
			}
		}
		if cfg.ImageKitPrivateKey != "" {
			if dec, err := a.crypto.Decrypt(cfg.ImageKitPrivateKey); err == nil {
				cfg.ImageKitPrivateKey = dec
			}
		}
		if cfg.GitHubToken != "" {
			if dec, err := a.crypto.Decrypt(cfg.GitHubToken); err == nil {
				cfg.GitHubToken = dec
			}
		}
	}

	target := body.Provider
	if target == "" {
		target = cfg.Provider
	}

	var res *storage.UploadResult
	switch strings.ToLower(target) {
	case "cloudinary":
		res, err = svc.TestConnection(r.Context(), "cloudinary")
	case "imagekit":
		res, err = svc.TestConnection(r.Context(), "imagekit")
	case "github":
		res, err = svc.TestConnection(r.Context(), "github")
	default:
		res, err = svc.TestConnection(r.Context(), "local")
	}

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":   "ok",
		"provider": res.Provider,
		"url":      res.URL,
		"file_id":  res.FileID,
	})
}

func (a *AdminHandler) UploadStorageFile(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(15 << 20); err != nil {
		http.Error(w, `{"error":"invalid multipart form or file too large"}`, http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, `{"error":"file field is required"}`, http.StatusBadRequest)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, `{"error":"failed to read uploaded file"}`, http.StatusInternalServerError)
		return
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	svc := storage.NewService(a.db, "")
	cfg, _ := svc.GetConfig(r.Context())
	if cfg != nil && a.crypto != nil {
		if cfg.CloudinaryAPISecret != "" {
			if dec, err := a.crypto.Decrypt(cfg.CloudinaryAPISecret); err == nil {
				cfg.CloudinaryAPISecret = dec
			}
		}
		if cfg.ImageKitPrivateKey != "" {
			if dec, err := a.crypto.Decrypt(cfg.ImageKitPrivateKey); err == nil {
				cfg.ImageKitPrivateKey = dec
			}
		}
		if cfg.GitHubToken != "" {
			if dec, err := a.crypto.Decrypt(cfg.GitHubToken); err == nil {
				cfg.GitHubToken = dec
			}
		}
	}

	res, err := svc.Upload(r.Context(), header.Filename, data, contentType)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}

func (a *AdminHandler) ServeUploads(w http.ResponseWriter, r *http.Request) {
	filename := chi.URLParam(r, "filename")
	clean := filepath.Clean(filename)
	if strings.Contains(clean, "..") || strings.HasPrefix(clean, "/") || strings.HasPrefix(clean, "\\") {
		http.Error(w, "invalid file path", http.StatusBadRequest)
		return
	}

	uploadDir := filepath.Join("data", "uploads")
	fullPath := filepath.Join(uploadDir, clean)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}

	http.ServeFile(w, r, fullPath)
}

func (a *AdminHandler) ListProviderCustomIcons(w http.ResponseWriter, r *http.Request) {
	svc := storage.NewService(a.db, "")
	icons, err := svc.ListCustomIcons(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to list custom icons"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(icons)
}

func (a *AdminHandler) SaveProviderCustomIcon(w http.ResponseWriter, r *http.Request) {
	providerID := chi.URLParam(r, "id")
	if providerID == "" {
		http.Error(w, `{"error":"provider id required"}`, http.StatusBadRequest)
		return
	}

	var body struct {
		IconURL         string `json:"icon_url"`
		DisplayName     string `json:"display_name"`
		StorageProvider string `json:"storage_provider"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	svc := storage.NewService(a.db, "")
	if err := svc.SaveCustomIcon(r.Context(), providerID, body.IconURL, body.DisplayName, body.StorageProvider); err != nil {
		http.Error(w, `{"error":"failed to save provider icon"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":      "ok",
		"provider_id": providerID,
		"icon_url":    body.IconURL,
	})
}

func (a *AdminHandler) DeleteProviderCustomIcon(w http.ResponseWriter, r *http.Request) {
	providerID := chi.URLParam(r, "id")
	if providerID == "" {
		http.Error(w, `{"error":"provider id required"}`, http.StatusBadRequest)
		return
	}

	svc := storage.NewService(a.db, "")
	if err := svc.DeleteCustomIcon(r.Context(), providerID); err != nil {
		http.Error(w, `{"error":"failed to delete provider icon"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":      "ok",
		"provider_id": providerID,
	})
}
