package routers

import (
	"gateway_api/app/global/consts"
	"gateway_api/app/global/variable"
	"gateway_api/app/http/controller/api"
	"gateway_api/app/http/controller/common"
	"gateway_api/app/http/middleware/cors"
	validatorFactory "gateway_api/app/http/validator/core/factory"
	"gateway_api/app/utils/gin_release"
	"net/http"

	"github.com/gin-contrib/pprof"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"go.uber.org/zap"
)

func InitApiRouter() *gin.Engine {
	var router *gin.Engine
	// 非调试模式（生产模式） 日志写到日志文件
	if !variable.ConfigYml.GetBool("AppDebug") {
		router = gin_release.ReleaseRouter()
	} else {
		// 调试模式，开启 pprof 包，便于开发阶段分析程序性能
		router = gin.Default()
		pprof.Register(router)
	}
	// 设置可信任的代理服务器列表,gin (2021-11-24发布的v1.7.7版本之后出的新功能)
	if variable.ConfigYml.GetInt("HttpServer.TrustProxies.IsOpen") == 1 {
		if err := router.SetTrustedProxies(variable.ConfigYml.GetStringSlice("HttpServer.TrustProxies.ProxyServerList")); err != nil {
			variable.ZapLog.Error(consts.GinSetTrustProxyError, zap.Error(err))
		}
	} else {
		variable.ZapLog.Info("test")
		_ = router.SetTrustedProxies(nil)
	}

	//根据配置进行设置跨域
	if variable.ConfigYml.GetBool("HttpServer.AllowCrossDomain") {
		router.Use(cors.Next())
	}

	// 使用requestid中间件
	router.Use(requestid.New(
		requestid.WithGenerator(func() string {
			return "glb-req-" + uuid.New().String()
		}),
		requestid.WithCustomHeaderStrKey("your-customer-key"),
	))

	router.GET("/", func(context *gin.Context) {

		context.String(http.StatusOK, "Api 模块接口 hello word！")
	})

	router.GET("/check", common.CheckMq)

	//  创建一个门户类接口路由组
	vApi := router.Group("/api/v1/")
	{
		// 模拟一个首页路由
		home := vApi.Group("home/")
		{
			home.GET("news", validatorFactory.Create(consts.ValidatorPrefix+"HomeNews"))
		}
		// 路由注册
		test := vApi.Group("test/")
		{
			test.GET("news", api.GetUserInfo)
			test.GET("news2", (&api.Home{}).News) // 跳过参数校验
		}
	}
	return router
}
