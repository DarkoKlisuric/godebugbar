package redact

import (
	"bytes"
	"encoding/json"
	"mime"
	"net/http"
	"net/url"
	"strings"
)

const (
	Mask          = "[redacted]"
	formMediaType = "application/x-www-form-urlencoded"
)

var (
	DefaultHeaders = []string{
		"Authorization",
		"Proxy-Authorization",
		"Cookie",
		"Set-Cookie",
		"X-Api-Key",
		"X-Auth-Token",
		"X-Csrf-Token",
		"X-Xsrf-Token",
	}
	DefaultQueryParams = []string{
		"token",
		"access_token",
		"refresh_token",
		"id_token",
		"password",
		"secret",
		"client_secret",
		"api_key",
		"apikey",
		"signature",
	}
	DefaultBodyFields = []string{
		"password",
		"password_confirmation",
		"current_password",
		"new_password",
		"token",
		"access_token",
		"refresh_token",
		"secret",
		"client_secret",
		"api_key",
	}
)

type Redactor struct {
	headers     map[string]bool
	queryParams map[string]bool
	bodyFields  map[string]bool
}

func New() *Redactor {
	r := &Redactor{
		headers:     make(map[string]bool),
		queryParams: make(map[string]bool),
		bodyFields:  make(map[string]bool),
	}
	r.AddHeaders(DefaultHeaders...)
	r.AddQueryParams(DefaultQueryParams...)
	r.AddBodyFields(DefaultBodyFields...)

	return r
}

func (r *Redactor) AddHeaders(names ...string) {
	for _, name := range names {
		if name = strings.TrimSpace(name); name != "" {
			r.headers[http.CanonicalHeaderKey(name)] = true
		}
	}
}

func (r *Redactor) AddQueryParams(names ...string) {
	addLower(r.queryParams, names)
}

func (r *Redactor) AddBodyFields(names ...string) {
	addLower(r.bodyFields, names)
}

func (r *Redactor) Header(h http.Header) http.Header {
	out := h.Clone()
	if out == nil {
		return http.Header{}
	}
	for name, values := range out {
		if r.headers[http.CanonicalHeaderKey(name)] {
			out[name] = maskAll(values)
		}
	}

	return out
}

func (r *Redactor) Query(values url.Values) url.Values {
	out := make(url.Values, len(values))
	for key, vals := range values {
		if r.queryParams[strings.ToLower(key)] {
			out[key] = maskAll(vals)
			continue
		}
		out[key] = append([]string(nil), vals...)
	}

	return out
}

func (r *Redactor) URL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	if u.User != nil {
		u.User = url.User(u.User.Username())
	}
	if u.RawQuery != "" {
		u.RawQuery = r.Query(u.Query()).Encode()
	}

	return u.String()
}

func IsStructured(contentType string) bool {
	mediaType, _, _ := mime.ParseMediaType(contentType)

	return isJSON(mediaType) || mediaType == formMediaType
}

func (r *Redactor) Body(contentType string, body []byte) string {
	mediaType, _, _ := mime.ParseMediaType(contentType)
	switch {
	case isJSON(mediaType):
		return r.jsonBody(body)
	case mediaType == formMediaType:
		values, err := url.ParseQuery(string(body))
		if err != nil {
			return string(body)
		}
		return r.formBody(values)
	default:
		return string(body)
	}
}

func (r *Redactor) jsonBody(body []byte) string {
	var decoded any
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	if err := decoder.Decode(&decoded); err != nil {
		return string(body)
	}

	out, err := json.MarshalIndent(r.walk(decoded), "", "  ")
	if err != nil {
		return string(body)
	}

	return string(out)
}

func (r *Redactor) formBody(values url.Values) string {
	for key, vals := range values {
		if r.bodyFields[strings.ToLower(key)] || r.queryParams[strings.ToLower(key)] {
			values[key] = maskAll(vals)
		}
	}

	return values.Encode()
}

func (r *Redactor) walk(v any) any {
	switch typed := v.(type) {
	case map[string]any:
		for key, value := range typed {
			if r.bodyFields[strings.ToLower(key)] {
				typed[key] = Mask
				continue
			}
			typed[key] = r.walk(value)
		}
		return typed
	case []any:
		for i, value := range typed {
			typed[i] = r.walk(value)
		}
		return typed
	default:
		return v
	}
}

func addLower(set map[string]bool, names []string) {
	for _, name := range names {
		if name = strings.ToLower(strings.TrimSpace(name)); name != "" {
			set[name] = true
		}
	}
}

func maskAll(values []string) []string {
	out := make([]string, len(values))
	for i := range out {
		out[i] = Mask
	}

	return out
}

func isJSON(mediaType string) bool {
	return mediaType == "application/json" || strings.HasSuffix(mediaType, "+json")
}
