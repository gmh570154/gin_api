# gin_api

go gin api project

# 已完成功能及待处理项

1.配置文件处理
2.activemq 集成
3.glb-request-id 集成到日志文件
4.redis 处理请求数据问题 -- doing,已优化 redis 连接 5.为啥 activemq 不用密码也能连接发送消息，待优化

# 消息传递方案设计

发送：
topic：gin:app:request
type: json
body:
{
func: function,
type: post/get,
request-id: glb-request_id,
body: params
}

接收：
topic：gin:app:response
type: json
body:
{
func: function,
type: post/get,
request-id: glb-request_id,
body: response
}

redis 消息保存：
redis.set(response:SHA2-256(func-type-sort(params)), body, 30min)

redis 消息获取(阻塞？)：如何重复请求相同数据
redis.get(response:SHA2-256(func-type-sort(params)))
重复获取的时候，需要把过期时间重置

流程图：

服务端：
发起请求 -> 请求数据 hash key 生成 -> 如果 redis 中 key 存在则获取返回
-> 如果 key 不存在 -> 发送数据到 mq，并等待 redis 中 key 有 response，需要设置超时，超时返回错误 -> 返回数据

mq 任务处理进程：
监听 mq -> 获取消息并解析，调用 app 获取数据 -> 将返回的数据插入到 mq 对应的 topic 中
