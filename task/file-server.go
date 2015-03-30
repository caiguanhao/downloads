package task

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"regexp"
	"sort"
	"strings"
)

type (
	FileServer struct {
		Name        string
		Source      string
		Grep        string
		GrepNamePos int
		GrepVerPos  []int
		SortByVer   bool

		files []string
	}

	ByVersion struct {
		Ver    [][]string
		VerPos []int
	}
)

func (bv ByVersion) Len() int      { return len(bv.Ver) }
func (bv ByVersion) Swap(i, j int) { bv.Ver[i], bv.Ver[j] = bv.Ver[j], bv.Ver[i] }
func (bv ByVersion) Less(i, j int) bool {
	var bvi1, bvj1, bvi2, bvj2, bvi3, bvj3 int
	fmt.Sscanf(bv.Ver[i][bv.VerPos[0]], "%d", &bvi1)
	fmt.Sscanf(bv.Ver[j][bv.VerPos[0]], "%d", &bvj1)
	fmt.Sscanf(bv.Ver[i][bv.VerPos[1]], "%d", &bvi2)
	fmt.Sscanf(bv.Ver[j][bv.VerPos[1]], "%d", &bvj2)
	fmt.Sscanf(bv.Ver[i][bv.VerPos[2]], "%d", &bvi3)
	fmt.Sscanf(bv.Ver[j][bv.VerPos[2]], "%d", &bvj3)
	if bvi1 == bvj1 {
		if bvi2 == bvj2 {
			return bvi3 < bvj3
		}
		return bvi2 < bvj2
	}
	return bvi1 < bvj1
}

func (fs *FileServer) GetLinks() error {
	log.Printf("Downloading page from %s\n", fs.Source)
	resp, err := http.Get(fs.Source)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	re := regexp.MustCompile(fs.Grep)
	links := re.FindAllStringSubmatch(string(content), -1)
	if fs.SortByVer {
		sort.Sort(ByVersion{
			Ver:    links,
			VerPos: fs.GrepVerPos,
		})
	}
	for _, link := range links {
		fs.files = append(fs.files, joinPath(fs.Source, link[fs.GrepNamePos]))
	}
	log.Printf("Found %d files.\n", len(fs.files))
	return nil
}

func (task *Task) AddFileServerLinks(fs FileServer) {
	for _, file := range fs.files {
		basename := path.Base(file)
		task.Downloads = append(task.Downloads, Download{
			Remote: file,
			Local:  fmt.Sprintf("%s/%s", fs.Name, basename),
		})
	}
}

func joinPath(source, fpath string) string {
	if !strings.HasSuffix(source, "/") {
		source += "/"
	}
	return source + fpath
}
