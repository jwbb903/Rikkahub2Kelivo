package converter

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/converter/backup-converter/models"
)

// ConvertKelivoToRikkaHub converts a Kelivo backup to a RikkaHub backup
func ConvertKelivoToRikkaHub(src *models.KelivoBackup) (*models.RikkaHubBackup, error) {
	if src == nil {
		return nil, fmt.Errorf("source backup is nil")
	}

	dst := &models.RikkaHubBackup{
		UploadFiles: make(map[string][]byte),
	}

	settings, modelIDMap, err := convertKelivoSettingsToRikkaHub(src.Settings)
	if err != nil {
		return nil, fmt.Errorf("failed to convert settings: %w", err)
	}
	dst.Settings = settings

	if src.Chats != nil {
		conversations, nodes := convertKelivoChatsToRikkaHub(src.Chats, modelIDMap, src.Settings)
		dst.Conversations = conversations
		dst.MessageNodes = nodes
	}

	return dst, nil
}

// modelIDMap maps "ProviderID::modelId" -> RikkaHub model UUID
type modelIDKey = string

func convertKelivoSettingsToRikkaHub(settings *models.KelivoSettings) (*models.RikkaHubSettings, map[modelIDKey]string, error) {
	if settings == nil {
		return nil, nil, fmt.Errorf("settings is nil")
	}

	modelIDMap := make(map[modelIDKey]string)

	// Convert providers
	providers := convertKelivoProviders(settings.ProviderConfigs, modelIDMap)

	// Find model IDs for special models
	chatModelID := findRikkaHubModelID(modelIDMap, settings)
	titleModelID := findRikkaHubModelIDByRaw(modelIDMap, settings.TitleModelRaw)
	translateModelID := findRikkaHubModelIDByRaw(modelIDMap, settings.TranslateModelRaw)
	ocrModelID := findRikkaHubModelIDByRaw(modelIDMap, settings.OCRModel)
	compressModelID := findRikkaHubModelIDByRaw(modelIDMap, settings.CompressModelRaw)
	suggestionModelID := chatModelID // Use same as chat model

	if chatModelID == "" && len(providers) > 0 && len(providers[0].Models) > 0 {
		chatModelID = providers[0].Models[0].ID
	}

	// Convert assistants
	assistants, defaultAssistantID := convertKelivoAssistants(settings.Assistants, settings.CurrentAssistantID, modelIDMap, settings.ProviderConfigs)

	// Convert MCP servers
	mcpServers := convertKelivoMCPServers(settings.MCPServers)

	// Convert search services
	searchServices := convertKelivoSearchServices(settings.SearchServices)

	// Convert lorebooks (world books)
	lorebooks := convertKelivoWorldBooks(settings.WorldBooks)

	// Convert quick messages
	quickMessages := convertKelivoQuickPhrases(settings.QuickPhrases)

	// Convert mode injections (instruction injections)
	modeInjections := convertKelivoInstructionInjections(settings.InstructionInjections)

	// Convert TTS services
	ttsProviders := convertKelivoTTSServices(settings.TTSServices)

	dst := &models.RikkaHubSettings{
		Init:          false,
		DynamicColor:  true,
		ThemeID:       "autumn",
		DeveloperMode: false,
		DisplaySetting: models.RikkaHubDisplaySetting{
			UserAvatar: models.RikkaHubAvatar{
				Type:    "me.rerere.rikkahub.data.model.Avatar.Emoji",
				Content: "😀",
			},
			UseAppIconStyleLoadingIndicator:       true,
			ShowUserAvatar:                        true,
			ShowAssistantBubble:                   false,
			ShowModelIcon:                         true,
			ShowModelName:                         true,
			ShowDateBelowName:                     settings.DisplayShowChatListDate,
			ShowTokenUsage:                        true,
			ShowThinkingContent:                   true,
			AutoCloseThinking:                     true,
			ShowUpdates:                           true,
			ShowMessageJumper:                     true,
			MessageJumperOnLeft:                   false,
			FontSizeRatio:                         1.0,
			EnableMessageGenerationHapticEffect:   settings.DisplayHapticsOnGenerate,
			SkipCropImage:                         true,
			EnableNotificationOnMessageGeneration: true,
			EnableLiveUpdateNotification:          true,
			CodeBlockAutoWrap:                     false,
			CodeBlockAutoCollapse:                 true,
			ShowLineNumbers:                       true,
			TTSOnlyReadQuoted:                     false,
			AutoPlayTTSAfterGeneration:            false,
			PasteLongTextAsFile:                   true,
			PasteLongTextThreshold:                1000,
			SendOnEnter:                           settings.DisplayEnterToSendOnMobile,
			EnableAutoScroll:                      true,
			EnableLatexRendering:                  true,
			EnableBlurEffect:                      true,
			ChatFontFamily:                        "default",
		},
		EnableWebSearch:         settings.SearchEnabled,
		FavoriteModels:          []string{},
		ChatModelID:             chatModelID,
		TitleModelID:            titleModelID,
		ImageGenModelID:         "",
		TitlePrompt:             defaultTitlePrompt(),
		TranslateModeID:         translateModelID,
		TranslatePrompt:         defaultTranslatePrompt(),
		TranslateThinkingBudget: 0,
		SuggestionModelID:       suggestionModelID,
		SuggestionPrompt:        defaultSuggestionPrompt(),
		OCRModelID:              ocrModelID,
		OCRPrompt:               defaultOCRPrompt(),
		CompressModelID:         compressModelID,
		CompressPrompt:          defaultCompressPrompt(),
		AssistantID:             defaultAssistantID,
		Providers:               providers,
		Assistants:              assistants,
		AssistantTags:           nil,
		SearchServices:          searchServices,
		SearchServiceSelected:   0,
		MCPServers:              mcpServers,
		Lorebooks:               lorebooks,
		QuickMessages:           quickMessages,
		ModeInjections:          modeInjections,
		TTSProviders:            ttsProviders,
		SelectedTTSProviderID:   "",
		WebServerEnabled:        false,
		WebServerPort:           8080,
		WebServerJwtEnabled:     false,
		WebServerLocalhostOnly:  true,
		LaunchCount:             0,
	}

	if chatModelID != "" {
		dst.FavoriteModels = []string{chatModelID}
	}

	return dst, modelIDMap, nil
}

