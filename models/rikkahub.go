package models

import "encoding/json"

// RikkaHubBackup represents the complete RikkaHub backup data
type RikkaHubBackup struct {
	Settings      *RikkaHubSettings
	Conversations []RikkaHubConversation
	MessageNodes  []RikkaHubMessageNode
	ManagedFiles  []RikkaHubManagedFile
	UploadFiles   map[string][]byte // filename -> content
}

// RikkaHubSettings represents settings.json in RikkaHub backup
type RikkaHubSettings struct {
	Init                    bool                    `json:"init"`
	DynamicColor            bool                    `json:"dynamicColor"`
	ThemeID                 string                  `json:"themeId"`
	DeveloperMode           bool                    `json:"developerMode"`
	DisplaySetting          RikkaHubDisplaySetting  `json:"displaySetting"`
	EnableWebSearch         bool                    `json:"enableWebSearch"`
	FavoriteModels          []string                `json:"favoriteModels"`
	ChatModelID             string                  `json:"chatModelId"`
	TitleModelID            string                  `json:"titleModelId"`
	ImageGenModelID         string                  `json:"imageGenerationModelId"`
	TitlePrompt             string                  `json:"titlePrompt"`
	TranslateModeID         string                  `json:"translateModeId"`
	TranslatePrompt         string                  `json:"translatePrompt"`
	TranslateThinkingBudget int                     `json:"translateThinkingBudget"`
	SuggestionModelID       string                  `json:"suggestionModelId"`
	SuggestionPrompt        string                  `json:"suggestionPrompt"`
	OCRModelID              string                  `json:"ocrModelId"`
	OCRPrompt               string                  `json:"ocrPrompt"`
	CompressModelID         string                  `json:"compressModelId"`
	CompressPrompt          string                  `json:"compressPrompt"`
	AssistantID             string                  `json:"assistantId"`
	Providers               []RikkaHubProvider      `json:"providers"`
	Assistants              []RikkaHubAssistant     `json:"assistants"`
	AssistantTags           []json.RawMessage       `json:"assistantTags"`
	SearchServices          []RikkaHubSearchService `json:"searchServices"`
	SearchCommonOptions     json.RawMessage         `json:"searchCommonOptions"`
	SearchServiceSelected   int                     `json:"searchServiceSelected"`
	MCPServers              []RikkaHubMCPServer     `json:"mcpServers"`
	WebDavConfig            json.RawMessage         `json:"webDavConfig"`
	S3Config                json.RawMessage         `json:"s3Config"`
	TTSProviders            []RikkaHubTTSProvider   `json:"ttsProviders"`
	SelectedTTSProviderID   string                  `json:"selectedTTSProviderId"`
	ModeInjections          []RikkaHubModeInjection `json:"modeInjections"`
	Lorebooks               []RikkaHubLorebook      `json:"lorebooks"`
	QuickMessages           []RikkaHubQuickMessage  `json:"quickMessages"`
	WebServerEnabled        bool                    `json:"webServerEnabled"`
	WebServerPort           int                     `json:"webServerPort"`
	WebServerJwtEnabled     bool                    `json:"webServerJwtEnabled"`
	WebServerAccessPassword string                  `json:"webServerAccessPassword"`
	WebServerLocalhostOnly  bool                    `json:"webServerLocalhostOnly"`
	BackupReminderConfig    json.RawMessage         `json:"backupReminderConfig"`
	LaunchCount             int                     `json:"launchCount"`
	SponsorAlertDismissedAt *int                    `json:"sponsorAlertDismissedAt"`
}

