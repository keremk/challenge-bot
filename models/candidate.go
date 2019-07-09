package models

type Candidate struct {
	Name        string `bson:"Name"`
	GithubAlias string `bson:"GithubAlias"`
	ResumeURL   string `bson:"ResumeURL"`
	ChallengeID string `bson:"ChallengeID"`
}

func NewCandidate(input map[string]string) Candidate {
	return Candidate{
		Name:        input["candidate_name"],
		GithubAlias: input["github_alias"],
		ResumeURL:   input["resume_URL"],
		ChallengeID: input["challenge_id"],
	}
}
