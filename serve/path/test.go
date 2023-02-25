package path

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"os"
	"time"
)

func GetHomepage(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("GetHomepage")
	_, err := io.WriteString(w, "home")
	if err != nil {
		log.Error().Msg("Failed GetHomepage")
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func GetHello(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("GetHello")
	_, err := io.WriteString(w, "hello from server")
	if err != nil {
		log.Error().Msg("Failed GetHello")
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func GetTest(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
}

type CompletionReq struct {
	Model            string   `json:"model"`
	Prompt           string   `json:"prompt"`
	Temperature      float64  `json:"temperature"`
	MaxTokens        int      `json:"max_tokens"`
	TopP             int      `json:"top_p"`
	FrequencyPenalty int      `json:"frequency_penalty"`
	PresencePenalty  float64  `json:"presence_penalty"`
	Stop             []string `json:"stop"`
}
type CompletionResp struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Text         string `json:"text"`
		Index        int    `json:"index"`
		Logprobs     any    `json:"logprobs"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type ChatPrompt struct {
	Prompt string `json:"prompt"`
}

func ChatWithGpt(w http.ResponseWriter, r *http.Request) {
	var chatPrompt ChatPrompt
	err := json.NewDecoder(r.Body).Decode(&chatPrompt)
	if err != nil || chatPrompt.Prompt == "" {
		log.Error().Msg(fmt.Sprintf("failed to parse chat %s", err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	payload := CompletionReq{
		Model:            "text-davinci-003",
		Prompt:           chatPrompt.Prompt,
		Temperature:      1,
		MaxTokens:        1024,
		TopP:             1,
		FrequencyPenalty: 0,
		PresencePenalty:  0,
		//Stop:             []string{"\n"},
	}

	payloadBytes, err := json.Marshal(payload)

	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest("POST", os.Getenv("OPENAI_API_URL"), body)
	if err != nil {
		log.Error().Msg(fmt.Sprintf("failed to create request %s", err))
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", os.ExpandEnv("Bearer "+os.Getenv("OPENAI_API_KEY")))

	res, err := http.DefaultClient.Do(req)
	if err != nil || res.StatusCode != 200 {
		log.Error().Msg(fmt.Sprintf("failed to do request to openai %d %s", res.StatusCode, err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(res.Body)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return
	}
	log.Info().Msg(fmt.Sprintf("received %s", string(raw)))
	var data CompletionResp
	err = json.Unmarshal(raw, &data)
	if err != nil {
		log.Error().Msg(fmt.Sprintf("failed parse response from openai %s", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	msgChan := make(chan string, 1024)
	timeout := time.After(3 * time.Second)
	flusher, _ := w.(http.Flusher)

	defer func() {
		close(msgChan)
		msgChan = nil
		log.Info().Msg("closed sse connection to server")
	}()

	// split to chunk -> send to channel
	go func() {
		text := data.Choices[0].Text
		chunkSize := 3
		for i := 0; i < len(text); i += chunkSize {
			end := i + chunkSize
			if end > len(text) {
				end = len(text)
			}
			msgChan <- text[i:end]
		}
	}()

	for {
		select {

		// message will be received here and printed
		case message := <-msgChan:
			log.Info().Msg(message)
			_, err := io.WriteString(w, message)
			if err != nil {
				log.Error().Msg("failed to response")
				return
			}
			time.Sleep(100 * time.Millisecond)
			flusher.Flush()

		// connection is closed then defer will be executed
		case <-r.Context().Done():
			log.Info().Msg("client end chat reply")
			return
		case <-timeout:
			log.Info().Msg("end chat reply")
			return
		}
	}
}
