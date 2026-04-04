package models

import "encoding/json"

// KelıvoBackup represents the complete Kelivo backup data
type KelivoBackup struct {
	Settings *KelivoSettings
	Chats    *KelıvoChats
}

// KelivoSettings represents settings.json in Kelivo backup
// Uses json.RawMessage for fields with mixed/uncertain types
type KelivoSettings struct {
	DisplayShowChatListDate      bool     `json:"display_show_chat_list_date_v1"`
	DisplayEnterToSendOnMobile   bool     `json:"display_enter_to_send_on_mobile_v1"`
	MCPRequestTimeoutMs          int      `json:"mcp_request_timeout_ms_v1"`
	OCRModel                     string   `json:"ocr_model_v1"`
	DisplayUsePureBackground     bool     `json:"display_use_pure_background_v1"`
	MCPServersRaw                string   `json:"mcp_servers_v1"`
	ProviderConfigsRaw           string   `json:"provider_configs_v1"`
	ProviderGroupCollapsedRaw    string   `json:"provider_group_collapsed_v1"`
	ThinkingBudget               int      `json:"thinking_budget_v1"`
	TitleModelRaw                string   `json:"title_model_v1"`
	DisplayHapticsOnGenerate     bool     `json:"display_haptics_on_generate_v1"`
	CurrentAssistantID           string   `json:"current_assistant_id_v1"`
	InstructionInjectionsRaw     string   `json:"instruction_injections_v1"`
	TranslateModelRaw            string   `json:"translate_model_v1"`
	ProvidersOrderList           []string `json:"providers_order_v1"`
	AppLocale                    string   `json:"app_locale_v1"`
	PinnedModelsList             []string `json:"pinned_models_v1"`
	DisplayDesktopShowTray       bool     `json:"display_desktop_show_tray_v1"`
	DisplayDesktopMinimizeToTray bool     `json:"display_desktop_minimize_to_tray_on_close_v1"`
	AssistantsRaw                string   `json:"assistants_v1"`
	ProviderGroupMapRaw          string   `json:"provider_group_map_v1"`
	SearchEnabled                bool     `json:"search_enabled_v1"`
	ProviderConfigsBackupRaw     string   `json:"provider_configs_backup_v1"`
	MigrationsVersion            int      `json:"migrations_version_v1"`
	QuickPhrasesRaw              string   `json:"quick_phrases_v1"`
	WorldBooksRaw                string   `json:"world_books_v1"`
	WorldBooksCollapsedRaw       string   `json:"world_books_collapsed_v1"`
	TTSServicesRaw               string   `json:"tts_services_v1"`
	TTSSelected                  int      `json:"tts_selected_v1"`
	SearchServicesRaw            string   `json:"search_services_v1"`
	SearchSelected               int      `json:"search_selected_v1"`
	SummaryModelRaw              string   `json:"summary_model_v1"`
	CompressModelRaw             string   `json:"compress_model_v1"`
	SelectedModelRaw             string   `json:"selected_model_v1"`

	// Fields with mixed/uncertain types stored as raw JSON
	GlobalProxyBypassRaw         json.RawMessage `json:"global_proxy_bypass_v1"`
	AndroidBackgroundChatModeRaw json.RawMessage `json:"android_background_chat_mode_v1"`

	// Parsed fields (not serialized directly)
	MCPServers            []KelivoMCPServer            `json:"-"`
	ProviderConfigs       map[string]KelivoProvider    `json:"-"`
	Assistants            []KelivoAssistant            `json:"-"`
	QuickPhrases          []KelivoQuickPhrase          `json:"-"`
	WorldBooks            []KelivoWorldBook            `json:"-"`
	SearchServices        []json.RawMessage            `json:"-"`
	TTSServices           []json.RawMessage            `json:"-"`
	InstructionInjections []KelivoInstructionInjection `json:"-"`
}

