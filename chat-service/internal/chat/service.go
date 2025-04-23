package chat

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nilspolek/DevOps/Chat/internal/utils"
)

type Guess struct {
	Word   string
	UserId uuid.UUID
}

type GuessingGame struct {
	words       []string
	gussedWords map[string]struct{}
	chatId      uuid.UUID
	guesses     []Guess
}

type ChatService struct {
	storage       utils.Storage
	auth          utils.AuthService
	ai            utils.AiService
	guessingNames map[uuid.UUID]GuessingGame
}

func New(storage utils.Storage, auth utils.AuthService, ai utils.AiService) ChatService {
	return ChatService{
		storage:       storage,
		auth:          auth,
		ai:            ai,
		guessingNames: make(map[uuid.UUID]GuessingGame), // map chat id to guessing game, initially empty
	}
}

func (c *ChatService) GetChats(user uuid.UUID) ([]utils.Chat, error) {
	chats, err := c.storage.GetChats(user)
	if err != nil {
		return nil, err
	}

	// check if the ai chat is included, if not append it
	contains := slices.ContainsFunc(chats, func(x utils.Chat) bool {
		return x.ID.String() == user.String()
	})

	if !contains {
		chats = append(chats, utils.Chat{
			ID:       user,
			Messages: []utils.Message{},
			Members:  []uuid.UUID{user},
			Name:     "AI",
		})
	}

	return chats, err
}

func (c *ChatService) GetMessages(userId uuid.UUID, chatId uuid.UUID, limit, offset int) ([]utils.Message, error) {

	if !c.MemberOfChat(userId, chatId) {
		return nil, utils.NewError("User is not a member of the chat", http.StatusUnauthorized)
	}

	return c.storage.GetChatMessages(chatId, limit, offset)
}

func (c *ChatService) MemberOfChat(userId uuid.UUID, chatId uuid.UUID) bool {
	return c.storage.MemberOfChat(userId, chatId) == nil
}

// the AI chat is a special chat where the an AI answers instead of a differnt user.
// the AI is not a user and has therefore not a user id
// to account for the missing user id the chat id with the ai will just be the targeted user id
func (c *ChatService) AnswerAiChat(userId uuid.UUID, content string) (*utils.Message, error) {
	// ensure that the chat exists
	_, err := c.storage.GetChat(userId)

	// chat does not exist yet => we have to create a new one
	if err != nil {

		chat := &utils.Chat{
			ID:         userId,
			Members:    []uuid.UUID{userId},
			CreatedAt:  time.Now(),
			LastActive: time.Now(),
			Name:       "AI",
		}

		err = c.storage.CreateOrUpdateChat(*chat)
		if err != nil {
			return nil, err
		}
	}

	// ask the ai for a response async and send the message
	go func() {
		messages, _ := c.storage.GetChatMessages(userId, 10, 0)

		context := ""
		for i, message := range messages {
			if i > 10 { // keep the context short
				break
			}
			if message.SenderID.String() == userId.String() {
				context += "User: \n" + message.Content + "\n\n"
			} else {
				context += "AI: \n" + message.Content + "\n\n"
			}
		}

		msg := utils.Message{
			ChatID:    userId,
			Content:   "",
			ID:        uuid.New(),
			SenderID:  uuid.MustParse(utils.AIChat),
			Timestamp: time.Now(),
			UpdatedAt: time.Now(),
			Read:      true,
		}

		// right now the ask ai is context unaware this is super shit, we definitly have to change that
		err := c.ai.AskAI(context+"User: "+content, func(response utils.GenerateResponse) {
			msg.Content = msg.Content + response.Response
			msg.UpdatedAt = time.Now()
			c.storage.UpdateMessage(msg)
		})

		if err != nil || msg.Content == "" {
			msg = utils.Message{
				ChatID:    userId,
				Content:   "Sorry, something went wrong.",
				ID:        uuid.New(),
				SenderID:  uuid.MustParse(utils.AIChat),
				Timestamp: time.Now(),
				UpdatedAt: time.Now(),
				Read:      true,
			}
		}
		c.storage.SaveMessage(msg)

	}()

	message := utils.Message{
		ID:        uuid.New(),
		ChatID:    userId,
		SenderID:  userId,
		Timestamp: time.Now(),
		UpdatedAt: time.Now(),
		Content:   content,
	}
	err = c.storage.SaveMessage(message)
	if err != nil {
		return nil, err
	}

	err = c.storage.UpdateChatActivity(userId)
	return &message, err
}

