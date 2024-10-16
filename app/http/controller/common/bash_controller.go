package common

import (
	"gateway_api/app/global/variable"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CheckMq(c *gin.Context) {
	res, err := variable.Check.Check(c)
	res["name"] = variable.Check.Name()

	variable.Log.Info(c, "test") //全局log，添加glb-request-id字段

	if err != nil {
		res["status"] = "Down"
	} else {
		res["status"] = "Up"
	}

	variable.MqCon.Send("test::test", "text/plain", []byte("Hello ActiveMQ!22"))
	c.JSON(http.StatusOK, res)
}
