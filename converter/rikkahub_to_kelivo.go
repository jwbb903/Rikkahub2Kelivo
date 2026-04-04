package converter

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/converter/backup-converter/models"
)

// ConvertRikkaHubToKelivo converts a RikkaHub backup to a Kelivo backup
func ConvertRikkaHubToKelivo(src *models.RikkaHubBackup) (*models.KelivoBackup, error) {
	if src == nil {
		return nil, fmt.Errorf("source backup is nil")
	}

	dst := &models.KelivoBackup{}

	// Build model ID mapping: RikkaHub model UUID -> "ProviderID::modelId"
	modelIDMap := make(map[string]string)      // RikkaHub model UUID -> KelivoProvider::modelId
	modelToProvider := make(map[string]string) // RikkaHub model UUID -> KelivoProviderID

	settings, err := convertRikkaHubSettingsToKelivo(src.Settings, modelIDMap, modelToProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to convert settings: %w", err)
	}
	dst.Settings = settings

	// Convert conversations from SQLite data
	if len(src.Conversations) > 0 {
		chats := convertRikkaHubChatsToKelivo(src.Conversations, src.MessageNodes, modelIDMap, modelToProvider)
		dst.Chats = chats
	} else {
		dst.Chats = &models.KelıvoChats{
			Version:       1,
			Conversations: []models.KelivoConversation{},
			Messages:      []models.KelivoMessage{},
		}
	}

	return dst, nil
}

func convertRikkaHubSettingsToKelivo(
	settings *models.RikkaHubSettings,
	modelIDMap map[string]string,
	modelToProvider map[string]string,
) (*models.KelivoSettings, error) {
	if settings == nil {
		return nil, fmt.Errorf("settings is nil")
	}

	// Convert providers
	providerConfigs, err := convertRikkaHubProviders(settings.Providers, modelIDMap, modelToProvider)
	if err != nil {
		return nil, err
	}

	// Find selected model
	selectedModel := resolveKelivoModelRef(settings.ChatModelID, modelIDMap, modelToProvider)
	titleModel := resolveKelivoModelRef(settings.TitleModelID, modelIDMap, modelToProvider)
	translateModel := resolveKelivoModelRef(settings.TranslateModeID, modelIDMap, modelToProvider)
	ocrModel := resolveKelivoModelRef(settings.OCRModelID, modelIDMap, modelToProvider)
	compressModel := resolveKelivoModelRef(settings.CompressModelID, modelIDMap, modelToProvider)
	summaryModel := compressModel

	// Convert assistants
	assistants := convertRikkaHubAssistants(settings.Assistants, modelIDMap, modelToProvider)

	// Convert MCP servers
	mcpServers := convertRikkaHubMCPServers(settings.MCPServers)

	// Convert world books
	worldBooks := convertRikkaHubLorebooks(settings.Lorebooks)

	// Convert quick phrases
	quickPhrases := convertRikkaHubQuickMessages(settings.QuickMessages)

	// Convert instruction injections
	instructionInjections := convertRikkaHubModeInjections(settings.ModeInjections)

	// Convert search services
	searchServices := convertRikkaHubSearchServices(settings.SearchServices)

	// Serialize sub-fields as JSON strings (Kelivo's format)
	mcpServersJSON, _ := json.Marshal(mcpServers)
	providerConfigsJSON, _ := json.Marshal(providerConfigs)
	assistantsJSON, _ := json.Marshal(assistants)
	quickPhrasesJSON, _ := json.Marshal(quickPhrases)
	worldBooksJSON, _ := json.Marshal(worldBooks)
	searchServicesJSON, _ := json.Marshal(searchServices)
	instructionInjectionsJSON, _ := json.Marshal(instructionInjections)

	// Find current assistant ID
	currentAssistantID := settings.AssistantID
	if currentAssistantID == "" && len(assistants) > 0 {
		currentAssistantID = assistants[0].ID
	}

	dst := &models.KelivoSettings{
		DisplayShowChatListDate:      settings.DisplaySetting.ShowDateBelowName,
		DisplayEnterToSendOnMobile:   settings.DisplaySetting.SendOnEnter,
		MCPRequestTimeoutMs:          30000,
		OCRModel:                     ocrModel,
		DisplayUsePureBackground:     false,
		MCPServersRaw:                string(mcpServersJSON),
		ProviderConfigsRaw:           string(providerConfigsJSON),
		ThinkingBudget:               0,
		TitleModelRaw:                titleModel,
		DisplayHapticsOnGenerate:     settings.DisplaySetting.EnableMessageGenerationHapticEffect,
		CurrentAssistantID:           currentAssistantID,
		InstructionInjectionsRaw:     string(instructionInjectionsJSON),
		TranslateModelRaw:            translateModel,
		AppLocale:                    "",
		DisplayDesktopShowTray:       false,
		DisplayDesktopMinimizeToTray: false,
		AssistantsRaw:                string(assistantsJSON),
		SearchEnabled:                settings.EnableWebSearch,
		MigrationsVersion:            1,
		QuickPhrasesRaw:              string(quickPhrasesJSON),
		WorldBooksRaw:                string(worldBooksJSON),
		SearchServicesRaw:            string(searchServicesJSON),
		SearchSelected:               0,
		SummaryModelRaw:              summaryModel,
		CompressModelRaw:             compressModel,
		SelectedModelRaw:             selectedModel,

		// Parsed fields
		MCPServers:            mcpServers,
		ProviderConfigs:       providerConfigs,
		Assistants:            assistants,
		QuickPhrases:          quickPhrases,
		WorldBooks:            worldBooks,
		InstructionInjections: instructionInjections,
	}

	_ = titleModel
	_ = translateModel
	_ = summaryModel

	return dst, nil
}