func convertKelivoProviders(providerConfigs map[string]models.KelivoProvider, modelIDMap map[string]string) []models.RikkaHubProvider {
	var providers []models.RikkaHubProvider

	for _, kp := range providerConfigs {
		rp := models.RikkaHubProvider{
			Type:    mapProviderType(kp.ProviderType),
			ID:      kp.ID,
			Enabled: kp.Enabled,
			Name:    kp.Name,
			APIKey:  kp.APIKey,
			BaseURL: kp.BaseURL,
			Headers: []models.RikkaHubHeader{},
			BalanceOption: models.RikkaHubBalanceOption{
				Enabled:    false,
				APIPath:    "/credits",
				ResultPath: "data.total_balance",
			},
		}

		// Build models list
		var models []models.RikkaHubModel
		for _, modelID := range kp.Models {
			override, hasOverride := kp.ModelOverrides[modelID]
			mID := uuid.New().String()
			mapKey := fmt.Sprintf("%s::%s", kp.ID, modelID)
			modelIDMap[mapKey] = mID

			rm := buildRikkaHubModel(modelID, mID, override, hasOverride)
			models = append(models, rm)
		}
		rp.Models = models
		providers = append(providers, rp)
	}

	return providers
}

func buildRikkaHubModel(modelID, id string, override models.KelivoModelOverride, hasOverride bool) models.RikkaHubModel {
	inputModalities := []string{"TEXT"}
	outputModalities := []string{"TEXT"}
	abilities := []string{"TOOL"}
	modelType := "CHAT"

	if hasOverride {
		inputModalities = mapModalities(override.Input)
		outputModalities = mapModalities(override.Output)
		abilities = mapAbilities(override.Abilities)
		modelType = strings.ToUpper(override.Type)
		if modelType == "" {
			modelType = "CHAT"
		}
	}

	return models.RikkaHubModel{
		ModelID:          modelID,
		DisplayName:      modelID,
		ID:               id,
		Type:             modelType,
		CustomHeaders:    []models.RikkaHubHeader{},
		CustomBodies:     []models.RikkaHubBody{},
		InputModalities:  inputModalities,
		OutputModalities: outputModalities,
		Abilities:        abilities,
		Tools:            []json.RawMessage{},
	}
}

