package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"sort"
	"strings"
	"time"
)

const (
	defaultAPIBaseURL   = "https://sexyvoice.ai/api/v1/speech"
	defaultDashboardURL = "https://sexyvoice.ai"

	ModelIDGPro    = "gpro"
	ModelIDOrpheus = "orpheus"
	ModelIDGrok    = "grok"

	// Legacy constants retained so older internal files still compile.
	ModelIDArcana   = "arcana"
	ModelIDArcanaV2 = "arcanav2"
	ModelIDArcanaV3 = "arcanav3"
	ModelIDMist     = "mist"
	ModelIDMistV2   = "mistv2"
	ModelIDMistV3   = "mistv3"
)

var validModelIDs = map[string]bool{
	ModelIDGPro:    true,
	ModelIDOrpheus: true,
	ModelIDGrok:    true,
}

type Client struct {
	baseURL          string
	apiKey           string
	authHeaderPrefix string
	userAgent        string
	client           *http.Client
}

type TTSRequest struct {
	Model          string  `json:"model"`
	Input          string  `json:"input"`
	Voice          string  `json:"voice"`
	ResponseFormat string  `json:"response_format,omitempty"`
	Style          *string `json:"style,omitempty"`
	Seed           *int    `json:"seed,omitempty"`

	// Legacy JSON fields retained so older tests and unused code still compile.
	Text                     string   `json:"text,omitempty"`
	Speaker                  string   `json:"speaker,omitempty"`
	ModelID                  string   `json:"modelId,omitempty"`
	Lang                     string   `json:"lang,omitempty"`
	RepetitionPenalty        *float64 `json:"repetition_penalty,omitempty"`
	Temperature              *float64 `json:"temperature,omitempty"`
	TopP                     *float64 `json:"top_p,omitempty"`
	MaxTokens                *int     `json:"max_tokens,omitempty"`
	SamplingRate             *int     `json:"samplingRate,omitempty"`
	SpeedAlpha               *float64 `json:"speedAlpha,omitempty"`
	PauseBetweenBrackets     *bool    `json:"pauseBetweenBrackets,omitempty"`
	PhonemizeBetweenBrackets *bool    `json:"phonemizeBetweenBrackets,omitempty"`
	InlineSpeedAlpha         *string  `json:"inlineSpeedAlpha,omitempty"`
	NoTextNormalization      *bool    `json:"noTextNormalization,omitempty"`
	SaveOovs                 *bool    `json:"saveOovs,omitempty"`
}

type TTSOptions struct {
	Speaker     string
	ModelID     string
	Lang        string
	AudioFormat string
	Style       *string
	Seed        *int

	// Legacy fields retained for compile compatibility.
	RepetitionPenalty        *float64
	Temperature              *float64
	TopP                     *float64
	MaxTokens                *int
	SamplingRate             *int
	SpeedAlpha               *float64
	PauseBetweenBrackets     *bool
	PhonemizeBetweenBrackets *bool
	InlineSpeedAlpha         *string
	NoTextNormalization      *bool
	SaveOovs                 *bool
}

type speechResponse struct {
	URL string `json:"url"`
}

type Voice struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Language      string   `json:"language"`
	Model         string   `json:"model"`
	Formats       []string `json:"formats"`
	SupportsStyle bool     `json:"supports_style"`
}

type voicesResponse struct {
	Data []Voice `json:"data"`
}