func (c *ChatService) Command(userId uuid.UUID, chatId uuid.UUID, content, command string) (*utils.Message, error) {
	if len(content) == 0 {
		return nil, utils.NewError("message content is empty", http.StatusBadRequest)
	}

	// handle ai chat
	if chatId.String() == userId.String() {
		return nil, utils.NewError("command not supported for AI chat", http.StatusBadRequest)
	}

	var asyncFunction func() = nil

	switch command {
	case "guess":
		asyncFunction = func() {

			time.Sleep(100 * time.Millisecond)

			game, exists := c.guessingNames[chatId]
			if !exists {

				startGameMsg := utils.Message{
					ID:        uuid.New(),
					ChatID:    chatId,
					SenderID:  uuid.MustParse(utils.AIChat),
					Timestamp: time.Now(),
					UpdatedAt: time.Now(),
					Content:   "Starting a new game. Please wait a moment...",
				}
				c.storage.SaveMessage(startGameMsg)

				// check if the chat is already in the guessing game
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)

				go func() {
					<-ctx.Done()
					game, ok := c.guessingNames[chatId]
					if !ok {
						cancel()
						return
					}

					// send the message that the user has guessed all words
					// Count guesses per user
					userGuesses := make(map[uuid.UUID]int)
					for _, guess := range game.guesses {
						userGuesses[guess.UserId]++
					}

					// Create leaderboard text
					leaderboard := "# Final Leaderboard:\n\n"
					for userId, count := range userGuesses {
						leaderboard += fmt.Sprintf("User @%s: %d words\n\n", userId.String(), count)
					}

					if len(game.guesses) == 0 {
						leaderboard = "No one guessed any words."
					}

					// send message that the game has ended
					message := utils.Message{
						ID:        uuid.New(),
						ChatID:    chatId,
						SenderID:  uuid.MustParse(utils.AIChat),
						Timestamp: time.Now(),
						UpdatedAt: time.Now(),
						Content:   fmt.Sprintf("# Game Over! ⌛\n\nThe guessing game has ended. No one guessed all words in time.\n\n**The words were:** %v \n\n%v", strings.Join(c.guessingNames[chatId].words, ", "), leaderboard),
					}
					c.storage.SaveMessage(message)
					delete(c.guessingNames, chatId)
					cancel()
				}()

				// create a new guessing game
				guesses, err := c.ai.GuessWords(content)
				if err != nil {
					message := utils.Message{
						ID:        uuid.New(),
						ChatID:    chatId,
						SenderID:  uuid.MustParse(utils.AIChat),
						Timestamp: time.Now(),
						UpdatedAt: time.Now(),
						Content:   "Failed to generate guesses for that topic. Please try again.",
					}
					c.storage.SaveMessage(message)
					return
				}

				c.guessingNames[chatId] = GuessingGame{
					words:       guesses,
					chatId:      chatId,
					guesses:     []Guess{},
					gussedWords: map[string]struct{}{},
				}

				// send a message that a new game has started
				message := utils.Message{
					ID:        uuid.New(),
					ChatID:    chatId,
					SenderID:  uuid.MustParse(utils.AIChat),
					Timestamp: time.Now(),
					UpdatedAt: time.Now(),
					Content:   fmt.Sprintf("# A new guessing game has started about **%s**. \nThere are **%v** words to guess.\n\nUse: `/guess [word]` - to guess the words.\n\nGood luck guessing!", content, len(guesses)),
				}
				c.storage.SaveMessage(message)

				return
			}

			// check if the guessed word is correct
			for _, word := range game.words {
				word = strings.ToLower(word)
				content = strings.ToLower(content)
				// check if the word has not been guessed yet and if the word is correct
				if _, exists := game.gussedWords[word]; !exists && word == content {
					// add the guess to the list of guesses
					game.guesses = append(game.guesses, Guess{
						Word:   content,
						UserId: userId,
					})

					c.guessingNames[chatId] = game

					// check if the user has guessed all words
					if len(game.guesses) == len(game.words) {
						// send the message that the user has guessed all words
						// Count guesses per user
						userGuesses := make(map[uuid.UUID]int)
						for _, guess := range game.guesses {
							userGuesses[guess.UserId]++
						}

						// Create leaderboard text
						leaderboard := "🏆 Game Over! Final Leaderboard:\n\n"
						for userId, count := range userGuesses {
							leaderboard += fmt.Sprintf("User @%s: %d words\n", userId.String(), count)
						}

						message := utils.Message{
							ID:        uuid.New(),
							ChatID:    chatId,
							SenderID:  uuid.MustParse(utils.AIChat),
							Timestamp: time.Now(),
							UpdatedAt: time.Now(),
							Content:   leaderboard + "\nCongratulations, all words have been guessed!",
						}

						// remove the game from the guessing games
						delete(c.guessingNames, chatId)
						c.storage.SaveMessage(message)
						return
					}

					// send the message that the user has guessed the word
					message := utils.Message{
						ID:        uuid.New(),
						ChatID:    chatId,
						SenderID:  uuid.MustParse(utils.AIChat),
						Timestamp: time.Now(),
						UpdatedAt: time.Now(),
						Content:   fmt.Sprintf("Congratulations, you have guessed one of the words. \n\nThere are still **%d** words left.", len(game.words)-len(game.guesses)),
					}
					c.storage.SaveMessage(message)
					return
				}
			}

			// send the message that the user has guessed the wrong word
			message := utils.Message{
				ID:        uuid.New(),
				ChatID:    chatId,
				SenderID:  uuid.MustParse(utils.AIChat),
				Timestamp: time.Now(),
				UpdatedAt: time.Now(),
				Content:   "Sorry, that is not a correct word. Please try again.",
			}
			c.storage.SaveMessage(message)

		}
	default:
		return nil, utils.NewError("command not supported", http.StatusBadRequest)
	}

	// check if the user is part of that chat
	member := c.MemberOfChat(userId, chatId)
	if !member {
		return nil, utils.NewError("User is not a member of the chat", http.StatusUnauthorized)
	}

	message := utils.Message{
		ID:        uuid.New(),
		ChatID:    chatId,
		SenderID:  userId,
		Timestamp: time.Now(),
		UpdatedAt: time.Now(),
		Content:   content,
		Command:   command,
	}

	err := c.storage.SaveMessage(message)
	if err != nil {
		return nil, err
	}

	go asyncFunction()

	err = c.storage.UpdateChatActivity(chatId)
	return &message, err
}

