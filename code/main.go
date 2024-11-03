package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	sdkginext "github.com/larksuite/oapi-sdk-gin"
	larkcard "github.com/larksuite/oapi-sdk-go/v3/card"
	larkevent "github.com/larksuite/oapi-sdk-go/v3/event"
	"github.com/sirupsen/logrus"
	"start-feishubot/handlers"
	"start-feishubot/initialization"
	"start-feishubot/logger"

	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	larkws "github.com/larksuite/oapi-sdk-go/v3/ws"
	"github.com/spf13/pflag"
	"start-feishubot/services/openai"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	//logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetReportCaller(true)
	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	//logrus.SetOutput(os.Stdout)

	// Only log the warning severity or above.
}

func main() {
	initialization.InitRoleList()
	pflag.Parse()
	config := initialization.GetConfig()
	initialization.LoadLarkClient(*config)
	gpt := openai.NewChatGPT(*config)
	handlers.InitHandlers(gpt, *config)
	go httpStart(config)
	clientStart(config)
}

func httpStart(config *initialization.Config) {
	eventHandler := dispatcher.NewEventDispatcher(
		config.FeishuAppVerificationToken, config.FeishuAppEncryptKey).
		OnP2MessageReceiveV1(handlers.Handler).
		OnP2MessageReadV1(func(ctx context.Context, event *larkim.P2MessageReadV1) error {
			logger.Debugf("http_req %v", event.RequestURI)
			return handlers.ReadHandler(ctx, event)
		})

	cardHandler := larkcard.NewCardActionHandler(
		config.FeishuAppVerificationToken, config.FeishuAppEncryptKey,
		handlers.CardHandler())

	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.POST("/webhook/event",
		sdkginext.NewEventHandlerFunc(eventHandler))
	r.POST("/webhook/card",
		sdkginext.NewCardActionHandlerFunc(
			cardHandler))

	if err := initialization.StartServer(*config, r); err != nil {
		logger.Fatalf("failed to start server: %v", err)
	}

}

// https://open.feishu.cn/document/server-docs/event-subscription-guide/event-subscription-configure-/request-url-configuration-case#d286cc88
func clientStart(config *initialization.Config) {
	// 注册事件回调，OnP2MessageReceiveV1 为接收消息 v2.0；OnCustomizedEvent 内的 message 为接收消息 v1.0。
	eventHandler := dispatcher.NewEventDispatcher("", "").
		OnP2MessageReceiveV1(handlers.Handler).
		OnP2MessageReadV1(func(ctx context.Context, event *larkim.P2MessageReadV1) error {
			logger.WithField("req", event.RequestURI).Info("from_websocket=============>>>")
			return handlers.ReadHandler(ctx, event)
		}).
		OnCustomizedEvent("这里填入你要自定义订阅的 event 的 key，例如 out_approval", func(ctx context.Context, event *larkevent.EventReq) error {
			fmt.Printf("[ OnCustomizedEvent access ], type: message, data: %s\n", string(event.Body))
			return nil
		})
	// 创建Client
	cli := larkws.NewClient(config.FeishuAppId, config.FeishuAppSecret,
		larkws.WithEventHandler(eventHandler),
		//larkws.WithLogLevel(larkcore.LogLevelDebug),
	)
	// 启动客户端
	err := cli.Start(context.Background())
	if err != nil {
		panic(err)
	}
}
