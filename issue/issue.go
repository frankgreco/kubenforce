package issue

import (
    "context"
    "os"
    "fmt"

    "github.com/google/go-github/github"
    "golang.org/x/oauth2"
    "k8s-audit/config"
)

func Create(issue *config.Issue) {

    fmt.Print("token: " + os.Getenv("ACCESS_TOKEN"))
    ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: "3388a9e50b8ebd590a896d21eb4bc4b338e5c2c9"},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
    issueReq := github.IssueRequest{
        Title: issue.Title,
        Body: issue.Body,
        State: issue.State,
    }
	_,_,err := client.Issues.Create(ctx, issue.Owner, issue.Repo, &issueReq)
    if err != nil {
        panic(err.Error())
    }

}