func (c *ChatService) SendMessage(userId uuid.UUID, chatId uuid.UUID, content string, media []uuid.UUID, replyTo *uuid.UUID) (*utils.Message, error) {

	if len(content) == 0 && len(media) == 0 {
		return nil, utils.NewError("message content is empty", http.StatusBadRequest)
	}

	// handle ai chat
	if chatId.String() == userId.String() {
		return c.AnswerAiChat(userId, content)
	}

	// check if the message exists if there is a reply
	if replyTo != nil {
		_, err := c.storage.GetMessage(*replyTo)
		if err != nil {
			return nil, utils.NewError("reply message not found", http.StatusNotFound)
		}
	}

	// check if the user is part of that chat
	member := c.MemberOfChat(userId, chatId)
	if !member {
		return nil, utils.NewError("User is not a member of the chat", http.StatusUnauthorized)
	}

	message := utils.Message{
		ID:        uuid.New(),
		ChatID:    chatId,
		SenderID:  userId,
		Timestamp: time.Now(),
		UpdatedAt: time.Now(),
		Media:     media,
		Content:   content,
		ReplyTo:   replyTo,
	}

	err := c.storage.SaveMessage(message)
	if err != nil {
		return nil, err
	}

	err = c.storage.UpdateChatActivity(chatId)
	return &message, err
}

func (c *ChatService) DeleteChat(userId uuid.UUID, chatId uuid.UUID) error {
	previousChat, err := c.storage.GetChat(chatId)
	if err != nil {
		return utils.NewError("Chat not found.", http.StatusNotFound)
	}

	contains := slices.ContainsFunc(previousChat.Members, func(i uuid.UUID) bool { return userId.String() == i.String() })
	if !contains {
		return utils.NewError("Updater is not present in members list of chat.", http.StatusBadRequest)
	}

	return c.storage.DeleteChat(chatId)
}

func (c *ChatService) UpdateMessage(userId uuid.UUID, message utils.Message) (utils.Message, error) {
	original, err := c.storage.GetMessage(message.ID)
	if err != nil {
		return utils.Message{}, utils.NewError("Message not found", http.StatusNotFound)
	}

	if original.SenderID.String() != userId.String() {
		return utils.Message{}, utils.NewError("Not the sender of the message", http.StatusUnauthorized)
	}

	original.Content = message.Content
	original.UpdatedAt = time.Now()
	if len(message.Media) > 0 {
		original.Media = message.Media
	}
	return original, c.storage.UpdateMessage(original)
}

func (c *ChatService) ReadMessage(userId uuid.UUID, messageId uuid.UUID) (utils.Message, error) {
	message, err := c.storage.GetMessage(messageId)
	if err != nil {
		return utils.Message{}, utils.NewError("message not found", http.StatusNotFound)
	}

	err = c.storage.MemberOfChat(userId, message.ChatID)
	if err != nil {
		return utils.Message{}, utils.NewError("not a member of that chat", http.StatusForbidden)
	}

	if userId.String() == message.SenderID.String() {
		return utils.Message{}, utils.NewError("cannot read own message", http.StatusBadRequest)
	}

	message.Read = true
	message.UpdatedAt = time.Now()
	return message, c.storage.UpdateMessage(message)
}

