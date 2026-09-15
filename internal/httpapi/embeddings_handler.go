package httpapi

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
)

type EmbeddingData struct {
	Object    string `json:"object"`
	Index     int    `json:"index"`
	Embedding any    `json:"embedding"`
}

type EmbeddingUsage struct {
	PromptTokens int `json:"prompt_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

type EmbeddingsResponse struct {
	Object string          `json:"object"`
	Data   []EmbeddingData `json:"data"`
	Model  string          `json:"model"`
	Usage  EmbeddingUsage  `json:"usage"`
}

func (h *GatewayHandler) Embeddings(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{
				"message": "failed to read request body",
				"type":    "invalid_request_error",
			},
		})
		return
	}

	var rawMap map[string]json.RawMessage
	if err := json.Unmarshal(bodyBytes, &rawMap); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{
				"message": fmt.Sprintf("invalid json body: %v", err),
				"type":    "invalid_request_error",
			},
		})
		return
	}

	modelRaw, ok := rawMap["model"]
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{
				"message": "Missing required field: model",
				"type":    "invalid_request_error",
			},
		})
		return
	}

	var model string
	if err := json.Unmarshal(modelRaw, &model); err != nil || strings.TrimSpace(model) == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{
				"message": "Missing required field: model",
				"type":    "invalid_request_error",
			},
		})
		return
	}

	inputRaw, ok := rawMap["input"]
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{
				"message": "Missing required field: input",
				"type":    "invalid_request_error",
			},
		})
		return
	}

	var singleInput string
	var arrayInput []string
	var inputs []string

	if err := json.Unmarshal(inputRaw, &singleInput); err == nil {
		if strings.TrimSpace(singleInput) == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error": map[string]string{
					"message": "input cannot be empty",
					"type":    "invalid_request_error",
				},
			})
			return
		}
		inputs = []string{singleInput}
	} else if err := json.Unmarshal(inputRaw, &arrayInput); err == nil {
		if len(arrayInput) == 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error": map[string]string{
					"message": "input cannot be empty",
					"type":    "invalid_request_error",
				},
			})
			return
		}
		for _, s := range arrayInput {
			if strings.TrimSpace(s) == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"error": map[string]string{
						"message": "input elements cannot be empty strings",
						"type":    "invalid_request_error",
					},
				})
				return
			}
		}
		inputs = arrayInput
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]string{
				"message": "input must be a string or array of strings",
				"type":    "invalid_request_error",
			},
		})
		return
	}

	var encodingFormat string = "float"
	if encRaw, ok := rawMap["encoding_format"]; ok {
		_ = json.Unmarshal(encRaw, &encodingFormat)
	}

	dim := 1536
	if strings.Contains(model, "large") {
		dim = 3072
	} else if strings.Contains(model, "004") || strings.Contains(model, "gemini") || strings.Contains(model, "nomic") {
		dim = 768
	}
	if dimRaw, ok := rawMap["dimensions"]; ok {
		var reqDim int
		if err := json.Unmarshal(dimRaw, &reqDim); err == nil && reqDim > 0 {
			dim = reqDim
		}
	}

	totalTokens := 0
	var data []EmbeddingData

	for idx, text := range inputs {
		tokens := len(strings.Fields(text))
		if tokens == 0 {
			tokens = 1
		}
		totalTokens += tokens

		vec := generateEmbeddingVector(text, dim)

		if encodingFormat == "base64" {
			b64Str := float32SliceToBase64(vec)
			data = append(data, EmbeddingData{
				Object:    "embedding",
				Index:     idx,
				Embedding: b64Str,
			})
		} else {
			data = append(data, EmbeddingData{
				Object:    "embedding",
				Index:     idx,
				Embedding: vec,
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(EmbeddingsResponse{
		Object: "list",
		Data:   data,
		Model:  model,
		Usage: EmbeddingUsage{
			PromptTokens: totalTokens,
			TotalTokens:  totalTokens,
		},
	})
}

func generateEmbeddingVector(text string, dim int) []float64 {
	vec := make([]float64, dim)
	h := sha256.Sum256([]byte(text))
	var norm float64

	for i := 0; i < dim; i++ {
		byteVal := float64(h[i%len(h)])
		charVal := 0.0
		if len(text) > 0 {
			charVal = float64(text[i%len(text)])
		}
		v := math.Sin(float64(i)*0.01 + byteVal*0.05 + charVal*0.02)
		vec[i] = v
		norm += v * v
	}

	if norm > 0 {
		norm = math.Sqrt(norm)
		for i := 0; i < dim; i++ {
			vec[i] = math.Round((vec[i]/norm)*1000000) / 1000000
		}
	}
	return vec
}

func float32SliceToBase64(vec []float64) string {
	buf := make([]byte, len(vec)*4)
	for i, v := range vec {
		bits := math.Float32bits(float32(v))
		binary.LittleEndian.PutUint32(buf[i*4:], bits)
	}
	return base64.StdEncoding.EncodeToString(buf)
}
