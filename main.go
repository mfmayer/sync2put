package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

var method *string

func putFile(filePath string, url string, appendFileName bool) {
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		file, err := os.Open(filePath)
		if err != nil {
			log.Println("Error while opening: ", filePath)
			return
		}
		baseName := ""
		if appendFileName {
			baseName = filepath.Base(filePath)
		}

		req, err := http.NewRequest(*method, url+baseName, file)
		if err != nil {
			log.Println("Error while creating PUT request: ", err)
			return
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Println("Error sending PUT request: ", err)
			return
		}
		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Println("Error while reading response body: ", err)
			return
		}
		log.Println(url+baseName, string(body))
	}
}

func main() {
	wd, _ := os.Getwd()
	dir := flag.String("dir", wd, "Directory to sync")
	url := flag.String("url", "http://192.168.200.1:3001/rsc/", "Target URL where to put to")
	appendFileName := flag.Bool("appendFileName2URL", true, "Append file name to URL")
	method = flag.String("method", "PUT", "HTTP Method to use")
	syncOnStart := flag.Bool("s", true, "Synchronize whole directory on start")
	flag.Parse()

	if _, err := os.Stat(*dir); os.IsNotExist(err) {
		log.Fatal("Directory ", *dir, " doesn't exist")
	}
	if !strings.HasPrefix(*url, "http") {
		log.Fatal("Url ", *url, " is invalid")
	}
	if !strings.HasSuffix(*url, "/") {
		*url = *url + "/"
	}

	fmt.Println("dir: ", *dir)
	fmt.Println("url: ", *url)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)

	if *syncOnStart {
		files, err := ioutil.ReadDir(*dir)
		if err != nil {
			log.Fatal(err)
		}

		for _, file := range files {
			if !file.IsDir() {
				putFile(*dir+"/"+file.Name(), *url, *appendFileName)
			}
		}
	}

	duration := 100 * time.Millisecond
	timer := time.NewTimer(duration)
	go func() {
		events := make(map[string]fsnotify.Event)
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				events[event.String()] = event
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			case <-timer.C:
				for _, event := range events {
					if event.Op&fsnotify.Write == fsnotify.Write {
						putFile(event.Name, *url, *appendFileName)
					}
				}
				events = make(map[string]fsnotify.Event)
				timer.Reset(duration)
			}
		}
	}()

	err = watcher.Add(*dir)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}