type apiErrorEnvelope struct {
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

func IsValidModelID(modelID string) bool {
	return validModelIDs[modelID]
}

func ValidModelIDs() []string {
	ids := make([]string, 0, len(validModelIDs))
	for id := range validModelIDs {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func IsArcanaModel(modelID string) bool {
	return modelID == ModelIDArcana || modelID == ModelIDArcanaV2 || modelID == ModelIDArcanaV3
}

func IsMistModel(modelID string) bool {
	return modelID == ModelIDMist || modelID == ModelIDMistV2 || modelID == ModelIDMistV3
}

func IsValidLang(_ string, _ string) bool {
	return true
}

func ValidLangsForModel(_ string) []string {
	return nil
}

func ValidateModelParams(opts *TTSOptions) error {
	if opts == nil {
		return nil
	}
	if opts.ModelID == "" {
		return fmt.Errorf("model is required")
	}
	if !IsValidModelID(opts.ModelID) {
		return fmt.Errorf("invalid model: %s (valid options: %s)", opts.ModelID, strings.Join(ValidModelIDs(), ", "))
	}
	if format := normalizeResponseFormat(opts.AudioFormat); format != "" && format != "wav" && format != "mp3" {
		return fmt.Errorf("unsupported format: %s (supported: wav, mp3)", opts.AudioFormat)
	}
	return nil
}

func GetAudioFormat(modelID string) string {
	if modelID == ModelIDGPro {
		return "audio/wav"
	}
	return "audio/mp3"
}

func GetAPIURL() string {
	if value := os.Getenv("SEXYVOICE_API_URL"); value != "" {
		return value
	}
	return defaultAPIBaseURL
}

func GetDashboardURL() string {
	if value := os.Getenv("SEXYVOICE_DASHBOARD_URL"); value != "" {
		return value
	}
	return defaultDashboardURL
}

type ClientOptions struct {
	APIKey           string
	APIURL           string
	AuthHeaderPrefix string
	Version          string
	Timeout          time.Duration
}

func NewClient(opts ClientOptions) *Client {
	client := &http.Client{}
	if opts.Timeout > 0 {
		client.Timeout = opts.Timeout
	}

	baseURL := opts.APIURL
	if baseURL == "" {
		baseURL = GetAPIURL()
	}

	authPrefix := opts.AuthHeaderPrefix
	if authPrefix == "" {
		if envPrefix := os.Getenv("SEXYVOICE_AUTH_HEADER_PREFIX"); envPrefix != "" {
			authPrefix = envPrefix
		} else {
			authPrefix = "Bearer"
		}
	}

	return &Client{
		baseURL:          baseURL,
		apiKey:           opts.APIKey,
		authHeaderPrefix: authPrefix,
		userAgent:        UserAgent(opts.Version),
		client:           client,
	}
}

func (c *Client) TTS(text string, opts *TTSOptions) ([]byte, error) {
	result, err := c.TTSStream(text, opts)
	if err != nil {
		return nil, err
	}
	defer result.Body.Close()

	audio, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read downloaded audio: %w", err)
	}
	return audio, nil
}

type TTSStreamResult struct {
	Body        io.ReadCloser
	ContentType string
	TTFB        time.Duration
}

func (c *Client) TTSStream(text string, opts *TTSOptions) (*TTSStreamResult, error) {
	if opts == nil || strings.TrimSpace(opts.Speaker) == "" {
		return nil, fmt.Errorf("voice is required")
	}
	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("input text is required")
	}
	if err := ValidateModelParams(opts); err != nil {
		return nil, err
	}

	reqBody, err := json.Marshal(TTSRequest{
		Model:          opts.ModelID,
		Input:          text,
		Voice:          opts.Speaker,
		ResponseFormat: normalizeResponseFormat(opts.AudioFormat),
		Style:          opts.Style,
		Seed:           opts.Seed,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	if c.apiKey != "" && c.authHeaderPrefix != "" {
		req.Header.Set("Authorization", fmt.Sprintf("%s %s", c.authHeaderPrefix, c.apiKey))
	}

	start := time.Now()
	resp, err := c.client.Do(req)
	ttfb := time.Since(start)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read API response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(apiErrorFromResponse(resp.StatusCode, respData))
	}

	var speechResp speechResponse
	if err := json.Unmarshal(respData, &speechResp); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}
	if speechResp.URL == "" {
		return nil, fmt.Errorf("API response did not include an audio URL")
	}

	downloadReq, err := http.NewRequest(http.MethodGet, speechResp.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create audio download request: %w", err)
	}
	downloadReq.Header.Set("User-Agent", c.userAgent)

	downloadResp, err := c.client.Do(downloadReq)
	if err != nil {
		return nil, fmt.Errorf("audio download failed: %w", err)
	}
	defer downloadResp.Body.Close()

	if downloadResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(downloadResp.Body)
		return nil, fmt.Errorf("audio download failed with status %d: %s", downloadResp.StatusCode, strings.TrimSpace(string(body)))
	}

	audioData, err := io.ReadAll(downloadResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read downloaded audio: %w", err)
	}

	contentType := downloadResp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = inferContentTypeFromURL(speechResp.URL, opts)
	}

	return &TTSStreamResult{
		Body:        io.NopCloser(bytes.NewReader(audioData)),
		ContentType: contentType,
		TTFB:        ttfb,
	}, nil
}

