package githubclient

import (
	"net/http"

	"github.com/google/go-github/v74/github"
)

type Client struct {
	*github.Client
}

func NewClient(token, baseURL string) *Client {
	var tc *http.Client
	if token != "" {
		tc = github.NewClient(nil).WithAuthToken(token).Client()
	}

	client, _ := github.NewClient(tc).WithEnterpriseURLs(baseURL, baseURL)
	return &Client{client}
}
