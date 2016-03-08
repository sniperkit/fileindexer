package main

import (
	"crypto/md5"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/olivere/elastic.v3"
)

var extensions []string
var client *elastic.Client
var indexName = "sources"

type fileEntry struct {
	Content string
	Path    string
}

func main() {
	exts := flag.String("exts", ".cs;.sql", "File extensions for search, sample: .txt;.sql")
	rootPath := flag.String("root", "empty", "Root path")
	flag.Parse()
	log.Printf("Search files with extensions: %s", *exts)
	var err error
	client, err = elastic.NewClient()
	if err != nil {
		panic(err)
	}

	checkIndex(client)
	extensions = strings.Split(*exts, ";")
	err = filepath.Walk(*rootPath, visit)
	if err != nil {
		log.Fatalf("error: %s", err)
	}
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func checkIndex(client *elastic.Client) {
	exists, err := client.IndexExists(indexName).Do()
	if err != nil {
		// Handle error
		panic(err)
	}
	// Create an index
	if !exists {
		_, err = client.CreateIndex(indexName).Do()
		if err != nil {
			// Handle error
			panic(err)
		}
	}
}

func visit(path string, f os.FileInfo, err error) error {
	if !f.IsDir() && contains(extensions, filepath.Ext(path)) {
		log.Printf("Visited: %s\n", path)
		fileContent := readFile(path)
		id := md5.Sum([]byte(path))
		fileItem := fileEntry{Content: fileContent, Path: path}
		log.Printf("Indexing: id %x, path %s", id, path)
		_, err := client.Index().Index(indexName).Type("source").Id(fmt.Sprintf("%x", id)).BodyJson(fileItem).Do()
		if err != nil {
			panic(err)
		}
	}
	return nil
}

func readFile(path string) string {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return string(file)
}