func mapProviderType(t string) string {
	switch strings.ToLower(t) {
	case "openai":
		return "openai"
	case "gemini", "google":
		return "google"
	case "anthropic", "claude":
		return "anthropic"
	default:
		return "openai"
	}
}

func mapModalities(kelivo []string) []string {
	var result []string
	for _, m := range kelivo {
		switch strings.ToLower(m) {
		case "text":
			result = append(result, "TEXT")
		case "image":
			result = append(result, "IMAGE")
		case "audio":
			result = append(result, "AUDIO")
		case "video":
			result = append(result, "VIDEO")
		default:
			result = append(result, strings.ToUpper(m))
		}
	}
	if len(result) == 0 {
		return []string{"TEXT"}
	}
	return result
}

func mapAbilities(kelivo []string) []string {
	var result []string
	for _, a := range kelivo {
		switch strings.ToLower(a) {
		case "tool":
			result = append(result, "TOOL")
		case "reasoning":
			result = append(result, "REASONING")
		case "search":
			result = append(result, "SEARCH")
		default:
			result = append(result, strings.ToUpper(a))
		}
	}
	return result
}

func findRikkaHubModelID(modelIDMap map[string]string, settings *models.KelivoSettings) string {
	// Try selected_model_v1 first
	return findRikkaHubModelIDByRaw(modelIDMap, settings.SelectedModelRaw)
}

// findRikkaHubModelIDByRaw resolves a model string like "ProviderID::modelId" to a RikkaHub model UUID
func findRikkaHubModelIDByRaw(modelIDMap map[string]string, raw string) string {
	if raw == "" {
		return ""
	}

	// Try formats: "ProviderID::modelId" or "ProviderType::modelId"
	parts := strings.SplitN(raw, "::", 2)
	if len(parts) == 2 {
		key := fmt.Sprintf("%s::%s", parts[0], parts[1])
		if id, ok := modelIDMap[key]; ok {
			return id
		}
		// Try matching by model ID only
		modelIDSuffix := parts[1]
		for k, v := range modelIDMap {
			if strings.HasSuffix(k, "::"+modelIDSuffix) {
				return v
			}
		}
	}

	// Try as model ID directly
	for k, v := range modelIDMap {
		if strings.HasSuffix(k, "::"+raw) {
			return v
		}
	}

	return ""
}

