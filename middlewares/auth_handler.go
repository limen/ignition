package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/limen/ignition/auth"
)

type AuthHandler struct {
	AuthProvider auth.UserAuthProviderInterface
	AbortFunc    func(ctx *gin.Context, err error)
}

func (ah AuthHandler) NewHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		username := ah.AuthProvider.GetContextUsername(ctx)
		credit := ah.AuthProvider.GetContextCredential(ctx)
		ok, err := ah.AuthProvider.Auth(username, credit)
		if !ok || err != nil {
			ah.AbortFunc(ctx, err)
			return
		}
		ctx.Next()
	}
}
