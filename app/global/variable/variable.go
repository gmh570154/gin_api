package variable

import (
	"gateway_api/app/global/my_errors"
	"gateway_api/app/utils/yml_config/ymlconfig_interf"
	"log"
	"os"
	"strings"

	"github.com/core-go/activemq"
	ah "github.com/core-go/health/activemq/v3"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/go-stomp/stomp/v3"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ginskeleton 封装的全局变量全部支持并发安全，请放心使用即可
// 开发者自行封装的全局变量，请做好并发安全检查与确认

type Config struct {
	Amq activemq.Config `mapstructure:"amq"`
}

var (
	BasePath           string                  // 定义项目的根目录
	EventDestroyPrefix = "Destroy_"            //  程序退出时需要销毁的事件前缀
	ConfigKeyPrefix    = "Config_"             //  配置文件键值缓存时，键的前缀
	DateFormat         = "2006-01-02 15:04:05" //  设置全局日期时间格式

	// 全局日志指针
	ZapLog *zap.Logger
	// 全局配置文件
	ConfigYml       ymlconfig_interf.YmlConfigInterf // 全局配置文件指针
	ConfigGormv2Yml ymlconfig_interf.YmlConfigInterf // 全局配置文件指针
	Check           *ah.HealthChecker
	MqCon           *stomp.Conn
	Log             *NewLogType
	Mqcfg           Config
)

type NewLogType struct{}

func (log *NewLogType) Info(c *gin.Context, msg string) {
	ZapLog.Info(msg, zapcore.Field{
		Key:    "request-id",
		Type:   zapcore.StringType,
		String: requestid.Get(c),
	})
}

func (log *NewLogType) Error(c *gin.Context, msg string) {
	ZapLog.Error(msg, zapcore.Field{
		Key:    "request-id",
		Type:   zapcore.StringType,
		String: requestid.Get(c),
	})
}

func (log *NewLogType) Debug(c *gin.Context, msg string) {
	ZapLog.Debug(msg, zapcore.Field{
		Key:    "request-id",
		Type:   zapcore.StringType,
		String: requestid.Get(c),
	})
}

func init() {
	// 1.初始化程序根目录
	if curPath, err := os.Getwd(); err == nil {
		// 路径进行处理，兼容单元测试程序程序启动时的奇怪路径
		if len(os.Args) > 1 && strings.HasPrefix(os.Args[1], "-test") {
			BasePath = strings.Replace(strings.Replace(curPath, `\test`, "", 1), `/test`, "", 1)
		} else {
			BasePath = curPath
		}
	} else {
		log.Fatal(my_errors.ErrorsBasePath)
	}
}
