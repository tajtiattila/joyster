package main

import (
	"code.google.com/p/go.net/websocket"
	"errors"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const webSocketPath = "/ws"

func WebGUI(addr string, sharedir string) {
	http.Handle("/lib/", http.StripPrefix("/lib/", http.FileServer(http.Dir(sharedir+"/lib"))))
	http.Handle(webSocketPath, websocket.Handler(handleWebSocket))
	http.Handle("/", templateReloader(sharedir+"/tmpl", "main", nil))
	go func() {
		err := http.ListenAndServe(addr, nil)
		log.Println("listenAndServe:", err)
	}()
}

func handleWebSocket(ws *websocket.Conn) {
	l := DefaultStatusDispatcher.NewListener()
	go io.Copy(ioutil.Discard, ws)
	for {
		data, err := l.Read()
		if err != nil {
			log.Println("handleWebSocket:", err)
			return
		}
		if err = websocket.JSON.Send(ws, data); err != nil {
			log.Println("handleWebSocket:", err)
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func templateReloader(dir, name string, config interface{}) http.Handler {
	tmpl, err := readTemplate(dir, name)
	if err != nil {
		log.Fatalln("templateReloader init:", err)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if tmplnew, err := readTemplate(dir, name); err == nil {
			tmpl = tmplnew
		} else {
			log.Printf("template reload error:", err)
		}
		config := &webConfig{
			WebSocketUrl: "ws://" + req.Host + webSocketPath,
		}
		tmpl.Execute(w, config)
	})
}

func readTemplate(dir, name string) (tmpl *template.Template, err error) {
	tmpl, err = template.ParseGlob(dir + "/*.tmpl")
	if err == nil {
		if tmpl = tmpl.Lookup(name); tmpl == nil {
			err = errors.New("missing template \"" + name + "\"")
		}
	}
	return
}

type webConfig struct {
	WebSocketUrl string
}
