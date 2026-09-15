package httpapi

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/dresar/ekarouter/internal/gateway"
)

func readJSON(r *http.Request, dst any) error {
	return json.NewDecoder(r.Body).Decode(dst)
}


type ProxyForwarder struct {
	gw *gateway.Gateway
	db *sql.DB
}

func NewProxyForwarder(gw *gateway.Gateway, db *sql.DB) *ProxyForwarder {
	return &ProxyForwarder{gw: gw, db: db}
}

type upstreamTarget struct {
	baseURL string
	apiKey  string
	model   string
}

func (p *ProxyForwarder) resolveTarget(ctx context.Context, modelStr string) (*upstreamTarget, error) {
	if p.db == nil {
		return nil, fmt.Errorf("no db configured")
	}
	var baseURL, apiKeyEnc, externalName string
	err := p.db.QueryRowContext(ctx, `
		SELECT COALESCE(pr.base_url,''), COALESCE(a.credentials,''), COALESCE(m.external_name,?)
		FROM routes r
		JOIN route_items ri ON ri.route_id = r.id
		JOIN providers pr ON pr.id = ri.provider_id
		JOIN accounts a ON a.id = ri.account_id
		LEFT JOIN models m ON m.id = ri.model_id
		WHERE r.name = ? AND r.enabled = 1 AND ri.enabled = 1
		ORDER BY ri.priority DESC
		LIMIT 1`, modelStr, modelStr).Scan(&baseURL, &apiKeyEnc, &externalName)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no route found for model: %s", modelStr)
	}
	if err != nil {
		return nil, err
	}
	return &upstreamTarget{baseURL: baseURL, apiKey: apiKeyEnc, model: externalName}, nil
}

func (p *ProxyForwarder) ForwardJSON(ctx context.Context, method, path string, body any, modelStr string) (*http.Response, error) {
	target, err := p.resolveTarget(ctx, modelStr)
	if err != nil {
		return nil, err
	}
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(b)
	}
	url := strings.TrimRight(target.baseURL, "/") + path
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if target.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+target.apiKey)
	}
	return http.DefaultClient.Do(req)
}

func (p *ProxyForwarder) ForwardMultipart(ctx context.Context, originalReq *http.Request, upstreamPath, modelStr string) (*http.Response, error) {
	target, err := p.resolveTarget(ctx, modelStr)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	if err := originalReq.ParseMultipartForm(32 << 20); err != nil {
		return nil, err
	}
	for key, vals := range originalReq.MultipartForm.Value {
		for _, v := range vals {
			_ = mw.WriteField(key, v)
		}
	}
	for key, files := range originalReq.MultipartForm.File {
		for _, fh := range files {
			f, err := fh.Open()
			if err != nil {
				return nil, err
			}
			part, err := mw.CreateFormFile(key, fh.Filename)
			if err != nil {
				f.Close()
				return nil, err
			}
			_, _ = io.Copy(part, f)
			f.Close()
		}
	}
	mw.Close()

	url := strings.TrimRight(target.baseURL, "/") + upstreamPath
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	if target.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+target.apiKey)
	}
	return http.DefaultClient.Do(req)
}
