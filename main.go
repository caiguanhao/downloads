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
	}

	nginx := task.FileServer{
		Name:   "nginx",
		Source: "http://nginx.org/download/",
	}

	if err := nginx.GetContent(); err != nil {
		log.Println(err)
	} else {
		if err := nginx.GetLinks(task.FileServerGetLinksOptions{
			Grep:        "\"(nginx-([1-9]+)\\.([0-9]+)\\.([0-9]+)\\.tar\\.gz)\"",
			GrepNamePos: 1,
			GrepVerPos:  []int{2, 3, 4},
			SortByVer:   true,
		}); err != nil {
			log.Println(err)
		} else {
			newTask.AddFileServerLinks(nginx)
		}
	}

	newTask.DownloadFiles(5)
}
