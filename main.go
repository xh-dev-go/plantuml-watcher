package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/xh-dev-go/xhUtils/flagUtils/flagBool"
	"github.com/xh-dev-go/xhUtils/flagUtils/flagString"
	"log"
	"net/http"
	"os"
	"strings"
)

func save(pumlUrl string, fileName string) {
	rawFileName := strings.Split(fileName, ".")[0]
	b, err := os.ReadFile(fileName)
	if err != nil {
		panic(err)
	}
	b = []byte(strings.ReplaceAll(string(b), "\n", ""))

	for n := 1; n <= 2; n++ {
		sEnc := "~h" + hex.EncodeToString(b)
		var outType string
		if n == 1 {
			outType = "png"
		}
		if n == 2 {
			outType = "svg"
		}
		resp, err := http.Get(fmt.Sprintf("%s/%s/%s", pumlUrl, outType, sEnc))
		if err != nil {
			log.Printf("Error processing: " + err.Error())
		}
		defer resp.Body.Close()
		if resp.StatusCode == 200 {
			f, e := os.Create(rawFileName + "." + outType)
			if e != nil {
				panic(e)
			}
			defer f.Close()
			f.ReadFrom(resp.Body)
		} else {
			println("Error: " + resp.Status)
		}
	}

}

type handler func(path string)

func wallThroughDirectory(dir string, fn handler) {
	fn(dir)
	files, err := os.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		if file.IsDir() {
			newFileName := dir + "/" + file.Name()
			wallThroughDirectory(newFileName, fn)
		}
	}
}

func main() {
	showOnlyFlag := flagBool.NewDefault("showOnly", "Show the directory to be add only", false).BindCmd()
	dirFlag := flagString.NewDefault("dir", ".", "The directory to be watched").BindCmd()
	urlFlag := flagString.NewDefault("url", "https://plantuml.com/plantuml", "The default url to use").BindCmd()
	flag.Parse()

	if showOnlyFlag.Value() {
		wallThroughDirectory(dirFlag.Value(), func(path string) {
			println("To be added path: " + path)
		})
		return
	}

	pumlUrl := urlFlag.Value()

	watcher, err := fsnotify.NewWatcher()
	wallThroughDirectory(dirFlag.Value(), func(path string) {
		//println("Adding path: " + path)
		watcher.Add(path)
	})

	for _, item := range watcher.WatchList() {
		println("Watch list: " + item)
	}

	if err != nil {
		log.Fatalln("NewWatcher failed: ", err)
	}

	defer watcher.Close()
	done := make(chan bool)

	go func() {
		defer close(done)

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if strings.HasSuffix(event.Name, ".puml") && (event.Op == fsnotify.Create || event.Op == fsnotify.Write) {
					log.Printf("%s %s\n", event.Name, event.Op)
					save(pumlUrl, event.Name)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add("./")
	if err != nil {
		log.Fatal("Add failed:", err)
	}
	<-done
}
