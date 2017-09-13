package main

import (
	"bufio"
	"context"
	"flag"
	"log"
	"os"
	"strings"
	"sync"
	"text/template"

	"golang.org/x/oauth2"

	"github.com/google/go-github/github"
)

func main() {
	var owner string
	flag.StringVar(&owner, "owner", "", "The repository's owner name.")

	var repo string
	flag.StringVar(&repo, "repo", "", "The repository name.")

	flag.Parse()

	if owner == "" {
		log.Println("Specify -owner flag correctly")
		return
	}

	if repo == "" {
		log.Println("Specify -repo flag correctly")
		return
	}

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

type post struct {
	Title string
	Date  string
	Body  string
}

const jekyllTemplate = `---
title: "{{.Title}}"
date: {{.Date}}
---

{{.Body}}
`

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

	title := issue.GetTitle()

	created := issue.GetCreatedAt()

	tpl := template.Must(template.New("jekyll_template").Parse(jekyllTemplate))
	p := post{
		Title: title,
		Date:  created.Format("2006-01-02 15:04:05"),
		Body:  astr,
	}

	filename := created.Format("2006-01-02") + "-" + strings.Replace(title, " ", "-", -1)
	f, err := os.Create("./articles/" + filename + ".md")
	if err != nil {
		// TODO: dump the number
		log.Printf("err: %v", err)
		return
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	if err := tpl.Execute(w, p); err != nil {
		log.Printf("err: %v", err)
		return
	}

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
