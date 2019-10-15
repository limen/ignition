package ignitor

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
)

var Response = response{}

type response struct{}

func (response) Success(ctx *gin.Context, data interface{}) {
	resp := map[string]interface{}{
		"status": "SUCCESS",
		"code":   "SUCCESS",
		"msg":    "success",
		"data":   data,
	}
	jsonData, err := json.Marshal(resp)
	if err != nil {
		panic(err)
	}
	ctx.Status(http.StatusOK)
	ctx.Header("Content-Type", "application/json; charset=utf-8")
	ctx.Writer.Write(jsonData)
}

func (response) Error(ctx *gin.Context, code, msg string, data interface{}) {
	resp := map[string]interface{}{
		"status": "SUCCESS",
		"code":   code,
		"msg":    "error",
		"data":   data,
	}
	jsonData, err := json.Marshal(resp)
	if err != nil {
		panic(err)
	}
	ctx.Status(http.StatusOK)
	ctx.Header("Content-Type", "application/json; charset=utf-8")
	ctx.Writer.Write(jsonData)
}
