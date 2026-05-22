package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/smtp"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	smallRecipient = "mailhog-smoke-small@example.test"
	largeRecipient = "mailhog-smoke-large@example.test"
	sender         = "mailhog-smoke@example.test"
)

type config struct {
	HTTPURL                string
	SMTPAddr               string
	SmallMessages          int
	SmallBytes             int
	LargeMessages          int
	LargeAttachmentBytes   int
	RPS                    int
	Workers                int
	LoadDuration           time.Duration
	RequestTimeout         time.Duration
	MaxCompactPayloadBytes int64
}

type messagesResponse struct {
	Total int                      `json:"total"`
	Count int                      `json:"count"`
	Start int                      `json:"start"`
	Items []map[string]interface{} `json:"items"`
}

type bodyChunk struct {
	ID         string       `json:"ID"`
	Content    *bodyContent `json:"Content"`
	Offset     int64        `json:"Offset"`
	NextOffset int64        `json:"NextOffset"`
	Limit      int64        `json:"Limit"`
	MaxSize    int64        `json:"MaxSize"`
	HasMore    bool         `json:"HasMore"`
	Truncated  bool         `json:"Truncated"`
	Source     string       `json:"Source"`
}

type bodyContent struct {
	Headers map[string][]string `json:"Headers"`
	Body    string              `json:"Body"`
	Size    int                 `json:"Size"`
}

type namedOp struct {
	name string
	run  func() error
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "perf smoke failed: %s\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := loadConfig()
	client := &http.Client{Timeout: cfg.RequestTimeout}

	if err := waitForHTTP(client, cfg); err != nil {
		return err
	}

	if err := sendDataset(cfg); err != nil {
		return err
	}

	list, listBytes, err := fetchMessages(client, endpoint(cfg.HTTPURL, "/api/v3/messages?limit=50"), cfg.MaxCompactPayloadBytes)
	if err != nil {
		return err
	}
	if list.Count == 0 {
		return fmt.Errorf("list endpoint returned no messages")
	}
	fmt.Printf("validated compact list response: messages=%d total=%d bytes=%d\n", list.Count, list.Total, listBytes)

	smallSearchURL := endpoint(cfg.HTTPURL, "/api/v3/search?kind=to&query="+url.QueryEscape(smallRecipient)+"&limit=50")
	smallMessages, searchBytes, err := fetchMessages(client, smallSearchURL, cfg.MaxCompactPayloadBytes)
	if err != nil {
		return err
	}
	smallID, err := firstMessageID(smallMessages)
	if err != nil {
		return fmt.Errorf("small message search: %w", err)
	}
	fmt.Printf("validated compact search response: messages=%d bytes=%d\n", smallMessages.Count, searchBytes)

	largeSearchURL := endpoint(cfg.HTTPURL, "/api/v3/search?kind=to&query="+url.QueryEscape(largeRecipient)+"&limit=50")
	largeMessages, _, err := fetchMessages(client, largeSearchURL, cfg.MaxCompactPayloadBytes)
	if err != nil {
		return err
	}
	largeID, err := firstMessageID(largeMessages)
	if err != nil {
		return fmt.Errorf("large message search: %w", err)
	}

	if err := checkBodyPreview(client, cfg, largeID, true); err != nil {
		return err
	}
	if err := checkDownload(client, cfg, smallID); err != nil {
		return err
	}

	if err := runLoad(client, cfg, smallID, largeID); err != nil {
		return err
	}

	fmt.Println("perf smoke passed")
	return nil
}

