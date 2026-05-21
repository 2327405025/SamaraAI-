// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package user

import (
	"context"

	"SamaraAI/common/code"
	"SamaraAI/internal/logic/resp"
	"SamaraAI/internal/svc"
	"SamaraAI/internal/types"
	usersvc "SamaraAI/service/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(req *types.RegisterReq) (*types.RegisterResp, error) {
	token, code_ := usersvc.Register(req.Email, req.Password, req.Captcha)
	if code_ != code.CodeSuccess {
		return &types.RegisterResp{BaseResp: resp.Base(code_)}, nil
	}
	return &types.RegisterResp{BaseResp: resp.Success(), Token: token}, nil
}
