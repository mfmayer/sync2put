package main

import (
	"crypto/tls"
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

func putFile(filePath string, url string, appendFileName bool, user, pwd string) {
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		file, err := os.Open(filePath)
		if err != nil {
			log.Println("Error while opening: ", filePath)
			return
		}
		baseName := ""
		if appendFileName {
			if !strings.HasSuffix(url, "/") {
				url = url + "/"
			}
			baseName = filepath.Base(filePath)
		}

		req, err := http.NewRequest(*method, url+baseName, file)
		if err != nil {
			log.Println("Error while creating PUT request: ", err)
			return
		}
		if len(user) > 0 || len(pwd) > 0 {
			req.SetBasicAuth(user, pwd)
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
	dir := flag.String("dir", "", fmt.Sprintf("Directory to sync (e.g. \"%v\")", wd))
	url := flag.String("url", "", "Target URL where to sync files to (e.g. \"http://192.168.200.1:3001/rsc/\")")
	auth := flag.String("auth", "", "Basic authentication in the form: \"<user>:<pwd>\"")
	appendFileName := flag.Bool("append", true, "Append file name to URL")
	method = flag.String("method", "PUT", "HTTP Method to use")
	syncOnStart := flag.Bool("s", true, "Synchronize whole directory on start")
	allowInsecure := flag.Bool("k", false, "Allow inscure connections with untrusted host certificates")
	flag.Parse()
	if *allowInsecure {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	if *dir == "" || *url == "" {
		fmt.Printf("Flags -dir and -url are needed. Try %v --help\n", os.Args[0])
		return
	}
	user := ""
	pwd := ""
	if len(*auth) > 0 {
		up := strings.Split(*auth, ":")
		if len(up) != 2 {
			fmt.Printf("auth not given in form \"<user>:<pwd>\". Try %v --help\n", os.Args[0])
			return
		}
		user = up[0]
		pwd = up[1]
	}

	if _, err := os.Stat(*dir); os.IsNotExist(err) {
		fmt.Printf("Directory %v doesn't exist. Try %v --help\n", *dir, os.Args[0])
		return
	}
	if !strings.HasPrefix(*url, "http") {
		fmt.Printf("url %v is invalid. Try %v --help\n", *url, os.Args[0])
		return
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
				putFile(*dir+"/"+file.Name(), *url, *appendFileName, user, pwd)
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
						putFile(event.Name, *url, *appendFileName, user, pwd)
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
