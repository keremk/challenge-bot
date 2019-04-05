package repo

import (
	"testing"

	"github.com/keremk/challenge-bot/config"
)

func TestCheckUser(t *testing.T) {
	env := config.GetEnvironment()
	userExists := checkUser("keremktest404", env.GithubToken)
	if userExists != false {
		t.Errorf("Checked non-existing user, and did not get false")
	}
}
