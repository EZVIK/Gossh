package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"os"
	"strconv"
)

type CMD struct {
	Namespace string              `validate:"required" json:"namespace"`
	IP        string              `validate:"required" json:"ip"`
	Command   string              `validate:"required" json:"cmd"`
	Result    map[string][]string `json:"result"`
}

func main() {
	fmt.Println("Starting Producer....")
	p, _ := rocketmq.NewProducer(
		producer.WithNsResolver(primitive.NewPassthroughResolver([]string{"159.75.82.148:9876"})),
		producer.WithRetry(2),
	)

	err := p.Start()
	if err != nil {
		fmt.Printf("start producer error: %s", err.Error())
		os.Exit(1)
	}

	topic := "SSH_REMOTE_CALL_COMMAND_LIST"
	ans := make(map[string][]string)
	for i := 0; i < 1; i++ {
		command := CMD{Namespace: strconv.Itoa(i), IP: "159.75.82.148", Command: "mkdir TTTEST;:;echo WTF;:;ls -ls" + strconv.Itoa(i), Result: ans}
		body, err := json.Marshal(command)
		if err != nil {
			fmt.Printf("Marshal error %s", err.Error())
			os.Exit(1)
		}
		msg := &primitive.Message{
			Topic: topic,
			Body:  body,
		}

		res, err := p.SendSync(context.Background(), msg)
		if err != nil {
			fmt.Printf("send message error: %s\n", err)
		} else {
			fmt.Printf("send message success: result=%s\n", res.String())
		}
	}

	err = p.Shutdown()
	if err != nil {
		fmt.Printf("shutdown producer error: %s", err.Error())
	}
}
