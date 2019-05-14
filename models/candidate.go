package models

type Candidate struct {
	Name          string
	GithubAlias   string
	ResumeURL     string
	ChallengeName string
}

func NewCandidate(input map[string]string) Candidate {
	return Candidate{
		Name:          input["candidate_name"],
		GithubAlias:   input["github_alias"],
		ResumeURL:     input["resume_URL"],
		ChallengeName: input["challenge_name"],
	}
}
