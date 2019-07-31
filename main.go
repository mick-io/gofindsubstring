package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

const (
	spacebar = " "
)

var nCores = runtime.NumCPU()

func isTextFile(f *os.File) bool {
	buffer := make([]byte, 512)
	_, err := f.Read(buffer)
	if err != nil {
		panic(err)
	}
	ct := http.DetectContentType(buffer)
	return strings.Contains(ct, "text")
}

func search(path, substr string) bool {
	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		panic(err)
	}
	if isTextFile(f) {
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			if strings.Contains(scanner.Text(), substr) {
				return true
			}
		}
	}
	return false
}

func searcher(substr string, fpaths chan string, matches []string, wg *sync.WaitGroup) {
	for fpath := range fpaths {
		if search(fpath, substr) {
			matches = append(matches, fpath)
		}
	}
	wg.Done()
}

func GoFindSubString(substr string, paths []string) []string {
	matches := make([]string, 0) // a slice of filepaths that contain the substring
	fpaths := make(chan string)
	wg := new(sync.WaitGroup)

	for i := 0; i < nCores; i++ {
		wg.Add(1)
		go searcher(substr, fpaths, matches, wg)
	}

	// TODO: Insure that input paths exist

	// Insuring that all paths are absolute
	for i, path := range paths {
		path, err := filepath.Abs(path)
		if err != nil {
			panic(err)
		}
		paths[i] = path
	}

	// Feeding input paths
	for i := 0; i < len(paths); i++ {
		path := paths[i]
		stat, err := os.Stat(path)
		if err != nil {
			panic(err)
		}
		if stat.IsDir() {
			// TODO: Consider reading a directory with a lower level function to improve performance.
			contents, err := ioutil.ReadDir(stat.Name())
			if err != nil {
				panic(err)
			}
			for _, item := range contents {
				p := filepath.Join(path, item.Name())
				paths = append(paths, p)
			}
			continue
		}
		if stat.Mode().IsRegular() {
			fpaths <- path
		}
	}

	close(fpaths)
	wg.Wait()
	return matches
}

func main() {
	substring := flag.String("substring", "", "The search string.")
	paths := flag.String("paths", "./", "A list paths the that will be searched separated by a space.")

	flag.Parse()

	if *substring == "" {
		panic("[FAIL] no substring argument was passed.")
	}

	matches := GoFindSubString(*substring, strings.Split(*paths, spacebar))

	for _, match := range matches {
		fmt.Println(match)
	}
}
