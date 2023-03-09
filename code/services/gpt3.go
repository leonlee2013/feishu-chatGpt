package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/atomic"
)

const (
	BASEURL     = "https://api.openai.com/v1/"
	maxTokens   = 2000
	temperature = 0.7
	engine      = "text-davinci-003"
)

// ChatGPTResponseBody 请求体
type ChatGPTResponseBody struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int                    `json:"created"`
	Model   string                 `json:"model"`
	Choices []ChoiceItem           `json:"choices"`
	Usage   map[string]interface{} `json:"usage"`
}

type ChatGPTTurboResponseBody struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int                    `json:"created"`
	Model   string                 `json:"model"`
	Choices []TurboChoiceItem      `json:"choices"`
	Usage   map[string]interface{} `json:"usage"`
}

type ChatGPTImageResponseBody struct {
	ID      string                 `json:"id"`
	Created int                    `json:"created"`
	Data    []ChoiceImageItem      `json:"data"`
	Usage   map[string]interface{} `json:"usage"`
}

type ChoiceImageItem struct {
	URL string `json:"url"`
}

type ChoiceItem struct {
	Text         string `json:"text"`
	Index        int    `json:"index"`
	Logprobs     int    `json:"logprobs"`
	FinishReason string `json:"finish_reason"`
}

type TurboChoiceItem struct {
	Message      TurboMessage `json:"message"`
	Index        int          `json:"index"`
	Logprobs     int          `json:"logprobs"`
	FinishReason string       `json:"finish_reason"`
}

// ChatGPTRequestBody 响应体
type ChatGPTRequestBody struct {
	Model            string  `json:"model"`
	Prompt           string  `json:"prompt"`
	MaxTokens        int     `json:"max_tokens"`
	Temperature      float32 `json:"temperature"`
	TopP             int     `json:"top_p"`
	FrequencyPenalty int     `json:"frequency_penalty"`
	PresencePenalty  int     `json:"presence_penalty"`
}

type ChatGPTTurboRequestBody struct {
	Model            string         `json:"model"`
	Messages         []TurboMessage `json:"messages"`
	MaxTokens        int            `json:"max_tokens"`
	Temperature      float32        `json:"temperature"`
	TopP             int            `json:"top_p"`
	FrequencyPenalty int            `json:"frequency_penalty"`
	PresencePenalty  int            `json:"presence_penalty"`
}

type TurboMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatGPTImageRequestBody struct {
	// Model            string  `json:"model"`
	// MaxTokens        int     `json:"max_tokens"`
	Prompt string `json:"prompt"`
	N      int    `json:"n"`
	Size   string `json:"size"`
	// Temperature      float32 `json:"temperature"`
	// TopP             int     `json:"top_p"`
	// FrequencyPenalty int     `json:"frequency_penalty"`
	// PresencePenalty  int     `json:"presence_penalty"`
}

// var is_api_key = atomic.NewBool(false)
var is_api_key = atomic.NewBool(true)

func IsAPIKey() bool {
	return is_api_key.Load()
}

func SwitchToAPIKey() {
	is_api_key.CAS(false, true)
}

func SwitchToBrowser() {
	is_api_key.CAS(true, false)
}

