package controllers

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"

	"strings"

	"github.com/garyburd/redigo/redis"
	"github.com/julienschmidt/httprouter"
	"github.com/watain666/ptt-alertor/connections"
	"github.com/watain666/ptt-alertor/models/counter"
	"github.com/watain666/ptt-alertor/models/top"
	"github.com/watain666/ptt-alertor/shorturl"
	"golang.org/x/net/websocket"

	"github.com/russross/blackfriday"
)

var tpls = []string{
	"public/docs.html",
	"public/top.html",
	"public/telegram.html",
	"public/messenger.html",
	"public/line.html",
	"public/tpls/head.tpl",
	"public/tpls/header.tpl",
	"public/tpls/slogan.tpl",
	"public/tpls/command.tpl",
	"public/tpls/counter.tpl",
	"public/tpls/footer.tpl",
	"public/tpls/script.tpl",
}

func markDowner(args ...interface{}) template.HTML {
	return template.HTML(blackfriday.MarkdownCommon([]byte(fmt.Sprintf("%s", args...))))
	// return template.HTML(strings.ToLower(fmt.Sprintf("%s", args...)))
}

var (
	templates = template.Must(template.ParseFiles(tpls...))
	wsHost    = os.Getenv("APP_WS_HOST")
	s3Domain  = os.Getenv("S3_DOMAIN")
)

// Index Handles router "/" request
func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	LineIndex(w, r, nil)
}

// LineIndex Handles router "/line" request
func LineIndex(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	err := templates.ExecuteTemplate(w, "line.html", struct {
		URI      string
		WSHost   string
		Count    []string
		S3Domain string
	}{"line", wsHost, count(), s3Domain})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// MessengerIndex Handles router "/messenger" request
func MessengerIndex(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	err := templates.ExecuteTemplate(w, "messenger.html", struct {
		URI      string
		WSHost   string
		Count    []string
		S3Domain string
	}{"messenger", wsHost, count(), s3Domain})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// TelegramIndex Handles router "/telegram" request
func TelegramIndex(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	err := templates.ExecuteTemplate(w, "telegram.html", struct {
		URI      string
		WSHost   string
		Count    []string
		S3Domain string
	}{"telegram", wsHost, count(), s3Domain})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func count() (counterStrs []string) {
	count, err := counter.Alert()
	if err != nil {
		return nil
	}
	countStrs := strings.Split((strconv.Itoa(count)), "")
	for index, num := range countStrs {
		counterStrs = append(counterStrs, num)
		if backIndex := len(countStrs) - index; backIndex != 1 && backIndex%3 == 1 {
			counterStrs = append(counterStrs, ",")
		}
	}
	return counterStrs
}

// Top Handles router "/top" request, it shows top rank of keywords, authors, pushsum
func Top(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	count := 100
	keywords := top.ListKeywordWithScore(count)
	authors := top.ListAuthorWithScore(count)
	pushsum := top.ListPushSumWithScore(count)
	data := struct {
		URI      string
		Keywords top.WordOrders
		Authors  top.WordOrders
		PushSum  top.WordOrders
		S3Domain string
	}{
		"top",
		keywords,
		authors,
		pushsum,
		s3Domain,
	}
	err := templates.ExecuteTemplate(w, "top.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Docs shows advanced intructions
func Docs(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	err := templates.ExecuteTemplate(w, "docs.html", struct {
		URI      string
		S3Domain string
	}{
		"docs",
		s3Domain,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Redirect redirects short url to original url
func Redirect(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	checksum := params.ByName("checksum")
	url := shorturl.Original(checksum)
	if url != "" {
		http.Redirect(w, r, url, http.StatusMovedPermanently)
	} else {
		t, err := template.ParseFiles("public/404.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		t.Execute(w, nil)
	}
}

// WebSocket upgrades http request to websocket
func WebSocket(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	websocket.Handler(counterHandler).ServeHTTP(w, r)
}

func counterHandler(ws *websocket.Conn) {
	conn := connections.Redis()
	defer conn.Close()
	psc := redis.PubSubConn{Conn: conn}
	psc.Subscribe("alert-counter")
	defer psc.Unsubscribe("alert-counter")
	for {
		switch v := psc.Receive().(type) {
		case redis.Message:
			_, err := ws.Write([]byte(v.Data))
			if err != nil {
				ws.Close()
				return
			}
		case redis.Subscription:
			// fmt.Printf("%s: %s %d\n", v.Channel, v.Kind, v.Count)
		case error:
			return
		}
	}
}
