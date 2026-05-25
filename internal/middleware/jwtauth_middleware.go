// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package middleware

import (
	"SamaraAI/common/code"
	"SamaraAI/internal/config"
	"SamaraAI/internal/contextkey"
	"SamaraAI/internal/types"
	"SamaraAI/utils/myjwt"
	"context"
	"net/http"
	"strings"

	"github.com/zeromicro/go-zero/rest/httpx"
)

type JwtAuthMiddleware struct{}

func NewJwtAuthMiddleware() *JwtAuthMiddleware {
	return &JwtAuthMiddleware{}
}

func (m *JwtAuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var token string
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		} else if config.IsDevMode() {
			token = r.URL.Query().Get("token")
		}

		if token == "" {
			httpx.OkJson(w, types.BaseResp{
				StatusCode: int64(code.CodeInvalidToken),
				StatusMsg:  code.CodeInvalidToken.Msg(),
			})
			return
		}

		userName, ok := myjwt.ParseToken(token)
		if !ok {
			httpx.OkJson(w, types.BaseResp{
				StatusCode: int64(code.CodeInvalidToken),
				StatusMsg:  code.CodeInvalidToken.Msg(),
			})
			return
		}

		ctx := context.WithValue(r.Context(), contextkey.UserName, userName)
		next(w, r.WithContext(ctx))
	}
}

func UserNameFromContext(ctx context.Context) string {
	v, _ := ctx.Value(contextkey.UserName).(string)
	return v
}
