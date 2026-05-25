package handler

import (
	"net/http"

	chat "SamaraAI/internal/handler/chat"
	"SamaraAI/internal/handler/file"
	"SamaraAI/internal/handler/image"
	user "SamaraAI/internal/handler/user"
	"SamaraAI/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	server.AddRoutes(
		rest.WithMiddlewares(
			[]rest.Middleware{serverCtx.JwtAuth},
			[]rest.Route{
				{Method: http.MethodPost, Path: "/chat/history", Handler: chat.ChatHistoryHandler(serverCtx)},
				{Method: http.MethodGet, Path: "/chat/sessions", Handler: chat.GetUserSessionsHandler(serverCtx)},
				{Method: http.MethodPost, Path: "/chat/send-stream-new-session", Handler: chat.CreateStreamSessionAndSendMessageHandler(serverCtx)},
				{Method: http.MethodPost, Path: "/chat/send-stream", Handler: chat.ChatStreamSendHandler(serverCtx)},
			}...,
		),
		rest.WithPrefix("/api/v1/AI"),
	)

	server.AddRoutes(
		[]rest.Route{
			{Method: http.MethodPost, Path: "/user/captcha", Handler: user.CaptchaHandler(serverCtx)},
			{Method: http.MethodPost, Path: "/user/login", Handler: user.LoginHandler(serverCtx)},
			{Method: http.MethodPost, Path: "/user/register", Handler: user.RegisterHandler(serverCtx)},
		},
		rest.WithPrefix("/api/v1"),
	)

	server.AddRoutes(
		rest.WithMiddlewares(
			[]rest.Middleware{serverCtx.JwtAuth},
			[]rest.Route{
				{Method: http.MethodPost, Path: "/recognize", Handler: image.RecognizeImageHandler},
			}...,
		),
		rest.WithPrefix("/api/v1/image"),
	)

	server.AddRoutes(
		rest.WithMiddlewares(
			[]rest.Middleware{serverCtx.JwtAuth},
			[]rest.Route{
				{Method: http.MethodPost, Path: "/upload", Handler: file.UploadRagFileHandler},
			}...,
		),
		rest.WithPrefix("/api/v1/file"),
	)
}
