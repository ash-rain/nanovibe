package routes

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"vibecodepc/server/db"
	"vibecodepc/server/httputil"
	"vibecodepc/server/services"
)

// RegisterAgentRoutes registers all agent API routes.
func RegisterAgentRoutes(r chi.Router) {
	r.Get("/api/agent/stream", getAgentStream)
	r.Post("/api/agent/message", postAgentMessage)
	r.Get("/api/agent/messages", getAgentMessages)
	r.Delete("/api/agent/messages", deleteAgentMessages)
}

// agentMessage mirrors the agent_messages table row.
type agentMessage struct {
	ID        int64   `json:"id"`
	Role      string  `json:"role"`
	Content   string  `json:"content"`
	CreatedAt int64   `json:"createdAt"`
	ProjectID *string `json:"projectId"`
}

func getAgentStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	// Get the last message ID so we only stream new messages.
	var lastID int64
	_ = db.DB().QueryRowContext(r.Context(), `SELECT COALESCE(MAX(id), 0) FROM agent_messages`).Scan(&lastID)

	ctx := r.Context()
	responseCh := services.WatchForResponses(ctx, lastID)

	// Ping ticker every 15 seconds.
	ping := time.NewTicker(15 * time.Second)
	defer ping.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-ping.C:
			fmt.Fprintf(w, "event: ping\ndata: {}\n\n")
			flusher.Flush()

		case msg, open := <-responseCh:
			if !open {
				return
			}
			data, err := json.Marshal(agentMessage{
				ID:        msg.ID,
				Role:      msg.Role,
				Content:   msg.Content,
				CreatedAt: msg.CreatedAt,
				ProjectID: nullableStr(msg.ProjectID),
			})
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "event: message\ndata: %s\n\n", data)
			flusher.Flush()
		}
	}
}

func postAgentMessage(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Content   string  `json:"content"`
		ProjectID *string `json:"projectId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Content == "" {
		httputil.WriteError(w, http.StatusBadRequest, "content is required")
		return
	}

	projectID := ""
	if body.ProjectID != nil {
		projectID = *body.ProjectID
	}

	// Store the user message in our own DB.
	now := time.Now().Unix()
	_, err := db.DB().ExecContext(r.Context(),
		`INSERT INTO agent_messages (role, content, created_at, project_id) VALUES (?, ?, ?, ?)`,
		"user", body.Content, now, nullableSQLStr(projectID),
	)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to store message")
		return
	}

	// Forward the message to nanoclaw.
	if err := services.InsertUserMessage(body.Content, projectID); err != nil {
		// Non-fatal: nanoclaw may not be running. Log it but continue.
		_ = err
	}

	w.WriteHeader(http.StatusNoContent)
}

func getAgentMessages(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB().QueryContext(r.Context(),
		`SELECT id, role, content, created_at, project_id FROM agent_messages ORDER BY id DESC LIMIT 50`,
	)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list messages")
		return
	}
	defer rows.Close()

	var messages []agentMessage
	for rows.Next() {
		var m agentMessage
		var projectID sql.NullString
		if err := rows.Scan(&m.ID, &m.Role, &m.Content, &m.CreatedAt, &projectID); err != nil {
			continue
		}
		if projectID.Valid {
			m.ProjectID = &projectID.String
		}
		messages = append(messages, m)
	}
	if messages == nil {
		messages = []agentMessage{}
	}

	// Return in chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{"data": messages})
}

func deleteAgentMessages(w http.ResponseWriter, r *http.Request) {
	_, err := db.DB().ExecContext(r.Context(), `DELETE FROM agent_messages`)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to delete messages")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func nullableStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func nullableSQLStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