func convertKelivoAssistants(
	assistants []models.KelivoAssistant,
	currentAssistantID string,
	modelIDMap map[string]string,
	providers map[string]models.KelivoProvider,
) ([]models.RikkaHubAssistant, string) {
	var result []models.RikkaHubAssistant
	defaultID := ""

	// Create a default assistant if none exist
	if len(assistants) == 0 {
		defaultAssistant := createDefaultRikkaHubAssistant("")
		result = append(result, defaultAssistant)
		defaultID = defaultAssistant.ID
		return result, defaultID
	}

	for _, ka := range assistants {
		// Resolve model ID
		modelKey := fmt.Sprintf("%s::%s", ka.ChatModelProvider, ka.ChatModelID)
		chatModelID := modelIDMap[modelKey]
		if chatModelID == "" {
			// Try looking by model ID alone
			for k, v := range modelIDMap {
				if strings.HasSuffix(k, "::"+ka.ChatModelID) {
					chatModelID = v
					break
				}
			}
		}

		// Map MCP server IDs (same IDs should work since we preserve them)
		mcpServers := ka.MCPServerIDs

		// Convert preset messages
		presetMessages := make([]models.RikkaHubPresetMessage, 0, len(ka.PresetMessages))
		for _, pm := range ka.PresetMessages {
			presetMessages = append(presetMessages, models.RikkaHubPresetMessage{
				Role:    pm.Role,
				Content: pm.Content,
			})
		}

		// Convert regex rules
		regexes := make([]models.RikkaHubRegex, 0, len(ka.RegexRules))
		for _, r := range ka.RegexRules {
			regexes = append(regexes, models.RikkaHubRegex{
				Pattern:     r.Pattern,
				Replacement: r.Replacement,
				Enabled:     r.Enabled,
			})
		}

		ra := models.RikkaHubAssistant{
			ID:                         ka.ID,
			ChatModelID:                chatModelID,
			Name:                       ka.Name,
			UseAssistantAvatar:         ka.UseAssistantAvatar,
			Tags:                       []string{},
			SystemPrompt:               ka.SystemPrompt,
			Temperature:                ka.Temperature,
			TopP:                       nil,
			ContextMessageSize:         ka.ContextMessageSize,
			StreamOutput:               ka.StreamOutput,
			EnableMemory:               ka.EnableMemory,
			UseGlobalMemory:            false,
			EnableRecentChatsReference: ka.EnableRecentChatsReference,
			MessageTemplate:            ka.MessageTemplate,
			PresetMessages:             presetMessages,
			QuickMessageIDs:            []string{},
			Regexes:                    regexes,
			ThinkingBudget:             ka.ThinkingBudget,
			MaxTokens:                  ka.MaxTokens,
			CustomHeaders:              []models.RikkaHubHeader{},
			CustomBodies:               []models.RikkaHubBody{},
			MCPServers:                 mcpServers,
			LocalTools:                 []models.RikkaHubLocalTool{},
			Background:                 nil,
			BackgroundOpacity:          1.0,
			ModeInjectionIDs:           []string{},
			LorebookIDs:                []string{},
			EnabledSkills:              []string{},
			EnableTimeReminder:         false,
		}

		if ka.TopP != 0 {
			topP := ka.TopP
			ra.TopP = &topP
		}

		// Handle avatar
		if ka.AvatarType != nil && ka.AvatarValue != nil {
			ra.Avatar = &models.RikkaHubAvatar{
				Type:    "me.rerere.rikkahub.data.model.Avatar.Emoji",
				Content: *ka.AvatarValue,
			}
		}

		result = append(result, ra)

		if ka.ID == currentAssistantID {
			defaultID = ka.ID
		}
	}

	if defaultID == "" && len(result) > 0 {
		defaultID = result[0].ID
	}

	return result, defaultID
}

func createDefaultRikkaHubAssistant(modelID string) models.RikkaHubAssistant {
	return models.RikkaHubAssistant{
		ID:                         uuid.New().String(),
		ChatModelID:                modelID,
		Name:                       "默认助手",
		UseAssistantAvatar:         false,
		Tags:                       []string{},
		SystemPrompt:               "",
		Temperature:                0.6,
		TopP:                       nil,
		ContextMessageSize:         64,
		StreamOutput:               true,
		EnableMemory:               false,
		UseGlobalMemory:            false,
		EnableRecentChatsReference: false,
		MessageTemplate:            "{{ message }}",
		PresetMessages:             []models.RikkaHubPresetMessage{},
		QuickMessageIDs:            []string{},
		Regexes:                    []models.RikkaHubRegex{},
		ThinkingBudget:             0,
		MaxTokens:                  nil,
		CustomHeaders:              []models.RikkaHubHeader{},
		CustomBodies:               []models.RikkaHubBody{},
		MCPServers:                 []string{},
		LocalTools:                 []models.RikkaHubLocalTool{},
		Background:                 nil,
		BackgroundOpacity:          1.0,
		ModeInjectionIDs:           []string{},
		LorebookIDs:                []string{},
		EnabledSkills:              []string{},
		EnableTimeReminder:         false,
	}
}

