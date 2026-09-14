package executor

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var varPattern = regexp.MustCompile(`\{\{\s*([a-zA-Z0-9_-]+)\s*\}\}`)

func InterpolateString(template string, vars map[string]string) string {
	if template == "" || len(vars) == 0 {
		return template
	}
	return varPattern.ReplaceAllStringFunc(template, func(m string) string {
		sub := varPattern.FindStringSubmatch(m)
		if len(sub) > 1 {
			k := sub[1]
			if v, ok := vars[k]; ok {
				return v
			}
		}
		return m
	})
}

func SanitizeHeaderKey(key string) (string, error) {
	k := strings.TrimSpace(key)
	if k == "" {
		return "", errors.New("empty header key")
	}
	if strings.ContainsAny(k, "\r\n:") {
		return "", fmt.Errorf("invalid characters in header key %q", k)
	}
	return k, nil
}

func SanitizeHeaderValue(value string) (string, error) {
	v := strings.TrimSpace(value)
	if strings.ContainsAny(v, "\r\n") {
		return "", errors.New("CRLF injection detected in header value")
	}
	return v, nil
}

func InterpolateHeaders(headers map[string]string, vars map[string]string) (map[string]string, error) {
	res := make(map[string]string, len(headers))
	for k, v := range headers {
		cleanK, err := SanitizeHeaderKey(InterpolateString(k, vars))
		if err != nil {
			return nil, err
		}
		cleanV, err := SanitizeHeaderValue(InterpolateString(v, vars))
		if err != nil {
			return nil, err
		}
		res[cleanK] = cleanV
	}
	return res, nil
}
