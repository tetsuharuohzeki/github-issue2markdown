package main

import (
	"bufio"
	"context"
	"log"
	"os"
	"strings"
	"sync"

	"golang.org/x/oauth2"

	"github.com/google/go-github/github"
)

func main() {
	token := os.Getenv("OAUTH_TOKEN")
	if token == "" {
		log.Println("Set OAUTH_TOKEN envvar")
		return
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: token,
		},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	client := github.NewClient(tc)

	var owner string = "karino2"
	var repo string = "karino2.github.io"

	ctx := context.Background()

	// TODO: may specify the since param by https://godoc.org/github.com/google/go-github/github#IssueListOptions
	list, _, err := client.Issues.ListByRepo(ctx, owner, repo, nil)
	if err != nil {
		log.Printf("could not get issue list")
		log.Printf("err: %v", err)
		return
	}

	//log.Printf("list: %v\n", list)

	var wg sync.WaitGroup
	for _, issue := range list {
		wg.Add(1)

		go fetchComment(ctx, &wg, client, owner, repo, issue)
	}
	wg.Wait()
}

func fetchComment(ctx context.Context, wg *sync.WaitGroup, client *github.Client, owner, repo string, issue *github.Issue) {
	defer wg.Done()
	log.Printf("start to fetch: %v\n", issue.GetNumber())

	// All issues have their number. This would not be zero.
	comments, _, err := client.Issues.ListComments(ctx, owner, repo, issue.GetNumber(), nil)
	if err != nil {
		// TODO: dump the number
		log.Printf("err: %v", err)
		return
	}
	article := []string{
		issue.GetBody(),
	}
	for _, comment := range comments {
		b := *comment.Body
		article = append(article, b)
	}

	astr := strings.Join(article, "\n\n")
	//log.Printf("%v", astr)

	filename := issue.GetCreatedAt().Format("2006-01-02T15_04_05")
	f, err := os.Create("./articles/" + filename + ".md")
	if err != nil {
		// TODO: dump the number
		log.Printf("err: %v", err)
		return
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	_, err = w.WriteString(astr)
	if err != nil {
		// TODO: dump the number
		log.Printf("err: %v", err)
		return
	}

	err = w.Flush()
	if err != nil {
		// TODO: dump the number
		log.Printf("err: %v", err)
		return
	}
}
