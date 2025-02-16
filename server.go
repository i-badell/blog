package main

import (
	"bytes"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/parser"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Post represents a blog post.
type Post struct {
	Title   string
	Summary string
	Image   string
	Tags    []string
	Content template.HTML // Converted HTML from Markdown
	Date    string
}

var postsTemplate *template.Template

// init loads the posts template from the file system.
func init() {
	var err error
	postsTemplate, err = template.ParseFiles("Templates/post.html")
	if err != nil {
		log.Fatalf("Error parsing posts template: %v", err)
	}
}

// loadPosts reads all Markdown files in the "posts" directory,
// converts them to HTML using goldmark, and returns a slice of Post objects.
func loadPosts() ([]Post, error) {
	var posts []Post
	files, err := filepath.Glob("posts/*.md")
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			log.Printf("Error reading file %s: %v", file, err)
			continue
		}

		markdown := goldmark.New(
			goldmark.WithExtensions(
				meta.Meta,
			),
		)

		var buf bytes.Buffer
		context := parser.NewContext()
		if err := markdown.Convert([]byte(data), &buf, parser.WithContext(context)); err != nil {
			panic(err)
		}
		metaData := meta.Get(context)

		var parsedTags []string
		tags := metaData["Tags"].([]interface{})
		for _, i := range tags {
			parsedTags = append(parsedTags, i.(string))
		}

		date, err := time.Parse("2006/01/02", metaData["Date"].(string))
		if err != nil {
			panic(err)
		}

		post := Post{
			Title:   metaData["Title"].(string),
			Summary: metaData["Summary"].(string),
			Image:   metaData["Image"].(string),
			Tags:    parsedTags,
			Content: template.HTML(buf.String()),
			Date:    date.Format("02/01/2006"),
		}
		posts = append(posts, post)
	}
	return posts, nil
}

func postsHandler(w http.ResponseWriter, r *http.Request) {
	posts, err := loadPosts()
	if err != nil {
		http.Error(w, "Unable to load posts", http.StatusInternalServerError)
		return
	}

	if err := postsTemplate.Execute(w, posts); err != nil {
		http.Error(w, "Error rendering posts template", http.StatusInternalServerError)
	}
}
func indexHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./Public/index.html")
}

func main() {
	http.HandleFunc("/posts", postsHandler)

	fs := http.FileServer(http.Dir("public"))
	http.Handle("/", fs)

	log.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
