package parser

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/converter/backup-converter/models"
)

// ParseKelivoBackup parses a Kelivo backup zip file
func ParseKelivoBackup(zipPath string) (*models.KelivoBackup, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open zip: %w", err)
	}
	defer r.Close()

	backup := &models.KelivoBackup{}
	var settingsData []byte
	var chatsData []byte

	for _, f := range r.File {
		data, err := readZipFile(f)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", f.Name, err)
		}

		switch f.Name {
		case "settings.json":
			settingsData = data
		case "chats.json":
			chatsData = data
		}
	}

	if settingsData != nil {
		settings, err := parseKelivoSettings(settingsData)
		if err != nil {
			return nil, fmt.Errorf("failed to parse settings: %w", err)
		}
		backup.Settings = settings
	}

	if chatsData != nil {
		chats, err := parseKelivoChats(chatsData)
		if err != nil {
			return nil, fmt.Errorf("failed to parse chats: %w", err)
		}
		backup.Chats = chats
	}

	return backup, nil
}

// ParseKelivoBackupFromBytes parses a Kelivo backup from bytes
func ParseKelivoBackupFromBytes(data []byte) (*models.KelivoBackup, error) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("failed to open zip: %w", err)
	}

	backup := &models.KelivoBackup{}
	var settingsData []byte
	var chatsData []byte

	for _, f := range r.File {
		fdata, err := readZipFile(f)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", f.Name, err)
		}

		switch f.Name {
		case "settings.json":
			settingsData = fdata
		case "chats.json":
			chatsData = fdata
		}
	}

	if settingsData != nil {
		settings, err := parseKelivoSettings(settingsData)
		if err != nil {
			return nil, fmt.Errorf("failed to parse settings: %w", err)
		}
		backup.Settings = settings
	}

	if chatsData != nil {
		chats, err := parseKelivoChats(chatsData)
		if err != nil {
			return nil, fmt.Errorf("failed to parse chats: %w", err)
		}
		backup.Chats = chats
	}

	return backup, nil
}

func parseKelivoSettings(data []byte) (*models.KelivoSettings, error) {
	settings := &models.KelivoSettings{}
	if err := json.Unmarshal(data, settings); err != nil {
		return nil, err
	}

	// Parse nested JSON string fields
	if settings.MCPServersRaw != "" {
		if err := json.Unmarshal([]byte(settings.MCPServersRaw), &settings.MCPServers); err != nil {
			// Try to ignore parse errors for unknown fields
			settings.MCPServers = []models.KelivoMCPServer{}
		}
	}

	if settings.ProviderConfigsRaw != "" {
		if err := json.Unmarshal([]byte(settings.ProviderConfigsRaw), &settings.ProviderConfigs); err != nil {
			settings.ProviderConfigs = map[string]models.KelivoProvider{}
		}
	}

	if settings.AssistantsRaw != "" {
		if err := json.Unmarshal([]byte(settings.AssistantsRaw), &settings.Assistants); err != nil {
			settings.Assistants = []models.KelivoAssistant{}
		}
	}

	if settings.QuickPhrasesRaw != "" {
		if err := json.Unmarshal([]byte(settings.QuickPhrasesRaw), &settings.QuickPhrases); err != nil {
			settings.QuickPhrases = []models.KelivoQuickPhrase{}
		}
	}

	if settings.WorldBooksRaw != "" {
		if err := json.Unmarshal([]byte(settings.WorldBooksRaw), &settings.WorldBooks); err != nil {
			settings.WorldBooks = []models.KelivoWorldBook{}
		}
	}

	if settings.SearchServicesRaw != "" {
		if err := json.Unmarshal([]byte(settings.SearchServicesRaw), &settings.SearchServices); err != nil {
			settings.SearchServices = []json.RawMessage{}
		}
	}

	if settings.TTSServicesRaw != "" {
		if err := json.Unmarshal([]byte(settings.TTSServicesRaw), &settings.TTSServices); err != nil {
			settings.TTSServices = []json.RawMessage{}
		}
	}

	if settings.InstructionInjectionsRaw != "" {
		if err := json.Unmarshal([]byte(settings.InstructionInjectionsRaw), &settings.InstructionInjections); err != nil {
			settings.InstructionInjections = []models.KelivoInstructionInjection{}
		}
	}

	return settings, nil
}

func parseKelivoChats(data []byte) (*models.KelıvoChats, error) {
	chats := &models.KelıvoChats{}
	if err := json.Unmarshal(data, chats); err != nil {
		return nil, err
	}
	return chats, nil
}

func readZipFile(f *zip.File) ([]byte, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return io.ReadAll(rc)
}