func (c *ChatService) DeleteMessage(userId uuid.UUID, messageId uuid.UUID) (utils.Message, error) {
	message, err := c.storage.GetMessage(messageId)
	if err != nil {
		return utils.Message{}, utils.NewError("Message not found", http.StatusNotFound)
	}

	if message.SenderID.String() != userId.String() {
		return utils.Message{}, utils.NewError("Not the sender of the message", http.StatusUnauthorized)
	}

	message.Content = ""
	message.Deleted = true
	message.UpdatedAt = time.Now()
	return message, c.storage.DeleteMessage(messageId)
}

func (c *ChatService) UpdateChat(userId uuid.UUID, chatId uuid.UUID, name string, members []uuid.UUID) (*utils.Chat, error) {
	previousChat, err := c.storage.GetChat(chatId)
	if err != nil {
		return nil, utils.NewError("Chat not found.", http.StatusNotFound)
	}

	contains := slices.ContainsFunc(previousChat.Members, func(i uuid.UUID) bool { return userId.String() == i.String() })
	if !contains {
		return nil, utils.NewError("Updater is not present in members list of chat.", http.StatusBadRequest)
	}

	if name == "AI" || name == "Direct Chat" {
		return nil, utils.NewError("invalid name", http.StatusBadRequest)
	}

	// filter out duplicate members
	membersMap := make(map[uuid.UUID]struct{})
	uniqueMember := []uuid.UUID{}

	for _, member := range members {
		if _, exists := membersMap[member]; !exists {
			membersMap[member] = struct{}{}
			uniqueMember = append(uniqueMember, member)
		}
	}

	if len(uniqueMember) < 2 {
		return nil, utils.NewError("Chat has to be between at least 2 persons", http.StatusBadRequest)
	}

	// check if all members exist
	exists, errc := c.auth.Exists(uniqueMember...)
	if errc != nil {
		return nil, errc
	}

	if !exists {
		return nil, utils.NewError("One or more members do not exist", http.StatusBadRequest)
	}

	chat := utils.Chat{
		ID:         chatId,
		Name:       name,
		Members:    uniqueMember,
		CreatedAt:  time.Now(),
		CreatorID:  userId,
		LastActive: time.Now(),
	}

	err = c.storage.CreateOrUpdateChat(chat)
	return &chat, err
}

func (c *ChatService) CreateChat(userId uuid.UUID, name string, members []uuid.UUID, initalMesage *string) (*utils.Chat, error) {

	contains := slices.ContainsFunc(members, func(i uuid.UUID) bool { return userId.String() == i.String() })

	if !contains {
		return nil, utils.NewError("Creator is not present in members list of chat.", http.StatusBadRequest)
	}

	// filter out duplicate members
	membersMap := make(map[uuid.UUID]struct{})
	uniqueMember := []uuid.UUID{}

	for _, member := range members {
		if _, exists := membersMap[member]; !exists {
			membersMap[member] = struct{}{}
			uniqueMember = append(uniqueMember, member)
		}
	}

	if len(uniqueMember) < 2 {
		return nil, utils.NewError("Chat has to be between at least 2 persons", http.StatusBadRequest)
	}

	// check if all members exist
	exists, errc := c.auth.Exists(uniqueMember...)
	if errc != nil {
		return nil, errc
	}

	if !exists {
		return nil, utils.NewError("One or more members do not exist", http.StatusBadRequest)
	}

	chat := utils.Chat{
		ID:         uuid.New(),
		Name:       name,
		Members:    uniqueMember,
		CreatedAt:  time.Now(),
		CreatorID:  userId,
		LastActive: time.Now(),
	}

	err := c.storage.CreateOrUpdateChat(chat)
	return &chat, err
}

func (c *ChatService) GetChat(id uuid.UUID) (*utils.Chat, error) {
	chat, err := c.storage.GetChat(id)
	if err != nil {
		return nil, utils.NewError("chat not found", http.StatusNotFound)
	}
	return chat, nil
}

func (c *ChatService) CreateDirectChat(userId uuid.UUID, receiver uuid.UUID, initialMessage *string) (*utils.Chat, error) {
	exists, err := c.auth.Exists(receiver)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, utils.NewError("Receiver does not exist", http.StatusBadRequest)
	}

	chat := utils.Chat{
		ID:         uuid.New(),
		Name:       "Direct Chat",
		Members:    []uuid.UUID{userId, receiver},
		CreatedAt:  time.Now(),
		CreatorID:  userId,
		LastActive: time.Now(),
	}

	err = c.storage.CreateOrUpdateChat(chat)

	if err != nil {
		return nil, err
	}

	if initialMessage != nil && len(*initialMessage) > 0 {
		message := utils.Message{
			ID:        uuid.New(),
			ChatID:    chat.ID,
			SenderID:  userId,
			Timestamp: time.Now(),
			UpdatedAt: time.Now(),
			Content:   *initialMessage,
		}
		chat.Messages = []utils.Message{message}

		err = c.storage.SaveMessage(message)
		if err != nil {
			return nil, err
		}
	}

	return &chat, nil
}
