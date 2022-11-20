package pubsub

const (
	//TopicUserUpdate - Topic used to notify user updates
	TopicUserUpdate string = "user-updated"

	//TopicUserCreation - Topic used to notify user creations
	TopicUserCreation string = "user-created"

	//TopicUserDeletion - Topic used to notify user deletions
	TopicUserDeletion string = "user-deleted"
)

// GetAllUsersTopics - returns a slice with all the users topics for easy
// "subscribe to all" functionality
func GetAllUsersTopics() []string {
	return []string{
		TopicUserUpdate,
		TopicUserCreation,
		TopicUserDeletion,
	}
}
