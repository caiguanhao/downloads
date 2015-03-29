package main

import (
	"log"
	"os"

	"github.com/caiguanhao/downloads/task"
)

func main() {
	newTask := task.Task{}

	github := task.GitHub{
		Name:        "docker-compose",
		Owner:       "docker",
		Repository:  "compose",
		AccessToken: os.Getenv("GITHUB_ACCESS_TOKEN"),
	}

	if err := github.GetGitHubReleases(); err != nil {
		log.Println(err)
	} else {
		newTask.AddGitHubSources(github)
		newTask.AddGitHubReleases(github)
		newTask.DownloadFiles(5)
	}
}
