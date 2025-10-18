package primary

// VersionService defines the interface for version-related use cases
type VersionService interface {
	SendStartupNotification(channelID int64) error
}
