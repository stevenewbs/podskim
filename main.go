package main

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"
	"log"
	"fmt"
)

var validWebPath = regexp.MustCompile("^/([a-zA-Z0-9]+)$")
var templates *template.Template

type Page struct {
	Name	string
	Title	string
	Date 	string
	Data	[]map[string]string
}

type Server struct {
	Config map[string]string
	T_DIR string //:= "./tmpl/"
	S_DIR string //:= "./static/"
	R_DIR string //:= "./res/"
}

func (srv *Server)LoadConfig() error {
	s, err := ioutil.ReadFile(srv.R_DIR + "config.json")
	if err != nil {
		log.Println("Error loading configuration: ", err)
		return err
	}
	err = json.Unmarshal(s, &srv.Config)
	if err != nil {
		log.Println("Error parsing config - please check config.json: ", err)
		return err
	}
	_, ok := srv.Config["Port"] 
	if ok != true {
		srv.Config["Port"] = "8080"
	}
	_, ok = srv.Config["Address"] 
	if ok != true {
		srv.Config["Address"] = "127.0.0.1"		
	}
	return nil
} 

func (srv *Server)loadPage(name string) (*Page, error) {
	now := time.Now()
	return &Page{Date: now.Format(time.RFC822)}, nil
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".tmpl", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func makeWebHandler(fn func (http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handling page ", r.URL.Path)
		t := "index"
		
		if r.URL.Path != "/" { // if it is just / then do index
			m := validWebPath.FindStringSubmatch(r.URL.Path)
			if m == nil {
				http.NotFound(w, r)
				return

			} else {
				t = m[1]	
			}
		}
		fn(w, r, t)
	}
}

func (srv *Server)dashHandler(w http.ResponseWriter, r *http.Request, name string) {
	p, err := srv.loadPage(name)
	//log.Println("Loading page ", name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	renderTemplate(w, "index", p)
}

func (srv *Server)StartServer() {
	err := srv.LoadConfig() 
	if err != nil {
		//log.Println(err);
		return
	}
	handlers := http.NewServeMux()
	handlers.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(srv.S_DIR))))
	handlers.Handle("/", makeWebHandler(srv.dashHandler))
	s := &http.Server{
		Addr:          	srv.Config["Address"] + ":" + srv.Config["Port"],
		Handler:        handlers,
	}
	templates = template.Must(template.ParseGlob(srv.T_DIR+"*.tmpl"))
	log.Fatal(s.ListenAndServe())
}

func main() {
	s := new(Server)
	s.T_DIR = "./tmpl/"
	s.S_DIR = "./static/"
	s.R_DIR = "./res/"
	go s.StartServer()
	log.Println("Web routine running...")
	var input string
	fmt.Scanln(&input)
}

