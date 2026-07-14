package domain

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nxdir-s/grl/internal/core/entity"
	"github.com/nxdir-s/grl/internal/core/valobj"
)

func TestSubstitute(t *testing.T) {
	subs := NewSubstitutions()
	vars := map[string]string{
		"host":  "api.example.com",
		"token": "secret",
		"a":     "1",
		"b":     "2",
		"_x":    "underscore",
	}

	cases := []struct {
		name     string
		input    string
		vars     map[string]string
		expected string
	}{
		{"basic", "https://{{host}}/users", vars, "https://api.example.com/users"},
		{"multiple", "{{a}} and {{b}}", vars, "1 and 2"},
		{"adjacent", "{{a}}{{b}}", vars, "12"},
		{"whitespace", "{{  host  }}", vars, "api.example.com"},
		{"unknown_var_untouched", "https://{{missing}}/x", vars, "https://{{missing}}/x"},
		{"mixed_known_unknown", "{{a}}-{{missing}}-{{b}}", vars, "1-{{missing}}-2"},
		{"underscore_name", "{{_x}}", vars, "underscore"},
		{"invalid_name_untouched", "{{1x}} {{a-b}}", vars, "{{1x}} {{a-b}}"},
		{"unclosed_untouched", "{{a", vars, "{{a"},
		{"no_placeholders", "plain text", vars, "plain text"},
		{"empty_string", "", vars, ""},
		{"nil_vars", "{{a}}", nil, "{{a}}"},
		{"empty_vars", "{{a}}", map[string]string{}, "{{a}}"},
		{"value_not_rescanned", "{{nested}}", map[string]string{"nested": "{{a}}"}, "{{a}}"},
		{"empty_value", "x{{empty}}y", map[string]string{"empty": ""}, "xy"},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, subs.Substitute(tt.input, tt.vars))
		})
	}
}

func TestSubstituteRequest(t *testing.T) {
	subs := NewSubstitutions()
	vars := map[string]string{"host": "api.example.com", "token": "secret"}

	req := &entity.Request{
		ID:     "req-1",
		Name:   "test",
		Method: valobj.MethodPost,
		URL:    "https://{{host}}/users",
		Headers: []valobj.Header{
			{Key: "Authorization", Value: "Bearer {{token}}", Enabled: true},
		},
		Params: []valobj.QueryParam{
			{Key: "q", Value: "{{token}}", Enabled: true},
		},
		Body: `{"host":"{{host}}"}`,
		Auth: valobj.Auth{Token: "{{token}}"},
	}

	out := subs.SubstituteRequest(req, vars)

	assert.Equal(t, "https://api.example.com/users", out.URL)
	assert.Equal(t, "Bearer secret", out.Headers[0].Value)
	assert.Equal(t, "secret", out.Params[0].Value)
	assert.Equal(t, `{"host":"api.example.com"}`, out.Body)
	assert.Equal(t, "secret", out.Auth.Token)

	assert.Equal(t, "https://{{host}}/users", req.URL, "input request must not be mutated")
}

func benchVars() map[string]string {
	return map[string]string{
		"host":    "api.example.com",
		"version": "v2",
		"token":   "abc123def456",
		"user":    "tester",
		"id":      "42",
	}
}

func BenchmarkSubstitute(b *testing.B) {
	subs := NewSubstitutions()
	vars := benchVars()

	cases := []struct {
		name  string
		input string
		vars  map[string]string
	}{
		{"url_5vars", "https://{{host}}/api/{{version}}/users/{{id}}?token={{token}}&user={{user}}", vars},
		{"body_10KB_with_vars", makeJSONWithVars(BenchMediumJSON), vars},
		{"body_10KB_no_placeholders", makeJSON(BenchMediumJSON), vars},
		{"body_10KB_no_vars", makeJSONWithVars(BenchMediumJSON), nil},
	}

	for _, bc := range cases {
		b.Run(bc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(bc.input)))

			for b.Loop() {
				_ = subs.Substitute(bc.input, bc.vars)
			}
		})
	}
}

func BenchmarkSubstituteRequest(b *testing.B) {
	subs := NewSubstitutions()
	vars := benchVars()

	headers := make([]valobj.Header, 0, 10)
	for i := 0; i < 10; i++ {
		headers = append(headers, valobj.Header{
			Key:     fmt.Sprintf("X-Header-%d", i),
			Value:   "value-{{token}}",
			Enabled: true,
		})
	}

	params := make([]valobj.QueryParam, 0, 5)
	for i := 0; i < 5; i++ {
		params = append(params, valobj.QueryParam{
			Key:     fmt.Sprintf("param%d", i),
			Value:   "{{id}}",
			Enabled: true,
		})
	}

	req := &entity.Request{
		ID:      "bench",
		Name:    "bench request",
		Method:  valobj.MethodPost,
		URL:     "https://{{host}}/api/{{version}}/users/{{id}}",
		Headers: headers,
		Params:  params,
		Body:    makeJSONWithVars(1024),
		Auth: valobj.Auth{
			Token: "{{token}}",
		},
	}

	b.ReportAllocs()

	for b.Loop() {
		_ = subs.SubstituteRequest(req, vars)
	}
}
