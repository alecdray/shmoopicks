package feed

import (
	"fmt"
	"log/slog"
	"shmoopicks/src/internal/core/contextx"
	"shmoopicks/src/internal/core/task"
)

type SyncSpotifyFeedTask struct {
	feed        FeedDTO
	feedService *Service
}

var _ task.Task = SyncSpotifyFeedTask{}

func NewSyncSpotifyFeedTask(feedService *Service, feed FeedDTO) task.Task {
	return SyncSpotifyFeedTask{feed: feed, feedService: feedService}
}

func (t SyncSpotifyFeedTask) Run(ctx contextx.ContextX) error {
	_, err := t.feedService.SyncSpotifyFeed(ctx, t.feed)
	return err
}

func (t SyncSpotifyFeedTask) Schedule() *task.CronExpression {
	return nil
}

func (t SyncSpotifyFeedTask) Name() string {
	return "sync_spotify_feed"
}

type SyncStaleSpotifyFeedsTask struct {
	feedService *Service
}

var _ task.Task = SyncStaleSpotifyFeedsTask{}

func NewSyncStaleSpotifyFeedsTask(feedService *Service) task.Task {
	return SyncStaleSpotifyFeedsTask{feedService: feedService}
}

func (t SyncStaleSpotifyFeedsTask) Run(ctx contextx.ContextX) error {
	staleFeeds, err := t.feedService.GetStaleSpotifyFeeds(ctx)
	if err != nil {
		err = fmt.Errorf("failed to get stale feeds: %w", err)
		return err
	}

	for _, feed := range staleFeeds {
		if feed.LastSyncStatus.IsSyncing() {
			continue
		}

		_, err := t.feedService.SyncSpotifyFeed(ctx, feed)
		if err != nil {
			err = fmt.Errorf("failed to sync spotify feed %s: %w", feed.ID, err)
			return err
		}

		slog.Debug("synced spotify feed", "id", feed.ID)
	}

	return nil
}

func (t SyncStaleSpotifyFeedsTask) Schedule() *task.CronExpression {
	schedule := task.CronExpression("* * * * *") // Every minute
	return &schedule
}

func (t SyncStaleSpotifyFeedsTask) Name() string {
	return "sync_stale_spotify_feeds"
}
