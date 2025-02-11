package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
)

func copyFile(src, dest string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}
	return destFile.Sync()
}

func worker(id int, wg *sync.WaitGroup, tasks <-chan [2]string) {
	defer wg.Done()
	for task := range tasks {
		src, dest := task[0], task[1]
		fmt.Printf("[Worker %d] Copying %s to %s \n", id, src, dest)
		err := copyFile(src, dest)
		if err != nil {
			log.Printf("[Worker %d] Error copying file %v \n", id, err)
		} else {
			fmt.Printf("[Worker %d] Copy completed: %s\n", id, dest)
		}
	}
}

func deletionWorker(id int, wg *sync.WaitGroup, deletions <-chan string) {
	defer wg.Done()
	for path := range deletions {
		fmt.Printf("[Deletion Worker %d] Removing: %s\n", id, path)
		err := os.Remove(path)
		if err != nil {
			log.Printf("[Deletion worker %d] Error removing file %v\n", id, err)
		} else {
			fmt.Printf("[Deletion Worker %d] File removed: %s\n", id, path)
		}
	}
}

func watchDirectory(srcDir, destDir string, tasks chan<- [2]string, deletions chan<- string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	err = watcher.Add(srcDir)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Watching directory:", srcDir)
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			fmt.Println("Event:", event)

			relPath, _ := filepath.Rel(srcDir, event.Name)
			targetPath := filepath.Join(destDir, relPath)

			switch {
			case event.Op&fsnotify.Create == fsnotify.Create:
				fmt.Println("File created:", event.Name)
				tasks <- [2]string{event.Name, targetPath}

			case event.Op&fsnotify.Write == fsnotify.Write:
				fmt.Println("File modified:", event.Name)
				tasks <- [2]string{event.Name, targetPath}

			case event.Op&fsnotify.Remove == fsnotify.Remove:
				fmt.Println("File removed:", event.Name)
				deletions <- targetPath

			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			fmt.Println("Error:", err)
		}

	}
}

func main() {
	source := os.Args[1]
	target := os.Args[2]
	tasks := make(chan [2]string, 10)
	deletions := make(chan string, 10)

	var wg sync.WaitGroup

	numWorkers := 4
	for i := 1; i <= numWorkers; i++ {
		wg.Add(1)
		go worker(i, &wg, tasks)
	}

	numDeletionWorkers := 2
	for i := 1; i <= numDeletionWorkers; i++ {
		wg.Add(1)
		go deletionWorker(i, &wg, deletions)
	}

	go watchDirectory(source, target, tasks, deletions)

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt)
	<-stopChan

	fmt.Println("\nShutting down...")

	close(tasks)
	close(deletions)

	wg.Wait()
	fmt.Println("File Synchronizer stopped.")

}
