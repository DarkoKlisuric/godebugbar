package redact_test

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/DarkoKlisuric/godebugbar/internal/redact"
)

func TestRedactorHeader(t *testing.T) {
	t.Parallel()

	r := redact.New()
	r.AddHeaders("x-tenant")
	in := http.Header{"Authorization": {"Bearer a"}, "X-Tenant": {"t1"}, "Accept": {"*/*"}}
	out := r.Header(in)

	if out.Get("Authorization") != redact.Mask || out.Get("X-Tenant") != redact.Mask || out.Get("Accept") != "*/*" {
		t.Fatalf("unexpected header redaction: %v", out)
	}
	if in.Get("Authorization") != "Bearer a" {
		t.Fatalf("expected the original header to stay untouched")
	}
}

func TestRedactorURLAndQuery(t *testing.T) {
	t.Parallel()

	r := redact.New()
	r.AddQueryParams("Session")

	cases := []struct {
		name string
		in   string
		want []string
		not  []string
	}{
		{name: "token param", in: "https://api.example/x?token=abc&page=2", want: []string{"page=2", "token=%5Bredacted%5D"}, not: []string{"abc"}},
		{name: "custom param case-insensitive", in: "https://api.example/x?SESSION=s1", want: []string{"redacted"}, not: []string{"s1"}},
		{name: "userinfo password", in: (&url.URL{Scheme: "https", User: url.UserPassword("bob", "pw"), Host: "api.example", Path: "/x"}).String(), want: []string{"bob@"}, not: []string{"pw"}},
		{name: "unparseable stays", in: "://bad", want: []string{"://bad"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := r.URL(tc.in)
			for _, want := range tc.want {
				if !strings.Contains(got, want) {
					t.Errorf("URL(%q) = %q, want it to contain %q", tc.in, got, want)
				}
			}
			for _, not := range tc.not {
				if strings.Contains(got, not) {
					t.Errorf("URL(%q) = %q, must not contain %q", tc.in, got, not)
				}
			}
		})
	}

	if q := r.Query(url.Values{"password": {"x"}}); q.Get("password") != redact.Mask {
		t.Fatalf("Query() did not redact password: %v", q)
	}
}

func TestRedactorBody(t *testing.T) {
	t.Parallel()

	r := redact.New()
	r.AddBodyFields("pin")

	cases := []struct {
		name        string
		contentType string
		body        string
		want        []string
		not         []string
	}{
		{name: "nested json", contentType: "application/json", body: `{"user":{"password":"p1","name":"ann"},"items":[{"pin":"1234"}],"n":1.50}`, want: []string{"ann", "1.50", redact.Mask}, not: []string{"p1", "1234"}},
		{name: "json suffix type", contentType: "application/vnd.api+json; charset=utf-8", body: `{"token":"t"}`, want: []string{redact.Mask}, not: []string{`"t"`}},
		{name: "form", contentType: "application/x-www-form-urlencoded", body: "email=a%40b.c&password=p1", want: []string{"email=a%40b.c"}, not: []string{"p1"}},
		{name: "invalid json stays raw", contentType: "application/json", body: `{"password":`, want: []string{`{"password":`}},
		{name: "plain text stays raw", contentType: "text/plain", body: "hello", want: []string{"hello"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := r.Body(tc.contentType, []byte(tc.body))
			for _, want := range tc.want {
				if !strings.Contains(got, want) {
					t.Errorf("Body() = %q, want it to contain %q", got, want)
				}
			}
			for _, not := range tc.not {
				if strings.Contains(got, not) {
					t.Errorf("Body() = %q, must not contain %q", got, not)
				}
			}
		})
	}

	if !redact.IsStructured("application/json") || redact.IsStructured("text/plain") {
		t.Fatalf("IsStructured gave the wrong answer")
	}
}
