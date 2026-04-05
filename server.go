package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
	_ "embed"

	"github.com/converter/backup-converter/converter"
	"github.com/converter/backup-converter/models"
	"github.com/converter/backup-converter/parser"
	"github.com/converter/backup-converter/writer"
)

//go:embed web/index.html
var webUI []byte

//go:embed web/css/style.css
var cssData []byte

//go:embed web/js/api.js
var jsApi []byte

//go:embed web/js/render.js
var jsRender []byte

//go:embed web/js/app.js
var jsApp []byte

type Server struct {
	mu        sync.RWMutex
	backupType string
	kelivo    *models.KelivoBackup
	rikkahub  *models.RikkaHubBackup
	uploadedAt time.Time
}

func StartServer(port string) {
	s := &Server{}
	mux := http.NewServeMux()

	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/css/style.css", s.handleStatic(cssData, "text/css"))
	mux.HandleFunc("/js/api.js", s.handleStatic(jsApi, "application/javascript"))
	mux.HandleFunc("/js/render.js", s.handleStatic(jsRender, "application/javascript"))
	mux.HandleFunc("/js/app.js", s.handleStatic(jsApp, "application/javascript"))
	mux.HandleFunc("/api/upload", s.handleUpload)
	mux.HandleFunc("/api/convert/k2r", s.handleConvertK2R)
	mux.HandleFunc("/api/convert/r2k", s.handleConvertR2K)
	mux.HandleFunc("/api/info", s.handleInfo)
	mux.HandleFunc("/api/providers", s.handleProviders)
	mux.HandleFunc("/api/assistants", s.handleAssistants)
	mux.HandleFunc("/api/mcp", s.handleMCP)
	mux.HandleFunc("/api/conversations", s.handleConversations)
	mux.HandleFunc("/api/lorebooks", s.handleLorebooks)
	mux.HandleFunc("/api/quick-messages", s.handleQuickMessages)
	mux.HandleFunc("/api/injections", s.handleInjections)

	addr := ":" + port
	fmt.Printf("Web 服务器已启动：http://localhost%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(webUI)
}

func (s *Server) handleStatic(data []byte, contentType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", contentType)
		w.Write(data)
	}
}

func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "POST only"})
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 200<<20)
	if err := r.ParseMultipartForm(200 << 20); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "解析表单失败: " + err.Error()})
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "未找到上传文件"})
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "读取文件失败"})
		return
	}

	btype := detectBackupTypeFromBytes(data)
	if btype == "unknown" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "无法识别的备份格式"})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.backupType = btype
	s.uploadedAt = time.Now()

	switch btype {
	case "kelivo":
		backup, err := parser.ParseKelivoBackupFromBytes(data)
		if err != nil {
			s.backupType = ""
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "解析 Kelivo 备份失败: " + err.Error()})
			return
		}
		s.kelivo = backup
		s.rikkahub = nil

		info := map[string]interface{}{
			"type":      "kelivo",
			"filename":  header.Filename,
			"size":      len(data),
			"uploadedAt": s.uploadedAt,
		}
		if backup.Settings != nil {
			settings := backup.Settings
			info["providers"] = len(settings.ProviderConfigs)
			info["assistants"] = len(settings.Assistants)
			info["mcpServers"] = len(settings.MCPServers)
			info["worldBooks"] = len(settings.WorldBooks)
			info["quickPhrases"] = len(settings.QuickPhrases)
			info["injections"] = len(settings.InstructionInjections)
		}
		if backup.Chats != nil {
			info["conversations"] = len(backup.Chats.Conversations)
			info["messages"] = len(backup.Chats.Messages)
		}
		writeJSON(w, http.StatusOK, info)

	case "rikkahub":
		backup, err := parser.ParseRikkaHubBackupFromBytes(data)
		if err != nil {
			s.backupType = ""
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "解析 RikkaHub 备份失败: " + err.Error()})
			return
		}
		s.rikkahub = backup
		s.kelivo = nil

		info := map[string]interface{}{
			"type":      "rikkahub",
			"filename":  header.Filename,
			"size":      len(data),
			"uploadedAt": s.uploadedAt,
		}
		if backup.Settings != nil {
			settings := backup.Settings
			info["providers"] = len(settings.Providers)
			info["assistants"] = len(settings.Assistants)
			info["mcpServers"] = len(settings.MCPServers)
			info["lorebooks"] = len(settings.Lorebooks)
			info["quickMessages"] = len(settings.QuickMessages)
			info["injections"] = len(settings.ModeInjections)
		}
		info["conversations"] = len(backup.Conversations)
		info["messageNodes"] = len(backup.MessageNodes)
		info["uploadFiles"] = len(backup.UploadFiles)
		info["managedFiles"] = len(backup.ManagedFiles)
		writeJSON(w, http.StatusOK, info)
	}
}