func convertRikkaHubProviders(
	providers []models.RikkaHubProvider,
	modelIDMap map[string]string,
	modelToProvider map[string]string,
) (map[string]models.KelivoProvider, error) {
	result := make(map[string]models.KelivoProvider)

	for _, rp := range providers {
		var modelIDs []string
		modelOverrides := make(map[string]models.KelivoModelOverride)

		for _, rm := range rp.Models {
			modelIDs = append(modelIDs, rm.ModelID)

			// Store mapping
			mapKey := fmt.Sprintf("%s::%s", rp.ID, rm.ModelID)
			modelIDMap[rm.ID] = mapKey
			modelToProvider[rm.ID] = rp.ID

			// Build model override
			override := models.KelivoModelOverride{
				Type:      strings.ToLower(rm.Type),
				Input:     mapModalitiesDown(rm.InputModalities),
				Output:    mapModalitiesDown(rm.OutputModalities),
				Abilities: mapAbilitiesDown(rm.Abilities),
			}
			modelOverrides[rm.ModelID] = override
		}

		kp := models.KelivoProvider{
			ID:             rp.ID,
			Enabled:        rp.Enabled,
			Name:           rp.Name,
			APIKey:         rp.APIKey,
			BaseURL:        rp.BaseURL,
			ProviderType:   mapProviderTypeDown(rp.Type),
			Models:         modelIDs,
			ModelOverrides: modelOverrides,
			ProxyEnabled:   false,
			ProxyPort:      "8080",
			APIKeys:        []string{},
			KeyManagement: models.KelivoKeyManagement{
				Strategy:                   "roundRobin",
				MaxFailuresBeforeDisable:   3,
				FailureRecoveryTimeMinutes: 5,
				EnableAutoRecovery:         true,
			},
		}

		// Convert headers to custom fields if any
		// (RikkaHub headers -> Kelivo doesn't have direct equivalent)

		result[rp.ID] = kp
	}

	return result, nil
}

func mapProviderTypeDown(t string) string {
	switch t {
	case "google":
		return "gemini"
	case "anthropic":
		return "anthropic"
	default:
		return "openai"
	}
}

func mapModalitiesDown(modalities []string) []string {
	var result []string
	for _, m := range modalities {
		result = append(result, strings.ToLower(m))
	}
	if len(result) == 0 {
		return []string{"text"}
	}
	return result
}

