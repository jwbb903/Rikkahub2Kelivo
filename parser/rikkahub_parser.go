package parser

import (
	"archive/zip"
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"

	"github.com/converter/backup-converter/models"
)

// ParseRikkaHubBackup parses a RikkaHub backup zip file
func ParseRikkaHubBackup(zipPath string) (*models.RikkaHubBackup, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open zip: %w", err)
	}
	defer r.Close()

	return parseRikkaHubZip(r.File)
}

// ParseRikkaHubBackupFromBytes parses a RikkaHub backup from bytes
func ParseRikkaHubBackupFromBytes(data []byte) (*models.RikkaHubBackup, error) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("failed to open zip: %w", err)
	}
	return parseRikkaHubZip(r.File)
}

func parseRikkaHubZip(files []*zip.File) (*models.RikkaHubBackup, error) {
	backup := &models.RikkaHubBackup{
		UploadFiles: make(map[string][]byte),
	}

	var settingsData []byte
	var dbData []byte
	var walData []byte

	for _, f := range files {
		data, err := readZipFile(f)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", f.Name, err)
		}

		switch f.Name {
		case "settings.json":
			settingsData = data
		case "rikka_hub.db":
			dbData = data
		case "rikka_hub-wal":
			walData = data
		case "rikka_hub-shm":
			// shm file not needed for reading
		default:
			// Store upload files
			if strings.HasPrefix(f.Name, "upload/") {
				fname := filepath.Base(f.Name)
				backup.UploadFiles[fname] = data
			}
		}
	}

	if settingsData != nil {
		settings, err := parseRikkaHubSettings(settingsData)
		if err != nil {
			return nil, fmt.Errorf("failed to parse settings: %w", err)
		}
		backup.Settings = settings
	}

	if dbData != nil {
		// Write db to temp file and query it
		conversations, messageNodes, managedFiles, err := parseRikkaHubDatabase(dbData, walData)
		if err != nil {
			return nil, fmt.Errorf("failed to parse database: %w", err)
		}
		backup.Conversations = conversations
		backup.MessageNodes = messageNodes
		backup.ManagedFiles = managedFiles
	}

	return backup, nil
}

func parseRikkaHubSettings(data []byte) (*models.RikkaHubSettings, error) {
	settings := &models.RikkaHubSettings{}
	if err := json.Unmarshal(data, settings); err != nil {
		return nil, err
	}
	return settings, nil
}

func parseRikkaHubDatabase(dbData []byte, walData []byte) (
	[]models.RikkaHubConversation,
	[]models.RikkaHubMessageNode,
	[]models.RikkaHubManagedFile,
	error,
) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "rikkahub_db_*")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "rikka_hub.db")
	if err := os.WriteFile(dbPath, dbData, 0644); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to write db: %w", err)
	}

	// Write WAL file if present (for consistency)
	if len(walData) > 0 {
		walPath := filepath.Join(tmpDir, "rikka_hub-wal")
		if err := os.WriteFile(walPath, walData, 0644); err != nil {
			return nil, nil, nil, fmt.Errorf("failed to write wal: %w", err)
		}
	}

	db, err := sql.Open("sqlite3", dbPath+"?mode=ro")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to open sqlite: %w", err)
	}
	defer db.Close()

	// Checkpoint WAL if present
	if len(walData) > 0 {
		_, _ = db.Exec("PRAGMA wal_checkpoint(FULL)")
	}

	conversations, err := queryConversations(db)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to query conversations: %w", err)
	}

	messageNodes, err := queryMessageNodes(db)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to query message nodes: %w", err)
	}

	managedFiles, err := queryManagedFiles(db)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to query managed files: %w", err)
	}

	return conversations, messageNodes, managedFiles, nil
}

func queryConversations(db *sql.DB) ([]models.RikkaHubConversation, error) {
	rows, err := db.Query(`
		SELECT id, assistant_id, title, nodes, create_at, update_at, suggestions, is_pinned
		FROM ConversationEntity
		ORDER BY create_at
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conversations []models.RikkaHubConversation
	for rows.Next() {
		var c models.RikkaHubConversation
		var nodesJSON string
		var suggestionsJSON string
		var isPinnedInt int

		err := rows.Scan(
			&c.ID, &c.AssistantID, &c.Title, &nodesJSON,
			&c.CreateAt, &c.UpdateAt, &suggestionsJSON, &isPinnedInt,
		)
		if err != nil {
			return nil, err
		}

		c.IsPinned = isPinnedInt != 0

		if err := json.Unmarshal([]byte(nodesJSON), &c.Nodes); err != nil {
			c.Nodes = []string{}
		}

		if err := json.Unmarshal([]byte(suggestionsJSON), &c.Suggestions); err != nil {
			c.Suggestions = []string{}
		}

		conversations = append(conversations, c)
	}
	return conversations, rows.Err()
}

func queryMessageNodes(db *sql.DB) ([]models.RikkaHubMessageNode, error) {
	rows, err := db.Query(`
		SELECT id, conversation_id, node_index, messages, select_index
		FROM message_node
		ORDER BY conversation_id, node_index
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []models.RikkaHubMessageNode
	for rows.Next() {
		var n models.RikkaHubMessageNode
		var messagesJSON string

		err := rows.Scan(
			&n.ID, &n.ConversationID, &n.NodeIndex, &messagesJSON, &n.SelectIndex,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(messagesJSON), &n.Messages); err != nil {
			n.Messages = []models.RikkaHubMessage{}
		}

		nodes = append(nodes, n)
	}
	return nodes, rows.Err()
}

func queryManagedFiles(db *sql.DB) ([]models.RikkaHubManagedFile, error) {
	rows, err := db.Query(`
		SELECT id, folder, relative_path, display_name, mime_type, size_bytes, created_at, updated_at
		FROM managed_files
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []models.RikkaHubManagedFile
	for rows.Next() {
		var f models.RikkaHubManagedFile
		err := rows.Scan(
			&f.ID, &f.Folder, &f.RelativePath, &f.DisplayName,
			&f.MimeType, &f.SizeBytes, &f.CreatedAt, &f.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	return files, rows.Err()
}

// writeZipFile helper for creating zip entries
func writeToZip(w *zip.Writer, name string, data []byte) error {
	f, err := w.Create(name)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, bytes.NewReader(data))
	return err
}
