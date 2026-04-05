package converter

// BuiltinProviderMapping defines the mapping between Kelivo and RikkaHub built-in providers.
// Both apps ship with preset providers that share the same API endpoints.
// During conversion, we match by baseUrl and only transfer the apiKey + user-added models,
// rather than copying the entire provider configuration.
type BuiltinProviderMapping struct {
	KelivoID   string // Kelivo provider key (e.g. "OpenAI", "SiliconFlow", "Gemini")
	KelivoType string // Kelivo providerType: "openai", "google", "claude"
	KelivoURL  string // Kelivo default baseUrl

	RikkaHubID   string // RikkaHub provider UUID
	RikkaHubName string // RikkaHub display name
	RikkaHubType string // RikkaHub provider type: "openai", "google", "anthropic"
}

// builtinProviderMappings lists all providers that exist in both apps.
// Matched by (baseUrl, providerType) in K→R direction, and by (ID) in R→K direction.
var builtinProviderMappings = []BuiltinProviderMapping{
	{
		KelivoID: "OpenAI", KelivoType: "openai", KelivoURL: "https://api.openai.com/v1",
		RikkaHubID: "1eeea727-9ee5-4cae-93e6-6fb01a4d051e", RikkaHubName: "OpenAI", RikkaHubType: "openai",
	},
	{
		KelivoID: "SiliconFlow", KelivoType: "openai", KelivoURL: "https://api.siliconflow.cn/v1",
		RikkaHubID: "56a94d29-c88b-41c5-8e09-38a7612d6cf8", RikkaHubName: "硅基流动", RikkaHubType: "openai",
	},
	{
		KelivoID: "Gemini", KelivoType: "google", KelivoURL: "https://generativelanguage.googleapis.com/v1beta",
		RikkaHubID: "6ab18148-c138-4394-a46f-1cd8c8ceaa6d", RikkaHubName: "Gemini", RikkaHubType: "google",
	},
	{
		KelivoID: "OpenRouter", KelivoType: "openai", KelivoURL: "https://openrouter.ai/api/v1",
		RikkaHubID: "d5734028-d39b-4d41-9841-fd648d65440e", RikkaHubName: "OpenRouter", RikkaHubType: "openai",
	},
	{
		KelivoID: "DeepSeek", KelivoType: "openai", KelivoURL: "https://api.deepseek.com/v1",
		RikkaHubID: "f099ad5b-ef03-446d-8e78-7e36787f780b", RikkaHubName: "DeepSeek", RikkaHubType: "openai",
	},
	{
		KelivoID: "AIhubmix", KelivoType: "openai", KelivoURL: "https://aihubmix.com/v1",
		RikkaHubID: "1b1395ed-b702-4aeb-8bc1-b681c4456953", RikkaHubName: "AiHubMix", RikkaHubType: "openai",
	},
	{
		KelivoID: "Aliyun", KelivoType: "openai", KelivoURL: "https://dashscope.aliyuncs.com/compatible-mode/v1",
		RikkaHubID: "f76cae46-069a-4334-ab8e-224e4979e58c", RikkaHubName: "阿里云百炼", RikkaHubType: "openai",
	},
	{
		KelivoID: "Zhipu AI", KelivoType: "openai", KelivoURL: "https://open.bigmodel.cn/api/paas/v4",
		RikkaHubID: "3bc40dc1-b11a-46fa-863b-6306971223be", RikkaHubName: "智谱AI开放平台", RikkaHubType: "openai",
	},
	{
		KelivoID: "Grok", KelivoType: "openai", KelivoURL: "https://api.x.ai/v1",
		RikkaHubID: "ff3cde7e-0f65-43d7-8fb2-6475c99f5990", RikkaHubName: "xAI", RikkaHubType: "openai",
	},
	{
		KelivoID: "ByteDance", KelivoType: "openai", KelivoURL: "https://ark.cn-beijing.volces.com/api/v3",
		RikkaHubID: "3dfd6f9b-f9d9-417f-80c1-ff8d77184191", RikkaHubName: "火山引擎", RikkaHubType: "openai",
	},
}

// findBuiltinByKelivo matches a Kelivo provider to a RikkaHub built-in provider
// by comparing base URL and provider type. Returns nil if no match.
func findBuiltinByKelivo(baseUrl string, providerType string) *BuiltinProviderMapping {
	normalizedURL := normalizeBaseUrl(baseUrl)
	for i := range builtinProviderMappings {
		if normalizeBaseUrl(builtinProviderMappings[i].KelivoURL) == normalizedURL &&
			equalProviderType(builtinProviderMappings[i].KelivoType, providerType) {
			return &builtinProviderMappings[i]
		}
	}
	return nil
}

// findBuiltinByRikkaHubID matches a RikkaHub provider ID to a Kelivo built-in provider.
// Returns nil if no match (i.e. it's a custom provider).
func findBuiltinByRikkaHubID(rikkahubID string) *BuiltinProviderMapping {
	for i := range builtinProviderMappings {
		if builtinProviderMappings[i].RikkaHubID == rikkahubID {
			return &builtinProviderMappings[i]
		}
	}
	return nil
}

// normalizeBaseUrl strips trailing slashes and normalizes a base URL for comparison
func normalizeBaseUrl(url string) string {
	result := url
	for len(result) > 0 && result[len(result)-1] == '/' {
		result = result[:len(result)-1]
	}
	return result
}

// equalProviderType compares provider types across the two apps
func equalProviderType(a, b string) bool {
	aNorm := normalizeProviderType(a)
	bNorm := normalizeProviderType(b)
	return aNorm == bNorm
}

func normalizeProviderType(t string) string {
	switch t {
	case "openai":
		return "openai"
	case "google", "gemini":
		return "google"
	case "claude", "anthropic":
		return "anthropic"
	default:
		return t
	}
}
