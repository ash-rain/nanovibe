package routes

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"

	"vibecodepc/server/db"
	"vibecodepc/server/services"
)

var wsUpgrader = websocket.Upgrader{
	CheckOrigin:     func(r *http.Request) bool { return true },
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
}

// RegisterTerminalRoutes registers the WebSocket terminal route.
func RegisterTerminalRoutes(r chi.Router) {
	r.Get("/ws/terminal/{projectId}", handleTerminal)
}

// wsMessage is the JSON structure exchanged over the WebSocket.
type wsMessage struct {
	Type string `json:"type"`
	Data string `json:"data,omitempty"`
	Cols uint16 `json:"cols,omitempty"`
	Rows uint16 `json:"rows,omitempty"`
	Code int    `json:"code,omitempty"`
}

func handleTerminal(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectId")

	// Look up the project path from the database.
	var projectPath string
	err := db.DB().QueryRowContext(r.Context(), `SELECT path FROM projects WHERE id = ?`, projectID).Scan(&projectPath)
	if err != nil {
		http.Error(w, "project not found", http.StatusNotFound)
		return
	}

	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("terminal: websocket upgrade: %v", err)
		return
	}
	defer conn.Close()

	// Get or start an opencode PTY session.
	ptmx, err := services.StartSession(projectID, projectPath)
	if err != nil {
		sendWSError(conn, "failed to start terminal: "+err.Error())
		return
	}

	// ptyDone is closed when the PTY reader goroutine exits.
	ptyDone := make(chan struct{})

	// PTY → WebSocket: continuously read PTY output and forward to the browser.
	go func() {
		defer close(ptyDone)
		buf := make([]byte, 4096)
		for {
			n, readErr := ptmx.Read(buf)
			if n > 0 {
				encoded := base64.StdEncoding.EncodeToString(buf[:n])
				msg := wsMessage{Type: "output", Data: encoded}
				data, _ := json.Marshal(msg)
				if writeErr := conn.WriteMessage(websocket.TextMessage, data); writeErr != nil {
					return
				}
			}
			if readErr != nil {
				if readErr == io.EOF {
					sendWSExit(conn, 0)
				} else {
					sendWSExit(conn, 1)
				}
				return
			}
		}
	}()

	// WebSocket → PTY: read client messages and act on them.
	// conn.ReadMessage blocks until a message arrives or the connection closes.
	for {
		_, raw, readErr := conn.ReadMessage()
		if readErr != nil {
			// Connection closed by client or network error — stop the PTY.
			_ = services.KillSession(projectID)
			return
		}

		var msg wsMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			continue
		}

		switch msg.Type {
		case "input":
			_, _ = ptmx.Write([]byte(msg.Data))
		case "resize":
			if msg.Cols > 0 && msg.Rows > 0 {
				_ = services.ResizeSession(projectID, msg.Cols, msg.Rows)
			}
		}

		// Check if PTY has exited.
		select {
		case <-ptyDone:
			return
		default:
		}
	}
}

func sendWSError(conn *websocket.Conn, msg string) {
	payload := "\r\nerror: " + msg + "\r\n"
	data, _ := json.Marshal(wsMessage{
		Type: "output",
		Data: base64.StdEncoding.EncodeToString([]byte(payload)),
	})
	_ = conn.WriteMessage(websocket.TextMessage, data)
}

func sendWSExit(conn *websocket.Conn, code int) {
	data, _ := json.Marshal(wsMessage{Type: "exit", Code: code})
	_ = conn.WriteMessage(websocket.TextMessage, data)
}
