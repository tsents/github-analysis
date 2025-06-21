package myjson


import (
	"fmt"
)

type TypeCatcher struct {
    EventType      string   `json:"type"`
}

// UnmarshalPayload unmarshals JSON payload into the correct struct based on event type.
func UnmarshalPayload(data []byte) (any, error) {
	var eventTypeCatch TypeCatcher;
	if err := json.Unmarshal(data, &eventTypeCatch); err != nil {
		return nil, fmt.Errorf("CommitCommentEvent: %w", err)
	}
	switch eventTypeCatch.EventType {}
	return nil, nil
}


type BaseEvent struct {
    ID        int      `json:"id"`
    Type      string   `json:"type"`
    Actor     Actor    `json:"actor"`
    Repo      Repo     `json:"repo"`
    Payload   Payload  `json:"payload"` // Payload is generic here, can be customized per event type
    Public    bool     `json:"public"`
    CreatedAt string   `json:"created_at"`
    Org       *Org     `json:"org,omitempty"` // Org is optional, so pointer with omitempty
}

type Actor struct {
    ID           int    `json:"id"`
    Login        string `json:"login"`
    DisplayLogin string `json:"display_login"`
    GravatarID   string `json:"gravatar_id"`
    URL          string `json:"url"`
    AvatarURL    string `json:"avatar_url"`
}

type Repo struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
    URL  string `json:"url"`
}

type Org struct {
    ID         int    `json:"id"`
    Login      string `json:"login"`
    GravatarID string `json:"gravatar_id"`
    URL        string `json:"url"`
    AvatarURL  string `json:"avatar_url"`
}

// Payload can be defined as any or a custom struct depending on event type
type Payload map[string]any

