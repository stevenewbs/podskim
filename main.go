package main

import (
	pr "github.com/stevenewbs/picorss"
	"encoding/json"
	"html/template"
	"io/ioutil"
	"net/http"
	"regexp"
	"log"
	"fmt"
	"strings"
	"os"
	"errors"
	"strconv"
)
var validWebPath = regexp.MustCompile("^/([a-zA-Z0-9]+)$")
var templates *template.Template

type Page struct {
	Casts	[]Cast
	Feed pr.Rss
}
type JSONResponse struct {
	Response string
	Message	string
}
type Server struct {
	S *http.Server
	Config map[string]string
	Casts map[string][]Cast
	MAIN string //:= "~/.podskim"
	T_DIR string //:= MAIN + "/tmpl/"
	S_DIR string //:= MAIN + "/static/"
	R_DIR string //:= MAIN + "/res/"
}
type Cast struct {
	Num	string
	Name	string
	Link	string
}

func DeleteCast(c []Cast, n string) []Cast {
	// helper function to remove casts from the main cast array
	r := []Cast{}
	s := len(c)
	for i := 0; i <= s-1; i++ {
		if c[i].Name == n {
			continue
		} else {
			r = append(r, c[i])
		}
	}
	return r
}

func FindCast(c []Cast, n string) (Cast, error) {
	// helper function to search through the array
	r := Cast{}
	s := len(c)
	for i := 0; i <= s-1; i++ {
		if c[i].Name == n {
			return c[i], nil
		}
	}
	return r, errors.New("Not found in this array")
}

func (srv *Server)CreateConfig() {
	err := os.MkdirAll(srv.R_DIR, 0755)
	if err != nil {
		log.Println("Cannot create config dir:", err)
	}
	srv.Config = make(map[string]string)
	srv.Config["Port"] = "8080"
	srv.Config["Address"] = "127.0.0.1"
	conf, _ := json.Marshal(srv.Config)
	log.Println("Writing to: ", srv.R_DIR + "config.json")
	err = ioutil.WriteFile(srv.R_DIR + "config.json", conf, 0644)
        if err != nil {
                log.Println(err)
        }
	log.Println("Config dir has been created")
}