func loadConfig() config {
	return config{
		HTTPURL:                strings.TrimRight(getenv("MAILHOG_HTTP_URL", "http://127.0.0.1:8025"), "/"),
		SMTPAddr:               getenv("MAILHOG_SMTP_ADDR", "127.0.0.1:1025"),
		SmallMessages:          getenvInt("MAILHOG_SMOKE_SMALL_MESSAGES", 20),
		SmallBytes:             getenvInt("MAILHOG_SMOKE_SMALL_BYTES", 1024*1024),
		LargeMessages:          getenvInt("MAILHOG_SMOKE_LARGE_MESSAGES", 3),
		LargeAttachmentBytes:   getenvInt("MAILHOG_SMOKE_LARGE_ATTACHMENT_BYTES", 8*1024*1024),
		RPS:                    getenvInt("MAILHOG_SMOKE_RPS", 10),
		Workers:                getenvInt("MAILHOG_SMOKE_WORKERS", 4),
		LoadDuration:           getenvDuration("MAILHOG_SMOKE_DURATION", 10*time.Second),
		RequestTimeout:         getenvDuration("MAILHOG_SMOKE_REQUEST_TIMEOUT", 10*time.Second),
		MaxCompactPayloadBytes: int64(getenvInt("MAILHOG_SMOKE_MAX_COMPACT_BYTES", 2*1024*1024)),
	}
}

func waitForHTTP(client *http.Client, cfg config) error {
	deadline := time.Now().Add(60 * time.Second)
	var lastErr error
	for time.Now().Before(deadline) {
		resp, err := client.Get(endpoint(cfg.HTTPURL, "/api/v3/messages?limit=1"))
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 500 {
				return nil
			}
			lastErr = fmt.Errorf("status %d", resp.StatusCode)
		} else {
			lastErr = err
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("MailHog HTTP endpoint did not become ready: %w", lastErr)
}

func sendDataset(cfg config) error {
	for i := 0; i < cfg.SmallMessages; i++ {
		subject := fmt.Sprintf("perf smoke small %03d", i)
		msg := buildTextMessage(smallRecipient, subject, cfg.SmallBytes)
		if err := sendMessageWithRetry(cfg.SMTPAddr, smallRecipient, msg); err != nil {
			return fmt.Errorf("send small message %d: %w", i, err)
		}
	}
	for i := 0; i < cfg.LargeMessages; i++ {
		subject := fmt.Sprintf("perf smoke large attachment %03d", i)
		msg := buildMIMEAttachmentMessage(largeRecipient, subject, cfg.LargeAttachmentBytes, i)
		if err := sendMessageWithRetry(cfg.SMTPAddr, largeRecipient, msg); err != nil {
			return fmt.Errorf("send large message %d: %w", i, err)
		}
	}
	fmt.Printf("sent dataset: small=%d large=%d\n", cfg.SmallMessages, cfg.LargeMessages)
	return nil
}

func buildTextMessage(to, subject string, bodyBytes int) []byte {
	var b strings.Builder
	b.Grow(bodyBytes + 512)
	writeCommonHeaders(&b, to, subject)
	b.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
	b.WriteString("\r\n")
	b.WriteString("MailHog performance smoke text message.\r\n")
	if bodyBytes > 0 {
		b.WriteString(strings.Repeat("x", bodyBytes))
	}
	b.WriteString("\r\n")
	return []byte(b.String())
}

func buildMIMEAttachmentMessage(to, subject string, attachmentBytes int, index int) []byte {
	boundary := fmt.Sprintf("mailhog-smoke-boundary-%d", index)
	var b strings.Builder
	b.Grow(attachmentBytes + attachmentBytes/2 + 2048)
	writeCommonHeaders(&b, to, subject)
	fmt.Fprintf(&b, "Content-Type: multipart/mixed; boundary=%q\r\n", boundary)
	b.WriteString("\r\n")
	fmt.Fprintf(&b, "--%s\r\n", boundary)
	b.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
	b.WriteString("\r\n")
	b.WriteString("Visible body for large attachment smoke message.\r\n")
	fmt.Fprintf(&b, "--%s\r\n", boundary)
	b.WriteString("Content-Type: application/octet-stream\r\n")
	b.WriteString("Content-Transfer-Encoding: base64\r\n")
	b.WriteString("Content-Disposition: attachment; filename=\"large.bin\"\r\n")
	b.WriteString("\r\n")

	raw := bytes.Repeat([]byte{byte('A' + index%26)}, attachmentBytes)
	encoded := make([]byte, base64.StdEncoding.EncodedLen(len(raw)))
	base64.StdEncoding.Encode(encoded, raw)
	writeWrappedBase64(&b, encoded)
	fmt.Fprintf(&b, "--%s--\r\n", boundary)
	return []byte(b.String())
}

func writeCommonHeaders(b *strings.Builder, to, subject string) {
	fmt.Fprintf(b, "From: %s\r\n", sender)
	fmt.Fprintf(b, "To: %s\r\n", to)
	fmt.Fprintf(b, "Subject: %s\r\n", subject)
	fmt.Fprintf(b, "Date: %s\r\n", time.Now().Format(time.RFC1123Z))
}

func writeWrappedBase64(b *strings.Builder, encoded []byte) {
	const width = 76
	for len(encoded) > 0 {
		n := width
		if len(encoded) < n {
			n = len(encoded)
		}
		b.Write(encoded[:n])
		b.WriteString("\r\n")
		encoded = encoded[n:]
	}
}

func sendMessageWithRetry(addr, to string, msg []byte) error {
	var lastErr error
	for i := 0; i < 20; i++ {
		if err := smtp.SendMail(addr, nil, sender, []string{to}, msg); err != nil {
			lastErr = err
			time.Sleep(500 * time.Millisecond)
			continue
		}
		return nil
	}
	return lastErr
}

func fetchMessages(client *http.Client, rawURL string, maxBytes int64) (*messagesResponse, int, error) {
	body, err := getLimited(client, rawURL, maxBytes)
	if err != nil {
		return nil, 0, err
	}

	var result messagesResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, len(body), fmt.Errorf("decode %s: %w", rawURL, err)
	}
	if err := validateCompactMessages(result.Items); err != nil {
		return nil, len(body), err
	}
	return &result, len(body), nil
}

