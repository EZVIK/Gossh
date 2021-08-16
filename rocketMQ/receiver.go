package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/EZVIK/Gossh/service"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"os"
	"strings"
	"time"
)

type CMD2 struct {
	Namespace string              `validate:"required" json:"namespace"`
	IP        string              `validate:"required" json:"ip"`
	Command   string              `validate:"required" json:"cmd"`
	Result    map[string][]string `json:"result"`
}

var runtimeMap = service.NewConnectMap()

func main() {

	c, _ := rocketmq.NewPushConsumer(
		consumer.WithGroupName("testGroup"),
		consumer.WithNsResolver(primitive.NewPassthroughResolver([]string{"159.75.82.148:9876"})),
	)

	//
	p, _ := rocketmq.NewProducer(
		producer.WithNsResolver(primitive.NewPassthroughResolver([]string{"159.75.82.148:9876"})),
		producer.WithRetry(2),
	)
	err1 := p.Start()
	if err1 != nil {
		fmt.Printf("start producer error: %s", err1.Error())
		os.Exit(1)
	}

	topic := "SSH_REMOTE_CALL_COMMAND_LIST"
	topicResult := "SSH_REMOTE_CALL_COMMAND_RESULT"

	// subscribe Command list waiting msg
	err := c.Subscribe(topic, consumer.MessageSelector{}, func(ctx context.Context, ext ...*primitive.MessageExt) (consumer.ConsumeResult, error) {

		for i := range ext {
			cmd := new(CMD2)
			json.Unmarshal(ext[i].Body, &cmd)
			commands := strings.Split(cmd.Command, ";:;")
			ans, err := runtimeMap.RunCmd(cmd.IP, commands)
			cmd.Result = ans
			body, err := json.Marshal(cmd)
			msg := &primitive.Message{
				Topic: topicResult,
				Body:  body,
			}

			// return result
			res, err := p.SendSync(context.Background(), msg)

			if err != nil {
				fmt.Printf("send message error: %s\n", err)
			} else {
				fmt.Printf("send message success: result=%s\n", res.String())
			}
		}

		return consumer.ConsumeSuccess, nil
	})

	//update offset to broker success ?
	if err != nil {
		fmt.Println(err.Error())
	}
	// Note: start after subscribe
	err = c.Start()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}

	// shutdown
	time.Sleep(time.Hour)
	err = c.Shutdown()
	if err != nil {
		fmt.Printf("shutdown Consumer error: %s", err.Error())
	}
	err = p.Shutdown()
	if err != nil {
		fmt.Printf("shutdown producer error: %s", err.Error())
	}

}
