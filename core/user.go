package core

import ()

type IdType struct {
	ID string `json:"id"`
}

type Dimensions struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

/* Node type */
type Node struct {
	DisplaySrc       string     `json:"display_src"`
	ThumbnailSrc     string     `json:"thumbnail_src"`
	Owner            IdType     `json:"owner"`
	ID               string     `json:"id"`
	Date             int64      `json:"date"`
	CommentsDisabled bool       `json:"comments_disabled"`
	Caption          string     `json:"caption"`
	Dimensions       Dimensions `json:"dimensions"`
	VideoViews       int        `json:"video_views"`
	Comments         Count      `json:"comments"`
	IsVideo          bool       `json:"is_video"`
	Code             string     `json:"code"`
	Likes            Count      `json:"likes"`
}

/* PageInfo type */
type PageInfo struct {
	EndCursor       string `json:"end_cursor"`
	HasPreviousPage bool   `json:"has_previous_page"`
	StartCursor     string `json:"start_cursor"`
	HasNextPage     bool   `json:"has_next_page"`
}

/* Media type */
type Media struct {
	Nodes    []Node   `json:"nodes"`
	PageInfo PageInfo `json:"page_info"`
	Count    int      `json:"count"`
}

/* Count type */
type Count struct {
	Count int `json:"count"`
}

/* Account type */
type Account struct {
	UserID string
	Node   Node
	Likes  int
}

/* User type */
type User struct {
	Media              Media  `json:"media"`
	Biography          string `json:"biography"`
	ID                 string `json:"id"`
	FullName           string `json:"full_name"`
	ExternalURL        string `json:"external_url"`
	Username           string `json:"username"`
	HasBlockedViewer   bool   `json:"has_blocked_viewer"`
	HasRequestedViewer bool   `json:"has_requested_viewer"`
	ConnectedFbPage    string `json:"connected_fb_page"`
	FollowedByViewer   bool   `json:"followed_by_viewer"`
	IsPrivate          bool   `json:"is_private"`
	BlockedByViewer    bool   `json:"blocked_by_viewer"`
	FollowsViewer      bool   `json:"follows_viewer"`
	Follows            Count  `json:"follows"`
	RequestedByViewer  bool   `json:requested_by_viewer""`
	FollowedBy         Count  `json:"followed_by"`
	IsVerified         bool   `json:"is_verified"`

	ExternalURLLinkshimmed string `json:""`
	ProfilePicURL          string `json:""`
}

/* UserJSON type*/
type UserJSON struct {
	User User `json:"user"`
}
