package storage

import (
	"bytes"
	"context"
	"crypto/sha1"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Config struct {
	Provider            string `json:"provider"`
	CloudinaryCloudName string `json:"cloudinary_cloud_name"`
	CloudinaryAPIKey    string `json:"cloudinary_api_key"`
	CloudinaryAPISecret string `json:"cloudinary_api_secret"`
	CloudinaryFolder    string `json:"cloudinary_folder"`
	ImageKitPublicKey   string `json:"imagekit_public_key"`
	ImageKitPrivateKey  string `json:"imagekit_private_key"`
	ImageKitURLEndpoint string `json:"imagekit_url_endpoint"`
	ImageKitFolder      string `json:"imagekit_folder"`
	CDNCustomDomain     string `json:"cdn_custom_domain"`
	GitHubToken         string `json:"github_token"`
	GitHubOwner         string `json:"github_owner"`
	GitHubRepo          string `json:"github_repo"`
	GitHubBranch        string `json:"github_branch"`
	GitHubFolder        string `json:"github_folder"`
	GitHubCDNDomain     string `json:"github_cdn_domain"`
}

type UploadResult struct {
	URL      string `json:"url"`
	Provider string `json:"provider"`
	FileID   string `json:"file_id"`
	Size     int64  `json:"size"`
	Width    int    `json:"width,omitempty"`
	Height   int    `json:"height,omitempty"`
	Format   string `json:"format,omitempty"`
}

type CustomIcon struct {
	ProviderID      string `json:"provider_id"`
	IconURL         string `json:"icon_url"`
	DisplayName     string `json:"display_name"`
	StorageProvider string `json:"storage_provider"`
	UpdatedAt       string `json:"updated_at"`
}

type Service struct {
	db        *sql.DB
	uploadDir string
	client    *http.Client
}

func NewService(db *sql.DB, uploadDir string) *Service {
	if uploadDir == "" {
		uploadDir = filepath.Join("data", "uploads")
	}
	_ = os.MkdirAll(uploadDir, 0755)
	return &Service{
		db:        db,
		uploadDir: uploadDir,
		client:    &http.Client{Timeout: 30 * time.Second},
	}
}

func (s *Service) GetConfig(ctx context.Context) (*Config, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT provider, cloudinary_cloud_name, cloudinary_api_key, cloudinary_api_secret, cloudinary_folder,
       imagekit_public_key, imagekit_private_key, imagekit_url_endpoint, imagekit_folder, cdn_custom_domain,
       COALESCE(github_token,''), COALESCE(github_owner,''), COALESCE(github_repo,''),
       COALESCE(github_branch,'main'), COALESCE(github_folder,'uploads'), COALESCE(github_cdn_domain,'')
FROM media_storage_configs WHERE id = 'active' LIMIT 1`)

	var cfg Config
	err := row.Scan(
		&cfg.Provider,
		&cfg.CloudinaryCloudName,
		&cfg.CloudinaryAPIKey,
		&cfg.CloudinaryAPISecret,
		&cfg.CloudinaryFolder,
		&cfg.ImageKitPublicKey,
		&cfg.ImageKitPrivateKey,
		&cfg.ImageKitURLEndpoint,
		&cfg.ImageKitFolder,
		&cfg.CDNCustomDomain,
		&cfg.GitHubToken,
		&cfg.GitHubOwner,
		&cfg.GitHubRepo,
		&cfg.GitHubBranch,
		&cfg.GitHubFolder,
		&cfg.GitHubCDNDomain,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &Config{
				Provider:         "local",
				CloudinaryFolder: "ekarouter",
				ImageKitFolder:   "ekarouter",
				GitHubBranch:     "main",
				GitHubFolder:     "uploads",
			}, nil
		}
		return nil, err
	}
	if cfg.CloudinaryFolder == "" {
		cfg.CloudinaryFolder = "ekarouter"
	}
	if cfg.ImageKitFolder == "" {
		cfg.ImageKitFolder = "ekarouter"
	}
	if cfg.GitHubBranch == "" {
		cfg.GitHubBranch = "main"
	}
	if cfg.GitHubFolder == "" {
		cfg.GitHubFolder = "uploads"
	}
	if cfg.Provider == "" {
		cfg.Provider = "local"
	}
	return &cfg, nil
}

func (s *Service) SaveConfig(ctx context.Context, cfg *Config) error {
	if cfg.Provider == "" {
		cfg.Provider = "local"
	}
	if cfg.GitHubBranch == "" {
		cfg.GitHubBranch = "main"
	}
	if cfg.GitHubFolder == "" {
		cfg.GitHubFolder = "uploads"
	}
	_, err := s.db.ExecContext(ctx, `
INSERT INTO media_storage_configs (
    id, provider, cloudinary_cloud_name, cloudinary_api_key, cloudinary_api_secret, cloudinary_folder,
    imagekit_public_key, imagekit_private_key, imagekit_url_endpoint, imagekit_folder, cdn_custom_domain,
    github_token, github_owner, github_repo, github_branch, github_folder, github_cdn_domain,
    updated_at
) VALUES ('active', ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
ON CONFLICT(id) DO UPDATE SET
    provider = excluded.provider,
    cloudinary_cloud_name = excluded.cloudinary_cloud_name,
    cloudinary_api_key = excluded.cloudinary_api_key,
    cloudinary_api_secret = CASE WHEN excluded.cloudinary_api_secret != '' THEN excluded.cloudinary_api_secret ELSE media_storage_configs.cloudinary_api_secret END,
    cloudinary_folder = excluded.cloudinary_folder,
    imagekit_public_key = excluded.imagekit_public_key,
    imagekit_private_key = CASE WHEN excluded.imagekit_private_key != '' THEN excluded.imagekit_private_key ELSE media_storage_configs.imagekit_private_key END,
    imagekit_url_endpoint = excluded.imagekit_url_endpoint,
    imagekit_folder = excluded.imagekit_folder,
    cdn_custom_domain = excluded.cdn_custom_domain,
    github_token = CASE WHEN excluded.github_token != '' THEN excluded.github_token ELSE media_storage_configs.github_token END,
    github_owner = excluded.github_owner,
    github_repo = excluded.github_repo,
    github_branch = excluded.github_branch,
    github_folder = excluded.github_folder,
    github_cdn_domain = excluded.github_cdn_domain,
    updated_at = CURRENT_TIMESTAMP`,
		cfg.Provider,
		cfg.CloudinaryCloudName,
		cfg.CloudinaryAPIKey,
		cfg.CloudinaryAPISecret,
		cfg.CloudinaryFolder,
		cfg.ImageKitPublicKey,
		cfg.ImageKitPrivateKey,
		cfg.ImageKitURLEndpoint,
		cfg.ImageKitFolder,
		cfg.CDNCustomDomain,
		cfg.GitHubToken,
		cfg.GitHubOwner,
		cfg.GitHubRepo,
		cfg.GitHubBranch,
		cfg.GitHubFolder,
		cfg.GitHubCDNDomain,
	)
	return err
}

func (s *Service) Upload(ctx context.Context, filename string, data []byte, contentType string) (*UploadResult, error) {
	cfg, err := s.GetConfig(ctx)
	if err != nil {
		return nil, err
	}

	switch strings.ToLower(cfg.Provider) {
	case "cloudinary":
		if cfg.CloudinaryCloudName != "" && cfg.CloudinaryAPIKey != "" && cfg.CloudinaryAPISecret != "" {
			res, err := s.uploadCloudinary(ctx, cfg, filename, data)
			if err == nil {
				return res, nil
			}
		}
		return s.uploadLocal(filename, data, cfg.CDNCustomDomain)

	case "imagekit":
		if cfg.ImageKitPrivateKey != "" && cfg.ImageKitURLEndpoint != "" {
			res, err := s.uploadImageKit(ctx, cfg, filename, data)
			if err == nil {
				return res, nil
			}
		}
		return s.uploadLocal(filename, data, cfg.CDNCustomDomain)

	case "github":
		if cfg.GitHubToken != "" && cfg.GitHubOwner != "" && cfg.GitHubRepo != "" {
			res, err := s.uploadGitHub(ctx, cfg, filename, data)
			if err == nil {
				return res, nil
			}
		}
		return s.uploadLocal(filename, data, cfg.CDNCustomDomain)

	default:
		return s.uploadLocal(filename, data, cfg.CDNCustomDomain)
	}
}

func (s *Service) TestConnection(ctx context.Context, targetProvider string) (*UploadResult, error) {
	cfg, err := s.GetConfig(ctx)
	if err != nil {
		return nil, err
	}

	testPNG := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D,
		0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4, 0x89, 0x00, 0x00, 0x00,
		0x0A, 0x49, 0x44, 0x41, 0x54, 0x78, 0x9C, 0x63, 0x00, 0x01, 0x00, 0x00,
		0x05, 0x00, 0x01, 0x0D, 0x0A, 0x2D, 0xB4, 0x00, 0x00, 0x00, 0x00, 0x49,
		0x45, 0x4E, 0x44, 0xAE, 0x42, 0x60, 0x82,
	}

	target := targetProvider
	if target == "" {
		target = cfg.Provider
	}

	switch strings.ToLower(target) {
	case "cloudinary":
		if cfg.CloudinaryCloudName == "" || cfg.CloudinaryAPIKey == "" || cfg.CloudinaryAPISecret == "" {
			return nil, errors.New("incomplete cloudinary credentials")
		}
		return s.uploadCloudinary(ctx, cfg, "test-connection.png", testPNG)

	case "imagekit":
		if cfg.ImageKitPrivateKey == "" || cfg.ImageKitURLEndpoint == "" {
			return nil, errors.New("incomplete imagekit credentials")
		}
		return s.uploadImageKit(ctx, cfg, "test-connection.png", testPNG)

	case "github":
		if cfg.GitHubToken == "" || cfg.GitHubOwner == "" || cfg.GitHubRepo == "" {
			return nil, errors.New("incomplete github credentials: token, owner, and repo are required")
		}
		return s.uploadGitHub(ctx, cfg, "test-connection.png", testPNG)

	default:
		return s.uploadLocal("test-connection.png", testPNG, cfg.CDNCustomDomain)
	}
}

func (s *Service) uploadLocal(filename string, data []byte, cdnDomain string) (*UploadResult, error) {
	ext := filepath.Ext(filename)
	if ext == "" {
		ext = ".png"
	}
	uniqueName := fmt.Sprintf("%s_%s%s", time.Now().Format("20060102150405"), uuid.NewString()[:8], ext)
	fullPath := filepath.Join(s.uploadDir, uniqueName)

	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return nil, fmt.Errorf("write local upload: %w", err)
	}

	urlPath := "/api/uploads/" + uniqueName
	if cdnDomain != "" {
		trimmed := strings.TrimRight(cdnDomain, "/")
		if !strings.HasPrefix(trimmed, "http://") && !strings.HasPrefix(trimmed, "https://") {
			trimmed = "https://" + trimmed
		}
		urlPath = trimmed + urlPath
	}

	return &UploadResult{
		URL:      urlPath,
		Provider: "local",
		FileID:   uniqueName,
		Size:     int64(len(data)),
		Format:   strings.TrimPrefix(ext, "."),
	}, nil
}

func (s *Service) uploadCloudinary(ctx context.Context, cfg *Config, filename string, data []byte) (*UploadResult, error) {
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	folder := cfg.CloudinaryFolder
	if folder == "" {
		folder = "ekarouter"
	}

	params := map[string]string{
		"folder":    folder,
		"timestamp": timestamp,
	}

	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var toSignParts []string
	for _, k := range keys {
		toSignParts = append(toSignParts, fmt.Sprintf("%s=%s", k, params[k]))
	}
	toSign := strings.Join(toSignParts, "&") + cfg.CloudinaryAPISecret

	hash := sha1.Sum([]byte(toSign))
	signature := hex.EncodeToString(hash[:])

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	_ = writer.WriteField("api_key", cfg.CloudinaryAPIKey)
	_ = writer.WriteField("timestamp", timestamp)
	_ = writer.WriteField("folder", folder)
	_ = writer.WriteField("signature", signature)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, err
	}
	if _, err := part.Write(data); err != nil {
		return nil, err
	}
	_ = writer.Close()

	endpoint := fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/image/upload", cfg.CloudinaryCloudName)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cloudinary upload request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("cloudinary upload failed (%d): %s", resp.StatusCode, string(respBody))
	}

	var cResp struct {
		SecureURL string `json:"secure_url"`
		URL       string `json:"url"`
		PublicID  string `json:"public_id"`
		Bytes     int64  `json:"bytes"`
		Width     int    `json:"width"`
		Height    int    `json:"height"`
		Format    string `json:"format"`
	}
	if err := json.Unmarshal(respBody, &cResp); err != nil {
		return nil, fmt.Errorf("unmarshal cloudinary response: %w", err)
	}

	retURL := cResp.SecureURL
	if retURL == "" {
		retURL = cResp.URL
	}

	return &UploadResult{
		URL:      retURL,
		Provider: "cloudinary",
		FileID:   cResp.PublicID,
		Size:     cResp.Bytes,
		Width:    cResp.Width,
		Height:   cResp.Height,
		Format:   cResp.Format,
	}, nil
}

func (s *Service) uploadImageKit(ctx context.Context, cfg *Config, filename string, data []byte) (*UploadResult, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	folder := cfg.ImageKitFolder
	if folder == "" {
		folder = "ekarouter"
	}
	if !strings.HasPrefix(folder, "/") {
		folder = "/" + folder
	}

	_ = writer.WriteField("fileName", filename)
	_ = writer.WriteField("folder", folder)
	_ = writer.WriteField("useUniqueFileName", "true")

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, err
	}
	if _, err := part.Write(data); err != nil {
		return nil, err
	}
	_ = writer.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://upload.imagekit.io/api/v1/files/upload", body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	auth := base64.StdEncoding.EncodeToString([]byte(cfg.ImageKitPrivateKey + ":"))
	req.Header.Set("Authorization", "Basic "+auth)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("imagekit upload request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("imagekit upload failed (%d): %s", resp.StatusCode, string(respBody))
	}

	var ikResp struct {
		FileID string `json:"fileId"`
		Name   string `json:"name"`
		URL    string `json:"url"`
		Size   int64  `json:"size"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
		Format string `json:"format"`
	}
	if err := json.Unmarshal(respBody, &ikResp); err != nil {
		return nil, fmt.Errorf("unmarshal imagekit response: %w", err)
	}

	return &UploadResult{
		URL:      ikResp.URL,
		Provider: "imagekit",
		FileID:   ikResp.FileID,
		Size:     ikResp.Size,
		Width:    ikResp.Width,
		Height:   ikResp.Height,
		Format:   ikResp.Format,
	}, nil
}

func (s *Service) uploadGitHub(ctx context.Context, cfg *Config, filename string, data []byte) (*UploadResult, error) {
	ext := filepath.Ext(filename)
	if ext == "" {
		ext = ".png"
	}
	uniqueName := fmt.Sprintf("%s_%s%s", time.Now().Format("20060102150405"), uuid.NewString()[:8], ext)

	folder := strings.Trim(cfg.GitHubFolder, "/")
	if folder == "" {
		folder = "uploads"
	}
	filePath := folder + "/" + uniqueName

	branch := cfg.GitHubBranch
	if branch == "" {
		branch = "main"
	}

	encoded := base64.StdEncoding.EncodeToString(data)

	payload := map[string]any{
		"message": "upload: " + uniqueName,
		"content": encoded,
		"branch":  branch,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal github payload: %w", err)
	}

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s",
		cfg.GitHubOwner, cfg.GitHubRepo, filePath)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, apiURL, bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.GitHubToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("github upload request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("github upload failed (%d): %s", resp.StatusCode, string(respBody))
	}

	cdnDomain := strings.TrimRight(cfg.GitHubCDNDomain, "/")
	var fileURL string
	if cdnDomain != "" {
		fileURL = cdnDomain + "/" + filePath
	} else {
		fileURL = fmt.Sprintf("https://cdn.jsdelivr.net/gh/%s/%s@%s/%s",
			cfg.GitHubOwner, cfg.GitHubRepo, branch, filePath)
	}

	return &UploadResult{
		URL:      fileURL,
		Provider: "github",
		FileID:   filePath,
		Size:     int64(len(data)),
		Format:   strings.TrimPrefix(ext, "."),
	}, nil
}

