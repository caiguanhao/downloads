package task

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

type (
	Task struct {
		Downloads []Download
	}

	Download struct {
		Remote  string
		Local   string
		Headers http.Header

		err error
	}
)

func (download Download) DownloadFile() (int64, error) {
	var err error

	var req *http.Request
	var resp *http.Response
	client := &http.Client{}
	req, err = http.NewRequest("GET", download.Remote, nil)
	if err != nil {
		return -1, err
	}

	for key, values := range download.Headers {
		for _, value := range values {
			req.Header.Set(key, value)
		}
	}

	resp, err = client.Do(req)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		var body []byte
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return -1, err
		}
		return -1, errors.New(tryToParseErrorMessage(body))
	}
	log.Printf("Downloading %s to %s\n", download.Remote, download.Local)

	err = mkdirp(download.Local)
	if err != nil {
		return -1, err
	}

	var fileinfo os.FileInfo
	fileinfo, err = os.Stat(download.Local)
	if err != nil {
		if !os.IsNotExist(err) {
			return -1, err
		}
	} else {
		if fileinfo.Size() == resp.ContentLength {
			log.Printf("%s exists and has the same file size. Ignored.\n", download.Local)
			return resp.ContentLength, nil
		}
	}

	var file *os.File
	file, err = os.Create(download.Local)
	if err != nil {
		return -1, err
	}
	defer file.Close()

	var written int64
	written, err = io.Copy(file, resp.Body)
	if err != nil {
		return -1, err
	}
	return written, nil
}

func (task Task) DownloadFiles(concurrency int) {
	done := make(chan struct{})
	defer close(done)

	downloadListChannel := make(chan Download)
	go func() {
		for _, download := range task.Downloads {
			select {
			case downloadListChannel <- download:
			case <-done:
				return
			}
		}
		close(downloadListChannel)
	}()

	downloadChannel := make(chan Download)
	var wg sync.WaitGroup
	wg.Add(concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			for download := range downloadListChannel {
				_, err := download.DownloadFile()
				if err != nil {
					download.err = err
				}
				select {
				case downloadChannel <- download:
				case <-done:
					return
				}
			}
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(downloadChannel)
	}()

	for download := range downloadChannel {
		if download.err != nil {
			log.Println(download.err)
		}
	}
}

func mkdirp(fpath string) error {
	return os.MkdirAll(filepath.Dir(fpath), 0755)
}

func tryToParseErrorMessage(content []byte) string {
	var githubErr GitHubAPIErrorMessage
	json.Unmarshal(content, &githubErr)
	if len(githubErr.Message) > 0 {
		return githubErr.Message
	}
	return string(content)
}
