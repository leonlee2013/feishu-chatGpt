package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"start-feishubot/initialization"
	"start-feishubot/services"
	"strings"

	"github.com/google/uuid"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

func helpText() string {
	var buffer bytes.Buffer
	buffer.WriteString("!help  - 帮助\n")
	buffer.WriteString("!reset - 重置上下文\n")
	buffer.WriteString("!info  - 信息（包括接入方式）\n")
	buffer.WriteString("!draw:+<prompt> - DALL·E画图")
	ret := buffer.String()
	log.Println(ret)
	return ret
}

func infoText() string {
	var buffer bytes.Buffer
	if services.IsAPIKey() {
		buffer.WriteString("接入方式：ChatGPTAPI Key接入\n")
		buffer.WriteString("优点：响应快, 稳定\n")
		buffer.WriteString("缺点：收费，质量不如网页版\n")
		buffer.WriteString("其他：当网页版被限流时使用\n")
	} else {
		buffer.WriteString("接入方式：ChatGPT Browser接入\n")
		buffer.WriteString("优点：网页版，真实的ChatGPT\n")
		buffer.WriteString("缺点：响应慢, 不稳定，经常被限流\n")
		buffer.WriteString("其他：被限流时，会切换到API Key")
	}
	ret := buffer.String()
	log.Println(ret)
	return ret
}

func replyMsg(ctx context.Context, msg string, msgId *string) error {
	fmt.Println("group-->", msgId)
	msg, i := processMessage(msg)
	if i != nil {
		return i
	}
	client := initialization.GetLarkClient()
	content := larkim.NewTextMsgBuilder().
		Text(msg).
		Build()

	resp, err := client.Im.Message.Reply(ctx, larkim.NewReplyMessageReqBuilder().
		MessageId(*msgId).
		Body(larkim.NewReplyMessageReqBodyBuilder().
			MsgType(larkim.MsgTypeText).
			Uuid(uuid.New().String()).
			Content(content).
			Build()).
		Build())

	// 处理错误
	if err != nil {
		fmt.Println(err)
		return err
	}

	// 服务端错误处理
	if !resp.Success() {
		fmt.Println(resp.Code, resp.Msg, resp.RequestId())
		return err
	}
	return nil

}
func sendMsg(ctx context.Context, msg string, chatId *string) error {
	fmt.Println("persional-->", chatId)
	msg, i := processMessage(msg)
	if i != nil {
		return i
	}
	client := initialization.GetLarkClient()
	content := larkim.NewTextMsgBuilder().
		Text(msg).
		Build()

	//fmt.Println("content", content)

	resp, err := client.Im.Message.Create(ctx, larkim.NewCreateMessageReqBuilder().
		ReceiveIdType(larkim.ReceiveIdTypeChatId).
		Body(larkim.NewCreateMessageReqBodyBuilder().
			MsgType(larkim.MsgTypeText).
			ReceiveId(*chatId).
			Content(content).
			Build()).
		Build())

	// 处理错误
	if err != nil {
		fmt.Println(err)
		return err
	}

	// 服务端错误处理
	if !resp.Success() {
		fmt.Println(resp.Code, resp.Msg, resp.RequestId())
		return err
	}
	return nil
}
func msgFilter(msg string) string {
	//replace @到下一个非空的字段 为 ''
	regex := regexp.MustCompile(`@[^ ]*`)
	return regex.ReplaceAllString(msg, "")

}
func parseContent(content string) string {
	//"{\"text\":\"@_user_1  hahaha\"}",
	//only get text content hahaha
	var contentMap map[string]interface{}
	err := json.Unmarshal([]byte(content), &contentMap)
	if err != nil {
		fmt.Println(err)
	}
	text := contentMap["text"].(string)
	return msgFilter(text)
}
func processMessage(msg interface{}) (string, error) {
	msg = strings.TrimSpace(msg.(string))
	msgB, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	msgStr := string(msgB)

	if len(msgStr) >= 2 {
		msgStr = msgStr[1 : len(msgStr)-1]
	}

	return msgStr, nil
}
