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

// TODO: Add count occurrences feature

const spacebar = " "

var (
	nCores             = runtime.NumCPU()
	filePathChan       = make(chan string, nCores)
	filesWithSubString = make([]string, 0)
)

func args() (string, []string) {
	substr := flag.String("substring", "", "The search string.")
	paths := flag.String("paths", "./", "A list paths the that will be searched separated by a space.")
	flag.Parse()
	if *substr == "" {
		panic("[FAIL] no substring argument was passed.")
	}
	return *substr, strings.Split(*paths, spacebar)
}

func main() {
	feedWG, workersWG := new(sync.WaitGroup), new(sync.WaitGroup)
	substring, searchPaths := args()
	feedWG.Add(1)
	go feedFilePathChannel(searchPaths, feedWG)

	// Starting workers
	for i := 0; i < nCores; i++ {
		workersWG.Add(1)
		go worker(substring, workersWG)
	}

	feedWG.Wait()
	close(filePathChan)
	workersWG.Wait()

	for _, fp := range filesWithSubString {
		fmt.Println(fp)
	}
}

func feedFilePathChannel(searchPaths []string, wg *sync.WaitGroup) {
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
			filePathChan <- p
		}
	}
	wg.Done()
}

func worker(substr string, wg *sync.WaitGroup) {
	for fp := range filePathChan {
		if !isTextFile(fp) {
			continue
		}
		if search(fp, substr) {
			filesWithSubString = append(filesWithSubString, fp)
		}
	}
	wg.Done()
}

func isTextFile(fp string) bool {
	cmd := exec.Command("file", fp)
	res, err := cmd.Output()
	if err != nil {
		panic(err)
	}
	return strings.Contains(string(res), "text")
}

// TODO: Alter function to work with multi-line substrings.
func search(path, substr string) bool {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), substr) {
			return true
		}
	}
	return false
}
