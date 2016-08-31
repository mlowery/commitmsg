package main

import (
	"log"
	"os"
	"net/url"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"bytes"
	"fmt"
	"io/ioutil"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

const (
	// https://help.github.com/articles/autolinked-references-and-urls/#issues-and-pull-requests
	FullRegexp = `^([^\/]+)\/([^#]+)#(\d+)$`
	// TODO(mlowery): support medium and short versions too
	MediumRegexp = `^([^#]+)#(\d+)$`
	ShortRegexp = `^#(\d+)$`


	sourceCommit = "commit"
	sourceTemplate = "template"
	sourceMerge = "merge"
)

func newClient(gitHubAPIBaseURL string, gitHubAPIAccessToken string) (*github.Client, error) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: gitHubAPIAccessToken},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)

	client := github.NewClient(tc)
	if !strings.HasSuffix("/", gitHubAPIBaseURL) {
		gitHubAPIBaseURL = gitHubAPIBaseURL + "/"
	}
	baseURL, err := url.Parse(gitHubAPIBaseURL)
	if err != nil {
		return nil, err
	}
	client.BaseURL = baseURL

	return client, nil
}

func processArgs() (string, string, string) {
	log.Printf("os.Args: %v\n", os.Args)
	file := os.Args[1]
	source := os.Args[2]
	var commit string
	if len(os.Args) >= 4 {
		commit = os.Args[3]
	}
	return file, source, commit
}

func currentBranch() (string, *exec.ExitError) {
	branch, err := exec.Command("git", "symbolic-ref", "-q", "--short", "HEAD").Output()
	if err != nil {
		return "", err.(*exec.ExitError)
	}
	return strings.TrimSpace(string(branch)), nil
}

func main() {
	file, source, commit := processArgs()
	log.Printf("file: %s, source: %s, commit: %s", file, source, commit)

	// some logic borrowed from https://gist.github.com/nyarly/0348b925ad1572777ba4

	if source != sourceTemplate {
		// Examples of when to exit without doing anything:
		// * commit is being amended
		log.Printf("source not 'template'; nothing to do\n")
		os.Exit(0)
	}

	branch, err1 := currentBranch()
	if err1 != nil {
		log.Printf("symbolic-ref returned error: %s (%s)\n", err1, err1.Stderr)
		log.Printf("symbolic-ref returned non-zero exit; nothing to do\n")
		// assume detached head; do nothing
		os.Exit(0)
	}
	
	re, err := regexp.Compile(FullRegexp)
	if err != nil {
		log.Fatalf("could not compile regexp: %s\n", err)
	}
	res := re.FindStringSubmatch(branch)
	if res == nil {
		log.Printf("branch %s does not match regexp; nothing to do\n", branch)
		os.Exit(0)
	}
	if len(res) != 4 {
		log.Fatalf("expected slice of size 4; got %v\n", res)
	}
	owner := res[1]
	repo := res[2]
	number, err := strconv.Atoi(res[3])
	if err != nil {
		log.Fatalf("could not parse int %s\n", res[3])
	}

	accessToken := os.Getenv("COMMITMSG_ACCESS_TOKEN")
	if accessToken == "" {
		log.Fatalf("env var COMMITMSG_ACCESS_TOKEN is required\n")
	}
	baseURL := os.Getenv("COMMITMSG_GITHUB_BASE_URL")
	if baseURL == "" {
		log.Fatalf("env var COMMITMSG_GITHUB_BASE_URL is required\n")
	}

	client, err := newClient(baseURL, accessToken)

	if err != nil {
		log.Fatalf("Could not create GitHub client: %s\n", err)
	}

	issue, _, err := client.Issues.Get(owner, repo, number)
	log.Printf("title: %s\n", *issue.Title)

	var buffer bytes.Buffer
	buffer.WriteString(*issue.Title)
	buffer.WriteString("\n\n")
	buffer.WriteString(fmt.Sprintf("Fixes %s\n\n", branch))
	
	lines := strings.Split(*issue.Body, "\n")
	for _, line := range lines {
		buffer.WriteString(fmt.Sprintf("# %s\n", line))
	}

	existingBytes, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatalf("Could not read file: %s\n", err)
	}

	buffer.Write(existingBytes)

	err = ioutil.WriteFile(file, buffer.Bytes(), 0600)
	if err != nil {
		log.Fatalf("Could not write file: %s\n", err)
	}
}
