package handlers

import (
	"context"
	"fmt"
	"log"
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
		return nil
	}
	//-------------------------------------------------------------
	if qParsed[:1] == "!" {
		cmd := qParsed[1:]
		switch cmd {
		case "help":
			sendMsg(ctx, helpText(), chatId)
			return nil
		case "reset":
			p.userCache.Clear(*openId)
			sendMsg(ctx, "🤖️：ChatGPT上下文已重置", chatId)
			return nil
		case "info":
			sendMsg(ctx, infoText(), chatId)
			return nil
		case "switch":
			var text string
			if services.IsAPIKey() {
				services.SwitchToBrowser()
				text = "ChatGPT从APIKey切换到Browser"
			} else {
				services.SwitchToAPIKey()
				text = "ChatGPT从Browser切换到APIKey"
			}
			log.Println(text)
			sendMsg(ctx, text, chatId)
			return nil
		default:
			if cmd[:5] == "draw:" {
				qParsed2 := cmd[5:]
				images, err := services.Images(qParsed2)
				if err != nil {
					sendErrorMsg(ctx, err, chatId)
					return nil
				}
				for _, image := range images {
					err := sendMsg(ctx, image, chatId)
					if err != nil {
						sendErrorMsg(ctx, err, chatId)
						return nil
					}
				}
				return nil
			}
		}
	}
	prompt := p.userCache.Get(*openId)
	if services.IsAPIKey() {
		prompt = fmt.Sprintf("%s\nQ:%s\nA:", prompt, qParsed)
		completions, err := services.Completions(prompt)
		if err != nil {
			sendErrorMsg(ctx, err, chatId)
			return nil
		}
		p.userCache.Set(*openId, qParsed, completions)
		err = sendMsg(ctx, completions, chatId)
		if err != nil {
			sendErrorMsg(ctx, err, chatId)
			return nil
		}
	} else {
		completions, newReply, err := services.HttpPostJson(qParsed, prompt)
		if err != nil {
			sendErrorMsg(ctx, err, chatId)
			return nil
		}
		p.userCache.Replace(*openId, newReply)
		err = sendMsg(ctx, completions, chatId)
		if err != nil {
			sendErrorMsg(ctx, err, chatId)
			return nil
		}
	}
	return nil
}

var _ MessageHandlerInterface = (*PersonalMessageHandler)(nil)

func sendErrorMsg(ctx context.Context, err error, chatId *string) {
	log.Printf("sendErrorMsg:%v msgId: %v", err, *chatId)
	sendMsg(ctx, fmt.Sprintf("🤖️：消息机器人出错，请稍后再试～\n错误信息: %v", err), chatId)
}

func NewPersonalMessageHandler() MessageHandlerInterface {
	return &PersonalMessageHandler{
		userCache: services.GetUserCache(),
		msgCache:  services.GetMsgCache(),
	}
}
