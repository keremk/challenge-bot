package config

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

type GithubChallengeReader struct{}

func NewGithubChallengeReader() *GithubChallengeReader {
	return &GithubChallengeReader{}
}

func (c *GithubChallengeReader) Read(url string, token string) ([]byte, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	tokenHeader := fmt.Sprintf("token %s", token)
	req.Header.Add("Authorization", tokenHeader)
	req.Header.Add("Accept", "application/vnd.github.v3.raw")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return body, nil
}
