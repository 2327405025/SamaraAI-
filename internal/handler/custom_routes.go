package handler

import (
	"SamaraAI/internal/handler/chat"
	"SamaraAI/internal/handler/file"
	"SamaraAI/internal/handler/image"
	"SamaraAI/internal/svc"
	"net/http"

	"github.com/zeromicro/go-zero/rest"
)

// RegisterCustomHandlers registers routes not covered by goctl api (SSE, multipart).
func RegisterCustomHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	server.AddRoutes(
		rest.WithMiddlewares(
			[]rest.Middleware{serverCtx.JwtAuth},
			[]rest.Route{
				{
					Method:  http.MethodPost,
					Path:    "/chat/send-stream-new-session",
					Handler: chat.CreateStreamSessionAndSendMessageHandler(serverCtx),
				},
				{
					Method:  http.MethodPost,
					Path:    "/chat/send-stream",
					Handler: chat.ChatStreamSendHandler(serverCtx),
				},
			}...,
		),
		rest.WithPrefix("/api/v1/AI"),
	)

	server.AddRoutes(
		rest.WithMiddlewares(
			[]rest.Middleware{serverCtx.JwtAuth},
			[]rest.Route{
				{
					Method:  http.MethodPost,
					Path:    "/recognize",
					Handler: image.RecognizeImageHandler,
				},
			}...,
		),
		rest.WithPrefix("/api/v1/image"),
	)

	server.AddRoutes(
		rest.WithMiddlewares(
			[]rest.Middleware{serverCtx.JwtAuth},
			[]rest.Route{
				{
					Method:  http.MethodPost,
					Path:    "/upload",
					Handler: file.UploadRagFileHandler,
				},
			}...,
		),
		rest.WithPrefix("/api/v1/file"),
	)
}
