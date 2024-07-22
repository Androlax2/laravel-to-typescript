package main

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"sync"
)

var (
	enumRegex = regexp.MustCompile(`(?i)enum\s+\w+`)
)

func isPhpFile(path string) bool {
	return filepath.Ext(path) == ".php"
}

func processFilesInPath(directoryPath string, phpFilePaths chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Printf("Processing files in path %s\n", directoryPath)

	err := filepath.Walk(directoryPath, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() && path != directoryPath {
			wg.Add(1)
			go processFilesInPath(path, phpFilePaths, wg)
			return filepath.SkipDir
		}

		if !isPhpFile(path) {
			return nil
		}

		fmt.Printf("Send %s to phpFilePaths\n", path)
		phpFilePaths <- path

		return nil
	})

	if err != nil {
		fmt.Printf("There was an error walking the directory path %s\n", directoryPath)
	}
}

func processPhpFiles(paths <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()

	for path := range paths {
		fmt.Printf("Process PHP file %s\n", path)
	}
}

func main() {
	fmt.Println("Starting processing files...")

	var wgDiscovery sync.WaitGroup
	var wgProcessing sync.WaitGroup

	phpFilePaths := make(chan string)

	wgProcessing.Add(1)
	go processPhpFiles(phpFilePaths, &wgProcessing)

	err := filepath.WalkDir("tests/fixtures/app", func(path string, di fs.DirEntry, err error) error {
		if di.IsDir() {
			wgDiscovery.Add(1)
			go processFilesInPath(path, phpFilePaths, &wgDiscovery)
			return filepath.SkipDir
		}

		return nil
	})

	if err != nil {
		fmt.Println("There was an error walking the path.")
		return
	}

	wgDiscovery.Wait()
	close(phpFilePaths)
	wgProcessing.Wait()

	fmt.Println("Typescript types has been generated.")

}
