package secondary

import (
	"bytes"
	"context"
	"crypto/tls"
	_ "embed"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/nxdir-s/grl/internal/core/entity"
	"github.com/nxdir-s/grl/internal/core/valobj"
	"github.com/stretchr/testify/assert"
)

const (
	BenchContentTypeJSON string = "application/json"
	BenchNumFormFields   int    = 10
	BenchRawBodySize     int64  = 1 * Kib
)

// newBenchServer returns a test server that always responds with http.StatusOK and payload
func newBenchServer(payload []byte) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)

		w.Header().Set(ContentTypeHeader, BenchContentTypeJSON)
		w.WriteHeader(http.StatusOK)
		w.Write(payload)
	}))
}

// benchRequest builds a request against url for the given body type
func benchRequest(url string, bodyType valobj.BodyType) *entity.Request {
	req := &entity.Request{
		Method:   valobj.MethodPost,
		URL:      url,
		BodyType: bodyType,
	}

	switch bodyType {
	case valobj.BodyTypeFormURL, valobj.BodyTypeFormData:
		req.FormFields = make([]valobj.FormField, 0, BenchNumFormFields)

		for i := 0; i < BenchNumFormFields; i++ {
			req.FormFields = append(req.FormFields, valobj.FormField{
				Key:     "field" + strconv.Itoa(i),
				Value:   "value" + strconv.Itoa(i),
				Enabled: true,
			})
		}
	default:
		req.Body = string(makeJSONPayload(BenchRawBodySize))
		req.Headers = []valobj.Header{
			{Key: ContentTypeHeader, Value: BenchContentTypeJSON, Enabled: true},
		}
	}

	return req
}

func BenchmarkHttpAdapterSend(b *testing.B) {
	cases := []struct {
		name     string
		bodyType valobj.BodyType
		respSize int64
	}{
		{"raw/small", valobj.BodyTypeRaw, BenchSmallBody},
		{"raw/medium", valobj.BodyTypeRaw, BenchMediumBody},
		{"raw/large", valobj.BodyTypeRaw, BenchLargeBody},
		{"form-url/small", valobj.BodyTypeFormURL, BenchSmallBody},
		{"form-data/small", valobj.BodyTypeFormData, BenchSmallBody},
	}

	for _, bm := range cases {
		b.Run(bm.name, func(b *testing.B) {
			ts := newBenchServer(makeJSONPayload(bm.respSize))
			defer ts.Close()

			adapter := NewHttpAdapter(&HttpConfig{}, benchLogger())
			req := benchRequest(ts.URL, bm.bodyType)

			b.ReportAllocs()

			if bm.bodyType == valobj.BodyTypeRaw {
				b.SetBytes(bm.respSize)
			}

			for b.Loop() {
				resp, err := adapter.Send(b.Context(), req)
				if err != nil {
					b.Fatal(err)
				}

				if resp.StatusCode != http.StatusOK {
					b.Fatalf("unexpected status code: %d", resp.StatusCode)
				}
			}
		})
	}
}

func BenchmarkHttpAdapterSendParallel(b *testing.B) {
	ts := newBenchServer(makeJSONPayload(BenchMediumBody))
	defer ts.Close()

	adapter := NewHttpAdapter(&HttpConfig{}, benchLogger())

	b.ReportAllocs()
	b.SetBytes(BenchMediumBody)

	b.RunParallel(func(pb *testing.PB) {
		// encodeBody writes req.ContentType, so each goroutine needs its own request
		req := benchRequest(ts.URL, valobj.BodyTypeRaw)

		for pb.Next() {
			resp, err := adapter.Send(b.Context(), req)
			if err != nil {
				b.Fatal(err)
			}

			if resp.StatusCode != http.StatusOK {
				b.Fatalf("unexpected status code: %d", resp.StatusCode)
			}
		}
	})
}

