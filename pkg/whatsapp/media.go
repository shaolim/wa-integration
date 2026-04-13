package whatsapp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const graphAPIBaseURL = "https://graph.facebook.com/v25.0"

type mediaInfoResponse struct {
	URL      string `json:"url"`
	MimeType string `json:"mime_type"`
	SHA256   string `json:"sha256"`
	FileSize int64  `json:"file_size"`
	ID       string `json:"id"`
}

// DownloadMedia fetches raw bytes for the given WhatsApp media ID.
// It returns the file bytes and the MIME type.
//
// WhatsApp media download is a two-step process:
//  1. GET /v25.0/{media-id} → resolves the temporary download URL
//  2. GET {url}             → streams the actual file bytes
func (w *WAClient) DownloadMedia(mediaID string) ([]byte, string, error) {
	info, err := w.getMediaInfo(mediaID)
	if err != nil {
		return nil, "", fmt.Errorf("get media info: %w", err)
	}

	data, err := w.fetchMediaBytes(info.URL)
	if err != nil {
		return nil, "", fmt.Errorf("fetch media bytes: %w", err)
	}

	return data, info.MimeType, nil
}

func (w *WAClient) getMediaInfo(mediaID string) (*mediaInfoResponse, error) {
	req, err := http.NewRequest(http.MethodGet, graphAPIBaseURL+"/"+mediaID, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+w.WAAccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("graph API returned %d: %s", resp.StatusCode, string(body))
	}

	var info mediaInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}

	return &info, nil
}

func (w *WAClient) fetchMediaBytes(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+w.WAAccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("media download returned %d: %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}

// mediaIDFromMessage extracts the media ID from a media-type message.
// Returns an empty string for non-media message types.
func MediaIDFromMessage(msg Message) string {
	switch msg.Type {
	case "image":
		if msg.Image != nil {
			return msg.Image.ID
		}
	case "video":
		if msg.Video != nil {
			return msg.Video.ID
		}
	case "audio":
		if msg.Audio != nil {
			return msg.Audio.ID
		}
	case "sticker":
		if msg.Sticker != nil {
			return msg.Sticker.ID
		}
	case "document":
		if msg.Document != nil {
			return msg.Document.ID
		}
	}
	return ""
}
