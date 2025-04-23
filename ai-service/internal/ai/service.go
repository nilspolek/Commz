package ai

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/url"

	"github.com/google/uuid"
	"github.com/ollama/ollama/api"
	"team6-managing.mni.thm.de/Commz/ai-service/internal/utils"
)

const (
	BIG_MODEL        = "llama3.2:1b"
	SMALL_MODEL      = "llama3.2:1b"
	SUMMARIZE_PROMPT = "Make a one sentence summary:\n"
	CORRECT_PROMPT   = "Fix spelling and grammar. Don't say what you did. Return only new version:\n\n"
	REWRITE_PROMPT   = "Rephrase this text, no matter what. Don't say what you did. Return only new version:\n\n"
	GEUSS_PROMPT     = "Generate a JSON list of names for that topic. Your response is ALWAYS an array of strings. Nothing else.\n\n Topic:"
)

var options = map[string]interface{}{
	"num_predict": 1024,
}

type AiService struct {
	client *api.Client
}

func New(ollamaURL *url.URL) AiService {
	client := api.NewClient(ollamaURL, http.DefaultClient)
	return AiService{client: client}
}

func (ai *AiService) CorrectText(line string, answerFunc api.GenerateResponseFunc) error {
	return ai.manipulateTextByPrompt(line, CORRECT_PROMPT, answerFunc, BIG_MODEL)
}

func (ai *AiService) RewriteText(line string, answerFunc api.GenerateResponseFunc) error {
	return ai.manipulateTextByPrompt(line, REWRITE_PROMPT, answerFunc, BIG_MODEL)
}

func (ai *AiService) manipulateTextByPrompt(line, prompt string, answerFunc api.GenerateResponseFunc, model string) error {
	var (
		ctx = context.Background()
		req = &api.GenerateRequest{
			Model:   model,
			Prompt:  prompt + line,
			Options: options,
		}
	)
	err := ai.client.Generate(ctx, req, answerFunc)
	if err != nil {
		return err
	}
	return nil
}

func (ai *AiService) GenerateGuessWords(topic string) ([]string, error) {
	var (
		ctx = context.Background()
		req = &api.GenerateRequest{
			Model:   SMALL_MODEL,
			Prompt:  GEUSS_PROMPT + topic,
			Options: options,
			Stream:  new(bool),
			Format: []byte(`{
				"type": "object",
				"properties": {
					"words": {
						"type": "array",
						"maxItems": 10,
						"items": {
							"type": "string"
						}
					}
				},
				"required": ["words"]
			}`),
		}
		jsonString string
		answerFunc = func(resp api.GenerateResponse) error {
			jsonString += resp.Response
			return nil
		}
	)
	err := ai.client.Generate(ctx, req, answerFunc)
	if err != nil {
		return nil, err
	}

	var guessWords struct {
		Words []string `json:"words"`
	}
	err = json.Unmarshal([]byte(jsonString), &guessWords)
	if err != nil {
		return nil, err
	}

	return guessWords.Words, nil
}

func (ai *AiService) SummarizeChat(chat utils.Chat, answerFunc api.GenerateResponseFunc) (err error) {
	if len(chat.Messages) == 0 {
		return utils.NewError("Not enough messages to summarize the chat.", http.StatusBadRequest)
	}

	chatHistory := ""
	// Get last 10 messages using slices
	messages := chat.Messages
	if len(messages) > 10 {
		messages = messages[len(messages)-10:]
	}
	// Build chat history string
	for _, message := range messages {
		if message.SenderID.String() != uuid.Nil.String() {
			chatHistory += "User: \n" + message.Content + "\n\n"
		} else {
			chatHistory += "AI: \n" + message.Content + "\n\n"
		}
	}

	var (
		ctx = context.Background()
		req = &api.GenerateRequest{
			Model:   SMALL_MODEL,
			Prompt:  SUMMARIZE_PROMPT + chatHistory,
			Options: options,
		}
	)
	err = ai.client.Generate(ctx, req, answerFunc)

	if err != nil {
		log.Fatal(err)
	}
	return
}
func (ai *AiService) AskAI(prompt string, answerFunc api.GenerateResponseFunc) (_ error) {
	var (
		ctx = context.Background()
		req = &api.GenerateRequest{
			Model:   SMALL_MODEL,
			Prompt:  prompt,
			Options: options,
		}
	)
	err := ai.client.Generate(ctx, req, answerFunc)
	if err != nil {
		return err
	}
	return
}