func validateCompactMessages(items []map[string]interface{}) error {
	for i, item := range items {
		if !isNilJSON(item["Raw"]) {
			return fmt.Errorf("item %d includes Raw payload", i)
		}
		if !isNilJSON(item["MIME"]) {
			return fmt.Errorf("item %d includes MIME payload", i)
		}
		content, ok := item["Content"].(map[string]interface{})
		if !ok {
			continue
		}
		if body, ok := content["Body"].(string); ok && body != "" {
			return fmt.Errorf("item %d includes Content.Body with %d bytes", i, len(body))
		}
		if !isNilJSON(content["MIME"]) {
			return fmt.Errorf("item %d includes Content.MIME payload", i)
		}
	}
	return nil
}

func checkBodyPreview(client *http.Client, cfg config, id string, verbose bool) error {
	bodyURL := endpoint(cfg.HTTPURL, "/api/v3/messages/"+url.PathEscape(id)+"/body?offset=0&limit=1048576")
	body, err := getLimited(client, bodyURL, 2*1024*1024)
	if err != nil {
		return err
	}
	var chunk bodyChunk
	if err := json.Unmarshal(body, &chunk); err != nil {
		return fmt.Errorf("decode body preview: %w", err)
	}
	if chunk.Content == nil || chunk.Content.Body == "" {
		return fmt.Errorf("body preview for %s is empty", id)
	}
	if len(chunk.Content.Body) > 1024*1024 {
		return fmt.Errorf("body preview returned %d bytes, expected <= 1MiB", len(chunk.Content.Body))
	}
	if verbose {
		fmt.Printf("validated body preview: id=%s bodyBytes=%d hasMore=%v truncated=%v source=%s\n", id, len(chunk.Content.Body), chunk.HasMore, chunk.Truncated, chunk.Source)
	}
	return nil
}