// KelivoMCPServer represents an MCP server configuration
type KelivoMCPServer struct {
	ID        string            `json:"id"`
	Enabled   bool              `json:"enabled"`
	Name      string            `json:"name"`
	Transport string            `json:"transport"`
	URL       string            `json:"url,omitempty"`
	Tools     []KelivoMCPTool   `json:"tools,omitempty"`
	Headers   map[string]string `json:"headers,omitempty"`
}

type KelivoMCPTool struct {
	Enabled     bool             `json:"enabled"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Params      []KelivoMCPParam `json:"params,omitempty"`
	Schema      json.RawMessage  `json:"schema,omitempty"`
}

type KelivoMCPParam struct {
	Name     string      `json:"name"`
	Required bool        `json:"required"`
	Type     string      `json:"type"`
	Default  interface{} `json:"default"`
}

// KelivoProvider represents an AI provider configuration
type KelivoProvider struct {
	ID                 string                         `json:"id"`
	Enabled            bool                           `json:"enabled"`
	Name               string                         `json:"name"`
	APIKey             string                         `json:"apiKey"`
	BaseURL            string                         `json:"baseUrl"`
	ProviderType       string                         `json:"providerType"`
	ChatPath           *string                        `json:"chatPath"`
	UseResponseAPI     bool                           `json:"useResponseApi"`
	VertexAI           interface{}                    `json:"vertexAI"`
	Location           interface{}                    `json:"location"`
	ProjectID          interface{}                    `json:"projectId"`
	ServiceAccountJSON interface{}                    `json:"serviceAccountJson"`
	Models             []string                       `json:"models"`
	ModelOverrides     map[string]KelivoModelOverride `json:"modelOverrides"`
	ProxyEnabled       bool                           `json:"proxyEnabled"`
	ProxyHost          string                         `json:"proxyHost"`
	ProxyPort          string                         `json:"proxyPort"`
	ProxyUsername      string                         `json:"proxyUsername"`
	ProxyPassword      string                         `json:"proxyPassword"`
	AvatarType         *string                        `json:"avatarType"`
	AvatarValue        *string                        `json:"avatarValue"`
	MultiKeyEnabled    bool                           `json:"multiKeyEnabled"`
	APIKeys            []string                       `json:"apiKeys"`
	KeyManagement      KelivoKeyManagement            `json:"keyManagement"`
}

type KelivoModelOverride struct {
	Type      string   `json:"type"`
	Input     []string `json:"input"`
	Output    []string `json:"output"`
	Abilities []string `json:"abilities"`
}

type KelivoKeyManagement struct {
	Strategy                   string      `json:"strategy"`
	MaxFailuresBeforeDisable   int         `json:"maxFailuresBeforeDisable"`
	FailureRecoveryTimeMinutes int         `json:"failureRecoveryTimeMinutes"`
	EnableAutoRecovery         bool        `json:"enableAutoRecovery"`
	RoundRobinIndex            interface{} `json:"roundRobinIndex"`
}

// KelivoAssistant represents an AI assistant configuration
type KelivoAssistant struct {
	ID                         string                `json:"id"`
	Name                       string                `json:"name"`
	Avatar                     interface{}           `json:"avatar"`
	AvatarType                 *string               `json:"avatarType,omitempty"`
	AvatarValue                *string               `json:"avatarValue,omitempty"`
	UseAssistantAvatar         bool                  `json:"useAssistantAvatar"`
	ChatModelProvider          string                `json:"chatModelProvider"`
	ChatModelID                string                `json:"chatModelId"`
	Temperature                float64               `json:"temperature"`
	TopP                       float64               `json:"topP"`
	ContextMessageSize         int                   `json:"contextMessageSize"`
	LimitContextMessages       bool                  `json:"limitContextMessages"`
	StreamOutput               bool                  `json:"streamOutput"`
	ThinkingBudget             int                   `json:"thinkingBudget"`
	MaxTokens                  *int                  `json:"maxTokens"`
	SystemPrompt               string                `json:"systemPrompt"`
	MessageTemplate            string                `json:"messageTemplate"`
	MCPServerIDs               []string              `json:"mcpServerIds"`
	Background                 interface{}           `json:"background"`
	Deletable                  bool                  `json:"deletable"`
	CustomHeaders              []KelivoCustomHeader  `json:"customHeaders"`
	CustomBody                 []KelivoCustomBody    `json:"customBody"`
	EnableMemory               bool                  `json:"enableMemory"`
	EnableRecentChatsReference bool                  `json:"enableRecentChatsReference"`
	PresetMessages             []KelivoPresetMessage `json:"presetMessages"`
	RegexRules                 []KelivoRegexRule     `json:"regexRules"`
}

type KelivoCustomHeader struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type KelivoCustomBody struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type KelivoPresetMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type KelivoRegexRule struct {
	Pattern     string `json:"pattern"`
	Replacement string `json:"replacement"`
	Enabled     bool   `json:"enabled"`
}

// KelivoQuickPhrase represents a quick phrase/message
type KelivoQuickPhrase struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Content     string  `json:"content"`
	IsGlobal    bool    `json:"isGlobal"`
	AssistantID *string `json:"assistantId"`
}

// KelivoWorldBook represents a world book (lorebook)
type KelivoWorldBook struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Enabled     bool                   `json:"enabled"`
	Entries     []KelivoWorldBookEntry `json:"entries"`
}

type KelivoWorldBookEntry struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Enabled        bool     `json:"enabled"`
	Priority       int      `json:"priority"`
	Position       string   `json:"position"`
	Content        string   `json:"content"`
	InjectDepth    int      `json:"injectDepth"`
	Role           string   `json:"role"`
	Keywords       []string `json:"keywords"`
	UseRegex       bool     `json:"useRegex"`
	CaseSensitive  bool     `json:"caseSensitive"`
	ScanDepth      int      `json:"scanDepth"`
	ConstantActive bool     `json:"constantActive"`
}

// KelivoInstructionInjection represents instruction injection config
type KelivoInstructionInjection struct {
	ID      string `json:"id"`
	Enabled bool   `json:"enabled"`
	Content string `json:"content"`
}

// KelıvoChats represents chats.json in Kelivo backup
type KelıvoChats struct {
	Version       int                  `json:"version"`
	Conversations []KelivoConversation `json:"conversations"`
	Messages      []KelivoMessage      `json:"messages"`
}

// KelivoConversation represents a chat conversation
type KelivoConversation struct {
	ID                         string         `json:"id"`
	Title                      string         `json:"title"`
	CreatedAt                  string         `json:"createdAt"`
	UpdatedAt                  string         `json:"updatedAt"`
	MessageIDs                 []string       `json:"messageIds"`
	IsPinned                   bool           `json:"isPinned"`
	MCPServerIDs               []string       `json:"mcpServerIds"`
	AssistantID                string         `json:"assistantId"`
	TruncateIndex              int            `json:"truncateIndex"`
	VersionSelections          map[string]int `json:"versionSelections"`
	Summary                    *string        `json:"summary"`
	LastSummarizedMessageCount int            `json:"lastSummarizedMessageCount"`
}

// KelivoMessage represents a chat message
type KelivoMessage struct {
	ID                    string   `json:"id"`
	Role                  string   `json:"role"`
	Content               string   `json:"content"`
	Timestamp             string   `json:"timestamp"`
	ModelID               *string  `json:"modelId"`
	ProviderID            *string  `json:"providerId"`
	TotalTokens           *int     `json:"totalTokens"`
	ConversationID        string   `json:"conversationId"`
	IsStreaming           bool     `json:"isStreaming"`
	ReasoningText         *string  `json:"reasoningText"`
	ReasoningStartAt      *string  `json:"reasoningStartAt"`
	ReasoningFinishedAt   *string  `json:"reasoningFinishedAt"`
	Translation           *string  `json:"translation"`
	ReasoningSegmentsJSON *string  `json:"reasoningSegmentsJson"`
	GroupID               string   `json:"groupId"`
	Version               int      `json:"version"`
	AttachmentIDs         []string `json:"attachmentIds,omitempty"`
}