func BenchmarkHttpAdapterEncodeBody(b *testing.B) {
	fileReq := &entity.Request{
		Method:   valobj.MethodPost,
		BodyType: valobj.BodyTypeFormData,
		FormFields: []valobj.FormField{
			{Key: "file", Value: "@testdata/response.json", Enabled: true},
		},
	}

	cases := []struct {
		name string
		req  *entity.Request
	}{
		{"raw", benchRequest(TestHost+TestEndpoint, valobj.BodyTypeRaw)},
		{"form-url", benchRequest(TestHost+TestEndpoint, valobj.BodyTypeFormURL)},
		{"form-data", benchRequest(TestHost+TestEndpoint, valobj.BodyTypeFormData)},
		{"form-data-file", fileReq},
	}

	adapter := NewHttpAdapter(&HttpConfig{}, benchLogger())

	for _, bm := range cases {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()

			for b.Loop() {
				body, err := adapter.encodeBody(bm.req)
				if err != nil {
					b.Fatal(err)
				}

				if _, err := io.Copy(io.Discard, body); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

//go:embed testdata/response.json
var testData []byte

const (
	TestHost                  string = "http://example.com"
	TestEndpoint              string = "/api"
	TestResponse              string = "{}"
	TestRateLimit             int    = 60
	TestErrMsg                string = "test error"
	TestPath                  string = "/home/document.pdf"
	TestTimeout               int    = 30
	TestRetryLimit            int    = 3
	TestMaxIdleConnections    int    = 10
	TestMaxConnectionsPerHost int    = 10
	TestReadByteLimit         int64  = 15 * Mib
)

type ErrTest struct{}

func (e *ErrTest) Error() string {
	return TestErrMsg
}

func TestSend(t *testing.T) {
	cases := []struct {
		cfg          *HttpConfig
		opts         []HttpOpt
		req          *entity.Request
		handler      func(w http.ResponseWriter, r *http.Request)
		expectedCode int
		expectedErr  error
	}{
		{
			opts: []HttpOpt{},
			cfg: &HttpConfig{
				TlsConfig: &tls.Config{
					MinVersion: tls.VersionTLS12,
				},
				RetryEnabled:          true,
				Timeout:               TestTimeout,
				RetryLimit:            TestRetryLimit,
				MaxIdleConnections:    TestMaxConnectionsPerHost,
				MaxConnectionsPerHost: TestMaxConnectionsPerHost,
				ReadByteLimit:         TestReadByteLimit,
			},
			req: &entity.Request{
				Method: valobj.MethodGet,
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write(testData)
			},
			expectedCode: http.StatusOK,
			expectedErr:  nil,
		},
		{
			opts: []HttpOpt{},
			cfg: &HttpConfig{
				TlsConfig: &tls.Config{
					MinVersion: tls.VersionTLS12,
				},
				RetryEnabled:          true,
				Timeout:               TestTimeout,
				RetryLimit:            TestRetryLimit,
				MaxIdleConnections:    TestMaxConnectionsPerHost,
				MaxConnectionsPerHost: TestMaxConnectionsPerHost,
				ReadByteLimit:         TestReadByteLimit,
			},
			req: &entity.Request{
				Method: valobj.MethodGet,
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"status": "error"}`))
			},
			expectedCode: http.StatusBadRequest,
			expectedErr:  nil,
		},
		{
			opts: []HttpOpt{},
			cfg: &HttpConfig{
				TlsConfig: &tls.Config{
					MinVersion: tls.VersionTLS12,
				},
				RetryEnabled:          true,
				Timeout:               TestTimeout,
				RetryLimit:            TestRetryLimit,
				MaxIdleConnections:    TestMaxConnectionsPerHost,
				MaxConnectionsPerHost: TestMaxConnectionsPerHost,
				ReadByteLimit:         TestReadByteLimit,
			},
			req: &entity.Request{
				Method: valobj.MethodGet,
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"status": "unauthorized"}`))
			},
			expectedCode: http.StatusUnauthorized,
			expectedErr:  nil,
		},
		{
			opts: []HttpOpt{},
			cfg: &HttpConfig{
				TlsConfig: &tls.Config{
					MinVersion: tls.VersionTLS12,
				},
				RetryEnabled:          true,
				Timeout:               TestTimeout,
				RetryLimit:            TestRetryLimit,
				MaxIdleConnections:    TestMaxConnectionsPerHost,
				MaxConnectionsPerHost: TestMaxConnectionsPerHost,
				ReadByteLimit:         TestReadByteLimit,
			},
			req: &entity.Request{
				Method: valobj.MethodGet,
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"status": "error"}`))
			},
			expectedCode: http.StatusInternalServerError,
			expectedErr:  nil,
		},
	}

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	for i, tt := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/", tt.handler)

			ts := httptest.NewServer(mux)

			tt.opts = append(tt.opts, WithHttpClient(ts.Client()))
			tt.req.URL = ts.URL

			adapter := NewHttpAdapter(tt.cfg, logger, tt.opts...)

			resp, err := adapter.Send(ctx, tt.req)

			assert.Equal(t, tt.expectedErr, err)
			assert.Equal(t, tt.expectedCode, resp.StatusCode)

			ts.Close()
		})
	}
}

func TestHttpErrors(t *testing.T) {
	var err error

	err = &ErrHttpCopy{&ErrTest{}}
	if len(err.Error()) == 0 {
		t.Error("missing error message for ErrHttpCopy")
	}

	err = &ErrHttpStatusCode{http.StatusInternalServerError, bytes.NewBufferString(TestErrMsg)}
	if len(err.Error()) == 0 {
		t.Error("missing error message for ErrHttpStatusCode")
	}

	err = &ErrReadRespBody{&ErrTest{}}
	if len(err.Error()) == 0 {
		t.Error("missing error message for ErrReadRespBody")
	}

	err = &ErrNilRequest{}
	if len(err.Error()) == 0 {
		t.Error("missing error message for ErrNilRequest")
	}

	err = &ErrAttachFile{TestPath, &ErrTest{}}
	if len(err.Error()) == 0 {
		t.Error("missing error message for ErrAttachFile")
	}
}
