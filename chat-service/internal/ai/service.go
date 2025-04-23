package ai

import (
	"strings"

	"github.com/nilspolek/DevOps/Chat/internal/utils"
)

type AiService struct {
	gateway string
}

func New(gateway string) AiService {
	return AiService{
		gateway: gateway,
	}
}

func (ai *AiService) AskAI(prompt string, response func(response utils.GenerateResponse)) error {
	gateway := strings.Replace(ai.gateway, "http://", "ws://", 1)
	return utils.WebsocketRequest(gateway+"/ai/ask", AskAiRequest{
		Prompt: prompt,
	}, response)
}

func (ai *AiService) GuessWords(topic string) ([]string, error) {
	result, err := utils.PostRequest[TextManipulationRequest, []string](ai.gateway+"/ai/guess", TextManipulationRequest{
		Text: topic,
	})
	if err != nil {
		return nil, err
	}
	return *result, nil
}
