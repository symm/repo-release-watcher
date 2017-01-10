package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/google/go-github/github"
	"github.com/olekukonko/tablewriter"
)

type ScrapedRelease struct {
	Repo    WatchedRepo
	Release github.RepositoryRelease
}

type WatchedRepo struct {
	UserName string `json:"userName"`
	RepoName string `json:"repoName"`
}

func getWatchlist() ([]WatchedRepo, error) {

	file, err := ioutil.ReadFile("watchlist.json")

	if err != nil {
		log.Fatal(err)
	}

	watchlist := []WatchedRepo{}
	if err := json.Unmarshal(file, &watchlist); err != nil {
		return watchlist, err
	}

	return watchlist, nil
}
func main() {

	resc, errc := make(chan ScrapedRelease), make(chan error)

	watchlist, err := getWatchlist()

	if err != nil {
		log.Fatal(err)
	}

	for _, v := range watchlist {
		go func(repo WatchedRepo) {
			latestRelease, err := getLatestRelease(repo.UserName, repo.RepoName)

			if err != nil {
				errc <- err
				return
			}

			resc <- ScrapedRelease{
				Repo:    repo,
				Release: latestRelease,
			}
		}(v)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(createReleaseHeader())

	for i := 0; i < len(watchlist); i++ {
		select {
		case res := <-resc:
			table.Append(createReleaseRow(res))
		case err := <-errc:
			fmt.Println(err)
		}
	}

	table.Render()
}

func createReleaseHeader() []string {
	return []string{"App", "Version", "Date", "Download"}
}
func createReleaseRow(res ScrapedRelease) []string {
	return []string{res.Repo.RepoName, *res.Release.Name, res.Release.CreatedAt.Format(time.RFC822), *res.Release.TarballURL}
}

func getLatestRelease(user string, repo string) (github.RepositoryRelease, error) {
	clientId := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")

	if clientId == "" || clientSecret == "" {
		log.Fatal("Please set a GitHub Oauth client ID and secret")
	}

	t := &github.UnauthenticatedRateLimitedTransport{
		ClientID:     clientId,
		ClientSecret: clientSecret,
	}

	client := github.NewClient(t.Client())

	releases, _, err := client.Repositories.ListReleases(user, repo, nil)

	if err != nil {
		return github.RepositoryRelease{}, nil
	}

	if len(releases) == 0 {
		return github.RepositoryRelease{}, errors.New(fmt.Sprintf("No releases found for %s/%s", user, repo))
	}

	return *releases[0], nil
}
