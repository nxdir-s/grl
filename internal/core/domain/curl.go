package domain

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/nxdir-s/grl/internal/core/entity"
	"github.com/nxdir-s/grl/internal/core/valobj"
)

type CurlCodec struct{}

func NewCurlCodec() *CurlCodec {
	return &CurlCodec{}
}

func (d *CurlCodec) Encode(req *entity.Request) string {
	var b strings.Builder
	b.WriteString("curl")

	method := req.Method.String()
	if len(method) != 0 && method != valobj.MethodGet.String() {
		b.WriteString(" -X ")
		b.WriteString(method)
	}

	for i := range req.Headers {
		if !req.Headers[i].Enabled || len(req.Headers[i].Key) == 0 {
			continue
		}

		b.WriteString(" \\\n  -H ")
		b.WriteString(d.shellQuote(req.Headers[i].Key + ": " + req.Headers[i].Value))
	}

	switch req.BodyType {
	case valobj.BodyTypeFormURL:
		for i := range req.FormFields {
			if !req.FormFields[i].Enabled || len(req.FormFields[i].Key) == 0 {
				continue
			}

			b.WriteString(" \\\n  --data-urlencode ")
			b.WriteString(d.shellQuote(req.FormFields[i].Key + "=" + req.FormFields[i].Value))
		}
	case valobj.BodyTypeFormData:
		for i := range req.FormFields {
			if !req.FormFields[i].Enabled || len(req.FormFields[i].Key) == 0 {
				continue
			}

			b.WriteString(" \\\n  --form ")
			b.WriteString(d.shellQuote(req.FormFields[i].Key + "=" + req.FormFields[i].Value))
		}
	default:
		if len(req.Body) != 0 {
			b.WriteString(" \\\n  -d ")
			b.WriteString(d.shellQuote(req.Body))
		}
	}

	finalURL := req.URL

	enabled := d.enabledParams(req.Params)
	if len(enabled) > 0 {
		qs := d.encodeParams(enabled)

		switch {
		case strings.Contains(finalURL, "?"):
			finalURL += "&" + qs
		default:
			finalURL += "?" + qs
		}
	}

	b.WriteString(" \\\n  ")
	b.WriteString(d.shellQuote(finalURL))

	return b.String()
}

func (d *CurlCodec) enabledParams(params []valobj.QueryParam) []valobj.QueryParam {
	out := make([]valobj.QueryParam, 0)

	for i := range params {
		if params[i].Enabled && len(params[i].Key) != 0 {
			out = append(out, params[i])
		}
	}

	return out
}

func (d *CurlCodec) encodeParams(params []valobj.QueryParam) string {
	v := url.Values{}

	for i := range params {
		v.Add(params[i].Key, params[i].Value)
	}

	return v.Encode()
}

func (d *CurlCodec) shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func (d *CurlCodec) Decode(s string) (*entity.Request, error) {
	s = stripLineContinuations(s)

	tokens, err := tokenize(s)
	if err != nil {
		return nil, err
	}

	if len(tokens) == 0 {
		return nil, errors.New("empty curl command")
	}

	if tokens[0] == "curl" {
		tokens = tokens[1:]
	}

	req := &entity.Request{
		Method: valobj.MethodGet,
	}

	for i := 0; i < len(tokens); i++ {
		tok := tokens[i]
		if tok == "" {
			continue
		}

		// Handle --flag=value form
		flagName, flagVal, hasEq := d.splitFlagValue(tok)

		next := func() (string, bool) {
			if hasEq {
				return flagVal, true
			}

			if i+1 >= len(tokens) {
				return "", false
			}

			i++

			return tokens[i], true
		}

		switch flagName {
		case "-X", "--request":
			if v, ok := next(); ok {
				req.Method = valobj.HTTPMethod(strings.ToUpper(v))
			}
		case "-H", "--header":
			if v, ok := next(); ok {
				if k, val, found := strings.Cut(v, ":"); found {
					req.Headers = append(req.Headers, valobj.Header{
						Key:     strings.TrimSpace(k),
						Value:   strings.TrimSpace(val),
						Enabled: true,
					})
				}
			}
		case "-d", "--data", "--data-raw", "--data-binary", "--data-ascii":
			if v, ok := next(); ok {
				req.Body = v
				req.BodyType = valobj.BodyTypeRaw

				if req.Method == valobj.MethodGet {
					req.Method = valobj.MethodPost
				}
			}
		case "--data-urlencode":
			if v, ok := next(); ok {
				k, val, _ := strings.Cut(v, "=")
				req.FormFields = append(req.FormFields, valobj.FormField{
					Key:     k,
					Value:   val,
					Enabled: true,
				})
				req.BodyType = valobj.BodyTypeFormURL

				if req.Method == valobj.MethodGet {
					req.Method = valobj.MethodPost
				}
			}
		case "-F", "--form", "--form-string":
			if v, ok := next(); ok {
				k, val, _ := strings.Cut(v, "=")
				req.FormFields = append(req.FormFields, valobj.FormField{
					Key:     k,
					Value:   val,
					Enabled: true,
				})
				req.BodyType = valobj.BodyTypeFormData

				if req.Method == valobj.MethodGet {
					req.Method = valobj.MethodPost
				}
			}
		case "-u", "--user":
			if v, ok := next(); ok {
				user, pass, _ := strings.Cut(v, ":")

				req.Auth = valobj.Auth{
					Type:     valobj.AuthBasic,
					Username: user,
					Password: pass,
				}
			}
		case "-A", "--user-agent":
			if v, ok := next(); ok {
				req.Headers = append(req.Headers, valobj.Header{
					Key: "User-Agent", Value: v, Enabled: true,
				})
			}
		case "-b", "--cookie":
			if v, ok := next(); ok {
				req.Headers = append(req.Headers, valobj.Header{
					Key: "Cookie", Value: v, Enabled: true,
				})
			}
		case "--url":
			if v, ok := next(); ok {
				req.URL = v
			}
		default:
			// Flag with value we don't care about — skip its argument if any
			if strings.HasPrefix(flagName, "-") && !hasEq {
				// Some flags take arguments (--max-time, --connect-timeout, etc.)
				// if next token isn't a flag, consume it.
				if i+1 < len(tokens) && !strings.HasPrefix(tokens[i+1], "-") {
					i++
				}

				continue
			}

			// Not a flag → treat as the URL if not set
			if len(req.URL) == 0 {
				req.URL = tok
			}
		}
	}

	if len(req.URL) == 0 {
		return nil, errors.New("no URL found in curl command")
	}

	d.promoteBearerAuth(req)
	d.extractQueryParams(req)

	return req, nil
}