func convertKelivoMCPServers(servers []models.KelivoMCPServer) []models.RikkaHubMCPServer {
	var result []models.RikkaHubMCPServer

	for _, ks := range servers {
		transport := "streamable_http"
		if ks.Transport == "inmemory" {
			// Skip built-in servers
			continue
		} else if ks.Transport == "sse" {
			transport = "sse"
		} else if ks.Transport == "stdio" {
			transport = "stdio"
		}

		tools := make([]models.RikkaHubMCPTool, 0, len(ks.Tools))
		for _, t := range ks.Tools {
			schema := t.Schema
			rt := models.RikkaHubMCPTool{
				Enable:        t.Enabled,
				Name:          t.Name,
				Description:   t.Description,
				InputSchema:   schema,
				NeedsApproval: false,
			}
			tools = append(tools, rt)
		}

		// Convert headers map to slice
		headers := []models.RikkaHubHeader{}
		for k, v := range ks.Headers {
			headers = append(headers, models.RikkaHubHeader{Name: k, Value: v})
		}

		rs := models.RikkaHubMCPServer{
			Type: transport,
			ID:   ks.ID,
			URL:  ks.URL,
			CommonOptions: models.RikkaHubMCPCommonOptions{
				Enable:  ks.Enabled,
				Name:    ks.Name,
				Headers: headers,
				Tools:   tools,
			},
		}
		result = append(result, rs)
	}

	return result
}

func convertKelivoSearchServices(services []json.RawMessage) []models.RikkaHubSearchService {
	// Convert raw JSON search services to RikkaHub format
	var result []models.RikkaHubSearchService
	for _, s := range services {
		var svc models.RikkaHubSearchService
		if err := json.Unmarshal(s, &svc); err == nil {
			result = append(result, svc)
		}
	}
	return result
}

func convertKelivoWorldBooks(books []models.KelivoWorldBook) []models.RikkaHubLorebook {
	var result []models.RikkaHubLorebook

	for _, kb := range books {
		entries := make([]models.RikkaHubLorebookEntry, 0, len(kb.Entries))
		for _, ke := range kb.Entries {
			entries = append(entries, models.RikkaHubLorebookEntry{
				ID:             ke.ID,
				Name:           ke.Name,
				Enabled:        ke.Enabled,
				Priority:       ke.Priority,
				Position:       ke.Position,
				Content:        ke.Content,
				InjectDepth:    ke.InjectDepth,
				Role:           ke.Role,
				Keywords:       ke.Keywords,
				UseRegex:       ke.UseRegex,
				CaseSensitive:  ke.CaseSensitive,
				ScanDepth:      ke.ScanDepth,
				ConstantActive: ke.ConstantActive,
			})
		}

		result = append(result, models.RikkaHubLorebook{
			ID:          kb.ID,
			Name:        kb.Name,
			Description: kb.Description,
			Enabled:     kb.Enabled,
			Entries:     entries,
		})
	}

	return result
}

func convertKelivoQuickPhrases(phrases []models.KelivoQuickPhrase) []models.RikkaHubQuickMessage {
	var result []models.RikkaHubQuickMessage

	for _, kp := range phrases {
		result = append(result, models.RikkaHubQuickMessage{
			ID:          kp.ID,
			Title:       kp.Title,
			Content:     kp.Content,
			IsGlobal:    kp.IsGlobal,
			AssistantID: kp.AssistantID,
		})
	}

	return result
}

func convertKelivoInstructionInjections(injections []models.KelivoInstructionInjection) []models.RikkaHubModeInjection {
	var result []models.RikkaHubModeInjection

	for _, ki := range injections {
		result = append(result, models.RikkaHubModeInjection{
			ID:      ki.ID,
			Enabled: ki.Enabled,
			Content: ki.Content,
		})
	}

	return result
}

func convertKelivoTTSServices(services []json.RawMessage) []models.RikkaHubTTSProvider {
	return []models.RikkaHubTTSProvider{}
}

