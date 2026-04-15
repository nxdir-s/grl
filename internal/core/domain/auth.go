package domain

import (
	"encoding/base64"

	"github.com/nxdir-s/grl/internal/core/entity"
	"github.com/nxdir-s/grl/internal/core/valobj"
)

const (
	AuthHeaderKey string = "Authorization"
	AuthBasic     string = "Basic "
	AuthBearer    string = "Bearer "
)

type Auth struct{}

func NewAuth() *Auth {
	return &Auth{}
}

func (d *Auth) Apply(req *entity.Request) *entity.Request {
	out := *req

	if len(req.Headers) > 0 {
		out.Headers = make([]valobj.Header, 0, len(req.Headers))
		out.Headers = append(out.Headers, req.Headers...)
	}

	if len(req.Params) > 0 {
		out.Params = make([]valobj.QueryParam, 0, len(req.Params))
		out.Params = append(out.Params, req.Params...)
	}

	switch req.Auth.Type {
	case valobj.AuthBasic:
		if len(req.Auth.Username) == 0 && len(req.Auth.Password) == 0 {
			return &out
		}

		raw := req.Auth.Username + ":" + req.Auth.Password
		encoded := base64.StdEncoding.EncodeToString([]byte(raw))

		out.Headers = append(out.Headers, valobj.Header{
			Key:     AuthHeaderKey,
			Value:   AuthBasic + encoded,
			Enabled: true,
		})
	case valobj.AuthBearer:
		if len(req.Auth.Token) == 0 {
			return &out
		}

		out.Headers = append(out.Headers, valobj.Header{
			Key:     AuthHeaderKey,
			Value:   AuthBearer + req.Auth.Token,
			Enabled: true,
		})
	case valobj.AuthAPIKey:
		if len(req.Auth.APIKeyName) == 0 {
			return &out
		}

		switch {
		case req.Auth.APIKeyIn == valobj.APIKeyInQuery:
			out.Params = append(out.Params, valobj.QueryParam{
				Key:     req.Auth.APIKeyName,
				Value:   req.Auth.APIKeyValue,
				Enabled: true,
			})
		default:
			out.Headers = append(out.Headers, valobj.Header{
				Key:     req.Auth.APIKeyName,
				Value:   req.Auth.APIKeyValue,
				Enabled: true,
			})
		}
	}

	return &out
}
