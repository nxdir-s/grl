package domain

import (
	"context"
	"net/url"

	"github.com/nxdir-s/grl/internal/core/entity"
	"github.com/nxdir-s/grl/internal/core/valobj"
	"github.com/nxdir-s/grl/internal/ports"
)

type ErrMissingURL struct{}

func (e *ErrMissingURL) Error() string {
	return "URL is required"
}

type ErrInvalidMethod struct{}

func (e *ErrInvalidMethod) Error() string {
	return "invalid HTTP method"
}

type ErrInvalidURL struct {
	err error
}

func (e *ErrInvalidURL) Error() string {
	return "invalid URL: " + e.err.Error()
}

type Requests struct {
	service       ports.RequestService
	environments  ports.Environments
	substitutions ports.Substitutions

	validMethods []valobj.HTTPMethod
}

func NewRequests(service ports.RequestService, environments ports.Environments, substitutions ports.Substitutions) *Requests {
	return &Requests{
		service:       service,
		environments:  environments,
		substitutions: substitutions,
		validMethods:  validMethods(),
	}
}

func (d *Requests) Send(ctx context.Context, req *entity.Request) (*entity.Response, error) {
	req = d.substitutions.SubstituteRequest(req, d.environments.ActiveVars(ctx))

	if err := d.Validate(req); err != nil {
		return nil, err
	}

	reqURL, err := d.BuildURL(req.URL, req.Params)
	if err != nil {
		return nil, err
	}

	built := *req
	built.URL = reqURL

	resp, err := d.service.Send(ctx, &built)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (d *Requests) Validate(req *entity.Request) error {
	if len(req.URL) == 0 {
		return &ErrMissingURL{}
	}

	if !d.validMethod(req.Method) {
		return &ErrInvalidMethod{}
	}

	if _, err := url.ParseRequestURI(req.URL); err != nil {
		return &ErrInvalidURL{err}
	}

	return nil
}

func (d *Requests) BuildURL(baseURL string, params []valobj.QueryParam) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", &ErrInvalidURL{err}
	}

	q := u.Query()

	for i := range params {
		if params[i].Enabled && len(params[i].Key) != 0 {
			q.Add(params[i].Key, params[i].Value)
		}
	}

	u.RawQuery = q.Encode()

	return u.String(), nil
}

func (d *Requests) validMethod(method valobj.HTTPMethod) bool {
	for i := range d.validMethods {
		if method == d.validMethods[i] {
			return true
		}
	}

	return false
}

func validMethods() []valobj.HTTPMethod {
	return []valobj.HTTPMethod{
		valobj.MethodGet,
		valobj.MethodPost,
		valobj.MethodPut,
		valobj.MethodDelete,
		valobj.MethodPatch,
		valobj.MethodHead,
		valobj.MethodOptions,
	}
}
