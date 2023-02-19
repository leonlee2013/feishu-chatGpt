package handlers

import (
	"context"
	"fmt"
	"start-feishubot/services"

	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

type PersonalMessageHandler struct {
	userCache services.UserCacheInterface
	msgCache  services.MsgCacheInterface
}

func (p PersonalMessageHandler) handle(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
	content := event.Event.Message.Content
	msgId := event.Event.Message.MessageId
	sender := event.Event.Sender
	openId := sender.SenderId.OpenId
	chatId := event.Event.Message.ChatId
	if p.msgCache.IfProcessed(*msgId) {
		fmt.Println("msgId", *msgId, "processed")
		return nil
	}
	p.msgCache.TagProcessed(*msgId)
	qParsed := parseContent(*content)
	if len(qParsed) == 0 {
		sendMsg(ctx, "🤖️：你想知道什么呢~", chatId)
		fmt.Println("msgId", *msgId, "message.text is empty")
		return nil
	}

	if qParsed == "/clear" || qParsed == "清除" {
		p.userCache.Clear(*openId)
		sendMsg(ctx, "🤖️：AI机器人已清除记忆", chatId)
		return nil
	}

	s := string([]rune(qParsed)[:3])
	if s == "画图：" {
		qParsed2 := string([]rune(qParsed[3:]))
		images, err := services.Images(qParsed2)
		ok := true
		if err != nil {
			sendMsg(ctx, fmt.Sprintf("🤖️：消息机器人摆烂了，请稍后再试～\n错误信息: %v", err), chatId)
			return nil
		}
		if len(images) == 0 {
			ok = false
		}
		if ok {
			// p.userCache.Set(*openId, qParsed, "")
			sendMsg(ctx, "画了2张图:", chatId)
			for _, image := range images {
				err := sendMsg(ctx, image, chatId)
				if err != nil {
					sendMsg(ctx, fmt.Sprintf("🤖️：消息机器人摆烂了，请稍后再试～\n错误信息: %v", err), chatId)
					return nil
				}
			}
		}
		return nil
	}

	prompt := p.userCache.Get(*openId)
	prompt = fmt.Sprintf("%s\nQ:%s\nA:", prompt, qParsed)
	completions, err := services.Completions(prompt)
	ok := true
	if err != nil {
		sendMsg(ctx, fmt.Sprintf("🤖️：消息机器人摆烂了，请稍后再试～\n错误信息: %v", err), chatId)
		return nil
	}
	if len(completions) == 0 {
		ok = false
	}
	if ok {
		p.userCache.Set(*openId, qParsed, completions)
		err := sendMsg(ctx, completions, chatId)
		if err != nil {
			sendMsg(ctx, fmt.Sprintf("🤖️：消息机器人摆烂了，请稍后再试～\n错误信息: %v", err), chatId)
			return nil
		}
	}
	return nil

}

var _ MessageHandlerInterface = (*PersonalMessageHandler)(nil)

func NewPersonalMessageHandler() MessageHandlerInterface {
	return &PersonalMessageHandler{
		userCache: services.GetUserCache(),
		msgCache:  services.GetMsgCache(),
	}
}
