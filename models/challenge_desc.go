package models

type ChallengeDesc struct {
	CandidateName     string
	GithubAlias       string
	ResumeURL         string
	ChallengeTemplate string
}

func NewChallengeDesc(input map[string]string) ChallengeDesc {
	return ChallengeDesc{
		CandidateName:     input["candidate_name"],
		GithubAlias:       input["github_alias"],
		ResumeURL:         input["resume_URL"],
		ChallengeTemplate: input["challenge_template"],
	}
}
