package domain

import (
	"regexp"
	"strings"

	"github.com/nxdir-s/grl/internal/core/entity"
	"github.com/nxdir-s/grl/internal/core/valobj"
)

type Substitutions struct {
	varPattern *regexp.Regexp
}

func NewSubstitutions() *Substitutions {
	return &Substitutions{
		varPattern: regexp.MustCompile(`\{\{\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*\}\}`),
	}
}

func (d *Substitutions) Substitute(str string, vars map[string]string) string {
	if len(vars) == 0 || !strings.Contains(str, "{{") {
		return str
	}

	matches := d.varPattern.FindAllStringSubmatchIndex(str, -1)
	if len(matches) == 0 {
		return str
	}

	var sb strings.Builder
	sb.Grow(len(str))

	last := 0
	for _, m := range matches {
		v, ok := vars[str[m[2]:m[3]]]
		if !ok {
			// leave unknown placeholders untouched
			continue
		}

		sb.WriteString(str[last:m[0]])
		sb.WriteString(v)
		last = m[1]
	}

	sb.WriteString(str[last:])

	return sb.String()
}

func (d *Substitutions) SubstituteRequest(req *entity.Request, vars map[string]string) *entity.Request {
	out := &entity.Request{
		ID:      req.ID,
		Name:    req.Name,
		Method:  req.Method,
		URL:     d.Substitute(req.URL, vars),
		Headers: make([]valobj.Header, 0, len(req.Headers)),
		Params:  make([]valobj.QueryParam, 0, len(req.Params)),
		Body:    d.Substitute(req.Body, vars),
		Auth: valobj.Auth{
			Type:        req.Auth.Type,
			Username:    d.Substitute(req.Auth.Username, vars),
			Password:    d.Substitute(req.Auth.Password, vars),
			Token:       d.Substitute(req.Auth.Token, vars),
			APIKeyName:  d.Substitute(req.Auth.APIKeyName, vars),
			APIKeyValue: d.Substitute(req.Auth.APIKeyValue, vars),
			APIKeyIn:    req.Auth.APIKeyIn,
		},
	}

	for i := range req.Headers {
		out.Headers = append(out.Headers, valobj.Header{
			Key:     req.Headers[i].Key,
			Value:   d.Substitute(req.Headers[i].Value, vars),
			Enabled: req.Headers[i].Enabled,
		})
	}

	for i := range req.Params {
		out.Params = append(out.Params, valobj.QueryParam{
			Key:     req.Params[i].Key,
			Value:   d.Substitute(req.Params[i].Value, vars),
			Enabled: req.Params[i].Enabled,
		})
	}

	return out
}
