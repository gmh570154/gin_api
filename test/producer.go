package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-stomp/stomp"
)

/**
生产者
*/

type Person struct {
	ID   uint
	Name string
	Age  int
}

func main() {

	//建立连接
	conn, err := stomp.Dial("tcp", "127.0.0.1:61613")
	if err != nil {
		fmt.Println("Connection erorr: " + err.Error())
	}

	//循环11次 向队列中发送数据
	for i := 0; i <= 1; i++ {
		// err = conn.Send("test::test", "text/plain", []byte(fmt.Sprintf("Hello ActiveMQ!"+string(i))))
		p := Person{
			ID:   1,
			Name: "Bruce",
			Age:  18,
		}
		output, err := json.Marshal(p)
		if err != nil {
			panic(err)
		}
		// json_body := `{"status":"UP","details":{"amq":{"status":"UP","data":{"version":"1.2"}},"mongo":{"status":"UP"}}}`
		// err = conn.Send("test::test", "text/plain", []byte(json_body))
		err = conn.Send("gin::app:response", "text/plain", []byte(output))
		if err != nil {
			fmt.Println("textMQ send ActiveMQ erorr: " + err.Error())
		}

		err = conn.Send("gin::app:request", "text/plain", []byte(output))
		if err != nil {
			fmt.Println("textMQ send ActiveMQ erorr: " + err.Error())
		}

		//需要一些时间去发送数据,否则可能会丢
		time.Sleep(time.Second)
	}

	fmt.Println("Send All data success!")

}