func Completions(msg string) (string, error) {
	requestBody := ChatGPTRequestBody{
		Model:            engine,
		Prompt:           msg,
		MaxTokens:        maxTokens,
		Temperature:      temperature,
		TopP:             1,
		FrequencyPenalty: 0,
		PresencePenalty:  0,
	}
	requestData, err := json.Marshal(requestBody)

	if err != nil {
		return "", err
	}
	log.Printf("request gtp json string : %v", string(requestData))
	req, err := http.NewRequest("POST", BASEURL+"completions", bytes.NewBuffer(requestData))
	if err != nil {
		return "", err
	}

	apiKey := viper.GetString("OPENAI_KEY")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	client := &http.Client{Timeout: 110 * time.Second}
	response, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode/2 != 100 {
		return "", fmt.Errorf("gtp api %s", response.Status)
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	gptResponseBody := &ChatGPTResponseBody{}
	log.Println(string(body))
	err = json.Unmarshal(body, gptResponseBody)
	if err != nil {
		return "", err
	}

	var reply string
	if len(gptResponseBody.Choices) > 0 {
		reply = gptResponseBody.Choices[0].Text
	}
	log.Printf("gpt response text: %s \n", reply)
	return reply, nil
}

func genMessages(msg, ask string) []TurboMessage {
	ret := []TurboMessage{}
	qaStrArr := strings.Split(msg, "Q:")
	for _, qaStr := range qaStrArr[1:] {
		qa := strings.Split(qaStr, "A:")
		if len(qa) != 2 {
			log.Printf("qaStr:%v", qaStr)
			return []TurboMessage{{"user", ask}}
		}
		question := strings.Trim(qa[0], "\n")
		answer := strings.Trim(qa[1], "\n")
		if question == "" || answer == "" {
			log.Printf("question:%v answer:%v", question, answer)
			return []TurboMessage{{"user", ask}}
		}
		ret = append(ret, TurboMessage{Role: "user", Content: question})
		ret = append(ret, TurboMessage{Role: "assistant", Content: answer})
	}
	ret = append(ret, TurboMessage{Role: "user", Content: ask})
	return ret
}

func Chat(msg, ask string) (string, error) {
	requestBody := ChatGPTTurboRequestBody{
		Model: "gpt-3.5-turbo",
		// Messages:         []TurboMessage{{"user", msg}},
		Messages:         genMessages(msg, ask),
		MaxTokens:        maxTokens,
		Temperature:      temperature,
		TopP:             1,
		FrequencyPenalty: 0,
		PresencePenalty:  0,
	}
	requestData, err := json.Marshal(requestBody)

	if err != nil {
		return "", err
	}
	log.Printf("request gtp turbo json string : %v", string(requestData))
	req, err := http.NewRequest("POST", BASEURL+"chat/completions", bytes.NewBuffer(requestData))
	if err != nil {
		return "", err
	}
	apiKey := viper.GetString("OPENAI_KEY")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	client := &http.Client{Timeout: 110 * time.Second}
	response, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode/2 != 100 {
		log.Printf("response======> %v", response)
		return "", fmt.Errorf("gpt-3.5-turbo api %s", response.Status)
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	gptTurboResponseBody := &ChatGPTTurboResponseBody{}
	log.Println(string(body))
	err = json.Unmarshal(body, gptTurboResponseBody)
	if err != nil {
		return "", err
	}

	var reply string
	if len(gptTurboResponseBody.Choices) > 0 {
		reply = gptTurboResponseBody.Choices[0].Message.Content
	}
	// log.Printf("gpt turbo response text: %s \n", reply)
	log.Printf("\nAsk:%s\nTurbo Reply:\n%s\n ", ask, reply)
	return reply, nil
}

func Images(msg string) ([]string, error) {
	requestBody := ChatGPTImageRequestBody{
		// Model:            engine,
		Prompt: msg,
		// MaxTokens:        maxTokens,
		// Temperature:      temperature,
		N:    1,
		Size: "512x512",
		// TopP:             1,
		// FrequencyPenalty: 0,
		// PresencePenalty:  0,
	}
	requestData, err := json.Marshal(requestBody)

	if err != nil {
		return nil, err
	}
	log.Printf("request gtp json string : %v", string(requestData))
	req, err := http.NewRequest("POST", BASEURL+"images/generations", bytes.NewBuffer(requestData))
	if err != nil {
		return nil, err
	}

	apiKey := viper.GetString("OPENAI_KEY")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	client := &http.Client{Timeout: 110 * time.Second}
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode/2 != 100 {
		fmt.Printf("response = %#v\n", response)
		return nil, fmt.Errorf("gpt api %s", response.Status)
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	gptResponseBody := &ChatGPTImageResponseBody{}
	log.Println(string(body))
	err = json.Unmarshal(body, gptResponseBody)
	if err != nil {
		return nil, err
	}

	// var reply string
	reply := make([]string, 0)
	if len(gptResponseBody.Data) > 0 {
		for i := 0; i < len(gptResponseBody.Data); i++ {
			reply = append(reply, gptResponseBody.Data[i].URL)
		}

		// reply = gptResponseBody.Data[0].URL
	}
	log.Printf("gpt response text: %s \n", reply)
	return reply, nil
}

func FormatQuestion(question string) string {
	return "Answer:" + question
}
