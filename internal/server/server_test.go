package server

import (
	"context"
	"database/sql"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/korobkovandrey/gmartloyalty/internal/infra/store"
	"github.com/korobkovandrey/gmartloyalty/test/dbtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestServerIntegration_registerRoutes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	accrual := &mockAccrualPusher{}

	tests := []struct {
		name string
		seed func(t *testing.T, db *sql.DB)
		testCase
	}{
		// register
		{
			name: "register ok",
			testCase: testCase{
				method:   http.MethodPost,
				url:      "/api/user/register",
				postBody: `{"login":"user1","password":"password"}`,
				headers: map[string]string{
					"Content-Type": "application/json",
				},
				wantHeaderAuthorization: true,
				wantCode:                http.StatusOK,
			},
		},
		{
			name: "register user already exists",
			seed: func(t *testing.T, db *sql.DB) {
				dbtest.SeedUsers(t, db)
			},
			testCase: testCase{
				method:   http.MethodPost,
				url:      "/api/user/register",
				postBody: `{"login":"user1","password":"password"}`,
				headers: map[string]string{
					"Content-Type": "application/json",
				},
				wantHeaderAuthorization: false,
				wantCode:                http.StatusConflict,
			},
		},
		{
			name: "register fail form",
			testCase: testCase{
				method:   http.MethodPost,
				url:      "/api/user/register",
				postBody: `{"fail":"form"}`,
				headers: map[string]string{
					"Content-Type": "application/json",
				},
				wantHeaderAuthorization: false,
				wantCode:                http.StatusBadRequest,
			},
		},
		// login
		{
			name: "login ok",
			seed: func(t *testing.T, db *sql.DB) {
				dbtest.SeedUsers(t, db)
			},
			testCase: testCase{
				method:   http.MethodPost,
				url:      "/api/user/login",
				postBody: `{"login":"user1","password":"password"}`,
				headers: map[string]string{
					"Content-Type": "application/json",
				},
				wantHeaderAuthorization: true,
				wantCode:                http.StatusOK,
			},
		},
		{
			name: "login user not exists",
			testCase: testCase{
				method:   http.MethodPost,
				url:      "/api/user/login",
				postBody: `{"login":"user1","password":"password"}`,
				headers: map[string]string{
					"Content-Type": "application/json",
				},
				wantHeaderAuthorization: false,
				wantCode:                http.StatusUnauthorized,
			},
		},
		{
			name: "login fail form",
			testCase: testCase{
				method:   http.MethodPost,
				url:      "/api/user/login",
				postBody: `{"fail":"form"}`,
				headers: map[string]string{
					"Content-Type": "application/json",
				},
				wantHeaderAuthorization: false,
				wantCode:                http.StatusBadRequest,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.Default()
			store.TestWithStore(t, func(t *testing.T, s *store.Store) {
				if tt.seed != nil {
					tt.seed(t, s.DB)
				}
				registerRoutes(router, "secret_key", 1, s, accrual)
				testHelper(t, router, tt.testCase)
			})
		})
	}
}

type mockAccrualPusher struct{}

func (m *mockAccrualPusher) PushOrder(orderNum string) {}

type testCase struct {
	method                  string
	url                     string
	postBody                string
	headers                 map[string]string
	wantCode                int
	wantHeaderAuthorization bool
	wantJSON                string
}

func testRequest(
	t *testing.T, ts *httptest.Server,
	method, path string, postBody io.Reader, headers map[string]string) (body []byte, statusCode int, headerAuthorization string) {
	t.Helper()
	req, err := http.NewRequestWithContext(context.TODO(), method, ts.URL+path, postBody)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	require.NoError(t, err)
	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, resp.Body.Close())
	}()
	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err)
	statusCode = resp.StatusCode
	headerAuthorization = resp.Header.Get("Authorization")
	return
}

func testHelper(t *testing.T, h http.Handler, tt testCase) {
	t.Helper()
	ts := httptest.NewServer(h)
	defer ts.Close()
	var postBody io.Reader
	if tt.postBody == "" {
		postBody = http.NoBody
	} else {
		postBody = strings.NewReader(tt.postBody)
	}
	gotBody, gotCode, headerAuthorization := testRequest(t, ts, tt.method, tt.url, postBody, tt.headers)
	gotBodyString := string(gotBody)
	require.Equal(t, tt.wantCode, gotCode)
	if tt.wantHeaderAuthorization {
		assert.NotEmpty(t, headerAuthorization)
	}
	if tt.wantJSON != "" {
		assert.JSONEq(t, tt.wantJSON, gotBodyString)
	}
}