func checkDownload(client *http.Client, cfg config, id string) error {
	downloadURL := endpoint(cfg.HTTPURL, "/api/v3/messages/"+url.PathEscape(id)+"/download")
	resp, err := client.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("GET %s: %w", downloadURL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("GET %s returned status %d", downloadURL, resp.StatusCode)
	}
	n, err := io.Copy(io.Discard, resp.Body)
	if err != nil {
		return fmt.Errorf("read download: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("download returned no bytes")
	}
	return nil
}

func runLoad(client *http.Client, cfg config, smallID, largeID string) error {
	if cfg.RPS <= 0 {
		return fmt.Errorf("MAILHOG_SMOKE_RPS must be positive")
	}
	if cfg.Workers <= 0 {
		return fmt.Errorf("MAILHOG_SMOKE_WORKERS must be positive")
	}

	listURL := endpoint(cfg.HTTPURL, "/api/v3/messages?limit=50")
	searchURL := endpoint(cfg.HTTPURL, "/api/v3/search?kind=to&query="+url.QueryEscape(smallRecipient)+"&limit=50")

	ops := []namedOp{
		{name: "list", run: func() error {
			_, _, err := fetchMessages(client, listURL, cfg.MaxCompactPayloadBytes)
			return err
		}},
		{name: "search", run: func() error {
			_, _, err := fetchMessages(client, searchURL, cfg.MaxCompactPayloadBytes)
			return err
		}},
		{name: "body", run: func() error {
			return checkBodyPreview(client, cfg, largeID, false)
		}},
		{name: "download", run: func() error {
			return checkDownload(client, cfg, smallID)
		}},
	}

	jobs := make(chan namedOp)
	var wg sync.WaitGroup
	var mu sync.Mutex
	counts := map[string]int{}
	var firstErr error

	for i := 0; i < cfg.Workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for op := range jobs {
				if err := op.run(); err != nil {
					mu.Lock()
					if firstErr == nil {
						firstErr = fmt.Errorf("%s request failed: %w", op.name, err)
					}
					mu.Unlock()
					continue
				}
				mu.Lock()
				counts[op.name]++
				mu.Unlock()
			}
		}()
	}

	interval := time.Second / time.Duration(cfg.RPS)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	deadline := time.NewTimer(cfg.LoadDuration)
	defer deadline.Stop()

	sent := 0
loop:
	for {
		select {
		case <-deadline.C:
			break loop
		case <-ticker.C:
			mu.Lock()
			err := firstErr
			mu.Unlock()
			if err != nil {
				break loop
			}
			jobs <- ops[sent%len(ops)]
			sent++
		}
	}
	close(jobs)
	wg.Wait()

	if firstErr != nil {
		return firstErr
	}
	minRequests := int(cfg.LoadDuration.Seconds() * float64(cfg.RPS) * 0.8)
	if sent < minRequests {
		return fmt.Errorf("sent %d requests, expected at least %d", sent, minRequests)
	}
	fmt.Printf("load complete: sent=%d counts=%v\n", sent, counts)
	return nil
}

func getLimited(client *http.Client, rawURL string, maxBytes int64) ([]byte, error) {
	resp, err := client.Get(rawURL)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", rawURL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("GET %s returned status %d", rawURL, resp.StatusCode)
	}
	limited := &io.LimitedReader{R: resp.Body, N: maxBytes + 1}
	body, err := ioutil.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", rawURL, err)
	}
	if int64(len(body)) > maxBytes {
		return nil, fmt.Errorf("GET %s returned more than %d bytes", rawURL, maxBytes)
	}
	return body, nil
}

func firstMessageID(result *messagesResponse) (string, error) {
	if result == nil || len(result.Items) == 0 {
		return "", fmt.Errorf("no messages returned")
	}
	id, ok := result.Items[0]["ID"].(string)
	if !ok || id == "" {
		return "", fmt.Errorf("first message has no ID")
	}
	return id, nil
}

func endpoint(base, path string) string {
	return strings.TrimRight(base, "/") + "/" + strings.TrimLeft(path, "/")
}

func isNilJSON(v interface{}) bool {
	return v == nil
}

func getenv(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func getenvInt(name string, fallback int) int {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getenvDuration(name string, fallback time.Duration) time.Duration {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	if parsed, err := time.ParseDuration(value); err == nil {
		return parsed
	}
	if seconds, err := strconv.Atoi(value); err == nil {
		return time.Duration(seconds) * time.Second
	}
	return fallback
}
