package slack

type ValidationError struct{}

func (e ValidationError) Error() string {
	return "Invalid validation token received from Slack server"
}
