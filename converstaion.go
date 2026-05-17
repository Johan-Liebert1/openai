package main

const Romaji = false

const DevPromptRomaji string = `
I'll provide you with a Japanese text, which is part of a conversation.
Your job is to convert the Japanese text to hiragana (with spaces, use "ㅤ" character for the space) plus romaji plus its English translation.
If the provided text is not Japanese, return it as is.
The text is supposed to be used as subtitles, so make sure it follows a conversational flow.
Do not include the original Japanese text, only the Hiragana, Romaji and the English translation.
Example - INPUT = "私", OUTPUT = "わたし\nwatashi\nI".
Only output the translation for the latest sentence in the chat, don't repeat translations.
ALWAYS CONVERT THE ENTIRE TEXT. DON'T GIVE ME MARKDOWN OR ANY OTHER FORMAT, I WANT THE ANSWER IN PLAIN TEXT FORMAT.
`

const DevPromptNonRomaji string = `
I'll provide you with a Japanese text, which is part of a conversation.
Your job is to convert the Japanese text to hiragana (with spaces, use "ㅤ" character for the space) plus its English translation.
If the provided text is not Japanese, return it as is. Use katakana whereever necessary.
Make sure particles like の, は, この etc, are separated by "ㅤ".
The text is supposed to be used as subtitles, so make sure it follows a conversational flow.
Do not include the original Japanese text, only the Hiragana/Katakana and the English translation.
Example - INPUT = "私はその島へ向かった", OUTPUT = "わたしㅤはㅤそのㅤしまㅤへㅤむかった\nI went towards that island".
Only output the translation for the latest sentence in the chat, don't repeat translations.
If there's something in the beginning of the sentence inside parenthesis, it's probably the name of the character, so translate that accordingly.
Remove any unnecessary newlines.
ALWAYS CONVERT THE ENTIRE TEXT. DON'T GIVE ME MARKDOWN OR ANY OTHER FORMAT, I WANT THE ANSWER IN PLAIN TEXT FORMAT.
`

const DevPromptGPT string = `
I'll provide you with a Japanese text, which is part of a conversation.
Your job is to convert the Japanese text to hiragana (with spaces, use "ㅤ" character for the space) plus its English translation.
If the provided text is not Japanese, return it as is. Use katakana whereever necessary.
Make sure particles like の, は, この etc, are separated by "ㅤ".
Do not include the original Japanese text, only the Hiragana/Katakana and the English translation.
Example - INPUT = "私はその島へ向かった", OUTPUT = "わたしㅤはㅤそのㅤしまㅤへㅤむかった\nI went towards that island".
If there's something in the beginning of the sentence inside parenthesis, it's probably the name of the character, so translate that accordingly.
Remove any unnecessary newlines, like if there is a single sentence that spans multiple lines, put it in one line.
ALWAYS CONVERT THE ENTIRE TEXT. I WANT THE ANSWER IN PLAIN TEXT FORMAT.
`

var DevPrompt string

type RequestMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIAPIRequest struct {
	Model    string           `json:"model"`
	Store    bool             `json:"store"`
	Messages []RequestMessage `json:"messages"`
}

// GPTResponse represents the structure of the API response
type GPTResponse struct {
	Choices           []Choice `json:"choices"`
	Created           float64  `json:"created"`
	ID                string   `json:"id"`
	Model             string   `json:"model"`
	Object            string   `json:"object"`
	ServiceTier       string   `json:"service_tier"`
	SystemFingerprint string   `json:"system_fingerprint"`
	Usage             Usage    `json:"usage"`
}

// Choice represents each choice returned in the response
type Choice struct {
	FinishReason string  `json:"finish_reason"`
	Index        int     `json:"index"`
	Logprobs     *string `json:"logprobs"`
	Message      Message `json:"message"`
}

// Message represents the message content in the response
type Message struct {
	Content string  `json:"content"`
	Refusal *string `json:"refusal"`
	Role    string  `json:"role"`
}

// Usage represents token usage details
type Usage struct {
	CompletionTokens        int          `json:"completion_tokens"`
	CompletionTokensDetails TokenDetails `json:"completion_tokens_details"`
	PromptTokens            int          `json:"prompt_tokens"`
	PromptTokensDetails     TokenDetails `json:"prompt_tokens_details"`
	TotalTokens             int          `json:"total_tokens"`
}

// TokenDetails represents details of token usage
type TokenDetails struct {
	AcceptedPredictionTokens int `json:"accepted_prediction_tokens"`
	AudioTokens              int `json:"audio_tokens"`
	ReasoningTokens          int `json:"reasoning_tokens"`
	RejectedPredictionTokens int `json:"rejected_prediction_tokens"`
	CachedTokens             int `json:"cached_tokens,omitempty"`
}

var chatMessages []RequestMessage

// Will send maximum of this many messages with rolling window
const MaxConvLen = 8

func GetConverstaionMessages(newMessage RequestMessage) []RequestMessage {
	// Keep the system message at index 0
	if len(chatMessages) > MaxConvLen {
		// Drop from index 1 (preserve system)
		// dropping idx 1 and 2 as idx 1 will be user prompt, and idx 2 will be assistant answer
		chatMessages = append(chatMessages[:1], chatMessages[3:]...)
	}

	chatMessages = append(chatMessages, newMessage)

	return chatMessages
}
