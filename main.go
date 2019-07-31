package main

import (
	"bufio"
	"errors"
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

const spacebar = " "

var nCores = runtime.NumCPU()

func isTextFile(fp string) bool {
	buffer := make([]byte, 512)
	f, err := os.Open(fp)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	f.Read(buffer)
	return strings.Contains(http.DetectContentType(buffer), "text")
}

func pathExist(p string) bool {
	if _, err := os.Stat(p); os.IsNotExist(err) {
		return false
	}
	return true
}

func search(path, substr string) bool {
	if !isTextFile(path) {
		return false
	}
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

func worker(substr string, fpaths chan string, matches []string, wg *sync.WaitGroup) {
	for fp := range fpaths {
		if search(fp, substr) {
			matches = append(matches, fp)
		}
	}
	wg.Done()
}

func GoFindSubString(substr string, paths []string) []string {
	matches := make([]string, 0) // a slice of filepaths that contain the substring
	fpaths := make(chan string)
	wg := new(sync.WaitGroup)

	// Insuring that all input filepaths exist.
	for _, p := range paths {
		if !pathExist(p) {
			m := fmt.Sprintf("[ERROR] %q does not exist", p)
			panic(errors.New(m))
		}
	}

	// Insuring that all paths are absolute
	for i, path := range paths {
		path, err := filepath.Abs(path)
		if err != nil {
			panic(err)
		}
		paths[i] = path
	}

	// Starting go routines
	for i := 0; i < nCores; i++ {
		wg.Add(1)
		go worker(substr, fpaths, matches, wg)
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
		} else {
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
