package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type ChatGPTBrowseReqBody struct {
	Message         string `json:"message"`
	ConversationId  string `json:"conversationId,omitempty"`
	ParentMessageId string `json:"parentMessageId,omitempty"`
}

type ChatGPTBrowseReplyBody struct {
	Answer         string `json:"answer"`
	MessageId      string `json:"messageId"`
	ConversationId string `json:"conversationId"`
}

func HttpPostJson(msg, lastReply string) (string, string, error) {
	reqBody := ChatGPTBrowseReqBody{
		Message: msg,
	}
	if lastReply != "" {
		lastReplyBody := &ChatGPTBrowseReplyBody{}
		err := json.Unmarshal([]byte(lastReply), lastReplyBody)
		if err != nil {
			log.Printf("err = %v\nlastReply = %v", err, lastReply)
		} else {
			reqBody.ConversationId = lastReplyBody.ConversationId
			reqBody.ParentMessageId = lastReplyBody.MessageId
		}
	}
	reqData, err := json.Marshal(reqBody)
	if err != nil {
		return "", "", err
	}
	log.Printf("request browser gtp json string : %v", string(reqData))

	jsonStr := reqData
	url := "http://localhost:8080/api/v1/ask"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	gptReplyBody := &ChatGPTBrowseReplyBody{}
	log.Println(string(body))
	err = json.Unmarshal(body, gptReplyBody)
	if err != nil {
		return "", "", err
	}
	if resp.StatusCode == 500 {
		if time.Now().Unix()%2 == 0 {
			SwitchToAPIKey()
			log.Printf("request: %v\n===========switch to api key==========", string(reqData))
		}
	}
	if resp.StatusCode/2 != 100 {
		statuscode := resp.StatusCode
		head := resp.Header
		fmt.Printf("statuscode = %v\n", statuscode)
		fmt.Printf("head = %v\n", head)
		fmt.Printf("resp = %#v\n", resp)
		return "", "", fmt.Errorf("ChatGPT Browser %s", resp.Status)
	}
	reply := gptReplyBody.Answer
	log.Printf("\nAsk:%s\nGPTBrowserGPT Reply:\n%s\n ", reqBody.Message, reply)
	return reply, string(body), nil
}
