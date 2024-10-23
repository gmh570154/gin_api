package bootstrap

import (
	// _ "gateway_api/app/core/destroy" // 监听程序退出信号，用于资源的释放

	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"gateway_api/app/global/my_errors"
	"gateway_api/app/global/variable"
	"gateway_api/app/http/validator/common/data_type/register_validator"

	"log"

	"gateway_api/app/service/sys_log_hook"
	"gateway_api/app/utils/validator_translation"
	"gateway_api/app/utils/yml_config"
	"gateway_api/app/utils/zap_factory"

	"os"

	redis "gateway_api/app/utils/redis_factory"

	"github.com/core-go/activemq"

	ch "gateway_api/app/core/activemq"

	"github.com/go-stomp/stomp/v3"
)

// 检查项目必须的非编译目录是否存在，避免编译后调用的时候缺失相关目录
func checkRequiredFolders() {
	//1.检查配置文件是否存在
	if _, err := os.Stat(variable.BasePath + "/config/config.yml"); err != nil {
		log.Fatal(my_errors.ErrorsConfigYamlNotExists + err.Error())
	}
	//2.检查storage/logs 目录是否存在
	if _, err := os.Stat(variable.BasePath + "/storage/logs/"); err != nil {
		log.Fatal(my_errors.ErrorsStorageLogsNotExists + err.Error())
	}

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
	ctx := context.Background()

	subscriberChecker := ch.NewHealthChecker(variable.Mqcfg.Amq, "proxy subscriber") // 第二个参数定义订阅的name
	go func() {                                                                      //异步处理接收的消息
		for {
			msg := <-sub.Subscription.C
			if msg.Err != nil {
				// TODO: 重连
				log.Printf("%s", msg.Err.Error())
			} else { // todo 需要将消息保存到redis中
				// 接受消息，执行业务逻辑
				log.Printf("mq msg hanlder start, msg: %s", string(msg.Body))

				var v variable.MqBody
				er1 := json.Unmarshal(msg.Body, &v) //转成json格式
				if er1 != nil {                     // 一层则打印日志，并忽略消息 --todo
					log.Printf("cannot unmarshal item: %s. Error: %s", msg.Body, er1.Error())
					continue
				}

				uid := fmt.Sprintf("%x", sha256.Sum256([]byte(v.Function_name+v.Method)))

				res, err3 := redis.RedisClient.Set(ctx, string(uid[:]), msg.Body, -1).Result()

				if err3 != nil {
					log.Printf("error: %s", err3)
				}

				log.Printf("mq msg hanlder end, uid: %s, result: %s", uid, res)
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
