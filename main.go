package main

import (
	"gateway_api/app/global/variable" // 常量包
	_ "gateway_api/bootstrap"         //初始化配置文件到全局变量
	"gateway_api/routers"
)

func main() {
	router := routers.InitApiRouter()
	_ = router.Run(variable.ConfigYml.GetString("HttpServer.Api.Port"))
}
