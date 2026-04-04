package writer

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/converter/backup-converter/models"
)

// WriteKelivoBackup writes a Kelivo backup to a zip file
func WriteKelivoBackup(backup *models.KelivoBackup, outputPath string) error {
	data, err := MarshalKelivoBackup(backup)
	if err != nil {
		return err
	}
	return os.WriteFile(outputPath, data, 0644)
}

// MarshalKelivoBackup serializes a Kelivo backup to zip bytes
func MarshalKelivoBackup(backup *models.KelivoBackup) ([]byte, error) {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	// Write settings.json
	if backup.Settings != nil {
		settingsData, err := marshalKelivoSettings(backup.Settings)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal settings: %w", err)
		}
		if err := writeZipEntry(w, "settings.json", settingsData); err != nil {
			return nil, fmt.Errorf("failed to write settings.json: %w", err)
		}
	}

	// Write chats.json
	if backup.Chats != nil {
		chatsData, err := json.MarshalIndent(backup.Chats, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal chats: %w", err)
		}
		if err := writeZipEntry(w, "chats.json", chatsData); err != nil {
			return nil, fmt.Errorf("failed to write chats.json: %w", err)
		}
	}

	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("failed to close zip: %w", err)
	}

	return buf.Bytes(), nil
}

func marshalKelivoSettings(settings *models.KelivoSettings) ([]byte, error) {
	// Re-serialize nested JSON string fields if they were parsed
	if len(settings.MCPServers) > 0 && settings.MCPServersRaw == "" {
		data, err := json.Marshal(settings.MCPServers)
		if err != nil {
			return nil, err
		}
		settings.MCPServersRaw = string(data)
	}

	if len(settings.ProviderConfigs) > 0 && settings.ProviderConfigsRaw == "" {
		data, err := json.Marshal(settings.ProviderConfigs)
		if err != nil {
			return nil, err
		}
		settings.ProviderConfigsRaw = string(data)
	}

	if len(settings.Assistants) > 0 && settings.AssistantsRaw == "" {
		data, err := json.Marshal(settings.Assistants)
		if err != nil {
			return nil, err
		}
		settings.AssistantsRaw = string(data)
	}

	if len(settings.QuickPhrases) > 0 && settings.QuickPhrasesRaw == "" {
		data, err := json.Marshal(settings.QuickPhrases)
		if err != nil {
			return nil, err
		}
		settings.QuickPhrasesRaw = string(data)
	}

	if len(settings.WorldBooks) > 0 && settings.WorldBooksRaw == "" {
		data, err := json.Marshal(settings.WorldBooks)
		if err != nil {
			return nil, err
		}
		settings.WorldBooksRaw = string(data)
	}

	if len(settings.SearchServices) > 0 && settings.SearchServicesRaw == "" {
		data, err := json.Marshal(settings.SearchServices)
		if err != nil {
			return nil, err
		}
		settings.SearchServicesRaw = string(data)
	}

	if len(settings.InstructionInjections) > 0 && settings.InstructionInjectionsRaw == "" {
		data, err := json.Marshal(settings.InstructionInjections)
		if err != nil {
			return nil, err
		}
		settings.InstructionInjectionsRaw = string(data)
	}

	return json.MarshalIndent(settings, "", "  ")
}

func writeZipEntry(w *zip.Writer, name string, data []byte) error {
	f, err := w.Create(name)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, bytes.NewReader(data))
	return err
}
