package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-stomp/stomp"
)

/*
*

消费者
*/
func main() {

	userName := "gmh"
	password := "1qaz!QAZ"
	clientID := "app handler1"
	sendTimeout := time.Duration(3)
	recvTimeout := time.Duration(3)
	options := []func(*stomp.Conn) error{
		stomp.ConnOpt.Login(userName, password),
		// stomp.ConnOpt.Host("/"),
		stomp.ConnOpt.Header("client-id", clientID),
		stomp.ConnOpt.HeartBeat(sendTimeout, recvTimeout),
	}

	fmt.Printf("%s", options)

	//Connect activemq
	conn, err := stomp.Dial("tcp", "127.0.0.1:61613", options...)
	if err != nil {
		fmt.Println("Connection erorr: " + err.Error())
	}

	//Subscribe creates a subscription on the STOMP server.
	//The first parameter "queue name"
	//The second parameter "The AckMode type is an enumeration of the acknowledgement modes for a"
	sub, err := conn.Subscribe("gin::app:request", stomp.AckMode(stomp.AckAuto))

	type MqBody struct {
		Function_name string
		Method        string
		Body          string
	}
	fmt.Println("Send All data success!111")
	go func() {
		//Circular acquisition
		for {
			fmt.Println("Send All data success!22")
			select {
			case v := <-sub.C:
				fmt.Println("Send All data success!")
				fmt.Println(string(v.Body))

				resp := MqBody{}
				json.Unmarshal(v.Body, &resp)
				resp.Body = "response data"
				back_context, _ := json.Marshal(resp)
				time.Sleep(time.Second * 2)
				conn.Send("gin::app:response", "text/plain", []byte(back_context))
			case <-time.After(time.Second * 3330):
				fmt.Println("Send All data success!4444")
				return
			}
		}
	}()

	for {
		time.Sleep(time.Second * 545)
		fmt.Println("Send All data success!333")
	}
}
