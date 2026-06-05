package provider

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/textproto"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"syntheticvision/internal/config"
)

// OpenAIProvider calls an OpenAI-compatible images/generations endpoint.
type OpenAIProvider struct {
	baseURL string
	apiKey  string
	model   string
	client  *http.Client
}

// NewOpenAIProvider builds an OpenAIProvider from config.
func NewOpenAIProvider(cfg config.Config) *OpenAIProvider {
	base := strings.TrimRight(cfg.OpenAIBaseURL, "/")
	return &OpenAIProvider{
		baseURL: base,
		apiKey:  cfg.OpenAIAPIKey,
		model:   cfg.ImageModel,
		client:  &http.Client{Timeout: 120 * time.Second},
	}
}

// Name implements Provider.
func (o *OpenAIProvider) Name() string { return "openai" }

type openAIRequest struct {
	Model          string `json:"model"`
	Prompt         string `json:"prompt"`
	Size           string `json:"size"`
	N              int    `json:"n"`
	ResponseFormat string `json:"response_format"`
}

type openAIResponse struct {
	Data []struct {
		B64JSON string `json:"b64_json"`
		URL     string `json:"url"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// Generate posts to {BaseURL}/v1/images/generations for text mode or
// {BaseURL}/v1/images/edits when a source image is present, then decodes the
// result. It prefers b64_json and falls back to downloading the url field. If
// the decoded bytes are not PNG, they are stored as-is with a detected MIME
// type.
func (o *OpenAIProvider) Generate(ctx context.Context, req GenerateRequest) (*GenerateResult, error) {
	if req.SourceImage != nil {
		return o.edit(ctx, req)
	}
	return o.generate(ctx, req)
}

// generate posts to {BaseURL}/v1/images/generations and decodes the result.
// It prefers b64_json and falls back to downloading the url field. If the
// decoded bytes are not PNG, they are stored as-is with a detected MIME type.
func (o *OpenAIProvider) generate(ctx context.Context, req GenerateRequest) (*GenerateResult, error) {
	if o.baseURL == "" {
		return nil, errors.New("openai base url not configured")
	}
	size := fmt.Sprintf("%dx%d", req.Width, req.Height)
	prompt := buildPrompt(req)
	body := openAIRequest{
		Model:          o.model,
		Prompt:         prompt,
		Size:           size,
		N:              1,
		ResponseFormat: "b64_json",
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	url := o.baseURL + "/v1/images/generations"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 64<<20))
	if err != nil {
		return nil, err
	}

	var parsed openAIResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("decode response (status %d): %w", resp.StatusCode, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if parsed.Error != nil && parsed.Error.Message != "" {
			return nil, fmt.Errorf("provider error: %s", parsed.Error.Message)
		}
		return nil, fmt.Errorf("provider returned status %d", resp.StatusCode)
	}
	if len(parsed.Data) == 0 {
		return nil, errors.New("provider returned no image data")
	}

	item := parsed.Data[0]
	if item.B64JSON != "" {
		imgBytes, err := base64.StdEncoding.DecodeString(item.B64JSON)
		if err != nil {
			return nil, fmt.Errorf("decode b64_json: %w", err)
		}
		return &GenerateResult{Image: imgBytes, MimeType: detectMime(imgBytes)}, nil
	}
	if item.URL != "" {
		return o.download(ctx, item.URL)
	}
	return nil, errors.New("provider returned neither b64_json nor url")
}

func (o *OpenAIProvider) edit(ctx context.Context, req GenerateRequest) (*GenerateResult, error) {
	if o.baseURL == "" {
		return nil, errors.New("openai base url not configured")
	}
	if req.SourceImage == nil || len(req.SourceImage.Data) == 0 {
		return nil, errors.New("source image is required for image edit")
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("model", o.model); err != nil {
		return nil, err
	}
	if err := writer.WriteField("prompt", buildPrompt(req)); err != nil {
		return nil, err
	}
	if req.Width > 0 && req.Height > 0 {
		if err := writer.WriteField("size", fmt.Sprintf("%dx%d", req.Width, req.Height)); err != nil {
			return nil, err
		}
	}
	if err := writeMultipartImage(writer, "image[]", req.SourceImage); err != nil {
		return nil, err
	}
	if req.MaskImage != nil && len(req.MaskImage.Data) > 0 {
		if err := writeMultipartImage(writer, "mask", req.MaskImage); err != nil {
			return nil, err
		}
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}

	url := o.baseURL + "/v1/images/edits"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &body)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", writer.FormDataContentType())
	httpReq.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 64<<20))
	if err != nil {
		return nil, err
	}

	var parsed openAIResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("decode response (status %d): %w", resp.StatusCode, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if parsed.Error != nil && parsed.Error.Message != "" {
			return nil, fmt.Errorf("provider error: %s", parsed.Error.Message)
		}
		return nil, fmt.Errorf("provider returned status %d", resp.StatusCode)
	}
	if len(parsed.Data) == 0 {
		return nil, errors.New("provider returned no image data")
	}

	item := parsed.Data[0]
	if item.B64JSON != "" {
		imgBytes, err := base64.StdEncoding.DecodeString(item.B64JSON)
		if err != nil {
			return nil, fmt.Errorf("decode b64_json: %w", err)
		}
		return &GenerateResult{Image: imgBytes, MimeType: detectMime(imgBytes)}, nil
	}
	if item.URL != "" {
		return o.download(ctx, item.URL)
	}
	return nil, errors.New("provider returned neither b64_json nor url")
}

func buildPrompt(req GenerateRequest) string {
	prompt := req.Prompt
	if req.Style != "" {
		prompt = req.Style + " style: " + prompt
	}
	if req.NegativePrompt != "" {
		prompt += "\nAvoid: " + req.NegativePrompt
	}
	return prompt
}

func writeMultipartImage(writer *multipart.Writer, field string, img *InputImage) error {
	filename := img.Filename
	if filename == "" {
		filename = "image" + extensionForMime(img.MimeType)
	}
	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, field, filepath.Base(filename)))
	if img.MimeType != "" {
		header.Set("Content-Type", img.MimeType)
	}
	part, err := writer.CreatePart(header)
	if err != nil {
		return err
	}
	_, err = part.Write(img.Data)
	return err
}

func extensionForMime(mime string) string {
	switch mime {
	case "image/jpeg":
		return ".jpg"
	case "image/webp":
		return ".webp"
	default:
		return ".png"
	}
}

func (o *OpenAIProvider) download(ctx context.Context, rawURL string) (*GenerateResult, error) {
	if err := validateDownloadURL(rawURL); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("download image: status %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 64<<20))
	if err != nil {
		return nil, err
	}
	return &GenerateResult{Image: data, MimeType: detectMime(data)}, nil
}

// validateDownloadURL guards the provider's url-return path against SSRF: the
// URL must be https and must not resolve to a loopback / private / link-local
// address (cloud metadata, internal services). Best-effort against an untrusted
// or compromised upstream that returns a malicious url instead of b64_json.
func validateDownloadURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid image url: %w", err)
	}
	if u.Scheme != "https" {
		return errors.New("image url must use https")
	}
	host := u.Hostname()
	if host == "" {
		return errors.New("image url has no host")
	}
	if ip := net.ParseIP(host); ip != nil {
		if isDisallowedIP(ip) {
			return errors.New("image url resolves to a disallowed address")
		}
		return nil
	}
	ips, err := net.LookupIP(host)
	if err != nil {
		return fmt.Errorf("resolve image host: %w", err)
	}
	for _, ip := range ips {
		if isDisallowedIP(ip) {
			return errors.New("image url resolves to a disallowed address")
		}
	}
	return nil
}

func isDisallowedIP(ip net.IP) bool {
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() ||
		ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast()
}

// detectMime sniffs common image signatures, defaulting to image/png.
func detectMime(b []byte) string {
	if len(b) >= 8 && bytes.Equal(b[:8], []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a}) {
		return "image/png"
	}
	if len(b) >= 3 && b[0] == 0xff && b[1] == 0xd8 && b[2] == 0xff {
		return "image/jpeg"
	}
	if len(b) >= 12 && bytes.Equal(b[:4], []byte("RIFF")) && bytes.Equal(b[8:12], []byte("WEBP")) {
		return "image/webp"
	}
	ct := http.DetectContentType(b)
	if strings.HasPrefix(ct, "image/") {
		return ct
	}
	return "image/png"
}
