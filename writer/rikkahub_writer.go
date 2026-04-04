package writer

import (
	"archive/zip"
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/converter/backup-converter/models"
)

// WriteRikkaHubBackup writes a RikkaHub backup to a zip file
func WriteRikkaHubBackup(backup *models.RikkaHubBackup, outputPath string) error {
	data, err := MarshalRikkaHubBackup(backup)
	if err != nil {
		return err
	}
	return os.WriteFile(outputPath, data, 0644)
}

// MarshalRikkaHubBackup serializes a RikkaHub backup to zip bytes
func MarshalRikkaHubBackup(backup *models.RikkaHubBackup) ([]byte, error) {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	// Write settings.json
	if backup.Settings != nil {
		settingsData, err := json.MarshalIndent(backup.Settings, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal settings: %w", err)
		}
		if err := writeZipEntry(w, "settings.json", settingsData); err != nil {
			return nil, fmt.Errorf("failed to write settings.json: %w", err)
		}
	}

	// Create SQLite database with conversations and messages
	dbData, err := createRikkaHubDatabase(backup)
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	if err := writeZipEntry(w, "rikka_hub.db", dbData); err != nil {
		return nil, fmt.Errorf("failed to write database: %w", err)
	}

	// Write upload files
	for filename, content := range backup.UploadFiles {
		if err := writeZipEntry(w, "upload/"+filename, content); err != nil {
			return nil, fmt.Errorf("failed to write upload file %s: %w", filename, err)
		}
	}

	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("failed to close zip: %w", err)
	}

	return buf.Bytes(), nil
}

func createRikkaHubDatabase(backup *models.RikkaHubBackup) ([]byte, error) {
	// Create temp directory and database
	tmpDir, err := os.MkdirTemp("", "rikkahub_write_*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "rikka_hub.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	// Create tables
	if err := createRikkaHubTables(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	// Insert conversations
	if err := insertConversations(db, backup.Conversations); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to insert conversations: %w", err)
	}

	// Insert message nodes
	if err := insertMessageNodes(db, backup.MessageNodes); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to insert message nodes: %w", err)
	}

	// Insert managed files
	if err := insertManagedFiles(db, backup.ManagedFiles); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to insert managed files: %w", err)
	}

	db.Close()

	// Read the database file
	return os.ReadFile(dbPath)
}

func createRikkaHubTables(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS "ConversationEntity" (
			id TEXT NOT NULL,
			assistant_id TEXT NOT NULL DEFAULT '0950e2dc-9bd5-4801-afa3-aa887aa36b4e',
			title TEXT NOT NULL,
			nodes TEXT NOT NULL,
			create_at INTEGER NOT NULL,
			update_at INTEGER NOT NULL,
			suggestions TEXT NOT NULL DEFAULT '[]',
			is_pinned INTEGER NOT NULL DEFAULT 0,
			PRIMARY KEY(id)
		)`,
		`CREATE TABLE IF NOT EXISTS message_node (
			id TEXT NOT NULL PRIMARY KEY,
			conversation_id TEXT NOT NULL,
			node_index INTEGER NOT NULL,
			messages TEXT NOT NULL,
			select_index INTEGER NOT NULL,
			FOREIGN KEY (conversation_id) REFERENCES ConversationEntity(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS index_message_node_conversation_id ON message_node(conversation_id)`,
		`CREATE TABLE IF NOT EXISTS managed_files (
			id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
			folder TEXT NOT NULL,
			relative_path TEXT NOT NULL,
			display_name TEXT NOT NULL,
			mime_type TEXT NOT NULL,
			size_bytes INTEGER NOT NULL,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS index_managed_files_relative_path ON managed_files(relative_path)`,
		`CREATE INDEX IF NOT EXISTS index_managed_files_folder ON managed_files(folder)`,
		`CREATE TABLE IF NOT EXISTS MemoryEntity (
			id TEXT NOT NULL PRIMARY KEY,
			conversation_id TEXT,
			content TEXT NOT NULL,
			created_at INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS GenMediaEntity (
			id TEXT NOT NULL PRIMARY KEY,
			conversation_id TEXT NOT NULL,
			message_id TEXT NOT NULL,
			type TEXT NOT NULL,
			data BLOB,
			created_at INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS favorites (
			id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
			type TEXT NOT NULL,
			content TEXT NOT NULL,
			created_at INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS room_master_table (
			id INTEGER PRIMARY KEY,
			identity_hash TEXT
		)`,
		`INSERT OR IGNORE INTO room_master_table (id, identity_hash) VALUES (42, 'rikkahub_converted')`,
		`CREATE TABLE IF NOT EXISTS android_metadata (locale TEXT)`,
		`INSERT OR IGNORE INTO android_metadata VALUES ('zh-CN')`,
	}

	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return fmt.Errorf("failed to execute: %s: %w", q[:min(len(q), 50)], err)
		}
	}

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func insertConversations(db *sql.DB, conversations []models.RikkaHubConversation) error {
	stmt, err := db.Prepare(`
		INSERT OR REPLACE INTO ConversationEntity 
		(id, assistant_id, title, nodes, create_at, update_at, suggestions, is_pinned)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, c := range conversations {
		nodesJSON, err := json.Marshal(c.Nodes)
		if err != nil {
			return err
		}

		suggestionsJSON, err := json.Marshal(c.Suggestions)
		if err != nil {
			return err
		}

		isPinnedInt := 0
		if c.IsPinned {
			isPinnedInt = 1
		}

		createAt := c.CreateAt
		if createAt == 0 {
			createAt = time.Now().UnixMilli()
		}
		updateAt := c.UpdateAt
		if updateAt == 0 {
			updateAt = time.Now().UnixMilli()
		}

		assistantID := c.AssistantID
		if assistantID == "" {
			assistantID = "0950e2dc-9bd5-4801-afa3-aa887aa36b4e"
		}

		_, err = stmt.Exec(
			c.ID, assistantID, c.Title, string(nodesJSON),
			createAt, updateAt, string(suggestionsJSON), isPinnedInt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert conversation %s: %w", c.ID, err)
		}
	}

	return nil
}

func insertMessageNodes(db *sql.DB, nodes []models.RikkaHubMessageNode) error {
	stmt, err := db.Prepare(`
		INSERT OR REPLACE INTO message_node
		(id, conversation_id, node_index, messages, select_index)
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, n := range nodes {
		messagesJSON, err := json.Marshal(n.Messages)
		if err != nil {
			return err
		}

		_, err = stmt.Exec(
			n.ID, n.ConversationID, n.NodeIndex, string(messagesJSON), n.SelectIndex,
		)
		if err != nil {
			return fmt.Errorf("failed to insert message node %s: %w", n.ID, err)
		}
	}

	return nil
}

func insertManagedFiles(db *sql.DB, files []models.RikkaHubManagedFile) error {
	if len(files) == 0 {
		return nil
	}

	stmt, err := db.Prepare(`
		INSERT OR REPLACE INTO managed_files
		(folder, relative_path, display_name, mime_type, size_bytes, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, f := range files {
		_, err = stmt.Exec(
			f.Folder, f.RelativePath, f.DisplayName, f.MimeType,
			f.SizeBytes, f.CreatedAt, f.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert managed file %s: %w", f.RelativePath, err)
		}
	}

	return nil
}
