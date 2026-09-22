package dto

import "time"

type SearchTracksRequest struct {
	Query    string
	Page     int
	PageSize int
}

type RequestTrackRequest struct {
	RoomUUID        string
	RequestedByUUID string
	YoutubeVideoID  string
}

type RemoveFromQueueRequest struct {
	QueueItemUUID string
	RequesterUUID string
}

type TrackResponse struct {
	UUID            string
	YoutubeVideoID  string
	Title           string
	Artist          string
	DurationSeconds int
	ThumbnailURL    string
}

type QueueItemResponse struct {
	UUID            string
	Track           TrackResponse
	RequestedByUUID string
	Status          string
	Position        int
	CreatedAt       time.Time
}
