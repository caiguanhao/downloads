package task

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type (
	GitHub struct {
		Name        string
		Owner       string
		Repository  string
		AccessToken string

		releases []struct {
			Version string `json:"tag_name"`
			Source  string `json:"tarball_url"`
			Files   []struct {
				Name string
				Url  string `json:"browser_download_url"`
			} `json:"assets"`
		}
	}

	GitHubAPIErrorMessage struct {
		Message string `json:"message"`
	}
)

func (task *Task) AddGitHubSources(github GitHub) {
	headers := github.accessTokenHeader()
	for _, release := range github.releases {
		task.Downloads = append(task.Downloads, Download{
			Remote:  release.Source,
			Local:   fmt.Sprintf("%s/%s/%s.tar.gz", github.Name, release.Version, github.Name),
			Headers: headers,
		})
	}
}

func (task *Task) AddGitHubReleases(github GitHub) {
	headers := github.accessTokenHeader()
	for _, release := range github.releases {
		for _, file := range release.Files {
			task.Downloads = append(task.Downloads, Download{
				Remote:  file.Url,
				Local:   fmt.Sprintf("%s/%s/%s", github.Name, release.Version, file.Name),
				Headers: headers,
			})
		}
	}
}

func (github *GitHub) GetGitHubReleases() error {
	releasesUrl := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", github.Owner, github.Repository)
	log.Printf("Downloading list from %s\n", releasesUrl)
	var err error
	var req *http.Request
	var resp *http.Response
	client := &http.Client{}
	req, err = http.NewRequest("GET", releasesUrl, nil)
	if err != nil {
		return err
	}
	if len(github.AccessToken) > 0 {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", github.AccessToken))
	}
	resp, err = client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	limit := resp.Header.Get("X-RateLimit-Limit")
	remaining := resp.Header.Get("X-RateLimit-Remaining")
	if len(limit) > 0 && len(remaining) > 0 {
		log.Printf("Rate limit: %s, %s remaining.\n", limit, remaining)
	}

	var body []byte
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return errors.New(tryToParseErrorMessage(body))
	}

	err = json.Unmarshal(body, &github.releases)
	if err != nil {
		return err
	}

	log.Println("List downloaded.")
	return nil
}

func (github GitHub) accessTokenHeader() http.Header {
	headers := http.Header{}
	if len(github.AccessToken) > 0 {
		headers.Add("Authorization", fmt.Sprintf("token %s", github.AccessToken))
	}
	return headers
}
