package providers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

type ErrorClass string

const (
	ErrorClassInvalidRequest ErrorClass = "invalid_request"
	ErrorClassAuth           ErrorClass = "authentication"
	ErrorClassQuota          ErrorClass = "quota"
	ErrorClassRateLimit      ErrorClass = "rate_limit"
	ErrorClassTimeout        ErrorClass = "timeout"
	ErrorClassNetwork        ErrorClass = "network"
	ErrorClassUpstream5xx    ErrorClass = "upstream_5xx"
	ErrorClassInternal       ErrorClass = "internal"
)

type ProviderError struct {
	StatusCode int
	Class      ErrorClass
	Message    string
	Err        error
}

func (e *ProviderError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %d: %s (%v)", e.Class, e.StatusCode, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %d: %s", e.Class, e.StatusCode, e.Message)
}

func (e *ProviderError) Unwrap() error {
	return e.Err
}

func (e *ProviderError) IsTransient() bool {
	switch e.Class {
	case ErrorClassRateLimit, ErrorClassQuota, ErrorClassTimeout, ErrorClassNetwork, ErrorClassUpstream5xx:
		return true
	default:
		return false
	}
}

func ClassifyHTTPError(statusCode int, body string) *ProviderError {
	lowerBody := strings.ToLower(body)
	if strings.Contains(lowerBody, "insufficient_quota") || strings.Contains(lowerBody, "quota_exceeded") || strings.Contains(lowerBody, "resource_exhausted") {
		return &ProviderError{StatusCode: statusCode, Class: ErrorClassQuota, Message: "upstream quota exceeded"}
	}
	if statusCode == 529 || strings.Contains(lowerBody, "overloaded") {
		return &ProviderError{StatusCode: statusCode, Class: ErrorClassUpstream5xx, Message: "upstream server overloaded"}
	}

	switch {
	case statusCode == http.StatusTooManyRequests:
		return &ProviderError{StatusCode: statusCode, Class: ErrorClassRateLimit, Message: "upstream rate limit exceeded"}
	case statusCode == http.StatusPaymentRequired:
		return &ProviderError{StatusCode: statusCode, Class: ErrorClassQuota, Message: "upstream quota exceeded"}
	case statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden:
		return &ProviderError{StatusCode: statusCode, Class: ErrorClassAuth, Message: "upstream authentication failed"}
	case statusCode >= 500:
		return &ProviderError{StatusCode: statusCode, Class: ErrorClassUpstream5xx, Message: fmt.Sprintf("upstream server error: %d", statusCode)}
	case statusCode >= 400:
		return &ProviderError{StatusCode: statusCode, Class: ErrorClassInvalidRequest, Message: fmt.Sprintf("upstream client error: %d", statusCode)}
	default:
		return &ProviderError{StatusCode: statusCode, Class: ErrorClassInternal, Message: body}
	}
}

func IsTransient(err error) bool {
	if err == nil {
		return false
	}
	var pe *ProviderError
	if errors.As(err, &pe) {
		return pe.IsTransient()
	}
	return false
}

type Registry struct {
	mu       sync.RWMutex
	adapters map[string]Adapter
}

func NewRegistry() *Registry {
	r := &Registry{
		adapters: make(map[string]Adapter),
	}
	return r
}

func (r *Registry) Register(kind string, adapter Adapter) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.adapters[kind] = adapter
}

func (r *Registry) Get(kind string) (Adapter, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.adapters[kind]
	if !ok {
		return nil, fmt.Errorf("unsupported provider kind: %s", kind)
	}
	return a, nil
}
