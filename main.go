package main

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
)

// Templates
var templates *template.Template

// Database Setup
var client *redis.Client

func main() {
	// Connect Database
	client = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	// Set Template Directory
	templates = template.Must(template.ParseGlob("templates/*.html"))
	// Routing
	r := mux.NewRouter()
	r.HandleFunc("/", getPosts).Methods("GET")
	r.HandleFunc("/posts", getPosts).Methods("GET")
	r.HandleFunc("/posts/new", newPost).Methods("GET")
	r.HandleFunc("/posts/create", createPost).Methods("POST")
	r.HandleFunc("/posts/{key}", deletePost).Methods("POST")
	r.HandleFunc("/posts/{key}/edit", editPost).Methods("GET")
	r.HandleFunc("/posts/{key}/update", updatePost).Methods("POST")
	r.HandleFunc("/posts/{key}", getPost).Methods("GET")
	// Create Server
	http.Handle("/", r)
	fmt.Println("Server Starting at 3000...")
	http.ListenAndServe(":3000", nil)
}

// Get All Posts
func getPosts(w http.ResponseWriter, r *http.Request) {
	// Get Posts Names
	postsNames, _, err := client.Scan(0, "posts:*", 1000).Result()
	// Get Posts Data
	var posts map[string]interface{}
	posts = make(map[string]interface{})
	if err != nil {
		fmt.Println(err)
	}
	for i, _ := range postsNames {
		postsItems, _ := client.HGetAll(postsNames[i]).Result()
		if postsNames[i] != "posts:next-id" {
			posts[postsNames[i]] = postsItems
		}
	}
	templates.ExecuteTemplate(w, "index.html", posts)

}

// Get One Post
func getPost(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	post, err := client.HGetAll(params["key"]).Result()
	if err != nil {
		return
	}
	data := make(map[string]interface{})
	data["post"] = post
	data["key"] = params["key"]
	templates.ExecuteTemplate(w, "show.html", data)
}

// New Post Page
func newPost(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "new.html", nil)
}

// New Post action
func createPost(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	title := r.PostForm.Get("title")
	body := r.PostForm.Get("body")
	if title != "" && body != "" {
		id, err := client.Incr("posts:next-id").Result()
		if err != nil {
			return
		}
		key := fmt.Sprintf("posts:%d", id)
		m := make(map[string]interface{})
		m["title"] = title
		m["body"] = body
		client.HMSet(key, m)
		http.Redirect(w, r, "/posts", 302)
	} else {
		http.Redirect(w, r, "/posts/new", 302)
	}
}

// Delete Post
func deletePost(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	client.Del(params["key"])
	http.Redirect(w, r, "/posts", 302)
}

// Edit Post
func editPost(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	post, err := client.HGetAll(params["key"]).Result()
	if err != nil {
		fmt.Println(err)
	}
	data := map[string]interface{}{
		"post": post,
		"key":  params["key"],
	}
	templates.ExecuteTemplate(w, "edit.html", data)
}

// Edit Post Action
func updatePost(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	params := mux.Vars(r)
	title := r.PostForm.Get("title")
	body := r.PostForm.Get("body")
	if title != "" && body != "" {
		m := map[string]interface{}{
			"title": r.PostForm.Get("title"),
			"body":  r.PostForm.Get("body"),
		}
		client.HMSet(params["key"], m)
		http.Redirect(w, r, "/posts/"+params["key"], 302)
	} else {
		http.Redirect(w, r, fmt.Sprintf("/posts/%s/edit", params["key"]), 302)
	}
}
