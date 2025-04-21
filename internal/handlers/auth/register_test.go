package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/korobkovandrey/gmartloyalty/internal/handlers/auth/mocks"
	"github.com/korobkovandrey/gmartloyalty/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestRegister(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name                    string
		requestBody             interface{}
		mockSetup               func(*mocks.MockregisterService)
		wantStatusCode          int
		wantHeaderAuthorization string
	}{
		{
			name:        "success",
			requestBody: map[string]interface{}{"login": "user1", "password": "pass123"},
			mockSetup: func(m *mocks.MockregisterService) {
				m.EXPECT().
					Register(gomock.Any(), gomock.Eq("user1"), gomock.Eq("pass123")).
					Return("jwt_token_123", nil).
					Times(1)
			},
			wantStatusCode:          http.StatusOK,
			wantHeaderAuthorization: "jwt_token_123",
		},
		{
			name:        "invalid json",
			requestBody: "invalid json",
			mockSetup: func(m *mocks.MockregisterService) {
				m.EXPECT().
					Register(gomock.Any(), gomock.Any(), gomock.Any()).
					MaxTimes(0)
			},
			wantStatusCode:          http.StatusBadRequest,
			wantHeaderAuthorization: "",
		},
		{
			name:        "invalid request",
			requestBody: map[string]interface{}{"login": "", "password": "pass123"},
			mockSetup: func(m *mocks.MockregisterService) {
				m.EXPECT().
					Register(gomock.Any(), gomock.Any(), gomock.Any()).
					MaxTimes(0)
			},
			wantStatusCode:          http.StatusBadRequest,
			wantHeaderAuthorization: "",
		},
		{
			name:        "user already exists",
			requestBody: map[string]interface{}{"login": "user1", "password": "pass123"},
			mockSetup: func(m *mocks.MockregisterService) {
				m.EXPECT().
					Register(gomock.Any(), gomock.Eq("user1"), gomock.Eq("pass123")).
					Return("", service.ErrAlreadyExists).
					Times(1)
			},
			wantStatusCode:          http.StatusConflict,
			wantHeaderAuthorization: "",
		},
		{
			name:        "service error",
			requestBody: map[string]interface{}{"login": "user1", "password": "pass123"},
			mockSetup: func(m *mocks.MockregisterService) {
				m.EXPECT().
					Register(gomock.Any(), gomock.Eq("user1"), gomock.Eq("pass123")).
					Return("", errors.New("database error")).
					Times(1)
			},
			wantStatusCode:          http.StatusInternalServerError,
			wantHeaderAuthorization: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body io.Reader
			if tt.requestBody != nil {
				if str, ok := tt.requestBody.(string); ok {
					body = strings.NewReader(str)
				} else {
					bodyBytes, err := json.Marshal(tt.requestBody)
					require.NoError(t, err)
					body = bytes.NewReader(bodyBytes)
				}
			} else {
				body = http.NoBody
			}
			mockRegisterService := mocks.NewMockregisterService(ctrl)
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(http.MethodPost, "/register", body)
			ctx.Request.Header.Set("Content-Type", "application/json")
			tt.mockSetup(mockRegisterService)
			NewRegisterHandler(mockRegisterService)(ctx)
			assert.Equal(t, tt.wantStatusCode, ctx.Writer.Status())
			assert.Equal(t, tt.wantHeaderAuthorization, ctx.Writer.Header().Get("Authorization"))
		})
	}
}
