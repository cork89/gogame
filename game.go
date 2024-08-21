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
	Value   string
	Left    *Node
	Right   *Node
	Forward *Node
}

var gameMap2 map[string]*Node

func hello(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/game/1", http.StatusSeeOther)
}

func changeNode(node Node, path string) *Node {
	if path == "left" {
		return node.Left
	}
	if path == "forward" {
		return node.Forward
	}
	return node.Right
}

func getNode(w http.ResponseWriter, r *http.Request) (*Node, error) {
	m := validPath.FindStringSubmatch(r.URL.Path)
	log.Printf("%s - %s - %s\n", r.URL.Query(), r.URL.Path, m)

	if m == nil {
		return nil, errors.New("invalid path")
	}

	tempNode, validNode := gameMap2[fmt.Sprintf("%s", m[2])]
	if !validNode {
		return nil, errors.New("invalid node")
	}

	if r.URL.Query().Has("path") {
		tempNode = changeNode(*tempNode, r.URL.Query().Get("path"))
		http.Redirect(w, r, fmt.Sprintf("/game/%s", tempNode.Value), http.StatusSeeOther)
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
	case gameMap2["5green"]:
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
	gameMap2 = make(map[string]*Node)
	node1 := Node{Value: "1"}
	node2 := Node{Value: "2"}
	node3 := Node{Value: "3"}
	node4 := Node{Value: "4"}
	node5 := Node{Value: "5green"}
	gameMap2["1"] = &Node{Value: node1.Value, Left: &node3, Right: &node4, Forward: &node2}
	gameMap2["2"] = &Node{Value: node2.Value, Left: &node1, Right: &node4}
	gameMap2["3"] = &Node{Value: node3.Value, Left: &node2, Right: &node1}
	gameMap2["4"] = &Node{Value: node4.Value, Left: &node5, Right: &node2, Forward: &node3}
	gameMap2["5green"] = &Node{Value: node5.Value}

	tmpl = make(map[string]*template.Template)

	funcMap := template.FuncMap{
		"IsForward": func(forward *Node) bool {
			if forward != nil && forward.Value != "" {
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
