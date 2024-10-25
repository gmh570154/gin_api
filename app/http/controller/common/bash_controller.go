package common

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"gateway_api/app/global/variable"
	redis "gateway_api/app/utils/redis_factory"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func CheckMq(c *gin.Context) {
	hander_name := c.HandlerName()
	method := c.Request.Method
	body, _ := c.GetRawData()
	variable.Log.Info(c, string(body[:]))
	// params := sort.Sort(sort.Interface(body))  // 参数排序功能，todo

	uid := fmt.Sprintf("%x", sha256.Sum256([]byte(hander_name+method)))

	res, err := variable.Check.Check(c)
	res["name"] = variable.Check.Name()

	// variable.Log.Info(c, "test") //全局log，添加glb-request-id字段

	if cmd, _ := redis.RedisClient.Exists(c, uid).Result(); cmd <= 0 { // 如果redis中没有结果，才会发送消息给mq获取异步请求结果
		mq_body := variable.MqBody{
			Function_name: hander_name,
			Method:        method,
			Body:          string(body),
		}

		res["status"] = "Up"
		if err != nil {
			res["status"] = "Down"
		}

		b, err := json.Marshal(mq_body)
		if err != nil {
			fmt.Println("JSON ERR:", err)
		}
		variable.MqCon.Send("gin::app:request", "text/plain", []byte(b))

		for {
			if cmd, _ := redis.RedisClient.Exists(c, uid).Result(); cmd > 0 {
				res["response"], _ = redis.RedisClient.Get(c, uid).Result()
				log.Printf("bash get response: %s, uid: %s", res["response"], uid)
				break
			} else {
				log.Println("bash sleep 1s")
				time.Sleep(time.Second)
			}
		}
	} else { // 直接从redis获取缓存结果
		res["response"], _ = redis.RedisClient.Get(c, uid).Result()
		log.Printf("bash get response: %s, uid: %s", res["response"], uid)
	}

	c.JSON(http.StatusOK, res)
}
