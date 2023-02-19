package path

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"os"
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

type CompletionPayload struct {
	Model            string   `json:"model"`
	Prompt           string   `json:"prompt"`
	Temperature      float64  `json:"temperature"`
	MaxTokens        int      `json:"max_tokens"`
	TopP             int      `json:"top_p"`
	FrequencyPenalty int      `json:"frequency_penalty"`
	PresencePenalty  float64  `json:"presence_penalty"`
	Stop             []string `json:"stop"`
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

	payload := CompletionPayload{
		Model:            "text-davinci-003",
		Prompt:           chatPrompt.Prompt,
		Temperature:      1,
		MaxTokens:        100,
		TopP:             1,
		FrequencyPenalty: 0,
		PresencePenalty:  0,
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
	if err != nil {
		log.Error().Msg(fmt.Sprintf("failed to do request %s", err))
		return
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(res.Body)

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return
	}

	log.Info().Msg(string(data))
	_, err = io.WriteString(w, string(data))
	if err != nil {
		log.Error().Msg("failed reply chat")
		return
	}

	//msgChan := make(chan string)
	//flusher, _ := w.(http.Flusher)
	//
	//defer func() {
	//	close(msgChan)
	//	msgChan = nil
	//	log.Info().Msg("closed sse connection")
	//}()
	//
	//for {
	//	select {
	//
	//	// message will be received here and printed
	//	case message := <-msgChan:
	//		log.Info().Msg(message)
	//		_, err := io.WriteString(w, message)
	//		if err != nil {
	//			log.Error().Msg("failed to response")
	//			return
	//		}
	//		flusher.Flush()
	//
	//	// connection is closed then defer will be executed
	//	case <-r.Context().Done():
	//		log.Info().Msg("end chat reply")
	//		return
	//	}
	//
	//}
}
