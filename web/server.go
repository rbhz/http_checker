package web

import (
	"encoding/json"
	"fmt"
	"github.com/rbhz/http_checker/watcher"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

// Server with rest api & static
type Server struct {
	watcher    watcher.Watcher
	staticPath string
	port       int
	sockets    map[string]*websocket.Conn
}

// Run web server
func (s *Server) Run() {
	fs := http.FileServer(http.Dir(s.staticPath))
	http.Handle("/", fs)
	http.HandleFunc("/api/list", s.list)
	http.HandleFunc("/ws", s.upgrade)
	fmt.Printf("Starting server on http://0.0.0.0:%v\n", s.port)
	err := http.ListenAndServe(fmt.Sprintf(":%v", s.port), nil)
	if err != nil {
		panic(err)
	}
}

func (s *Server) list(w http.ResponseWriter, r *http.Request) {
	data, _ := json.Marshal(s.watcher.GetUrls())
	w.Write(data)
}

func (s *Server) upgrade(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	remoteAddr := c.RemoteAddr().String()
	defer delete(s.sockets, remoteAddr)
	s.sockets[remoteAddr] = c
	for {
		_, _, err := c.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
	}
}

// Broadcast message to all active sockets
func (s *Server) Broadcast(message []byte) {
	for _, conn := range s.sockets {
		go conn.WriteMessage(websocket.TextMessage, message)
	}
}

// GetServer returns web server
func GetServer(w watcher.Watcher, static string, port int) Server {
	return Server{
		watcher:    w,
		staticPath: static,
		port:       port,
		sockets:    make(map[string]*websocket.Conn),
	}
}