func mapAbilitiesDown(abilities []string) []string {
	var result []string
	for _, a := range abilities {
		result = append(result, strings.ToLower(a))
	}
	return result
}

// resolveKelivoModelRef resolves a RikkaHub model UUID to a Kelivo model reference string
func resolveKelivoModelRef(modelUUID string, modelIDMap map[string]string, modelToProvider map[string]string) string {
	if modelUUID == "" {
		return ""
	}
	if ref, ok := modelIDMap[modelUUID]; ok {
		// ref is "providerID::modelId"
		parts := strings.SplitN(ref, "::", 2)
		if len(parts) == 2 {
			return fmt.Sprintf("%s::%s", parts[0], parts[1])
		}
	}
	return modelUUID
}

func convertRikkaHubAssistants(
	assistants []models.RikkaHubAssistant,
	modelIDMap map[string]string,
	modelToProvider map[string]string,
) []models.KelivoAssistant {
	var result []models.KelivoAssistant

	for _, ra := range assistants {
		// Resolve model reference
		chatModelProvider := ""
		chatModelID := ra.ChatModelID

		if ref, ok := modelIDMap[ra.ChatModelID]; ok {
			parts := strings.SplitN(ref, "::", 2)
			if len(parts) == 2 {
				chatModelProvider = parts[0]
				chatModelID = parts[1]
			}
		}

		// Convert preset messages
		presetMessages := make([]models.KelivoPresetMessage, 0, len(ra.PresetMessages))
		for _, pm := range ra.PresetMessages {
			presetMessages = append(presetMessages, models.KelivoPresetMessage{
				Role:    pm.Role,
				Content: pm.Content,
			})
		}

		// Convert regexes
		regexRules := make([]models.KelivoRegexRule, 0, len(ra.Regexes))
		for _, r := range ra.Regexes {
			regexRules = append(regexRules, models.KelivoRegexRule{
				Pattern:     r.Pattern,
				Replacement: r.Replacement,
				Enabled:     r.Enabled,
			})
		}

		topP := 1.0
		if ra.TopP != nil {
			topP = *ra.TopP
		}

		ka := models.KelivoAssistant{
			ID:                         ra.ID,
			Name:                       ra.Name,
			UseAssistantAvatar:         ra.UseAssistantAvatar,
			ChatModelProvider:          chatModelProvider,
			ChatModelID:                chatModelID,
			Temperature:                ra.Temperature,
			TopP:                       topP,
			ContextMessageSize:         ra.ContextMessageSize,
			LimitContextMessages:       true,
			StreamOutput:               ra.StreamOutput,
			ThinkingBudget:             ra.ThinkingBudget,
			MaxTokens:                  ra.MaxTokens,
			SystemPrompt:               ra.SystemPrompt,
			MessageTemplate:            ra.MessageTemplate,
			MCPServerIDs:               ra.MCPServers,
			Deletable:                  true,
			CustomHeaders:              []models.KelivoCustomHeader{},
			CustomBody:                 []models.KelivoCustomBody{},
			EnableMemory:               ra.EnableMemory,
			EnableRecentChatsReference: ra.EnableRecentChatsReference,
			PresetMessages:             presetMessages,
			RegexRules:                 regexRules,
		}

		// Handle avatar
		if ra.Avatar != nil {
			avatarType := "emoji"
			avatarValue := ra.Avatar.Content
			ka.AvatarType = &avatarType
			ka.AvatarValue = &avatarValue
		}

		result = append(result, ka)
	}

	return result
}

