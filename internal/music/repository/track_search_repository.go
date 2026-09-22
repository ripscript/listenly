package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"listenly-backend/internal/music/models"

	"github.com/elastic/go-elasticsearch/v9"
)

const trackIndexName = "tracks"

type TrackSearchRepository interface {
	IndexTrack(ctx context.Context, track *models.Track) error
	Search(ctx context.Context, query string, limit, offset int) ([]string, int64, error) // return list of track UUID
	DeleteIndex(ctx context.Context, trackUUID string) error
}

type trackSearchRepository struct {
	es *elasticsearch.Client
}

func NewTrackSearchRepository(es *elasticsearch.Client) TrackSearchRepository {
	return &trackSearchRepository{es: es}
}

type trackDocument struct {
	UUID            string `json:"uuid"`
	YoutubeVideoID  string `json:"youtube_video_id"`
	Title           string `json:"title"`
	Artist          string `json:"artist"`
	DurationSeconds int    `json:"duration_seconds"`
	ThumbnailURL    string `json:"thumbnail_url"`
}

func (r *trackSearchRepository) IndexTrack(ctx context.Context, track *models.Track) error {
	doc := trackDocument{
		UUID:            track.UUID.String(),
		YoutubeVideoID:  track.YoutubeVideoID,
		Title:           track.Title,
		Artist:          track.Artist,
		DurationSeconds: track.DurationSeconds,
		ThumbnailURL:    track.ThumbnailURL,
	}

	body, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	res, err := r.es.Index(
		trackIndexName,
		bytes.NewReader(body),
		r.es.Index.WithDocumentID(doc.UUID),
		r.es.Index.WithContext(ctx),
		r.es.Index.WithRefresh("true"), // supaya langsung searchable, cocok untuk skala kecil
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("elasticsearch index error: %s", res.String())
	}
	return nil
}

func (r *trackSearchRepository) Search(ctx context.Context, query string, limit, offset int) ([]string, int64, error) {
	searchQuery := map[string]interface{}{
		"from": offset,
		"size": limit,
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":     query,
				"fields":    []string{"title^2", "artist"}, // title lebih diprioritaskan (boost x2)
				"fuzziness": "AUTO",                        // toleransi typo
			},
		},
	}

	body, err := json.Marshal(searchQuery)
	if err != nil {
		return nil, 0, err
	}

	res, err := r.es.Search(
		r.es.Search.WithContext(ctx),
		r.es.Search.WithIndex(trackIndexName),
		r.es.Search.WithBody(bytes.NewReader(body)),
	)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, 0, fmt.Errorf("elasticsearch search error: %s", res.String())
	}

	var result struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source trackDocument `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, 0, err
	}

	uuids := make([]string, 0, len(result.Hits.Hits))
	for _, hit := range result.Hits.Hits {
		uuids = append(uuids, hit.Source.UUID)
	}

	return uuids, result.Hits.Total.Value, nil
}

func (r *trackSearchRepository) DeleteIndex(ctx context.Context, trackUUID string) error {
	res, err := r.es.Delete(trackIndexName, trackUUID, r.es.Delete.WithContext(ctx))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return nil
}