func convertKelivoChatsToRikkaHub(
	chats *models.KelıvoChats,
	modelIDMap map[string]string,
	settings *models.KelivoSettings,
) ([]models.RikkaHubConversation, []models.RikkaHubMessageNode) {
	var conversations []models.RikkaHubConversation
	var messageNodes []models.RikkaHubMessageNode

	// Build message lookup
	messageByID := make(map[string]models.KelivoMessage)
	for _, m := range chats.Messages {
		messageByID[m.ID] = m
	}

	for _, kc := range chats.Conversations {
		// Find assistant ID in RikkaHub
		assistantID := kc.AssistantID

		// Parse timestamps
		createAt := parseTimestamp(kc.CreatedAt)
		updateAt := parseTimestamp(kc.UpdatedAt)

		rc := models.RikkaHubConversation{
			ID:          kc.ID,
			AssistantID: assistantID,
			Title:       kc.Title,
			Nodes:       []string{},
			CreateAt:    createAt,
			UpdateAt:    updateAt,
			Suggestions: []string{},
			IsPinned:    kc.IsPinned,
		}

		// Group messages by groupId and version - create message nodes
		// In Kelivo: messages are flat with groupId+version indicating branching
		// In RikkaHub: messages are in nodes where each node is a "position" in the conversation
		nodes := buildMessageNodes(kc, messageByID, modelIDMap)
		for _, n := range nodes {
			rc.Nodes = append(rc.Nodes, n.ID)
			messageNodes = append(messageNodes, n)
		}

		conversations = append(conversations, rc)
	}

	return conversations, messageNodes
}

// buildMessageNodes converts Kelivo's flat message list with versioning to RikkaHub's node structure
func buildMessageNodes(
	conv models.KelivoConversation,
	messageByID map[string]models.KelivoMessage,
	modelIDMap map[string]string,
) []models.RikkaHubMessageNode {
	var nodes []models.RikkaHubMessageNode

	// Group messages by groupId
	type msgVersion struct {
		msg     models.KelivoMessage
		version int
	}
	groupMap := make(map[string][]msgVersion)
	groupOrder := []string{}
	seenGroups := make(map[string]bool)

	for _, msgID := range conv.MessageIDs {
		msg, ok := messageByID[msgID]
		if !ok {
			continue
		}
		gid := msg.GroupID
		if gid == "" {
			gid = msg.ID
		}
		if !seenGroups[gid] {
			seenGroups[gid] = true
			groupOrder = append(groupOrder, gid)
		}
		groupMap[gid] = append(groupMap[gid], msgVersion{msg: msg, version: msg.Version})
	}

	// Create a node for each group position
	for nodeIdx, groupID := range groupOrder {
		versions := groupMap[groupID]

		// Sort by version
		// Simple insertion sort for small slices
		for i := 1; i < len(versions); i++ {
			for j := i; j > 0 && versions[j].version < versions[j-1].version; j-- {
				versions[j], versions[j-1] = versions[j-1], versions[j]
			}
		}

		// Convert each version to a RikkaHub message
		var rMsgs []models.RikkaHubMessage
		for _, v := range versions {
			rm := convertKelivoMessageToRikkaHub(v.msg, modelIDMap)
			rMsgs = append(rMsgs, rm)
		}

		// Determine selected index (from versionSelections)
		selectedIdx := 0
		if selIdx, ok := conv.VersionSelections[groupID]; ok {
			selectedIdx = selIdx
			if selectedIdx >= len(rMsgs) {
				selectedIdx = len(rMsgs) - 1
			}
		}

		node := models.RikkaHubMessageNode{
			ID:             uuid.New().String(),
			ConversationID: conv.ID,
			NodeIndex:      nodeIdx,
			Messages:       rMsgs,
			SelectIndex:    selectedIdx,
		}
		nodes = append(nodes, node)
	}

	return nodes
}

