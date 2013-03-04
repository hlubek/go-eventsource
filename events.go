package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

var requestCounter = 0

func main() {
	http.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		requestCounter++
		requestId := requestCounter

		if r.Method == "OPTIONS" {
			h := w.Header()
			h.Set("Access-Control-Allow-Origin", "*")
			h.Set("Access-Control-Allow-Methods", "GET")
			h.Set("Access-Control-Allow-Headers", "Last-Event-ID, Cache-Control")
			h.Set("Access-Control-Max-Age", "86400")
		} else if r.Method == "GET" {
			f, ok := w.(http.Flusher)
			if !ok {
				http.Error(w, "streaming unsupported", http.StatusInternalServerError)
				return
			}
			c, ok := w.(http.CloseNotifier)
			if !ok {
				http.Error(w, "close notification unsupported", http.StatusInternalServerError)
				return
			}

			log.Printf("%d: Responding to GET", requestId)

			h := w.Header()

			h.Set("Content-Type", "text/event-stream")
			h.Set("Cache-Control", "no-cache")
			h.Set("Connection", "keep-alive")
			h.Set("Access-Control-Allow-Origin", "*")

			w.WriteHeader(http.StatusOK)

			padding := ":" + strings.Repeat(" ", 2049) + "\n"
			w.Write([]byte(padding))
			fmt.Fprint(w, "retry: 2000\n")
			fmt.Fprintf(w, "data: %d\n\n", time.Now().Unix())
			f.Flush()

			log.Printf("%d: Ticking...", requestId)

			t := time.NewTicker(1 * time.Second)
			defer t.Stop()

			closer := c.CloseNotify()
			for {
				select {
				case i := <-t.C:
					log.Printf("%d: Tick.", requestId)
					fmt.Fprintf(w, "data: %d\n\n", i.Unix())
					f.Flush()
				case <-closer:
					log.Printf("%d: Closing connection...", requestId)
					return
				}
			}
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}

	})

	http.Handle("/", http.FileServer(http.Dir("static")))

	s := &http.Server{
		Addr: ":8080",
	}
	log.Fatal(s.ListenAndServe())
}
