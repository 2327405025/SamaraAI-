// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package svc

import (
	"SamaraAI/internal/config"
	"SamaraAI/internal/middleware"
	"github.com/zeromicro/go-zero/rest"
)

type ServiceContext struct {
	Config  config.Config
	JwtAuth rest.Middleware
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:  c,
		JwtAuth: middleware.NewJwtAuthMiddleware().Handle,
	}
}
