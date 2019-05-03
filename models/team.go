package models

type ChallengeTeam struct {
	ID               string
	Name             string
	GithubOwner      string
	GithubOrg        string
	TrackingRepoName string
}

type Organization struct {
	ID   string
	Name string
}

type SlackTeam struct {
	Name           string
	SlackID        string
	SlackBotToken  string
	SlackBotUserID string
}
