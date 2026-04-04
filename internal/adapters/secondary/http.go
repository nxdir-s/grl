package secondary

import (
	"bytes"
	"context"
	"crypto/tls"
	"io"
	"math"
	"net"
	"net/http"
	"net/http/httptrace"
	"strings"
	"time"

	"github.com/nxdir-s/grl/internal/core/entity"
	"github.com/nxdir-s/grl/internal/core/valobj"
)

const (
	_ int64 = 1 << (10 * iota)
	Kib
	Mib
	Gib
)

const (
	DefaultTimeout         int   = 10
	DefaultMaxIdleConns    int   = 100
	DefaultMaxConnsPerHost int   = 100
	DefaultReadByteLimit   int64 = 15 * Mib
	DefaultRetryLimit      int   = 3

	ContentTypeHeader string = "Content-Type"
)

type ErrHttpCopy struct {
	err error
}

func (e *ErrHttpCopy) Error() string {
	return "failed to copy request body: " + e.err.Error()
}

type ErrHttpStatusCode struct {
	code int
	msg  *bytes.Buffer
}

func (e *ErrHttpStatusCode) Error() string {
	return "recieved bad status code: " + e.msg.String()
}

type ErrReadReqBody struct {
	err error
}

func (e *ErrReadReqBody) Error() string {
	return "failed to read request body: " + e.err.Error()
}

type ErrReadRespBody struct {
	err error
}

func (e *ErrReadRespBody) Error() string {
	return "failed to read response body: " + e.err.Error()
}

type HttpOpt func(a *HttpAdapter)

func WithHttpClient(client *http.Client) HttpOpt {
	return func(a *HttpAdapter) {
		a.http = client
	}
}

type HttpConfig struct {
	TlsConfig             *tls.Config
	RetryEnabled          bool
	Timeout               int
	RetryLimit            int
	MaxIdleConnections    int
	MaxConnectionsPerHost int
	ReadByteLimit         int64
}

type HttpAdapter struct {
	http  *http.Client
	limit int64
}

func NewHttpAdapter(cfg *HttpConfig, opts ...HttpOpt) *HttpAdapter {
	var timeout int = DefaultTimeout * int(time.Second)
	if cfg.Timeout != 0 {
		timeout = cfg.Timeout
	}

	var byteLimit int64 = DefaultReadByteLimit
	if cfg.ReadByteLimit != 0 {
		byteLimit = cfg.ReadByteLimit
	}

	var maxIdleConns int = DefaultMaxIdleConns
	if cfg.MaxIdleConnections != 0 {
		maxIdleConns = cfg.MaxIdleConnections
	}

	var maxConnsPerHost int = DefaultMaxConnsPerHost
	if cfg.MaxConnectionsPerHost != 0 {
		maxConnsPerHost = cfg.MaxConnectionsPerHost
	}

	defaultTransport := &http.Transport{
		Dial: (&net.Dialer{
			Timeout: time.Duration(timeout),
		}).Dial,
		TLSClientConfig:     cfg.TlsConfig,
		MaxIdleConns:        maxIdleConns,
		MaxConnsPerHost:     maxConnsPerHost,
		MaxIdleConnsPerHost: maxConnsPerHost,
		IdleConnTimeout:     time.Duration(timeout),
		TLSHandshakeTimeout: time.Duration(timeout),
	}

	var transport http.RoundTripper = defaultTransport

	if cfg.RetryEnabled {
		transport = NewRetryTransport(defaultTransport, cfg.RetryLimit)
	}

	adapter := &HttpAdapter{
		http: &http.Client{
			Timeout:   time.Duration(timeout),
			Transport: transport,
		},
		limit: byteLimit,
	}

	for _, opt := range opts {
		opt(adapter)
	}

	return adapter
}

