package handlers

import "team6-managing.mni.thm.de/Commz/ai-service/internal/utils"

// Summarize Chat represents the request body for creating a summary of a chat
type SummarzeChatReques struct {
	ChatID string `json:"chat_id"`
}

// SummarizeChatResponse represents the response body for creating a summary of a chat
type SummarizeChatResponse struct {
	Chat utils.Chat `json:"chat"`
}

// AskAiRequest represents the request body for asking the AI a question
type AskAiRequest struct {
	Prompt string `json:"prompt"`
}

// AskAiResponse represents the response body for asking the AI a question
type TextManipulationRequest struct {
	Text string `json:"text"`
}

// TextManipulationResponse represents the response body for asking the AI a question
type TextManipulationResponse struct {
	Text string `json:"text"`
}