type RikkaHubDisplaySetting struct {
	UserAvatar                            RikkaHubAvatar `json:"userAvatar"`
	UserNickname                          string         `json:"userNickname"`
	UseAppIconStyleLoadingIndicator       bool           `json:"useAppIconStyleLoadingIndicator"`
	ShowUserAvatar                        bool           `json:"showUserAvatar"`
	ShowAssistantBubble                   bool           `json:"showAssistantBubble"`
	ShowModelIcon                         bool           `json:"showModelIcon"`
	ShowModelName                         bool           `json:"showModelName"`
	ShowDateBelowName                     bool           `json:"showDateBelowName"`
	ShowTokenUsage                        bool           `json:"showTokenUsage"`
	ShowThinkingContent                   bool           `json:"showThinkingContent"`
	AutoCloseThinking                     bool           `json:"autoCloseThinking"`
	ShowUpdates                           bool           `json:"showUpdates"`
	ShowMessageJumper                     bool           `json:"showMessageJumper"`
	MessageJumperOnLeft                   bool           `json:"messageJumperOnLeft"`
	FontSizeRatio                         float64        `json:"fontSizeRatio"`
	EnableMessageGenerationHapticEffect   bool           `json:"enableMessageGenerationHapticEffect"`
	SkipCropImage                         bool           `json:"skipCropImage"`
	EnableNotificationOnMessageGeneration bool           `json:"enableNotificationOnMessageGeneration"`
	EnableLiveUpdateNotification          bool           `json:"enableLiveUpdateNotification"`
	CodeBlockAutoWrap                     bool           `json:"codeBlockAutoWrap"`
	CodeBlockAutoCollapse                 bool           `json:"codeBlockAutoCollapse"`
	ShowLineNumbers                       bool           `json:"showLineNumbers"`
	TTSOnlyReadQuoted                     bool           `json:"ttsOnlyReadQuoted"`
	AutoPlayTTSAfterGeneration            bool           `json:"autoPlayTTSAfterGeneration"`
	PasteLongTextAsFile                   bool           `json:"pasteLongTextAsFile"`
	PasteLongTextThreshold                int            `json:"pasteLongTextThreshold"`
	SendOnEnter                           bool           `json:"sendOnEnter"`
	EnableAutoScroll                      bool           `json:"enableAutoScroll"`
	EnableLatexRendering                  bool           `json:"enableLatexRendering"`
	EnableBlurEffect                      bool           `json:"enableBlurEffect"`
	ChatFontFamily                        string         `json:"chatFontFamily"`
}

type RikkaHubAvatar struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

// RikkaHubProvider represents an AI provider configuration
type RikkaHubProvider struct {
	Type          string                `json:"type"`
	ID            string                `json:"id"`
	Enabled       bool                  `json:"enabled"`
	Name          string                `json:"name"`
	Models        []RikkaHubModel       `json:"models"`
	BalanceOption RikkaHubBalanceOption `json:"balanceOption"`
	APIKey        string                `json:"apiKey"`
	BaseURL       string                `json:"baseUrl"`
	Headers       []RikkaHubHeader      `json:"headers"`
}

type RikkaHubModel struct {
	ModelID           string            `json:"modelId"`
	DisplayName       string            `json:"displayName"`
	ID                string            `json:"id"`
	Type              string            `json:"type"`
	CustomHeaders     []RikkaHubHeader  `json:"customHeaders"`
	CustomBodies      []RikkaHubBody    `json:"customBodies"`
	InputModalities   []string          `json:"inputModalities"`
	OutputModalities  []string          `json:"outputModalities"`
	Abilities         []string          `json:"abilities"`
	Tools             []json.RawMessage `json:"tools"`
	ProviderOverwrite *json.RawMessage  `json:"providerOverwrite"`
}

type RikkaHubBalanceOption struct {
	Enabled    bool   `json:"enabled"`
	APIPath    string `json:"apiPath"`
	ResultPath string `json:"resultPath"`
}

type RikkaHubHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type RikkaHubBody struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// RikkaHubAssistant represents an AI assistant configuration
type RikkaHubAssistant struct {
	ID                         string                  `json:"id"`
	ChatModelID                string                  `json:"chatModelId"`
	Name                       string                  `json:"name"`
	Avatar                     *RikkaHubAvatar         `json:"avatar"`
	UseAssistantAvatar         bool                    `json:"useAssistantAvatar"`
	Tags                       []string                `json:"tags"`
	SystemPrompt               string                  `json:"systemPrompt"`
	Temperature                float64                 `json:"temperature"`
	TopP                       *float64                `json:"topP"`
	ContextMessageSize         int                     `json:"contextMessageSize"`
	StreamOutput               bool                    `json:"streamOutput"`
	EnableMemory               bool                    `json:"enableMemory"`
	UseGlobalMemory            bool                    `json:"useGlobalMemory"`
	EnableRecentChatsReference bool                    `json:"enableRecentChatsReference"`
	MessageTemplate            string                  `json:"messageTemplate"`
	PresetMessages             []RikkaHubPresetMessage `json:"presetMessages"`
	QuickMessageIDs            []string                `json:"quickMessageIds"`
	Regexes                    []RikkaHubRegex         `json:"regexes"`
	ThinkingBudget             int                     `json:"thinkingBudget"`
	MaxTokens                  *int                    `json:"maxTokens"`
	CustomHeaders              []RikkaHubHeader        `json:"customHeaders"`
	CustomBodies               []RikkaHubBody          `json:"customBodies"`
	MCPServers                 []string                `json:"mcpServers"`
	LocalTools                 []RikkaHubLocalTool     `json:"localTools"`
	Background                 *string                 `json:"background"`
	BackgroundOpacity          float64                 `json:"backgroundOpacity"`
	ModeInjectionIDs           []string                `json:"modeInjectionIds"`
	LorebookIDs                []string                `json:"lorebookIds"`
	EnabledSkills              []string                `json:"enabledSkills"`
	EnableTimeReminder         bool                    `json:"enableTimeReminder"`
}

type RikkaHubPresetMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type RikkaHubRegex struct {
	Pattern     string `json:"pattern"`
	Replacement string `json:"replacement"`
	Enabled     bool   `json:"enabled"`
}

type RikkaHubLocalTool struct {
	Type string `json:"type"`
}

// RikkaHubMCPServer represents an MCP server configuration
type RikkaHubMCPServer struct {
	Type          string                   `json:"type"`
	ID            string                   `json:"id"`
	CommonOptions RikkaHubMCPCommonOptions `json:"commonOptions"`
	URL           string                   `json:"url,omitempty"`
	Command       string                   `json:"command,omitempty"`
	Args          []string                 `json:"args,omitempty"`
	Env           map[string]string        `json:"env,omitempty"`
}

type RikkaHubMCPCommonOptions struct {
	Enable  bool              `json:"enable"`
	Name    string            `json:"name"`
	Headers []RikkaHubHeader  `json:"headers"`
	Tools   []RikkaHubMCPTool `json:"tools"`
}

type RikkaHubMCPTool struct {
	Enable        bool            `json:"enable"`
	Name          string          `json:"name"`
	Description   string          `json:"description"`
	InputSchema   json.RawMessage `json:"inputSchema"`
	NeedsApproval bool            `json:"needsApproval"`
}

// RikkaHubSearchService represents a search service configuration
type RikkaHubSearchService struct {
	Type           string `json:"type"`
	ID             string `json:"id"`
	AcceptLanguage string `json:"acceptLanguage,omitempty"`
	APIKey         string `json:"apiKey,omitempty"`
	URL            string `json:"url,omitempty"`
	Region         string `json:"region,omitempty"`
}

// RikkaHubTTSProvider represents a TTS provider configuration
type RikkaHubTTSProvider struct {
	ID      string          `json:"id"`
	Name    string          `json:"name"`
	Type    string          `json:"type"`
	Options json.RawMessage `json:"options"`
}

// RikkaHubModeInjection represents a mode injection configuration
type RikkaHubModeInjection struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Content string `json:"content"`
	Enabled bool   `json:"enabled"`
}

