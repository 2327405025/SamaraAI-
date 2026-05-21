package resp

import (
	"SamaraAI/common/code"
	"SamaraAI/internal/types"
)

func Base(c code.Code) types.BaseResp {
	return types.BaseResp{
		StatusCode: int64(c),
		StatusMsg:  c.Msg(),
	}
}

func Success() types.BaseResp {
	return Base(code.CodeSuccess)
}
