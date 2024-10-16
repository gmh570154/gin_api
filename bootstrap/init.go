package bootstrap

import (
	// _ "gateway_api/app/core/destroy" // 监听程序退出信号，用于资源的释放

	"context"
	"encoding/json"
	"fmt"
	"gateway_api/app/global/my_errors"
	"gateway_api/app/global/variable"
	"gateway_api/app/http/validator/common/data_type/register_validator"

	"log"

	"gateway_api/app/service/sys_log_hook"
	// "gateway_api/app/utils/casbin_v2"
	// "gateway_api/app/utils/gorm_v2"
	// "gateway_api/app/utils/snow_flake"
	"gateway_api/app/utils/validator_translation"
	// "gateway_api/app/utils/websocket/core"
	"gateway_api/app/utils/yml_config"
	"gateway_api/app/utils/zap_factory"

	// "log"
	"os"

	"github.com/core-go/activemq"
	ah "github.com/core-go/health/activemq/v3"
	"github.com/go-stomp/stomp/v3"
)

// 检查项目必须的非编译目录是否存在，避免编译后调用的时候缺失相关目录
func checkRequiredFolders() {
	//1.检查配置文件是否存在
	if _, err := os.Stat(variable.BasePath + "/config/config.yml"); err != nil {
		log.Fatal(my_errors.ErrorsConfigYamlNotExists + err.Error())
	}
	// if _, err := os.Stat(variable.BasePath + "/config/gorm_v2.yml"); err != nil {
	// 	log.Fatal(my_errors.ErrorsConfigGormNotExists + err.Error())
	// }
	//2.检查public目录是否存在
	// if _, err := os.Stat(variable.BasePath + "/public/"); err != nil {
	// 	log.Fatal(my_errors.ErrorsPublicNotExists + err.Error())
	// }
	//3.检查storage/logs 目录是否存在
	if _, err := os.Stat(variable.BasePath + "/storage/logs/"); err != nil {
		log.Fatal(my_errors.ErrorsStorageLogsNotExists + err.Error())
	}
	// 4.自动创建软连接、更好的管理静态资源
	// if _, err := os.Stat(variable.BasePath + "/public/storage"); err == nil {
	// 	if err = os.RemoveAll(variable.BasePath + "/public/storage"); err != nil {
	// 		log.Fatal(my_errors.ErrorsSoftLinkDeleteFail + err.Error())
	// 	}
	// }
	// if err := os.Symlink(variable.BasePath+"/storage/app", variable.BasePath+"/public/storage"); err != nil {
	// 	log.Fatal(my_errors.ErrorsSoftLinkCreateFail + err.Error())
	// }
}

func init_mq() {
	variable.Mqcfg = variable.Config{
		Amq: activemq.Config{
			Addr:             variable.ConfigYml.GetString("amq.addr"),
			UserName:         variable.ConfigYml.GetString("amq.username"),
			Password:         variable.ConfigYml.GetString("amq.password"),
			DestinationName:  variable.ConfigYml.GetString("amq.destination_name"),
			SubscriptionName: variable.ConfigYml.GetString("amq.subscription_name"),
		},
	}

	logError := func(ctx context.Context, msg string) { //错误执行
		log.Println(msg)
	}

	sub, er2 := activemq.NewSubscriberByConfig(variable.Mqcfg.Amq, stomp.AckAuto, logError, true)

	if er2 != nil {
		log.Fatal("Cannot create a new subscriber. Error: " + er2.Error())
	}
	subscriberChecker := ah.NewHealthChecker(variable.Mqcfg.Amq.Addr, "amq_subscriber") // 第二个参数定义订阅的name
	go func() {                                                                         //异步处理接收的消息
		for {
			msg := <-sub.Subscription.C
			if msg.Err != nil {
				// TODO: 重连
				log.Printf("%s", msg.Err.Error())
			} else { // todo 需要将消息保存到redis中
				// 接受消息，执行业务逻辑
				fmt.Println(string(msg.Body))
				log.Println("start")

				var v any
				er1 := json.Unmarshal(msg.Body, &v) //转成json格式
				if er1 != nil {                     // 一层则打印日志，并忽略消息 --todo
					log.Printf("cannot unmarshal item: %s. Error: %s", msg.Body, er1.Error())
					continue
				}
				log.Println("end")
			}
		}

	}()
	variable.Check = subscriberChecker
	variable.MqCon = sub.Conn
}

func init() {
	// 1. 初始化 项目根路径，参见 variable 常量包，相关路径：app\global\variable\variable.go

	//2.检查配置文件以及日志目录等非编译性的必要条件
	checkRequiredFolders()

	//3.初始化表单参数验证器，注册在容器（Web、Api共用容器）
	// register_validator.WebRegisterValidator()
	register_validator.ApiRegisterValidator()

	// 4.启动针对配置文件(confgi.yml、gorm_v2.yml)变化的监听， 配置文件操作指针，初始化为全局变量
	variable.ConfigYml = yml_config.CreateYamlFactory()
	variable.ConfigYml.ConfigFileChangeListen()

	init_mq()
	// 5.初始化全局日志句柄，并载入日志钩子处理函数
	variable.ZapLog = zap_factory.CreateZapFactory(sys_log_hook.ZapLogHandler)

	//10.全局注册 validator 错误翻译器,zh 代表中文，en 代表英语
	if err := validator_translation.InitTrans("zh"); err != nil {
		log.Fatal(my_errors.ErrorsValidatorTransInitFail + err.Error())
	}
}
