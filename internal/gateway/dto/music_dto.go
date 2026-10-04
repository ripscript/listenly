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

type StreamURLHTTPResponse struct {
	StreamURL string `json:"stream_url"`
	ExpiresAt string `json:"expires_at"`
}

type YouTubeResultHTTPResponse struct {
	YoutubeVideoID  string `json:"youtube_video_id"`
	Title           string `json:"title"`
	Channel         string `json:"channel"`
	DurationSeconds int    `json:"duration_seconds"`
	ThumbnailURL    string `json:"thumbnail_url"`
}

type SearchYouTubeHTTPResponse struct {
	Results []YouTubeResultHTTPResponse `json:"results"`
}

type AdvanceQueueHTTPRequest struct {
	CurrentQueueItemUUID string `json:"current_queue_item_uuid"`
}

type AdvanceQueueHTTPResponse struct {
	HasNext  bool                   `json:"has_next"`
	NextItem *QueueItemHTTPResponse `json:"next_item,omitempty"`
}
