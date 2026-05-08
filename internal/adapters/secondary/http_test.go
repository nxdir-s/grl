package secondary

import (
	"bytes"
	"context"
	"crypto/tls"
	_ "embed"
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
		expectedCode int
		expectedErr  error
		endpoint     string
		response     *bytes.Buffer
	}{
		{
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
			opts: []HttpOpt{},
			req: &entity.Request{
				Method: valobj.MethodGet,
				URL:    TestHost + TestEndpoint,
			},
			expectedCode: http.StatusOK,
			expectedErr:  nil,
			endpoint:     TestEndpoint,
			response:     bytes.NewBuffer(testData),
		},
	}

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	for i, tt := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.expectedCode)
				w.Write(tt.response.Bytes())
			})

			ts := httptest.NewServer(mux)

			tt.opts = append(tt.opts, WithHttpClient(ts.Client()))

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
