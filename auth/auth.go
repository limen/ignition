package auth

import "github.com/gin-gonic/gin"

type UserInterface interface {
	GetUsername() interface{}
	GetPassword() string
}

type UserProviderInterface interface {
	Create(user UserInterface) error
	FindByUsername(username interface{}) UserInterface
	ValidatePassword(username interface{}, password string) (bool, error)
}

type UserAuthProviderInterface interface {
	Create(username interface{}, credential interface{}) error
	Auth(username interface{}, credential interface{}) (bool, error)
	GetContextUsername(ctx *gin.Context) interface{}
	GetContextCredential(ctx *gin.Context) interface{}
}
