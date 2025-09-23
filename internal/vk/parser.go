package vk

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"vk-normalizer/internal/contract"
)

const NormalizedMessageSource = "vk"
const NormalizedMessageImageType = "photo"

type Update struct {
	GroupID int64  `json:"group_id"`
	Type    string `json:"type"`
	EventID string `json:"event_id"`
	V       string `json:"v"`

	Object struct {
		Message VKMessage `json:"message"`
	} `json:"object"`

	Secret *string `json:"secret,omitempty"`
}

type VKMessage struct {
	Date        int64        `json:"date"`
	FromID      int64        `json:"from_id"`
	ID          int64        `json:"id"`
	PeerID      int64        `json:"peer_id"`
	Text        string       `json:"text"`
	Attachments []Attachment `json:"attachments"`
}

type Attachment struct {
	Type  string `json:"type"`
	Photo *Photo `json:"photo,omitempty"`
}

type Photo struct {
	ID      int64       `json:"id"`
	OwnerID int64       `json:"owner_id"`
	Sizes   []PhotoSize `json:"sizes"`
}

type PhotoSize struct {
	Height int    `json:"height"`
	Width  int    `json:"width"`
	Type   string `json:"type"`
	URL    string `json:"url"`
}

type VkMessageParser struct{}

func NewVkMessageParser() *VkMessageParser { return &VkMessageParser{} }

func (*VkMessageParser) ParseVKMessage(data []byte) (contract.NormalizedMessage, error) {
	var upd Update
	if err := json.Unmarshal(data, &upd); err != nil {
		return contract.NormalizedMessage{}, fmt.Errorf("parsing vk message: %w", err)
	}
	msg := upd.Object.Message

	var textPtr *string
	if msg.Text != "" {
		t := msg.Text
		textPtr = &t
	}

	n := contract.NormalizedMessage{
		Source:         NormalizedMessageSource,
		UserID:         msg.FromID,
		Username:       nil,
		ChatID:         msg.PeerID,
		Text:           textPtr,
		Timestamp:      time.Unix(msg.Date, 0).UTC(),
		Media:          []contract.MediaObject{},
		OriginalUpdate: json.RawMessage(data),
	}

	for _, a := range msg.Attachments {
		if a.Type == "photo" && a.Photo != nil {
			originalID := fmt.Sprintf("photo%d_%d", a.Photo.OwnerID, a.Photo.ID)

			sizes := a.Photo.Sizes
			if len(sizes) == 0 {
				continue
			}
			sort.Slice(sizes, func(i, j int) bool {
				return sizes[i].Width*sizes[i].Height > sizes[j].Width*sizes[j].Height
			})
			bestURL := sizes[0].URL

			n.Media = append(n.Media, contract.MediaObject{
				Type:           NormalizedMessageImageType,
				OriginalFileID: originalID,
				MimeType:       "image/jpeg",
				VkUrl:          bestURL,
			})
		}
	}

	return n, nil
}