func convertRikkaHubMCPServers(servers []models.RikkaHubMCPServer) []models.KelivoMCPServer {
	var result []models.KelivoMCPServer

	for _, rs := range servers {
		transport := "http"
		if rs.Type == "sse" {
			transport = "sse"
		} else if rs.Type == "stdio" {
			transport = "stdio"
		}

		tools := make([]models.KelivoMCPTool, 0, len(rs.CommonOptions.Tools))
		for _, rt := range rs.CommonOptions.Tools {
			tools = append(tools, models.KelivoMCPTool{
				Enabled:     rt.Enable,
				Name:        rt.Name,
				Description: rt.Description,
				Schema:      rt.InputSchema,
			})
		}

		// Convert headers to map
		headers := make(map[string]string)
		for _, h := range rs.CommonOptions.Headers {
			if h.Name != "" {
				headers[h.Name] = h.Value
			}
		}

		ks := models.KelivoMCPServer{
			ID:        rs.ID,
			Enabled:   rs.CommonOptions.Enable,
			Name:      rs.CommonOptions.Name,
			Transport: transport,
			URL:       rs.URL,
			Tools:     tools,
			Headers:   headers,
		}
		result = append(result, ks)
	}

	return result
}

func convertRikkaHubLorebooks(lorebooks []models.RikkaHubLorebook) []models.KelivoWorldBook {
	var result []models.KelivoWorldBook

	for _, rl := range lorebooks {
		entries := make([]models.KelivoWorldBookEntry, 0, len(rl.Entries))
		for _, re := range rl.Entries {
			entries = append(entries, models.KelivoWorldBookEntry{
				ID:             re.ID,
				Name:           re.Name,
				Enabled:        re.Enabled,
				Priority:       re.Priority,
				Position:       re.Position,
				Content:        re.Content,
				InjectDepth:    re.InjectDepth,
				Role:           re.Role,
				Keywords:       re.Keywords,
				UseRegex:       re.UseRegex,
				CaseSensitive:  re.CaseSensitive,
				ScanDepth:      re.ScanDepth,
				ConstantActive: re.ConstantActive,
			})
		}

		result = append(result, models.KelivoWorldBook{
			ID:          rl.ID,
			Name:        rl.Name,
			Description: rl.Description,
			Enabled:     rl.Enabled,
			Entries:     entries,
		})
	}

	return result
}

func convertRikkaHubQuickMessages(msgs []models.RikkaHubQuickMessage) []models.KelivoQuickPhrase {
	var result []models.KelivoQuickPhrase

	for _, rm := range msgs {
		result = append(result, models.KelivoQuickPhrase{
			ID:          rm.ID,
			Title:       rm.Title,
			Content:     rm.Content,
			IsGlobal:    rm.IsGlobal,
			AssistantID: rm.AssistantID,
		})
	}

	return result
}

func convertRikkaHubModeInjections(injections []models.RikkaHubModeInjection) []models.KelivoInstructionInjection {
	var result []models.KelivoInstructionInjection

	for _, ri := range injections {
		result = append(result, models.KelivoInstructionInjection{
			ID:      ri.ID,
			Enabled: ri.Enabled,
			Content: ri.Content,
		})
	}

	return result
}

func convertRikkaHubSearchServices(services []models.RikkaHubSearchService) []interface{} {
	if len(services) == 0 {
		return []interface{}{}
	}
	var result []interface{}
	for _, s := range services {
		result = append(result, s)
	}
	return result
}

