package main

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"regexp"
)

var tmpl map[string]*template.Template

var validPath = regexp.MustCompile("^/(game)/([a-zA-Z0-9]+)$")

type Node struct {
	Node    string
	Left    string
	Right   string
	Forward string
}

var gameMap map[string]*Node

func hello(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/game/1", http.StatusSeeOther)
}

func changeNode(node Node, path string) *Node {
	if path == "left" {
		return gameMap[node.Left]
	}
	if path == "forward" {
		return gameMap[node.Forward]
	}
	return gameMap[node.Right]
}

func getNode(w http.ResponseWriter, r *http.Request) (*Node, error) {
	m := validPath.FindStringSubmatch(r.URL.Path)
	log.Printf("%s - %s - %s\n", r.URL.Query(), r.URL.Path, m)

	if m == nil {
		return nil, errors.New("invalid path")
	}

	tempNode, validNode := gameMap[fmt.Sprintf("%s", m[2])]
	if !validNode {
		return nil, errors.New("invalid node")
	}

	if r.URL.Query().Has("path") {
		tempNode = changeNode(*tempNode, r.URL.Query().Get("path"))
		http.Redirect(w, r, fmt.Sprintf("/game/%s", tempNode.Node), http.StatusSeeOther)
		return tempNode, nil
	}
	return tempNode, nil
}

func gameHandler(w http.ResponseWriter, r *http.Request) {
	node, err := getNode(w, r)

	if err != nil {
		log.Println(err.Error())
	}

	switch node {
	case nil:
		err = tmpl["notfound"].ExecuteTemplate(w, "base", node)
	case gameMap["5green"]:
		err = tmpl["winner"].ExecuteTemplate(w, "base", node)
	default:
		err = tmpl["game"].ExecuteTemplate(w, "base", node)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {
	gameMap = make(map[string]*Node)
	gameMap["1"] = &Node{Node: "1", Left: "3", Right: "4", Forward: "2"}
	gameMap["2"] = &Node{Node: "2", Left: "1", Right: "4", Forward: ""}
	gameMap["3"] = &Node{Node: "3", Left: "2", Right: "1", Forward: ""}
	gameMap["4"] = &Node{Node: "4", Left: "5green", Right: "2", Forward: "3"}
	gameMap["5green"] = &Node{Node: "5green", Left: "", Right: "", Forward: ""}
	gameMap["404"] = &Node{}

	tmpl = make(map[string]*template.Template)

	funcMap := template.FuncMap{
		"IsForward": func(forward string) bool {
			if forward != "" {
				return true
			}
			return false
		},
	}

	tmpl["notfound"] = template.Must(template.ParseFiles("static/notfound.html", "static/base.html"))
	tmpl["winner"] = template.Must(template.ParseFiles("static/winner.html", "static/base.html"))
	tmpl["game"] = template.Must(template.New("game").Funcs(funcMap).ParseFiles("static/game.html", "static/base.html"))

	http.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		http.ServeFile(w, r, "static/style.css")
	})

	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("images"))))

	http.HandleFunc("/game/", gameHandler)
	http.HandleFunc("/", hello)
	http.ListenAndServe(":8090", nil)
}