func (s *Service) SaveCustomIcon(ctx context.Context, providerID, iconURL, displayName, storageProvider string) error {
	_, err := s.db.ExecContext(ctx, `
INSERT INTO provider_custom_icons (provider_id, icon_url, display_name, storage_provider, updated_at)
VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
ON CONFLICT(provider_id) DO UPDATE SET
    icon_url = excluded.icon_url,
    display_name = excluded.display_name,
    storage_provider = excluded.storage_provider,
    updated_at = CURRENT_TIMESTAMP`,
		providerID, iconURL, displayName, storageProvider,
	)
	return err
}

func (s *Service) GetCustomIcon(ctx context.Context, providerID string) (*CustomIcon, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT provider_id, icon_url, display_name, storage_provider, COALESCE(updated_at, '')
FROM provider_custom_icons WHERE provider_id = ? LIMIT 1`, providerID)

	var icon CustomIcon
	if err := row.Scan(&icon.ProviderID, &icon.IconURL, &icon.DisplayName, &icon.StorageProvider, &icon.UpdatedAt); err != nil {
		return nil, err
	}
	return &icon, nil
}

func (s *Service) ListCustomIcons(ctx context.Context) (map[string]CustomIcon, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT provider_id, icon_url, display_name, storage_provider, COALESCE(updated_at, '')
FROM provider_custom_icons`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]CustomIcon)
	for rows.Next() {
		var icon CustomIcon
		if err := rows.Scan(&icon.ProviderID, &icon.IconURL, &icon.DisplayName, &icon.StorageProvider, &icon.UpdatedAt); err == nil {
			result[icon.ProviderID] = icon
		}
	}
	return result, nil
}

func (s *Service) DeleteCustomIcon(ctx context.Context, providerID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM provider_custom_icons WHERE provider_id = ?`, providerID)
	return err
}
