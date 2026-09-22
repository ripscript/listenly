package service

import (
	"context"

	"listenly-backend/internal/music/repository"
)

func ReindexAllTracks(ctx context.Context, repo repository.MusicRepository, searchRepo repository.TrackSearchRepository) (int, error) {
	tracks, _, err := repo.SearchTracks(ctx, "", 1000, 0)
	if err != nil {
		return 0, err
	}
	for _, t := range tracks {
		if err := searchRepo.IndexTrack(ctx, &t); err != nil {
			return 0, err
		}
	}
	return len(tracks), nil
}
