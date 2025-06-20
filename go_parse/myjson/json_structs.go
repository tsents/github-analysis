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
	switch eventTypeCatch.EventType {
	case "CommitCommentEvent":
		var p CommitCommentEventPayload
		if err := json.Unmarshal(data, &p); err != nil {
			return nil, fmt.Errorf("CommitCommentEvent: %w", err)
		}
		return p, nil

	case "CreateEvent":
		var p CreateEventPayload
		if err := json.Unmarshal(data, &p); err != nil {
			return nil, fmt.Errorf("CreateEvent: %w", err)
		}
		return p, nil

	case "DeleteEvent":
		var p DeleteEventPayload
		if err := json.Unmarshal(data, &p); err != nil {
			return nil, fmt.Errorf("DeleteEvent: %w", err)
		}
		return p, nil

	case "ForkEvent":
		var p ForkEventPayload
		if err := json.Unmarshal(data, &p); err != nil {
			return nil, fmt.Errorf("ForkEvent: %w", err)
		}
		return p, nil

	case "GollumEvent":
		var p GollumEventPayload
		if err := json.Unmarshal(data, &p); err != nil {
			return nil, fmt.Errorf("GollumEvent: %w", err)
		}
		return p, nil

	case "IssueCommentEvent":
		var p IssueCommentEventPayload
		if err := json.Unmarshal(data, &p); err != nil {
			return nil, fmt.Errorf("IssueCommentEvent: %w", err)
		}
		return p, nil

	case "IssuesEvent":
		var p IssuesEventPayload
		if err := json.Unmarshal(data, &p); err != nil {
			return nil, fmt.Errorf("IssuesEvent: %w", err)
		}
		return p, nil

	case "MemberEvent":
		var p MemberEventPayload
		if err := json.Unmarshal(data, &p); err != nil {
			return nil, fmt.Errorf("MemberEvent: %w", err)
		}
		return p, nil

	case "PublicEvent":
		var p PublicEventPayload
		if err := json.Unmarshal(data, &p); err != nil {
			return nil, fmt.Errorf("PublicEvent: %w", err)
		}
		return p, nil

	case "PullRequestEvent":
		var p PullRequestEventPayload
		if err := json.Unmarshal(data, &p); err != nil {
			return nil, fmt.Errorf("PullRequestEvent: %w", err)
		}
		return p, nil

	case "PullRequestReviewEvent":
		var p PullRequestReviewEventPayload
		if err := json.Unmarshal(data, &p); err != nil {
			return nil, fmt.Errorf("PullRequestReviewEvent: %w", err)
		}
		return p, nil

	case "PullRequestReviewCommentEvent":
		var p PullRequestReviewCommentEventPayload
		if err := json.Unmarshal(data, &p); err != nil {
			return nil, fmt.Errorf("PullRequestReviewCommentEvent: %w", err)
		}
		return p, nil

	case "PullRequestReviewThreadEvent":
		var p PullRequestReviewThreadEventPayload
		if err := json.Unmarshal(data, &p); err != nil {
			return nil, fmt.Errorf("PullRequestReviewThreadEvent: %w", err)
		}
		return p, nil

	case "PushEvent":
		var p PushEventPayload
		if err := json.Unmarshal(data, &p); err != nil {
			return nil, fmt.Errorf("PushEvent: %w", err)
		}
		return p, nil

	case "ReleaseEvent":
		var p ReleaseEventPayload
		if err := json.Unmarshal(data, &p); err != nil {
			return nil, fmt.Errorf("ReleaseEvent: %w", err)
		}
		return p, nil

	case "SponsorshipEvent":
		var p SponsorshipEventPayload
		if err := json.Unmarshal(data, &p); err != nil {
			return nil, fmt.Errorf("SponsorshipEvent: %w", err)
		}
		return p, nil

	case "WatchEvent":
		var p WatchEventPayload
		if err := json.Unmarshal(data, &p); err != nil {
			return nil, fmt.Errorf("WatchEvent: %w", err)
		}
		return p, nil

	default:
		var p map[string]any
		if err := json.Unmarshal(data, &p); err != nil {
			return nil, fmt.Errorf("Unknown event type (%s): %w", eventTypeCatch.EventType, err)
		}
		return p, nil
	}
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


type CommitCommentEventPayload struct {
    Action  string         `json:"action"`
    Comment CommitComment  `json:"comment"`
}

type CommitComment struct {
    ID        int    `json:"id"`
    Body      string `json:"body"`
    CommitID  string `json:"commit_id"`
    URL       string `json:"url"`
    CreatedAt string `json:"created_at"`
    UpdatedAt string `json:"updated_at"`
    // Add other commit comment fields as needed
}

type CreateEventPayload struct {
    Ref          *string `json:"ref"`           // can be null
    RefType      string  `json:"ref_type"`      // branch, tag, or repository
    MasterBranch string  `json:"master_branch"`
    Description  string  `json:"description"`
    PusherType   string  `json:"pusher_type"`   // user or deploy key
}

type DeleteEventPayload struct {
    Ref     string `json:"ref"`
    RefType string `json:"ref_type"` // branch or tag
}

type ForkEventPayload struct {
    Forkee Repo `json:"forkee"` // created repository resource
}

//
type GollumEventPayload struct {
    Pages []GollumPage `json:"pages"`
}