func (srv *Server)LoadConfig() error {
	if _, err := os.Stat(srv.R_DIR + "config.json"); os.IsNotExist(err)  {
		log.Println("Creating config")
		srv.CreateConfig()
	}
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

func (srv *Server)LoadCasts() error {
	c := map[string][]Cast{}
	s, err := ioutil.ReadFile(srv.R_DIR + "urls.json")
	if os.IsNotExist(err) {
		srv.Casts = c
		return nil
	}
	if err != nil {
		log.Println("Error loading urls: ", err)
		return err
	}
	err = json.Unmarshal(s, &c)
	if err != nil {
		log.Println("Error parsing urls - please check urls.json: ", err)
		return err
	}
	srv.Casts = c
	return nil
}
func (srv *Server)WriteBackCasts() error {
	jsonobj, err := json.Marshal(srv.Casts)
	//log.Println(jsonobj)
	if err != nil {
		log.Println(err)
		return err
	}
	err = ioutil.WriteFile(srv.R_DIR + "urls.json", jsonobj, 0644|os.ModeAppend)
	if err != nil {
		log.Println(err)
		return err
	}
	return err
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".tmpl", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
func renderJson(w http.ResponseWriter, j JSONResponse) {
	jm, err := json.Marshal(j)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jm)
}

func makeWebHandler(fn func (http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//log.Println("Handling page ", r.URL.Path)
		if r.URL.Path != "/" { // if it is just / then do index
			m := validWebPath.FindStringSubmatch(r.URL.Path)
			if m == nil {
				http.NotFound(w, r)
				return
			}
		}
		fn(w, r)
	}
}

func (srv *Server)DashHandler(w http.ResponseWriter, r *http.Request) {
	p := &Page{Casts: srv.Casts["casts"]}
	renderTemplate(w, "index", p)
}

func (srv *Server)AddHandler(w http.ResponseWriter, r *http.Request) {
	jo := JSONResponse{}
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		jo.Response = "Error"
		jo.Message = fmt.Sprint(err)
	} else {
		url := r.Form["newurl"][0] // parse form assumes multi value for each - we only need the first one
		name := r.Form["name"][0]
		//log.Println(url)
		if strings.HasPrefix(url, "http") {
			num := strconv.Itoa(len(srv.Casts["casts"]) + 1)
			c := Cast {num, name, url}
			srv.Casts["casts"] = append(srv.Casts["casts"], c)
			err = srv.WriteBackCasts()
			if err != nil {
				log.Println(err)
				jo.Response = "Error"
				jo.Message = fmt.Sprint(err)
			}
		} else {
			log.Println("invalid url")
			jo.Response = "Error"
			jo.Message = "Invalid URL"
		}
		if jo.Response != "Error" {
			jo.Response = "Success"
			jo.Message = fmt.Sprint("Added " + url)
		}
		//log.Println(srv.Casts)
	}
	renderJson(w, jo)
}
func (srv *Server)DeleteHandler(w http.ResponseWriter, r *http.Request) {
	jo := JSONResponse{}
	jo.Response = "Error"
	err := r.ParseForm()
	if  err != nil {
		log.Println(err)
		jo.Message = fmt.Sprint(err)
	} else {
		name := r.Form["name"][0] // parse form assumes multi value for each - we only need the first one
		srv.Casts["casts"] = DeleteCast(srv.Casts["casts"], name)
		err = srv.WriteBackCasts()
		if err != nil {
			log.Println(err)
			jo.Message = fmt.Sprint(err)
		} else {
		  jo.Response = "Success"
		  jo.Message = fmt.Sprint("Removed " + name)
		}
	}
	renderJson(w, jo)
}
func (srv *Server)FeedHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	p := &Page{}
	if err != nil {
		http.NotFound(w, r)
		return
	} else {
		n := r.Form["name"][0]
		a, perr := strconv.Atoi(r.Form["amount"][0])
		if perr != nil {
			a = 1
		}
		log.Println("Going to get ", n, " - top ", a)
		c, err := FindCast(srv.Casts["casts"], n)
		if err != nil {
			http.NotFound(w, r)
		}
		response, herr := http.Get(c.Link)
		if herr != nil {
			log.Println(herr)
		}
		defer response.Body.Close()
		//log.Println(response)
		rss, err := pr.ResponseToRss(response)
		if err != nil {
			log.Println("Error parsing feed: ", err)
		} else {
			//log.Println(" TITLE:", rss.Channel.Items[0].Enclosure)

			if len(rss.Channel.Items) > 0 {
				if a > 0 && len(rss.Channel.Items) >= a { // only retrieve the amount we asked for
					rss.Channel.Items = rss.Channel.Items[:a]
				}
			}
			p.Feed = rss
		}
	}
	renderTemplate(w, "feed", p)
}

func (srv *Server)StartServer() {
	err := srv.LoadConfig()
	if err != nil {
		//log.Println(err);
		return
	}
	err = srv.LoadCasts()
	if err != nil {
		log.Println(err)
		return
	}
	cont := true
	cont, err = exists(srv.S_DIR)
	if err != nil {
		log.Println("Error with Static Directory:", err)
	}
	if cont == false {log.Println("Static dir not found")}
	cont, err = exists(srv.T_DIR)
	if err != nil {
		log.Println("Error with Template Directory:", err)
	}
	if cont == false { log.Println("Template dir not found") }
	if cont {
		handlers := http.NewServeMux()
		handlers.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(srv.S_DIR))))
		handlers.Handle("/add", makeWebHandler(srv.AddHandler))
		handlers.Handle("/delete", makeWebHandler(srv.DeleteHandler))
		handlers.Handle("/feed", makeWebHandler(srv.FeedHandler))
		handlers.Handle("/", makeWebHandler(srv.DashHandler))
		srv.S = &http.Server{
			Addr:		srv.Config["Address"] + ":" + srv.Config["Port"],
			Handler:        handlers,
		}
		templates = template.Must(template.ParseGlob(srv.T_DIR+"*.tmpl"))
		log.Printf("Launch http://%v:%v in browser", srv.Config["Address"], srv.Config["Port"])
		log.Fatal(srv.S.ListenAndServe())
	}
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func main() {
	s := new(Server)
	s.MAIN = os.Getenv("HOME") + "/.podskim/"
	s.T_DIR = s.MAIN + "tmpl/"
	s.S_DIR = s.MAIN + "static/"
	s.R_DIR = s.MAIN + "res/"
  log.Println("Web routine running...")
	s.StartServer()
}