func (s *Server) handleConvertK2R(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "POST only"})
		return
	}

	s.mu.RLock()
	if s.kelivo == nil {
		s.mu.RUnlock()
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "没有已上传的 Kelivo 备份"})
		return
	}
	src := s.kelivo
	s.mu.RUnlock()

	result, err := converter.ConvertKelivoToRikkaHub(src)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "转换失败: " + err.Error()})
		return
	}

	data, err := writer.MarshalRikkaHubBackup(result)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "序列化失败: " + err.Error()})
		return
	}

	timestamp := time.Now().Format("20060102_150405")
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="rikkahub_converted_%s.zip"`, timestamp))
	w.Write(data)
}

func (s *Server) handleConvertR2K(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "POST only"})
		return
	}

	s.mu.RLock()
	if s.rikkahub == nil {
		s.mu.RUnlock()
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "没有已上传的 RikkaHub 备份"})
		return
	}
	src := s.rikkahub
	s.mu.RUnlock()

	result, err := converter.ConvertRikkaHubToKelivo(src)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "转换失败: " + err.Error()})
		return
	}

	data, err := writer.MarshalKelivoBackup(result)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "序列化失败: " + err.Error()})
		return
	}

	timestamp := time.Now().Format("20060102_150405")
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="kelivo_converted_%s.zip"`, timestamp))
	w.Write(data)
}

func (s *Server) handleInfo(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.backupType == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "没有已上传的备份"})
		return
	}

	info := map[string]interface{}{
		"type":       s.backupType,
		"uploadedAt": s.uploadedAt,
	}

	switch s.backupType {
	case "kelivo":
		b := s.kelivo
		if b.Settings != nil {
			st := b.Settings
			info["providers"] = len(st.ProviderConfigs)
			info["assistants"] = len(st.Assistants)
			info["mcpServers"] = len(st.MCPServers)
			info["worldBooks"] = len(st.WorldBooks)
			info["quickPhrases"] = len(st.QuickPhrases)
			info["injections"] = len(st.InstructionInjections)
		}
		if b.Chats != nil {
			info["conversations"] = len(b.Chats.Conversations)
			info["messages"] = len(b.Chats.Messages)
		}
	case "rikkahub":
		b := s.rikkahub
		if b.Settings != nil {
			st := b.Settings
			info["providers"] = len(st.Providers)
			info["assistants"] = len(st.Assistants)
			info["mcpServers"] = len(st.MCPServers)
			info["lorebooks"] = len(st.Lorebooks)
			info["quickMessages"] = len(st.QuickMessages)
			info["injections"] = len(st.ModeInjections)
		}
		info["conversations"] = len(b.Conversations)
		info["messageNodes"] = len(b.MessageNodes)
		info["uploadFiles"] = len(b.UploadFiles)
		info["managedFiles"] = len(b.ManagedFiles)
	}

	writeJSON(w, http.StatusOK, info)
}

func (s *Server) handleProviders(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.backupType == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "没有已上传的备份"})
		return
	}

	switch s.backupType {
	case "kelivo":
		if s.kelivo.Settings == nil {
			writeJSON(w, http.StatusOK, []interface{}{})
			return
		}
		providers := make([]map[string]interface{}, 0, len(s.kelivo.Settings.ProviderConfigs))
		for id, p := range s.kelivo.Settings.ProviderConfigs {
			providers = append(providers, map[string]interface{}{
				"id":           id,
				"name":         p.Name,
				"baseUrl":      p.BaseURL,
				"providerType": p.ProviderType,
				"enabled":      p.Enabled,
				"models":       p.Models,
				"hasApiKey":    p.APIKey != "",
			})
		}
		writeJSON(w, http.StatusOK, providers)
	case "rikkahub":
		if s.rikkahub.Settings == nil {
			writeJSON(w, http.StatusOK, []interface{}{})
			return
		}
		providers := make([]map[string]interface{}, 0, len(s.rikkahub.Settings.Providers))
		for _, p := range s.rikkahub.Settings.Providers {
			models := make([]string, 0, len(p.Models))
			for _, m := range p.Models {
				models = append(models, m.DisplayName)
			}
			providers = append(providers, map[string]interface{}{
				"id":        p.ID,
				"name":      p.Name,
				"baseUrl":   p.BaseURL,
				"type":      p.Type,
				"enabled":   p.Enabled,
				"models":    models,
				"hasApiKey": p.APIKey != "",
			})
		}
		writeJSON(w, http.StatusOK, providers)
	}
}

