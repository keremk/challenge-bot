package models

type Candidate struct {
	Name        string
	GithubAlias string
	ResumeURL   string
	ChallengeID string
}

func NewCandidate(input map[string]string) Candidate {
	return Candidate{
		Name:        input["candidate_name"],
		GithubAlias: input["github_alias"],
		ResumeURL:   input["resume_URL"],
		ChallengeID: input["challenge_id"],
	}
}
