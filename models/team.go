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
	ID        string
	Name      string
	BotToken  string
	BotUserID string
}
