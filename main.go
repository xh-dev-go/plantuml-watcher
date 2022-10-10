package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/fsnotify/fsnotify"
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
		}
	}

}

func main() {
	urlFlag := flagString.NewDefault("url", "https://plantuml.com/", "The default url to use").BindCmd()
	//urlFlag := flagString.NewDefault("url", "http://localhost:4567/", "The default url to use").BindCmd()
	flag.Parse()
	pumlUrl := urlFlag.Value()

	watcher, err := fsnotify.NewWatcher()

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