func convertKelivoMessageToRikkaHub(msg models.KelivoMessage, modelIDMap map[string]string) models.RikkaHubMessage {
	// Convert content to parts
	parts := []models.RikkaHubMessagePart{}

	if msg.Content != "" {
		parts = append(parts, models.RikkaHubMessagePart{
			Type:     "text",
			Text:     msg.Content,
			Priority: 0,
		})
	}

	// Resolve model ID
	var modelID *string
	if msg.ModelID != nil && msg.ProviderID != nil {
		key := fmt.Sprintf("%s::%s", *msg.ProviderID, *msg.ModelID)
		if rID, ok := modelIDMap[key]; ok {
			modelID = &rID
		} else {
			// Keep original model ID as fallback
			modelID = msg.ModelID
		}
	}

	// Parse timestamps
	createdAt := msg.Timestamp
	if createdAt == "" {
		createdAt = time.Now().Format(time.RFC3339Nano)
	}

	var usage *models.RikkaHubUsage
	if msg.TotalTokens != nil && *msg.TotalTokens > 0 {
		usage = &models.RikkaHubUsage{
			TotalTokens: *msg.TotalTokens,
		}
	}

	return models.RikkaHubMessage{
		ID:          msg.ID,
		Role:        msg.Role,
		Parts:       parts,
		Annotations: []json.RawMessage{},
		CreatedAt:   createdAt,
		FinishedAt:  nil,
		ModelID:     modelID,
		Usage:       usage,
		Translation: msg.Translation,
	}
}

func parseTimestamp(ts string) int64 {
	if ts == "" {
		return time.Now().UnixMilli()
	}

	// Try various formats
	formats := []string{
		"2006-01-02T15:04:05.999",
		"2006-01-02T15:04:05",
		time.RFC3339,
		time.RFC3339Nano,
	}

	for _, format := range formats {
		t, err := time.ParseInLocation(format, ts, time.UTC)
		if err == nil {
			return t.UnixMilli()
		}
	}

	return time.Now().UnixMilli()
}

// Default prompts for RikkaHub

func defaultTitlePrompt() string {
	return `I will give you some dialogue content in the ` + "`<content>`" + ` block.
You need to summarize the conversation between user and assistant into a short title.
1. The title language should be consistent with the user's primary language
2. Do not use punctuation or other special symbols
3. Reply directly with the title
4. Summarize using {locale} language
5. The title should not exceed 10 characters

<content>
{content}
</content>`
}

func defaultTranslatePrompt() string {
	return `You are a translation expert, skilled in translating various languages, and maintaining accuracy, faithfulness, and elegance in translation.
Next, I will send you text. Please translate it into {target_lang}, and return the translation result directly, without adding any explanations or other content.

Please translate the <source_text> section:

<source_text>
{source_text}
</source_text>`
}

func defaultSuggestionPrompt() string {
	return `I will provide you with some chat content in the ` + "`<content>`" + ` block, including conversations between the User and the AI assistant.
You need to act as the **User** to reply to the assistant, generating 3~5 appropriate and contextually relevant responses to the assistant.

Rules:
1. Reply directly with suggestions, do not add any formatting, and separate suggestions with newlines, no need to add markdown list formats.
2. Use {locale} language.
3. Ensure each suggestion is valid.
4. Each suggestion should not exceed 10 characters.
5. Imitate the user's previous conversational style.
6. Act as a User, not an Assistant!

<content>
{content}
</content>`
}

func defaultOCRPrompt() string {
	return `You are an OCR assistant.

Extract all visible text from the image and also describe any non-text elements (icons, shapes, arrows, objects, symbols, or emojis).

For each element, specify:
- The exact text (for text) or a short description (for non-text).
- For document-type content, please use markdown and latex format.
- If there are objects like buildings or characters, try to identify who they are.
- Its approximate position in the image (e.g., 'top left', 'center right', 'bottom middle').
- Its spatial relationship to nearby elements (e.g., 'above', 'below', 'next to', 'on the left of').

Keep the original reading order and layout structure as much as possible.
Do not interpret or translate—only transcribe and describe what is visually present.`
}

func defaultCompressPrompt() string {
	return `You are a conversation compression assistant. Compress the following conversation into a concise summary.

Requirements:
1. Preserve key facts, decisions, and important context that would be needed to continue the conversation
2. Keep the summary in the same language as the original conversation
3. Target approximately {target_tokens} tokens
4. Output the summary directly without any explanations or meta-commentary
5. Format the summary as context information that can be used to continue the conversation
6. Use {locale} language
7. Start the output with a clear indicator that this is a summary

{additional_context}

<conversation>
{content}
</conversation>`
}