func (a *HttpAdapter) Send(ctx context.Context, req *entity.Request) (*entity.Response, error) {
	var bodyReader io.Reader
	if len(req.Body) != 0 {
		bodyReader = strings.NewReader(req.Body)
	}

	httpReq, err := http.NewRequestWithContext(ctx, req.Method.String(), req.URL, bodyReader)
	if err != nil {
		return nil, err
	}

	for _, h := range req.Headers {
		if h.Enabled && len(h.Key) != 0 {
			httpReq.Header.Set(h.Key, h.Value)
		}
	}

	var timing valobj.Timing
	var dnsStart, connectStart, tlsStart, gotConn time.Time
	start := time.Now()

	trace := &httptrace.ClientTrace{
		DNSStart: func(_ httptrace.DNSStartInfo) {
			dnsStart = time.Now()
		},
		DNSDone: func(_ httptrace.DNSDoneInfo) {
			timing.DNSLookup = time.Since(dnsStart)
		},
		ConnectStart: func(_, _ string) {
			connectStart = time.Now()
		},
		ConnectDone: func(_, _ string, _ error) {
			timing.TCPConnect = time.Since(connectStart)
		},
		TLSHandshakeStart: func() {
			tlsStart = time.Now()
		},
		TLSHandshakeDone: func(_ tls.ConnectionState, _ error) {
			timing.TLSHandshake = time.Since(tlsStart)
		},
		GotConn: func(_ httptrace.GotConnInfo) {
			gotConn = time.Now()
		},
		GotFirstResponseByte: func() {
			if !gotConn.IsZero() {
				timing.TTFB = time.Since(gotConn)
			}
		},
	}

	httpReq = httpReq.WithContext(httptrace.WithClientTrace(httpReq.Context(), trace))

	httpResp, err := a.http.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode/10 != 20 {
		errBody := &bytes.Buffer{}

		if _, err := io.Copy(errBody, io.LimitReader(httpResp.Body, a.limit)); err != nil {
			return nil, &ErrReadRespBody{err}
		}

		return nil, &ErrHttpStatusCode{httpResp.StatusCode, errBody}
	}

	var body *bytes.Buffer
	if _, err := io.Copy(body, io.LimitReader(httpResp.Body, a.limit)); err != nil {
		return nil, &ErrReadRespBody{err}
	}

	timing.Total = time.Since(start)

	headers := make([]valobj.Header, 0)
	for key, values := range httpResp.Header {
		for _, val := range values {
			headers = append(headers, valobj.Header{
				Key:     key,
				Value:   val,
				Enabled: true,
			})
		}
	}

	return &entity.Response{
		StatusCode:    httpResp.StatusCode,
		Status:        httpResp.Status,
		Headers:       headers,
		Body:          body,
		ContentType:   httpResp.Header.Get(ContentTypeHeader),
		ContentLength: httpResp.ContentLength,
		Timing:        timing,
	}, nil
}

type RetryTransport struct {
	transport http.RoundTripper
	retryMax  int
}

// NewRetryTransport wraps the supplied http transport with a retryable implementation
func NewRetryTransport(transport *http.Transport, limit int) *RetryTransport {
	var retryLimit int = DefaultRetryLimit
	if limit != 0 {
		retryLimit = limit
	}

	return &RetryTransport{
		transport: transport,
		retryMax:  retryLimit,
	}
}

// RoundTrip implements the http.RoundTripper interface with retries
func (t *RetryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var bodyBytes []byte
	var err error

	if req.Body != nil {
		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, &ErrHttpCopy{err}
		}

		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	resp, err := t.transport.RoundTrip(req)

	retries := 0
	for shouldRetry(resp, err) && retries < t.retryMax {
		time.Sleep(backoff(retries))

		if resp.Body != nil {
			drainBody(resp.Body)
		}

		if req.Body != nil {
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		resp, err = t.transport.RoundTrip(req)

		retries++
	}

	return resp, err
}

func drainBody(body io.ReadCloser) error {
	defer body.Close()

	if _, err := io.ReadAll(body); err != nil {
		return err
	}

	return nil
}

// shouldRetry checks for errors and non 2XX status codes to determine whether to retry
func shouldRetry(resp *http.Response, err error) bool {
	if err != nil {
		return true
	}

	if resp.StatusCode/10 != 20 {
		return true
	}

	return false
}

// backoff doubles the delay
func backoff(retries int) time.Duration {
	return time.Duration(math.Pow(2, float64(retries))) * time.Second
}
