package common

import (
	"gateway_api/app/global/variable"
	redis "gateway_api/app/utils/redis_factory"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CheckMq(c *gin.Context) {
	res, err := variable.Check.Check(c)
	res["name"] = variable.Check.Name()

	variable.Log.Info(c, "test") //全局log，添加glb-request-id字段
	// 从连接池获取一个连接
	cmd := redis.RedisClient.Set(c, "a", "123", -1)

	if cmd.Err() != nil {
		variable.Log.Error(c, cmd.Err().Error())
	}
	variable.Log.Info(c, cmd.Val())

	if err != nil {
		res["status"] = "Down"
	} else {
		res["status"] = "Up"
	}

	variable.MqCon.Send("test::test", "text/plain", []byte("Hello ActiveMQ!22"))
	c.JSON(http.StatusOK, res)
}
