package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

const (
	spacebar = " "
)

// TODO: remove package variables
var (
	nCores  = runtime.NumCPU()
	matches = make([]string, 0)
	channel = make(chan string, 0)
	wg      = new(sync.WaitGroup)
)

func search(substring, path string) bool {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), substring) {
			return true
		}
	}
	return false
}

func searcher(s string) {
	for path := range channel {
		if path == "" {
			break
		}

		// TODO: Consider moving the creation of an absolute path to improve performance.
		path, err := filepath.Abs(path)
		if err != nil {
			panic(err)
		}

		stat, err := os.Stat(path)
		if err != nil {
			panic(err)
		}

		if stat.IsDir() {
			content, err := ioutil.ReadDir(path)
			if err != nil {
				panic(err)
			}

			for _, f := range content {
				channel <- filepath.Join(f.Name(), path)
			}
			continue
		}
		if stat.Mode().IsRegular() && search(s, path) {
			matches = append(matches, path)
		}
	}
	wg.Done()
}

func GoFindSubString(s string, paths []string, recurse bool) []string {
	for i := 0; i < nCores; i++ {
		wg.Add(1)
		go searcher(s)
	}

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
				paths = append(paths, item.Name())
			}
		}
	}

	close(channel)
	wg.Wait()

	return matches
}

func main() {
	recusrive := flag.Bool("recursive", false, "If set to true the subdirectories will be recursively searched.")
	substring := flag.String("substring", "", "The search string.")
	paths := flag.String("paths", "./", "A list paths the that will be searched separated by a space.")

	flag.Parse()

	if *substring == "" {
		panic("[FAIL] no substring argument was passed.")
	}

	GoFindSubString(*substring, strings.Split(*paths, spacebar), *recusrive)

	for _, match := range matches {
		fmt.Println(match)
	}
}
