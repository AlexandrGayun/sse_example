package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"log"
	"net/http"
	"sse_example/config"
	"sse_example/lib/openai"
	"time"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("can't load .env file")
	}
	var cfg config.Config
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatal("can't read OS env")
	}
	cm := openai.PipeConnectionManager{
		InputCh:     make(chan string, 20),
		OutputCh:    make(chan string, 20),
		InputClosed: make(chan struct{}),
	}
	r := mux.NewRouter()
	r.HandleFunc("/chat", attachInputPipe(&cm))
	r.HandleFunc("/healthcheck", okHandler)
	r.HandleFunc("/sse", sseHandler(&cm))
	httpClient := http.Client{Timeout: 10 * time.Second}
	go openai.Process(&cm, httpClient, cfg)
	http.ListenAndServe(fmt.Sprintf(":%s", cfg.HttpApiPort), r)
}

// attachInputPipe upgrades connection to websocket and continuously read values to given channel
func attachInputPipe(cm *openai.PipeConnectionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-cm.InputClosed:
			cm.InputCh = make(chan string, 20)
		default:
		}
		defer func() {
			close(cm.InputCh)
			cm.InputClosed <- struct{}{}
		}()
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("can't upgrade connection to websocket: %v", err)
			return
		}
		for {
			messageType, p, err := conn.ReadMessage()
			if err != nil {
				log.Printf("can't read message: %v", err)
				return
			}
			if messageType == websocket.CloseMessage {
				return
			}
			if messageType != websocket.TextMessage {
				log.Println("non-text message type, skip")
				continue
			}
			if len(cm.InputCh) == cap(cm.InputCh) {
				err := conn.WriteMessage(websocket.TextMessage, []byte("processing queue is full, try again later"))
				if err != nil {
					log.Printf("can't write message: %v", err)
					return
				}
			} else {
				cm.InputCh <- string(p)
			}
		}
	}
}

func sseHandler(cm *openai.PipeConnectionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "SSE not supported", http.StatusInternalServerError)
			log.Println("sse flusher not supported")
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Type")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Content-Type", "text/event-stream")
		for {
			select {
			case v, ok := <-cm.OutputCh:
				if !ok {
					return
				}
				_, err := fmt.Fprint(w, fmt.Sprintf("%v\n\n", v))
				if err != nil {
					log.Printf("can't write data to response writer: %v", err)
				}
				flusher.Flush()
			case <-r.Context().Done():
				return
			}
		}
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func okHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("hello"))
	if err != nil {
		log.Println(err)
	}
}
