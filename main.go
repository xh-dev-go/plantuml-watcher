package main

import (
	"bufio"
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

func refresh(pumlUrl, dir string) {
	wallThroughDirectory(dir, func(path string) {}, func(path string) {
		if strings.HasSuffix(path, ".puml") {
			save(pumlUrl, path)
		}
	})
}

func save(pumlUrl string, fileName string) {
	fmt.Println("Process: " + fileName)
	rawFileName := fileName[0 : len(fileName)-5]
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

func wallThroughDirectory(dir string, handleDirectory handler, handleFile handler) {
	handleDirectory(dir)
	files, err := os.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		newFileName := dir + "/" + file.Name()
		if file.IsDir() {
			wallThroughDirectory(newFileName, handleDirectory, handleFile)
		} else {
			handleFile(newFileName)
		}
	}
}

func main() {
	showOnlyFlag := flagBool.NewDefault("showOnly", "Show the directory to be add only", false).BindCmd()
	dirFlag := flagString.NewDefault("dir", ".", "The directory to be watched").BindCmd()
	urlFlag := flagString.NewDefault("url", "https://plantuml.com/plantuml", "The default url to use").BindCmd()
	flag.Parse()

	if showOnlyFlag.Value() {
		wallThroughDirectory(dirFlag.Value(),
			func(path string) {
				println("To be added path: " + path)
			},
			func(path string) {},
		)
		return
	}

	pumlUrl := urlFlag.Value()

	fmt.Println("Refresh all puml files before start watching")
	refresh(pumlUrl, dirFlag.Value())
	fmt.Println("Complete - Refresh all puml files before start watching")

	watcher, err := fsnotify.NewWatcher()
	wallThroughDirectory(dirFlag.Value(), func(path string) {
		watcher.Add(path)
	}, func(path string) {})

	for _, item := range watcher.WatchList() {
		println("Watch list: " + item)
	}

	if err != nil {
		log.Fatalln("NewWatcher failed: ", err)
	}

	defer watcher.Close()

	go func() {

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
				file, err := os.Open(event.Name)
				if err != nil {
					println(err.Error())
				}
				fs, err := file.Stat()
				if err != nil {
					println(err.Error())
				}
				if event.Op == fsnotify.Create && fs.IsDir() {
					println("Runtime add directory: " + event.Name)
					watcher.Add(event.Name)
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

	input := bufio.NewScanner(os.Stdin)

	fmt.Println("Input \"refresh\" to refresh all *.puml files")
	fmt.Println("Input \"quit\" to exit program")
	for input.Scan() {
		text := input.Text()
		if text == "refresh" {
			refresh(pumlUrl, dirFlag.Value())
		} else if text == "quit" {
			return
		}
		fmt.Println("Input \"refresh\" to refresh all *.puml files")
		fmt.Println("Input \"quit\" to exit program")
	}
}
