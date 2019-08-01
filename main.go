package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// TODO: Add count occurrences feature.
// TODO: Write unit test.
// TODO: Write readme.

const spacebar = " "

var filesWithSubString = make([]string, 0)

func main() {
	nCores := runtime.NumCPU()
	filePathChan := make(chan string, nCores)

	feedWG, workersWG := new(sync.WaitGroup), new(sync.WaitGroup)
	substring, searchPaths := args()

	feedWG.Add(1)
	go feedFilePathChannel(searchPaths, filePathChan, feedWG)

	// Starting workers
	for i := 0; i < nCores; i++ {
		workersWG.Add(1)
		go worker(substring, filePathChan, workersWG)
	}

	feedWG.Wait()
	close(filePathChan)
	workersWG.Wait()

	for _, fp := range filesWithSubString {
		fmt.Println(fp)
	}
}

func args() (string, []string) {
	substring := flag.String("substring", "", "The search string.")
	paths := flag.String("paths", "./", "A list paths the that will be searched separated by a space.")
	flag.Parse()
	if *substring == "" {
		panic("[FAIL] no substring argument was passed.")
	}
	return *substring, strings.Split(*paths, spacebar)
}

func feedFilePathChannel(searchPaths []string, fpChan chan string, wg *sync.WaitGroup) {
	for i := 0; i < len(searchPaths); i++ {
		p := searchPaths[i]
		fi, err := os.Stat(p)
		if err != nil {
			panic(err)
		}
		if fi.IsDir() {
			stats, err := ioutil.ReadDir(p)
			if err != nil {
				panic(err)
			}
			for _, stat := range stats {
				subpath := filepath.Join(p, stat.Name())
				searchPaths = append(searchPaths, subpath)
			}
		} else {
			fpChan <- p
		}
	}
	wg.Done()
}

func worker(substring string, fpChan chan string, wg *sync.WaitGroup) {
	for fp := range fpChan {
		if !isTextFile(fp) {
			continue
		}
		if search(fp, substring) {
			filesWithSubString = append(filesWithSubString, fp)
		}
	}
	wg.Done()
}

// TODO: Make Windows friendly
func isTextFile(fp string) bool {
	cmd := exec.Command("file", fp)
	res, err := cmd.Output()
	if err != nil {
		panic(err)
	}
	return strings.Contains(string(res), "text")
}

// TODO: Alter function to work with multi-line substrings.
func search(path, substring string) bool {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), substring) {
			return true
		}
	}
	return false
}
