package issue

import (
	"context"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type Issue struct {
	Owner string
	Repo  string
	Title *string
	Body  *string
	State *string
}

func (issue *Issue) Create() {

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: "< you're access token >"}, // <-- add this as a cmd flag
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	issueReq := github.IssueRequest{
		Title: issue.Title,
		Body:  issue.Body,
		State: issue.State,
	}
	_, _, err := client.Issues.Create(ctx, issue.Owner, issue.Repo, &issueReq)
	if err != nil {
		panic(err.Error())
	}

}
