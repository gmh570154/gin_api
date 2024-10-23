package activemq

import (
	"context"

	"time"

	"github.com/core-go/activemq"
	"github.com/go-stomp/stomp/v3"
)

type HealthChecker struct {
	name     string
	addr     string
	username string
	password string
}

func NewHealthChecker(mqCnf activemq.Config, name string) *HealthChecker {
	return &HealthChecker{name, mqCnf.Addr, mqCnf.UserName, mqCnf.Password}
}

func (s *HealthChecker) Name() string {
	return s.name
}

func (s *HealthChecker) Check(ctx context.Context) (map[string]interface{}, error) {
	var conn *stomp.Conn
	sendTimeout, recvTimeout := time.Duration(3), time.Duration(3) // 发送请求超时3s
	options := []func(*stomp.Conn) error{
		stomp.ConnOpt.Login(s.username, s.password),
		stomp.ConnOpt.HeartBeat(sendTimeout, recvTimeout),
	}

	res := make(map[string]interface{})
	var err error
	conn, err = stomp.Dial("tcp", s.addr, options...)
	if err != nil {
		return res, err
	}

	res["version"] = conn.Version()
	return res, err
}
