package secondary

import (
	"bytes"
	"context"
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
	TestHost      string = "http://example.com"
	TestEndpoint  string = "/api"
	TestResponse  string = "{}"
	TestRateLimit int    = 60
)

func TestSend(t *testing.T) {
	cases := []struct {
		opts         []HttpOpt
		req          *entity.Request
		expectedCode int
		expectedErr  error
		endpoint     string
		response     *bytes.Buffer
	}{
		{
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

			adapter := NewHttpAdapter(&HttpConfig{}, logger, tt.opts...)

			resp, err := adapter.Send(ctx, tt.req)

			assert.Equal(t, tt.expectedErr, err)
			assert.Equal(t, tt.expectedCode, resp.StatusCode)

			ts.Close()
		})
	}
}