// RikkaHubLorebook represents a lorebook (world book)
type RikkaHubLorebook struct {
	ID          string                  `json:"id"`
	Name        string                  `json:"name"`
	Description string                  `json:"description"`
	Enabled     bool                    `json:"enabled"`
	Entries     []RikkaHubLorebookEntry `json:"entries"`
}

type RikkaHubLorebookEntry struct {
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

// RikkaHubQuickMessage represents a quick message configuration
type RikkaHubQuickMessage struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Content     string  `json:"content"`
	IsGlobal    bool    `json:"isGlobal"`
	AssistantID *string `json:"assistantId"`
}

// RikkaHubConversation represents a chat conversation from SQLite
type RikkaHubConversation struct {
	ID          string   `json:"id"`
	AssistantID string   `json:"assistant_id"`
	Title       string   `json:"title"`
	Nodes       []string `json:"nodes"` // serialized as JSON array
	CreateAt    int64    `json:"create_at"`
	UpdateAt    int64    `json:"update_at"`
	Suggestions []string `json:"suggestions"`
	IsPinned    bool     `json:"is_pinned"`
}

// RikkaHubMessageNode represents a message node from SQLite
type RikkaHubMessageNode struct {
	ID             string            `json:"id"`
	ConversationID string            `json:"conversation_id"`
	NodeIndex      int               `json:"node_index"`
	Messages       []RikkaHubMessage `json:"messages"`
	SelectIndex    int               `json:"select_index"`
}

// RikkaHubMessage represents a single message in a node
type RikkaHubMessage struct {
	ID          string                `json:"id"`
	Role        string                `json:"role"`
	Parts       []RikkaHubMessagePart `json:"parts"`
	Annotations []json.RawMessage     `json:"annotations"`
	CreatedAt   string                `json:"createdAt"`
	FinishedAt  *string               `json:"finishedAt"`
	ModelID     *string               `json:"modelId"`
	Usage       *RikkaHubUsage        `json:"usage"`
	Translation *string               `json:"translation"`
}

// RikkaHubMessagePart represents a part of a message (union type based on type field)
// type can be: "text", "image", "video", "audio", "document", "reasoning", "tool"
type RikkaHubMessagePart struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
	// For image/video/audio parts
	URL string `json:"url,omitempty"`
	// For document parts
	FileName string `json:"fileName,omitempty"`
	MimeType string `json:"mimeType,omitempty"`
	// For reasoning parts
	Reasoning      string          `json:"reasoning,omitempty"`
	ReasoningSteps json.RawMessage `json:"steps,omitempty"`
	// For tool parts
	ToolCallID    string          `json:"toolCallId,omitempty"`
	ToolName      string          `json:"toolName,omitempty"`
	ToolInput     string          `json:"input,omitempty"`
	ToolOutput    json.RawMessage `json:"output,omitempty"`
	ApprovalState string          `json:"approvalState,omitempty"`
	// Common
	FileID   string          `json:"fileId,omitempty"`
	Data     string          `json:"data,omitempty"`
	Metadata json.RawMessage `json:"metadata,omitempty"`
	Priority int             `json:"priority,omitempty"`
}

type RikkaHubUsage struct {
	PromptTokens     int `json:"promptTokens"`
	CompletionTokens int `json:"completionTokens"`
	CachedTokens     int `json:"cachedTokens"`
	TotalTokens      int `json:"totalTokens"`
}

// RikkaHubManagedFile represents a managed file record from SQLite
type RikkaHubManagedFile struct {
	ID           int64  `json:"id"`
	Folder       string `json:"folder"`
	RelativePath string `json:"relative_path"`
	DisplayName  string `json:"display_name"`
	MimeType     string `json:"mime_type"`
	SizeBytes    int64  `json:"size_bytes"`
	CreatedAt    int64  `json:"created_at"`
	UpdatedAt    int64  `json:"updated_at"`
}