func (c *Client) ValidateAPIKey() error {
	validateURL := strings.TrimSuffix(strings.TrimSuffix(c.baseURL, "/"), "/speech") + "/models"

	req, err := http.NewRequest(http.MethodGet, validateURL, nil)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	if c.apiKey != "" && c.authHeaderPrefix != "" {
		req.Header.Set("Authorization", fmt.Sprintf("%s %s", c.authHeaderPrefix, c.apiKey))
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("authentication failed: invalid API key")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(apiErrorFromResponse(resp.StatusCode, body))
	}
	return nil
}

func (c *Client) ListVoices() ([]Voice, error) {
	voicesURL := strings.TrimSuffix(strings.TrimSuffix(c.baseURL, "/"), "/speech") + "/voices"

	req, err := http.NewRequest(http.MethodGet, voicesURL, nil)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	if c.apiKey != "" && c.authHeaderPrefix != "" {
		req.Header.Set("Authorization", fmt.Sprintf("%s %s", c.authHeaderPrefix, c.apiKey))
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read API response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(apiErrorFromResponse(resp.StatusCode, body))
	}

	var payload voicesResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}
	return payload.Data, nil
}

func EffectiveOpts(opts *TTSOptions) (speaker, modelId, lang string) {
	speaker = opts.Speaker
	modelId = opts.ModelID
	lang = opts.Lang
	return
}

func normalizeResponseFormat(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "", "audio/wav":
		if strings.TrimSpace(strings.ToLower(value)) == "audio/wav" {
			return "wav"
		}
		return ""
	case "wav":
		return "wav"
	case "audio/mp3", "audio/mpeg":
		return "mp3"
	case "mp3":
		return "mp3"
	default:
		return strings.TrimSpace(strings.ToLower(value))
	}
}

func inferContentTypeFromURL(rawURL string, opts *TTSOptions) string {
	parsed, err := url.Parse(rawURL)
	if err == nil {
		switch strings.ToLower(path.Ext(parsed.Path)) {
		case ".wav":
			return "audio/wav"
		case ".mp3":
			return "audio/mp3"
		}
	}

	if format := normalizeResponseFormat(opts.AudioFormat); format == "wav" {
		return "audio/wav"
	} else if format == "mp3" {
		return "audio/mp3"
	}
	return GetAudioFormat(opts.ModelID)
}

func apiErrorFromResponse(statusCode int, body []byte) string {
	var envelope apiErrorEnvelope
	if err := json.Unmarshal(body, &envelope); err == nil && envelope.Error.Message != "" {
		if statusCode == http.StatusUnauthorized {
			return "authentication failed: invalid API key"
		}
		return envelope.Error.Message
	}

	switch statusCode {
	case http.StatusUnauthorized:
		return "authentication failed: invalid API key"
	case http.StatusTooManyRequests:
		return "rate limited: too many requests"
	default:
		trimmed := strings.TrimSpace(string(body))
		if trimmed == "" {
			return fmt.Sprintf("API error %d", statusCode)
		}
		return fmt.Sprintf("API error %d: %s", statusCode, trimmed)
	}
}