func convertRikkaHubChatsToKelivo(
	conversations []models.RikkaHubConversation,
	messageNodes []models.RikkaHubMessageNode,
	modelIDMap map[string]string,
	modelToProvider map[string]string,
) *models.KelıvoChats {
	chats := &models.KelıvoChats{
		Version:       1,
		Conversations: []models.KelivoConversation{},
		Messages:      []models.KelivoMessage{},
	}

	// Build node lookup by conversation ID
	nodesByConv := make(map[string][]models.RikkaHubMessageNode)
	for _, n := range messageNodes {
		nodesByConv[n.ConversationID] = append(nodesByConv[n.ConversationID], n)
	}

	// Sort nodes by index
	for convID := range nodesByConv {
		nodes := nodesByConv[convID]
		for i := 1; i < len(nodes); i++ {
			for j := i; j > 0 && nodes[j].NodeIndex < nodes[j-1].NodeIndex; j-- {
				nodes[j], nodes[j-1] = nodes[j-1], nodes[j]
			}
		}
		nodesByConv[convID] = nodes
	}

	for _, rc := range conversations {
		nodes := nodesByConv[rc.ID]

		// Convert nodes to flat messages with group IDs
		var messageIDs []string
		versionSelections := make(map[string]int)
		var allMessages []models.KelivoMessage

		for _, node := range nodes {
			// Each node represents a "position" in the conversation
			// Messages in a node are versions (branches)
			groupID := uuid.New().String()

			for vIdx, rmsg := range node.Messages {
				km := convertRikkaHubMessageToKelivo(rmsg, rc.ID, groupID, vIdx, modelIDMap, modelToProvider)
				allMessages = append(allMessages, km)
				messageIDs = append(messageIDs, km.ID)
			}

			// The selected version
			if node.SelectIndex > 0 && node.SelectIndex < len(node.Messages) {
				versionSelections[groupID] = node.SelectIndex
			}
		}

		// Dedup message IDs (only add first occurrence per group)
		seenGroups := make(map[string]bool)
		var dedupIDs []string
		for _, km := range allMessages {
			if !seenGroups[km.GroupID] {
				seenGroups[km.GroupID] = true
				// Add only the first message in each group (version 0 = group representative)
				if km.Version == 0 {
					dedupIDs = append(dedupIDs, km.ID)
				}
			}
		}
		_ = messageIDs

		// Use the first message from each group as representative
		messageIDs = dedupIDs

		createdAt := millisecondsToTimestamp(rc.CreateAt)
		updatedAt := millisecondsToTimestamp(rc.UpdateAt)

		kc := models.KelivoConversation{
			ID:                         rc.ID,
			Title:                      rc.Title,
			CreatedAt:                  createdAt,
			UpdatedAt:                  updatedAt,
			MessageIDs:                 messageIDs,
			IsPinned:                   rc.IsPinned,
			MCPServerIDs:               []string{},
			AssistantID:                rc.AssistantID,
			TruncateIndex:              -1,
			VersionSelections:          versionSelections,
			LastSummarizedMessageCount: 0,
		}

		chats.Conversations = append(chats.Conversations, kc)
		chats.Messages = append(chats.Messages, allMessages...)
	}

	return chats
}

func convertRikkaHubMessageToKelivo(
	msg models.RikkaHubMessage,
	convID string,
	groupID string,
	version int,
	modelIDMap map[string]string,
	modelToProvider map[string]string,
) models.KelivoMessage {
	// Build content from parts
	var contentParts []string
	for _, part := range msg.Parts {
		if part.Type == "text" && part.Text != "" {
			contentParts = append(contentParts, part.Text)
		}
	}
	content := strings.Join(contentParts, "\n")

	// Resolve model and provider IDs
	var modelID *string
	var providerID *string

	if msg.ModelID != nil {
		if ref, ok := modelIDMap[*msg.ModelID]; ok {
			parts := strings.SplitN(ref, "::", 2)
			if len(parts) == 2 {
				pID := parts[0]
				mID := parts[1]
				providerID = &pID
				modelID = &mID
			}
		} else {
			modelID = msg.ModelID
		}
	}

	// Parse timestamps
	timestamp := msg.CreatedAt
	if timestamp == "" {
		timestamp = time.Now().Format("2006-01-02T15:04:05.000")
	}

	var totalTokens *int
	if msg.Usage != nil {
		t := msg.Usage.TotalTokens
		totalTokens = &t
	}

	return models.KelivoMessage{
		ID:             msg.ID,
		Role:           msg.Role,
		Content:        content,
		Timestamp:      timestamp,
		ModelID:        modelID,
		ProviderID:     providerID,
		TotalTokens:    totalTokens,
		ConversationID: convID,
		IsStreaming:    false,
		GroupID:        groupID,
		Version:        version,
		Translation:    msg.Translation,
	}
}

func millisecondsToTimestamp(ms int64) string {
	t := time.UnixMilli(ms).UTC()
	return t.Format("2006-01-02T15:04:05.000")
}
