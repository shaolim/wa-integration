package whatsapp

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

// WebhookPayload is the top-level structure sent by the WhatsApp Cloud API.
type WebhookPayload struct {
	Object string  `json:"object"`
	Entry  []Entry `json:"entry"`
}

// Parse unmarshals a raw webhook JSON body into a WebhookPayload.
func Parse(body []byte) (*WebhookPayload, error) {
	var payload WebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	return &payload, nil
}

type Entry struct {
	ID      string   `json:"id"`
	Changes []Change `json:"changes"`
}

type Change struct {
	Field string `json:"field"`
	Value Value  `json:"value"`
}

type Value struct {
	MessagingProduct string    `json:"messaging_product"`
	Metadata         Metadata  `json:"metadata"`
	Contacts         []Contact `json:"contacts,omitempty"`
	Messages         []Message `json:"messages,omitempty"`
	Statuses         []Status  `json:"statuses,omitempty"`
}

type Metadata struct {
	DisplayPhoneNumber string `json:"display_phone_number"`
	PhoneNumberID      string `json:"phone_number_id"`
}

type Contact struct {
	Profile Profile `json:"profile"`
	WAID    string  `json:"wa_id"`
}

type Profile struct {
	Name string `json:"name"`
}

// Message represents an inbound WhatsApp message. The Type field indicates
// which of the optional payload fields is populated.
type Message struct {
	From      string `json:"from"`
	ID        string `json:"id"`
	Timestamp string `json:"timestamp"`
	Type      string `json:"type"`

	Text        *TextMessage        `json:"text,omitempty"`
	Image       *MediaMessage       `json:"image,omitempty"`
	Video       *MediaMessage       `json:"video,omitempty"`
	Audio       *MediaMessage       `json:"audio,omitempty"`
	Sticker     *MediaMessage       `json:"sticker,omitempty"`
	Document    *DocumentMessage    `json:"document,omitempty"`
	Location    *LocationMessage    `json:"location,omitempty"`
	Reaction    *ReactionMessage    `json:"reaction,omitempty"`
	Button      *ButtonMessage      `json:"button,omitempty"`
	Interactive *InteractiveMessage `json:"interactive,omitempty"`
	Context     *MessageContext     `json:"context,omitempty"`
}

type TextMessage struct {
	Body string `json:"body"`
}

// MediaMessage covers image, video, audio, and sticker message types.
type MediaMessage struct {
	ID       string `json:"id"`
	MimeType string `json:"mime_type"`
	SHA256   string `json:"sha256"`
	Caption  string `json:"caption,omitempty"`
}

type DocumentMessage struct {
	ID       string `json:"id"`
	MimeType string `json:"mime_type"`
	SHA256   string `json:"sha256"`
	Filename string `json:"filename,omitempty"`
	Caption  string `json:"caption,omitempty"`
}

type LocationMessage struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Name      string  `json:"name,omitempty"`
	Address   string  `json:"address,omitempty"`
}

type ReactionMessage struct {
	MessageID string `json:"message_id"`
	Emoji     string `json:"emoji"`
}

// ButtonMessage is set when the user taps a quick-reply button.
type ButtonMessage struct {
	Text    string `json:"text"`
	Payload string `json:"payload"`
}

// InteractiveMessage is set for interactive button/list replies.
type InteractiveMessage struct {
	Type        string       `json:"type"`
	ButtonReply *ButtonReply `json:"button_reply,omitempty"`
	ListReply   *ListReply   `json:"list_reply,omitempty"`
}

type ButtonReply struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type ListReply struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

// MessageContext is present when the inbound message is a reply to another message.
type MessageContext struct {
	From string `json:"from"`
	ID   string `json:"id"`
}

// Status represents a delivery/read receipt for an outbound message.
type Status struct {
	ID           string        `json:"id"`
	RecipientID  string        `json:"recipient_id"`
	Status       string        `json:"status"`
	Timestamp    string        `json:"timestamp"`
	Conversation *Conversation `json:"conversation,omitempty"`
	Pricing      *Pricing      `json:"pricing,omitempty"`
}

type Conversation struct {
	ID     string              `json:"id"`
	Origin *ConversationOrigin `json:"origin,omitempty"`
}

type ConversationOrigin struct {
	Type string `json:"type"`
}

type Pricing struct {
	Billable     bool   `json:"billable"`
	PricingModel string `json:"pricing_model"`
	Category     string `json:"category"`
}

func (w *WAClient) ValidateSignature(body []byte, signature string) bool {
	if signature == "" {
		return false
	}

	if len(signature) < 7 || signature[:7] != "sha256=" {
		return false
	}
	expectedHash := signature[7:]

	mac := hmac.New(sha256.New, []byte(w.WAAppSecret))
	mac.Write(body)
	computedHash := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(computedHash), []byte(expectedHash))
}