func (d *CurlCodec) promoteBearerAuth(req *entity.Request) {
	if len(req.Auth.Type) != 0 && req.Auth.Type != valobj.AuthNone {
		return
	}

	for i, h := range req.Headers {
		if !strings.EqualFold(h.Key, "Authorization") {
			continue
		}

		fields := strings.SplitN(h.Value, " ", 2)
		if len(fields) == 2 && strings.EqualFold(fields[0], "Bearer") {
			req.Auth = valobj.Auth{
				Type:  valobj.AuthBearer,
				Token: fields[1],
			}

			req.Headers = append(req.Headers[:i], req.Headers[i+1:]...)

			return
		}
	}
}

func (d *CurlCodec) extractQueryParams(req *entity.Request) {
	u, err := url.Parse(req.URL)
	if err != nil || u.RawQuery == "" {
		return
	}

	for k, vs := range u.Query() {
		for _, v := range vs {
			req.Params = append(req.Params, valobj.QueryParam{
				Key: k, Value: v, Enabled: true,
			})
		}
	}

	u.RawQuery = ""
	req.URL = u.String()
}

func (d *CurlCodec) splitFlagValue(tok string) (flag string, value string, hasEq bool) {
	if strings.HasPrefix(tok, "--") {
		if i := strings.Index(tok, "="); i > 0 {
			return tok[:i], tok[i+1:], true
		}
	}

	return tok, "", false
}

// stripLineContinuations removes "\\\n" and "\\\r\n" sequences.
func stripLineContinuations(s string) string {
	s = strings.ReplaceAll(s, "\\\r\n", " ")
	s = strings.ReplaceAll(s, "\\\n", " ")

	return s
}

// tokenize splits a shell-like string honoring single and double quotes.
// Inside single quotes everything is literal. Inside double quotes,
// backslash escapes " \ $ and newline.
func tokenize(s string) ([]string, error) {
	var (
		tokens  []string
		cur     strings.Builder
		inSQ    bool
		inDQ    bool
		pending bool
	)

	flush := func() {
		if pending || cur.Len() > 0 {
			tokens = append(tokens, cur.String())
			cur.Reset()
			pending = false
		}
	}

	for i := 0; i < len(s); i++ {
		ch := s[i]
		switch {
		case inSQ:
			if ch == '\'' {
				inSQ = false
				pending = true
				continue
			}

			cur.WriteByte(ch)
		case inDQ:
			if ch == '\\' && i+1 < len(s) {
				n := s[i+1]
				if n == '"' || n == '\\' || n == '$' || n == '`' || n == '\n' {
					cur.WriteByte(n)
					i++
					continue
				}
			}

			if ch == '"' {
				inDQ = false
				pending = true
				continue
			}

			cur.WriteByte(ch)
		default:
			if ch == '\'' {
				inSQ = true
				pending = true
				continue
			}

			if ch == '"' {
				inDQ = true
				pending = true
				continue
			}

			if ch == '\\' && i+1 < len(s) {
				cur.WriteByte(s[i+1])
				i++
				continue
			}

			if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
				flush()
				continue
			}

			cur.WriteByte(ch)
		}
	}

	if inSQ || inDQ {
		return nil, fmt.Errorf("unterminated quote")
	}

	flush()

	return tokens, nil
}