type GollumPage struct {
    PageName string `json:"page_name"`
    Title    string `json:"title"`
    Action   string `json:"action"`   // created or edited
    Sha      string `json:"sha"`
    HTMLURL  string `json:"html_url"`
}


type IssueCommentEventPayload struct {
    Action  string         `json:"action"` // created, edited, deleted
    Changes *IssueChanges  `json:"changes,omitempty"`
    Issue   Issue          `json:"issue"`
    Comment Comment        `json:"comment"`
}

type IssueChanges struct {
    Body *ChangeFrom `json:"body,omitempty"`
}

type ChangeFrom struct {
    From string `json:"from"`
}

type Issue struct {
    ID     int    `json:"id"`
    Title  string `json:"title"`
    Body   string `json:"body"`
    State  string `json:"state"`
    // add more fields as needed
}

type Comment struct {
    ID   int    `json:"id"`
    Body string `json:"body"`
    // add more fields as needed
}

type IssuesEventPayload struct {
    Action   string         `json:"action"` // opened, edited, closed, etc.
    Issue    Issue          `json:"issue"`
    Changes  *IssueChanges  `json:"changes,omitempty"`
    Assignee *User          `json:"assignee,omitempty"`
    Label    *Label         `json:"label,omitempty"`
}

type User struct {
    ID    int    `json:"id"`
    Login string `json:"login"`
    // other fields as needed
}

type Label struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Color string `json:"color"`
}

type MemberEventPayload struct {
    Action  string           `json:"action"`
    Member  User             `json:"member"`
    Changes *MemberChanges   `json:"changes,omitempty"`
}

type MemberChanges struct {
    OldPermission *ChangeFrom `json:"old_permission,omitempty"`
}

type PublicEventPayload struct {
    // Empty payload for PublicEvent
}

type PullRequestEventPayload struct {
    Action      string         `json:"action"`
    Number      int            `json:"number"`
    Changes     *PRChanges     `json:"changes,omitempty"`
    PullRequest PullRequest    `json:"pull_request"`
    Reason      string         `json:"reason,omitempty"`
}

type PRChanges struct {
    Title *ChangeFrom `json:"title,omitempty"`
    Body  *ChangeFrom `json:"body,omitempty"`
}

type PullRequest struct {
    ID     int    `json:"id"`
    Title  string `json:"title"`
    Body   string `json:"body"`
    State  string `json:"state"`
    // add other fields as needed
}

type PullRequestReviewEventPayload struct {
    Action      string      `json:"action"`
    PullRequest PullRequest `json:"pull_request"`
    Review      Review      `json:"review"`
}

type Review struct {
    ID     int    `json:"id"`
    Body   string `json:"body"`
    State  string `json:"state"`
    // add other fields as needed
}

type PullRequestReviewCommentEventPayload struct {
    Action      string          `json:"action"`
    Changes     *CommentChanges `json:"changes,omitempty"`
    PullRequest PullRequest     `json:"pull_request"`
    Comment     Comment         `json:"comment"`
}

type CommentChanges struct {
    Body *ChangeFrom `json:"body,omitempty"`
}

type PullRequestReviewThreadEventPayload struct {
    Action      string      `json:"action"` // resolved, unresolved
    PullRequest PullRequest `json:"pull_request"`
    Thread      Thread      `json:"thread"`
}

type Thread struct {
    ID      int    `json:"id"`
    Comments []Comment `json:"comments"`
    // add other fields as needed
}

type PushEventPayload struct {
    PushID      int       `json:"push_id"`
    Size        int       `json:"size"`
    DistinctSize int      `json:"distinct_size"`
    Ref         string    `json:"ref"`
    Head        string    `json:"head"`
    Before      string    `json:"before"`
    Commits     []Commit  `json:"commits"`
}

type Commit struct {
    Sha     string `json:"sha"`
    Message string `json:"message"`
    Author  Author `json:"author"`
    URL     string `json:"url"`
    Distinct bool  `json:"distinct"`
}

type Author struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

type ReleaseEventPayload struct {
    Action  string          `json:"action"`
    Changes *ReleaseChanges `json:"changes,omitempty"`
    Release Release         `json:"release"`
}

type ReleaseChanges struct {
    Body *ChangeFrom `json:"body,omitempty"`
    Name *ChangeFrom `json:"name,omitempty"`
}

type Release struct {
    ID          int    `json:"id"`
    TagName     string `json:"tag_name"`
    Name        string `json:"name"`
    Body        string `json:"body"`
    Draft       bool   `json:"draft"`
    Prerelease  bool   `json:"prerelease"`
    CreatedAt   string `json:"created_at"`
    PublishedAt string `json:"published_at"`
    // add more fields as needed
}

type SponsorshipEventPayload struct {
    Action        string             `json:"action"`
    EffectiveDate string             `json:"effective_date,omitempty"`
    Changes       *SponsorshipChange `json:"changes,omitempty"`
}

type SponsorshipChange struct {
    Tier         *TierChange `json:"tier,omitempty"`
    PrivacyLevel *ChangeFrom `json:"privacy_level,omitempty"`
}

type TierChange struct {
    From any `json:"from"` // Could be more detailed type
}

type WatchEventPayload struct {
    Action string `json:"action"` // currently only "started"
}
