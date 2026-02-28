package models

type FeedKind string

const (
	FeedKindSpotify FeedKind = "spotify"
)

type FeedSyncStatus string

const (
	FeedSyncStatusNone    FeedSyncStatus = "none"
	FeedSyncStatusPending FeedSyncStatus = "pending"
	FeedSyncStatusSuccess FeedSyncStatus = "success"
	FeedSyncStatusFailure FeedSyncStatus = "failure"
)

type ReleaseFormat string

const (
	ReleaseFormatDigital  ReleaseFormat = "digital"
	ReleaseFormatVinyl    ReleaseFormat = "vinyl"
	ReleaseFormatCD       ReleaseFormat = "cd"
	ReleaseFormatCassette ReleaseFormat = "cassette"
)
