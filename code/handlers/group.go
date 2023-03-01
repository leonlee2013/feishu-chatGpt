package handlers

import (
	"context"
	"fmt"
	"log"
	"start-feishubot/services"

	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/spf13/viper"
)

type GroupMessageHandler struct {
	userCache services.UserCacheInterface
	msgCache  services.MsgCacheInterface
}

func (p GroupMessageHandler) handle(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
	ifMention := judgeIfMentionMe(event)
	if !ifMention {
		return nil
	}
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

	//-------------------------------------------------------------
	if qParsed[:2] == " !" {
		cmd := qParsed[2:]
		switch cmd {
		case "help":
			// sendMsg(ctx, helpText(), chatId)
			replyMsg(ctx, helpText(), msgId)
			return nil
		case "reset":
			p.userCache.Clear(*openId)
			// sendMsg(ctx, "🤖️：ChatGPT上下文已重置", chatId)
			replyMsg(ctx, "🤖️：ChatGPT上下文已重置", msgId)
			return nil
		case "info":
			// sendMsg(ctx, infoText(), chatId)
			replyMsg(ctx, infoText(), msgId)
			return nil
		default:
			if cmd[:5] == "draw:" {
				qParsed2 := cmd[5:]
				images, err := services.Images(qParsed2)
				if err != nil {
					// sendErrorMsg(ctx, err, chatId)
					replyErrorMsg(ctx, err, msgId)
					return nil
				}
				for _, image := range images {
					err := sendMsg(ctx, image, chatId)
					if err != nil {
						// sendErrorMsg(ctx, err, chatId)
						replyErrorMsg(ctx, err, msgId)
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
			replyErrorMsg(ctx, err, msgId)
			return nil
		}
		p.userCache.Set(*openId, qParsed, completions)
		err = replyMsg(ctx, completions, msgId)
		if err != nil {
			replyErrorMsg(ctx, err, msgId)
			return nil
		}
	} else {
		completions, newReply, err := services.HttpPostJson(qParsed, prompt)
		if err != nil {
			replyErrorMsg(ctx, err, msgId)
			return nil
		}
		p.userCache.Replace(*openId, newReply)
		err = replyMsg(ctx, completions, msgId)
		if err != nil {
			replyErrorMsg(ctx, err, msgId)
			return nil
		}
	}
	return nil
}

var _ MessageHandlerInterface = (*PersonalMessageHandler)(nil)

func replyErrorMsg(ctx context.Context, err error, msgId *string) {
	log.Printf("replyErrorMsg:%v msgId: %v", err, *msgId)
	replyMsg(ctx, fmt.Sprintf("🤖️：消息机器人出错，请稍后再试～\n错误信息: %v", err), msgId)
}

func NewGroupMessageHandler() MessageHandlerInterface {
	return &GroupMessageHandler{
		userCache: services.GetUserCache(),
		msgCache:  services.GetMsgCache(),
	}
}

func judgeIfMentionMe(event *larkim.P2MessageReceiveV1) bool {
	mention := event.Event.Message.Mentions
	if len(mention) != 1 {
		return false
	}
	return *mention[0].Name == viper.GetString("BOT_NAME")
}
