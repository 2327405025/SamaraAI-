// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package user

import (
	"context"

	"SamaraAI/common/code"
	"SamaraAI/internal/logic/resp"
	"SamaraAI/internal/svc"
	"SamaraAI/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginReq) (*types.LoginResp, error) {
	token, code_ := login(req.Username, req.Password)
	if code_ != code.CodeSuccess {
		return &types.LoginResp{BaseResp: resp.Base(code_)}, nil
	}
	return &types.LoginResp{BaseResp: resp.Success(), Token: token}, nil
}
