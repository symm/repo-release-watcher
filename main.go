package main

import (
	"errors"
	"fmt"
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
	UserName string
	RepoName string
}

func getWatchlist() []WatchedRepo {
	return []WatchedRepo{
		WatchedRepo{UserName: "Robz8", RepoName: "TWLoader"},
		WatchedRepo{UserName: "TiniVi", RepoName: "safehax"},
		WatchedRepo{UserName: "nedwill", RepoName: "fasthax"},
		WatchedRepo{UserName: "d0k3", RepoName: "Decrypt9WIP"},
		WatchedRepo{UserName: "d0k3", RepoName: "EmuNAND9"},
		WatchedRepo{UserName: "d0k3", RepoName: "OTPHelper"},
		WatchedRepo{UserName: "d0k3", RepoName: "Hourglass9"},
		WatchedRepo{UserName: "d0k3", RepoName: "CTRXplorer"},
		WatchedRepo{UserName: "d0k3", RepoName: "A9NC"},
		WatchedRepo{UserName: "d0k3", RepoName: "GodMode9"},
		WatchedRepo{UserName: "AuroraWright", RepoName: "Luma3DS"},
		WatchedRepo{UserName: "AuroraWright", RepoName: "SafeA9LHInstaller"},
		WatchedRepo{UserName: "javimadgit", RepoName: "TinyFormat"},
		WatchedRepo{UserName: "Steveice10", RepoName: "FBI"},
		WatchedRepo{UserName: "yellows8", RepoName: "hblauncher_loader"},
		WatchedRepo{UserName: "meladroit", RepoName: "svdt"},
		WatchedRepo{UserName: "roxas75", RepoName: "rxTools"},
		WatchedRepo{UserName: "mid-kid", RepoName: "CakesForeveryWan"},
		WatchedRepo{UserName: "FloatingStar", RepoName: "FTP-GMX"},
		WatchedRepo{UserName: "llakssz", RepoName: "CIAngel"},
	}
}
func main() {

	resc, errc := make(chan ScrapedRelease), make(chan error)

	watchlist := getWatchlist()

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
