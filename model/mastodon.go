package model

import "time"

type MastodonStatus struct {
	CreatedAt        time.Time         `json:"created_at"`
	Visibility       string            `json:"visibility"`
	Language         string            `json:"language,omitempty"`
	Uri              string            `json:"uri,omitempty"`
	Url              string            `json:"url,omitempty"`
	Content          string            `json:"content"`
	Sensitive        bool              `json:"sensitive"`
	Account          MastodonAccount   `json:"account"`
	Tags             []Tag             `json:"tags"`
	MediaAttachments []MediaAttachment `json:"media_attachments"`
}

type MastodonAccount struct {
	Id             string `json:"id"`
	Acct           string `json:"acct"`
	Discoverable   bool   `json:"discoverable"`
	DisplayName    string `json:"display_name"`
	Indexable      *bool  `json:"indexable,omitempty"` // sometimes it's missing
	Locked         bool   `json:"locked"`
	Noindex        bool   `json:"noindex"`
	Note           string `json:"note"`
	Uri            string `json:"uri"`
	Url            string `json:"url"`
	FollowersCount uint32 `json:"followers_count"`
	StatusesCount  uint32 `json:"statuses_count"`
	Tags           []Tag  `json:"tags"`
}

type Tag struct {
	Name string `json:"name"`
}

type MediaAttachment struct {
	Type       string `json:"type"`
	Url        string `json:"url"`
	PreviewUrl string `json:"preview_url"`
}
