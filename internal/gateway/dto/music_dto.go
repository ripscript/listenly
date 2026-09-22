package dto

type SearchTracksHTTPRequest struct {
	Query string `query:"q" validate:"required,min=1"`
}

type RequestTrackHTTPRequest struct {
	YoutubeVideoID string `json:"youtube_video_id" validate:"required"`
}

type TrackHTTPResponse struct {
	UUID            string `json:"uuid"`
	YoutubeVideoID  string `json:"youtube_video_id"`
	Title           string `json:"title"`
	Artist          string `json:"artist"`
	DurationSeconds int    `json:"duration_seconds"`
	ThumbnailURL    string `json:"thumbnail_url"`
}

type QueueItemHTTPResponse struct {
	UUID            string            `json:"uuid"`
	Track           TrackHTTPResponse `json:"track"`
	RequestedByUUID string            `json:"requested_by_uuid"`
	Status          string            `json:"status"`
	Position        int               `json:"position"`
	CreatedAt       string            `json:"created_at"`
}

type SearchTracksHTTPResponse struct {
	Tracks     []TrackHTTPResponse `json:"tracks"`
	TotalItems int                 `json:"total_items"`
}