func (s *Server) handleAssistants(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.backupType == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "没有已上传的备份"})
		return
	}

	switch s.backupType {
	case "kelivo":
		if s.kelivo.Settings == nil {
			writeJSON(w, http.StatusOK, []interface{}{})
			return
		}
		assistants := make([]map[string]interface{}, 0, len(s.kelivo.Settings.Assistants))
		for _, a := range s.kelivo.Settings.Assistants {
			assistants = append(assistants, map[string]interface{}{
				"id":                    a.ID,
				"name":                  a.Name,
				"chatModelProvider":     a.ChatModelProvider,
				"chatModelId":           a.ChatModelID,
				"temperature":           a.Temperature,
				"topP":                  a.TopP,
				"contextMessageSize":    a.ContextMessageSize,
				"streamOutput":          a.StreamOutput,
				"systemPrompt":          truncateStr(a.SystemPrompt, 500),
				"mcpServerIds":          a.MCPServerIDs,
				"enableMemory":          a.EnableMemory,
				"thinkingBudget":        a.ThinkingBudget,
				"presetMessages":        a.PresetMessages,
			})
		}
		writeJSON(w, http.StatusOK, assistants)
	case "rikkahub":
		if s.rikkahub.Settings == nil {
			writeJSON(w, http.StatusOK, []interface{}{})
			return
		}
		assistants := make([]map[string]interface{}, 0, len(s.rikkahub.Settings.Assistants))
		for _, a := range s.rikkahub.Settings.Assistants {
			assistants = append(assistants, map[string]interface{}{
				"id":                 a.ID,
				"name":               a.Name,
				"chatModelId":        a.ChatModelID,
				"temperature":        a.Temperature,
				"topP":               a.TopP,
				"contextMessageSize": a.ContextMessageSize,
				"streamOutput":       a.StreamOutput,
				"systemPrompt":       truncateStr(a.SystemPrompt, 500),
				"mcpServers":         a.MCPServers,
				"enableMemory":       a.EnableMemory,
				"useGlobalMemory":    a.UseGlobalMemory,
				"thinkingBudget":     a.ThinkingBudget,
				"presetMessages":     a.PresetMessages,
				"lorebookIds":        a.LorebookIDs,
				"modeInjectionIds":   a.ModeInjectionIDs,
				"tags":               a.Tags,
			})
		}
		writeJSON(w, http.StatusOK, assistants)
	}
}

func (s *Server) handleMCP(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.backupType == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "没有已上传的备份"})
		return
	}

	switch s.backupType {
	case "kelivo":
		if s.kelivo.Settings == nil {
			writeJSON(w, http.StatusOK, []interface{}{})
			return
		}
		writeJSON(w, http.StatusOK, s.kelivo.Settings.MCPServers)
	case "rikkahub":
		if s.rikkahub.Settings == nil {
			writeJSON(w, http.StatusOK, []interface{}{})
			return
		}
		writeJSON(w, http.StatusOK, s.rikkahub.Settings.MCPServers)
	}
}

