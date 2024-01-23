package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
)

type ChatCompletion struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int         `json:"index"`
		Message      Message     `json:"message"`
		Logprobs     interface{} `json:"logprobs"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	SystemFingerprint interface{} `json:"system_fingerprint"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Subtitle struct {
	Number int
	Time   string
	Quote  string
}

func CheckFile(s *Settings) error {
	if _, err := os.Stat(*s.InputFile); err != nil {
		return errors.New(fmt.Sprintf("input file \"%s\" not exists", *s.InputFile))
	}
	return nil
}

func ParseStrings(s *Settings) (error, []Subtitle) {
	text, err := os.ReadFile(*s.InputFile)
	if err != nil {
		return errors.New("unable to read input file"), nil
	}
	filtered := bytes.ReplaceAll(text, []byte("\r"), []byte{})
	pattern := regexp.MustCompile(`(?m)(\d+)\n(.+\s-->\s.+)\n([\s\S]*?)\n($|\z)`)
	matches := pattern.FindAllStringSubmatch(string(filtered), -1)

	subtitles := make([]Subtitle, len(matches))

	var i = 0

	for _, match := range matches {
		number, _ := strconv.Atoi(string(match[1]))

		time := string(match[2])
		quote := string(match[3])
		quote = string(bytes.ReplaceAll([]byte(quote), []byte("\n"), []byte{}))

		subtitles[i] = Subtitle{
			Number: number,
			Time:   time,
			Quote:  quote,
		}
		i++
	}
	return nil, subtitles
}

func CallAPI(s *Settings, subs *[]Subtitle) (error, []Subtitle) {
	result := make([]Subtitle, len(*subs))
	client := &http.Client{}
	var i = 0
	for _, sub := range *subs {
		requestBody := fmt.Sprintf(`{"messages": [{"role": "user", "content": "%s"}, {"role": "system", "content": "%s"}], "model": "gpt-3.5-turbo", "temperature": 0.5, "max_tokens": 1024}`, sub.Quote, string(*s.Prompt))
		req, err := http.NewRequest("POST", s.ApiURL, bytes.NewBuffer([]byte(requestBody)))
		reqBody := req.Body
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+string(*s.ApiKey))
		if err != nil {
			return err, nil
		}

		resp, err := client.Do(req)
		if err != nil {
			return err, nil
		}
		defer resp.Body.Close()

		fmt.Printf("Translating line %d: %s\n\n", i, sub.Quote)

		responseBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err, nil
		}

		var chatData ChatCompletion

		err = json.Unmarshal([]byte(responseBody), &chatData)
		if err != nil {
			return errors.New(fmt.Sprintf("ERROR with: %s \n\nresponse: %s", reqBody, responseBody)), nil
		}

		if len(chatData.Choices) == 0 {
			fmt.Printf("ERROR with: %s \n\nresponse: %s", reqBody, responseBody)
		} else {
			result[i] = Subtitle{Number: sub.Number, Time: sub.Time, Quote: chatData.Choices[0].Message.Content}
		}
		i++

	}
	return nil, result
}
