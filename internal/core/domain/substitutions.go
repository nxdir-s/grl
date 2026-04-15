package domain

import (
	"regexp"

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
	if len(vars) == 0 {
		return str
	}

	return d.varPattern.ReplaceAllStringFunc(str, func(match string) string {
		m := d.varPattern.FindStringSubmatch(match)
		if len(m) < 2 {
			return match
		}

		if v, ok := vars[m[1]]; ok {
			return v
		}

		return match
	})
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