func (s *Server) handleConversations(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.backupType == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "没有已上传的备份"})
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/conversations")

	switch s.backupType {
	case "kelivo":
		if path == "" || path == "/" {
			if s.kelivo.Chats == nil {
				writeJSON(w, http.StatusOK, []interface{}{})
				return
			}
			msgMap := make(map[string]int)
			if s.kelivo.Chats != nil {
				for _, m := range s.kelivo.Chats.Messages {
					msgMap[m.ConversationID]++
				}
			}
			convs := make([]map[string]interface{}, 0, len(s.kelivo.Chats.Conversations))
			for _, c := range s.kelivo.Chats.Conversations {
				convs = append(convs, map[string]interface{}{
					"id":                c.ID,
					"title":             c.Title,
					"assistantId":       c.AssistantID,
					"isPinned":          c.IsPinned,
					"createdAt":         c.CreatedAt,
					"updatedAt":         c.UpdatedAt,
					"messageCount":      msgMap[c.ID],
					"mcpServerIds":      c.MCPServerIDs,
				})
			}
			writeJSON(w, http.StatusOK, convs)
		} else {
			convID := strings.TrimPrefix(path, "/")
			if s.kelivo.Chats == nil {
				writeJSON(w, http.StatusNotFound, map[string]string{"error": "未找到会话"})
				return
			}
			var conv *models.KelivoConversation
			for _, c := range s.kelivo.Chats.Conversations {
				if c.ID == convID {
					conv = &c
					break
				}
			}
			if conv == nil {
				writeJSON(w, http.StatusNotFound, map[string]string{"error": "未找到会话"})
				return
			}
			var msgs []models.KelivoMessage
			if s.kelivo.Chats != nil {
				for _, m := range s.kelivo.Chats.Messages {
					if m.ConversationID == convID {
						msgs = append(msgs, m)
					}
				}
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"conversation": conv,
				"messages":     msgs,
			})
		}

	case "rikkahub":
		if path == "" || path == "/" {
			if len(s.rikkahub.Conversations) == 0 {
				writeJSON(w, http.StatusOK, []interface{}{})
				return
			}
			nodeCount := make(map[string]int)
			for _, n := range s.rikkahub.MessageNodes {
				nodeCount[n.ConversationID]++
			}
			convs := make([]map[string]interface{}, 0, len(s.rikkahub.Conversations))
			for _, c := range s.rikkahub.Conversations {
				convs = append(convs, map[string]interface{}{
					"id":           c.ID,
					"title":        c.Title,
					"assistantId":  c.AssistantID,
					"isPinned":     c.IsPinned,
					"createdAt":    c.CreateAt,
					"updatedAt":    c.UpdateAt,
					"nodeCount":    nodeCount[c.ID],
				})
			}
			writeJSON(w, http.StatusOK, convs)
		} else {
			convID := strings.TrimPrefix(path, "/")
			var conv *models.RikkaHubConversation
			for i := range s.rikkahub.Conversations {
				if s.rikkahub.Conversations[i].ID == convID {
					conv = &s.rikkahub.Conversations[i]
					break
				}
			}
			if conv == nil {
				writeJSON(w, http.StatusNotFound, map[string]string{"error": "未找到会话"})
				return
			}
			var nodes []models.RikkaHubMessageNode
			for _, n := range s.rikkahub.MessageNodes {
				if n.ConversationID == convID {
					nodes = append(nodes, n)
				}
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"conversation": conv,
				"messageNodes": nodes,
			})
		}
	}
}

func (s *Server) handleLorebooks(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.backupType == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "没有已上传的备份"})
		return
	}

	switch s.backupType {
	case "kelivo":
		if s.kelivo.Settings == nil {
			writeJSON(w, http.StatusOK, []interface{}{})
			return
		}
		writeJSON(w, http.StatusOK, s.kelivo.Settings.WorldBooks)
	case "rikkahub":
		if s.rikkahub.Settings == nil {
			writeJSON(w, http.StatusOK, []interface{}{})
			return
		}
		writeJSON(w, http.StatusOK, s.rikkahub.Settings.Lorebooks)
	}
}

func (s *Server) handleQuickMessages(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.backupType == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "没有已上传的备份"})
		return
	}

	switch s.backupType {
	case "kelivo":
		if s.kelivo.Settings == nil {
			writeJSON(w, http.StatusOK, []interface{}{})
			return
		}
		writeJSON(w, http.StatusOK, s.kelivo.Settings.QuickPhrases)
	case "rikkahub":
		if s.rikkahub.Settings == nil {
			writeJSON(w, http.StatusOK, []interface{}{})
			return
		}
		writeJSON(w, http.StatusOK, s.rikkahub.Settings.QuickMessages)
	}
}

func (s *Server) handleInjections(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.backupType == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "没有已上传的备份"})
		return
	}

	switch s.backupType {
	case "kelivo":
		if s.kelivo.Settings == nil {
			writeJSON(w, http.StatusOK, []interface{}{})
			return
		}
		writeJSON(w, http.StatusOK, s.kelivo.Settings.InstructionInjections)
	case "rikkahub":
		if s.rikkahub.Settings == nil {
			writeJSON(w, http.StatusOK, []interface{}{})
			return
		}
		writeJSON(w, http.StatusOK, s.rikkahub.Settings.ModeInjections)
	}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
