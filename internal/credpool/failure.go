package credpool

import (
	"errors"
	"net"
	"strings"

	"github.com/dresar/ekarouter/internal/providers"
)

type FailureAction struct {
	ShouldRotate bool
	ShouldRetry  bool
	IsPermanent  bool
	Reason       string
}

func ClassifyForRotation(err error) FailureAction {
	if err == nil {
		return FailureAction{}
	}

	var pe *providers.ProviderError
	if errors.As(err, &pe) {
		return classifyProviderError(pe)
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return FailureAction{
				ShouldRotate: true,
				ShouldRetry:  true,
				IsPermanent:  false,
				Reason:       "network_timeout",
			}
		}
		return FailureAction{
			ShouldRotate: true,
			ShouldRetry:  true,
			IsPermanent:  false,
			Reason:       "network_error",
		}
	}

	errMsg := strings.ToLower(err.Error())

	if strings.Contains(errMsg, "connection refused") ||
		strings.Contains(errMsg, "connection reset") ||
		strings.Contains(errMsg, "broken pipe") ||
		strings.Contains(errMsg, "eof") {
		return FailureAction{
			ShouldRotate: true,
			ShouldRetry:  true,
			IsPermanent:  false,
			Reason:       "connection_error",
		}
	}

	if strings.Contains(errMsg, "context canceled") ||
		strings.Contains(errMsg, "context deadline exceeded") {
		return FailureAction{
			ShouldRotate: false,
			ShouldRetry:  false,
			IsPermanent:  false,
			Reason:       "client_cancelled",
		}
	}

	return FailureAction{
		ShouldRotate: false,
		ShouldRetry:  false,
		IsPermanent:  true,
		Reason:       "unknown_error",
	}
}

func classifyProviderError(pe *providers.ProviderError) FailureAction {
	switch pe.Class {
	case providers.ErrorClassRateLimit:
		return FailureAction{
			ShouldRotate: true,
			ShouldRetry:  true,
			IsPermanent:  false,
			Reason:       "rate_limited",
		}

	case providers.ErrorClassQuota:
		return FailureAction{
			ShouldRotate: true,
			ShouldRetry:  false,
			IsPermanent:  false,
			Reason:       "quota_exhausted",
		}

	case providers.ErrorClassTimeout:
		return FailureAction{
			ShouldRotate: true,
			ShouldRetry:  true,
			IsPermanent:  false,
			Reason:       "upstream_timeout",
		}

	case providers.ErrorClassNetwork:
		return FailureAction{
			ShouldRotate: true,
			ShouldRetry:  true,
			IsPermanent:  false,
			Reason:       "upstream_network_error",
		}

	case providers.ErrorClassUpstream5xx:
		return FailureAction{
			ShouldRotate: true,
			ShouldRetry:  true,
			IsPermanent:  false,
			Reason:       "upstream_server_error",
		}

	case providers.ErrorClassAuth:
		return FailureAction{
			ShouldRotate: false,
			ShouldRetry:  false,
			IsPermanent:  true,
			Reason:       "authentication_failed",
		}

	case providers.ErrorClassInvalidRequest:
		return FailureAction{
			ShouldRotate: false,
			ShouldRetry:  false,
			IsPermanent:  true,
			Reason:       "invalid_request",
		}

	default:
		return FailureAction{
			ShouldRotate: false,
			ShouldRetry:  false,
			IsPermanent:  true,
			Reason:       "unclassified_provider_error",
		}
	}
}

func SanitizeError(err error) string {
	if err == nil {
		return ""
	}

	var pe *providers.ProviderError
	if errors.As(err, &pe) {
		return string(pe.Class) + ": " + pe.Message
	}

	msg := err.Error()
	if strings.Contains(msg, "Bearer ") ||
		strings.Contains(msg, "sk-") ||
		strings.Contains(msg, "key_") ||
		strings.Contains(msg, "token=") {
		return "credential_error_redacted"
	}
	return msg
}
