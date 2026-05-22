package storage

import (
	"bufio"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"mime/multipart"
	"net/textproto"
	"os"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/mailhog/data"
)

const (
	maildirRefreshInterval = 5 * time.Second
	maildirTempPrefix      = ".mailhog-tmp-"
	maildirFlushInterval   = 8 * 1024 * 1024
)

type maildirCacheEntry struct {
	id         string
	modTime    time.Time
	message    data.Message
	searchTo   []string
	searchFrom []string
}

type maildirMessageWriter struct {
	maildir     *Maildir
	id          string
	path        string
	tempPath    string
	file        maildirDataFile
	raw         *data.SMTPMessage
	headers     map[string][]string
	lastHeader  string
	inHeaders   bool
	contentSize int
	dirtyBytes  int
	committed   bool
}

type maildirDataFile interface {
	io.Writer
	Sync() error
	Close() error
}

// Maildir is a maildir storage backend
type Maildir struct {
	Path      string
	mu        sync.RWMutex
	entries   []*maildirCacheEntry
	entryByID map[string]*maildirCacheEntry
}

// CreateMaildir creates a new maildir storage backend
func CreateMaildir(path string) *Maildir {
	if len(path) == 0 {
		dir, err := ioutil.TempDir("", "mailhog")
		if err != nil {
			panic(err)
		}
		path = dir
	}
	if _, err := os.Stat(path); err != nil {
		err := os.MkdirAll(path, 0770)
		if err != nil {
			panic(err)
		}
	}
	log.Println("Maildir path is", path)
	maildir := &Maildir{
		Path:      path,
		entryByID: make(map[string]*maildirCacheEntry),
	}
	go maildir.refreshLoop()
	return maildir
}

// Store stores a message and returns its storage ID
func (maildir *Maildir) Store(m *data.Message) (string, error) {
	id := string(m.ID)
	path := filepath.Join(maildir.Path, id)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0660)
	if err != nil {
		return "", err
	}
	if err := writeSMTPMessage(file, m.Raw); err != nil {
		file.Close()
		os.Remove(path)
		return "", err
	}
	if err := file.Sync(); err != nil {
		file.Close()
		os.Remove(path)
		return "", err
	}
	dropFileCache(file)
	if err := file.Close(); err != nil {
		os.Remove(path)
		return "", err
	}

	modTime := time.Now()
	if fileinfo, err := os.Stat(path); err == nil {
		modTime = fileinfo.ModTime()
	}
	maildir.putCacheEntry(newMaildirCacheEntry(id, modTime, m))

	return id, nil
}

// CreateMessageWriter stores SMTP DATA incrementally without retaining the
// full body or attachments in memory.
func (maildir *Maildir) CreateMessageWriter(m *data.SMTPMessage, hostname string) (MessageWriter, error) {
	id, err := data.NewMessageID(hostname)
	if err != nil {
		return nil, err
	}

	writer := &maildirMessageWriter{
		maildir:   maildir,
		id:        string(id),
		path:      filepath.Join(maildir.Path, string(id)),
		tempPath:  filepath.Join(maildir.Path, maildirTempPrefix+string(id)),
		raw:       cloneSMTPMessageEnvelope(m),
		headers:   make(map[string][]string),
		inHeaders: true,
	}

	file, err := openMaildirDataFile(writer.tempPath)
	if err != nil {
		return nil, err
	}
	writer.file = file

	if err := writer.writeEnvelope(); err != nil {
		writer.Abort()
		return nil, err
	}

	return writer, nil
}

func (writer *maildirMessageWriter) WriteLine(line string) error {
	if writer.file == nil {
		return os.ErrClosed
	}
	if err := writeString(writer.file, line); err != nil {
		return err
	}
	if err := writeString(writer.file, "\r\n"); err != nil {
		return err
	}
	lineSize := len(line) + len("\r\n")
	writer.contentSize += lineSize
	writer.dirtyBytes += lineSize
	writer.captureHeaderLine(line)
	if writer.dirtyBytes >= maildirFlushInterval {
		return writer.flush()
	}
	return nil
}

func (writer *maildirMessageWriter) Commit() (string, *data.Message, error) {
	if writer.file == nil {
		return "", nil, os.ErrClosed
	}
	if err := writer.flush(); err != nil {
		writer.Abort()
		return "", nil, err
	}
	if err := writer.file.Close(); err != nil {
		writer.file = nil
		os.Remove(writer.tempPath)
		return "", nil, err
	}
	writer.file = nil

	if err := os.Rename(writer.tempPath, writer.path); err != nil {
		os.Remove(writer.tempPath)
		return "", nil, err
	}
	writer.committed = true

	modTime := time.Now()
	if fileinfo, err := os.Stat(writer.path); err == nil {
		modTime = fileinfo.ModTime()
	}

	msg := &data.Message{
		ID:      data.MessageID(writer.id),
		From:    data.PathFromString(writer.raw.From),
		To:      maildirPathsFromStrings(writer.raw.To),
		Created: modTime,
		Content: &data.Content{
			Headers: writer.headers,
			Size:    writer.contentSize,
		},
	}
	writer.maildir.putCacheEntry(newMaildirCacheEntry(writer.id, modTime, msg))
	if writer.contentSize >= maildirFlushInterval {
		debug.FreeOSMemory()
	}
	return writer.id, msg, nil
}

func (writer *maildirMessageWriter) Abort() error {
	if writer.committed {
		return nil
	}
	if writer.file != nil {
		writer.file.Close()
		writer.file = nil
	}
	return os.Remove(writer.tempPath)
}

func (writer *maildirMessageWriter) flush() error {
	if writer.file == nil {
		return os.ErrClosed
	}
	if err := writer.file.Sync(); err != nil {
		return err
	}
	writer.dirtyBytes = 0
	return nil
}

func (writer *maildirMessageWriter) writeEnvelope() error {
	if err := writeString(writer.file, "HELO:<"+writer.raw.Helo+">\r\n"); err != nil {
		return err
	}
	if err := writeString(writer.file, "FROM:<"+writer.raw.From+">\r\n"); err != nil {
		return err
	}
	for _, t := range writer.raw.To {
		if err := writeString(writer.file, "TO:<"+t+">\r\n"); err != nil {
			return err
		}
	}
	return writeString(writer.file, "\r\n")
}

func (writer *maildirMessageWriter) captureHeaderLine(line string) {
	if !writer.inHeaders {
		return
	}
	if line == "" {
		writer.inHeaders = false
		return
	}
	writer.lastHeader = appendMailHeaderLine(writer.headers, writer.lastHeader, line)
}

func writeSMTPMessage(w io.Writer, m *data.SMTPMessage) error {
	if err := writeString(w, "HELO:<"+m.Helo+">\r\n"); err != nil {
		return err
	}
	if err := writeString(w, "FROM:<"+m.From+">\r\n"); err != nil {
		return err
	}
	for _, t := range m.To {
		if err := writeString(w, "TO:<"+t+">\r\n"); err != nil {
			return err
		}
	}
	if err := writeString(w, "\r\n"); err != nil {
		return err
	}
	return writeString(w, m.Data)
}

func writeString(w io.Writer, value string) error {
	_, err := io.WriteString(w, value)
	return err
}

func cloneSMTPMessageEnvelope(m *data.SMTPMessage) *data.SMTPMessage {
	if m == nil {
		return &data.SMTPMessage{}
	}
	return &data.SMTPMessage{
		From: m.From,
		To:   append([]string(nil), m.To...),
		Helo: m.Helo,
	}
}

func maildirPathsFromStrings(paths []string) []*data.Path {
	items := make([]*data.Path, 0, len(paths))
	for _, path := range paths {
		items = append(items, data.PathFromString(path))
	}
	return items
}

func appendMailHeaderLine(headers map[string][]string, lastHeader string, line string) string {
	if lastHeader != "" && (strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t")) {
		headers[lastHeader][len(headers[lastHeader])-1] = headers[lastHeader][len(headers[lastHeader])-1] + line
		return lastHeader
	}
	if strings.Contains(line, ": ") {
		parts := strings.SplitN(line, ": ", 2)
		headers[parts[0]] = []string{parts[1]}
		return parts[0]
	}
	return lastHeader
}

// Count returns the number of stored messages
func (maildir *Maildir) Count() int {
	maildir.mu.RLock()
	defer maildir.mu.RUnlock()
	return len(maildir.entries)
}

// Search finds messages matching the query
func (maildir *Maildir) Search(kind, query string, start, limit int) (*data.Messages, int, error) {
	if start < 0 {
		start = 0
	}
	if limit <= 0 {
		messages := data.Messages(make([]data.Message, 0))
		return &messages, 0, nil
	}

	query = strings.ToLower(query)

	if kind == "containing" {
		return maildir.searchContaining(query, start, limit)
	}

	maildir.mu.RLock()
	defer maildir.mu.RUnlock()

	messages := make([]data.Message, 0)
	matched := 0
	for _, entry := range maildir.entries {
		if !entry.matches(kind, query) {
			continue
		}
		if matched >= start && len(messages) < limit {
			messages = append(messages, entry.message)
		}
		matched++
	}

	msgs := data.Messages(messages)
	return &msgs, matched, nil
}

func (maildir *Maildir) searchContaining(query string, start, limit int) (*data.Messages, int, error) {
	entries := maildir.snapshotEntries()
	messages := make([]data.Message, 0)
	matched := 0

	for _, entry := range entries {
		msg, err := maildir.Load(entry.id)
		if err != nil {
			log.Println(err)
			continue
		}

		if msg.Raw != nil && strings.Contains(strings.ToLower(msg.Raw.Data), query) {
			if matched >= start && len(messages) < limit {
				messages = append(messages, compactMaildirMessage(msg))
			}
			matched++
		}

		if matched >= start+limit {
			break
		}
	}

	msgs := data.Messages(messages)
	return &msgs, matched, nil
}

// List lists stored messages by index
func (maildir *Maildir) List(start, limit int) (*data.Messages, error) {
	log.Println("Listing messages in", maildir.Path)
	messages := make([]data.Message, 0)
	msgs := data.Messages(messages)

	if start < 0 {
		start = 0
	}
	if limit <= 0 {
		return &msgs, nil
	}

	maildir.mu.RLock()
	defer maildir.mu.RUnlock()
	if len(maildir.entries) == 0 || start >= len(maildir.entries) {
		return &msgs, nil
	}

	end := start + limit
	if end > len(maildir.entries) {
		end = len(maildir.entries)
	}

	for _, entry := range maildir.entries[start:end] {
		messages = append(messages, entry.message)
	}

	log.Printf("Found %d messages", len(messages))
	msgs = data.Messages(messages)
	return &msgs, nil
}

// DeleteOne deletes an individual message by storage ID
func (maildir *Maildir) DeleteOne(id string) error {
	err := os.Remove(filepath.Join(maildir.Path, id))
	if err == nil {
		maildir.removeCacheEntry(id)
	}
	return err
}

// DeleteAll deletes all in memory messages
func (maildir *Maildir) DeleteAll() error {
	err := os.RemoveAll(maildir.Path)
	if err != nil {
		return err
	}
	err = os.Mkdir(maildir.Path, 0770)
	if err == nil {
		maildir.clearCache()
	}
	return err
}

// DeleteOlderThan deletes maildir messages older than cutoff.
func (maildir *Maildir) DeleteOlderThan(cutoff time.Time) error {
	files, err := ioutil.ReadDir(maildir.Path)
	if err != nil {
		return err
	}

	var firstErr error
	for _, fileinfo := range files {
		if fileinfo.IsDir() || strings.HasPrefix(fileinfo.Name(), maildirTempPrefix) || !fileinfo.ModTime().Before(cutoff) {
			continue
		}
		err := os.Remove(filepath.Join(maildir.Path, fileinfo.Name()))
		if err != nil && !os.IsNotExist(err) && firstErr == nil {
			firstErr = err
		}
	}

	maildir.refreshCache()
	return firstErr
}

// Load returns an individual message by storage ID
func (maildir *Maildir) Load(id string) (*data.Message, error) {
	b, err := ioutil.ReadFile(filepath.Join(maildir.Path, id))
	if err != nil {
		return nil, err
	}
	// FIXME domain
	m := data.FromBytes(b).Parse("mailhog.example")
	m.ID = data.MessageID(id)
	if fileinfo, err := os.Stat(filepath.Join(maildir.Path, id)); err == nil {
		m.Created = fileinfo.ModTime()
	}
	return m, nil
}

// LoadPreview returns a full message for small files and a compact message for
// large files so UI preview cannot load huge attachments into memory.
func (maildir *Maildir) LoadPreview(id string, maxSize int64) (*data.Message, error) {
	path := filepath.Join(maildir.Path, id)
	fileinfo, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if maxSize <= 0 || fileinfo.Size() <= maxSize {
		return maildir.Load(id)
	}
	return maildir.loadCompactMessage(fileinfo)
}

// LoadBodyChunk returns a bounded chunk of the displayable message body. For
// multipart messages it walks MIME parts and skips attachments.
func (maildir *Maildir) LoadBodyChunk(id string, offset int64, limit int64, maxSize int64) (*MessageBodyChunk, error) {
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = 1024 * 1024
	}
	if maxSize <= 0 {
		maxSize = limit
	}

	file, err := os.Open(filepath.Join(maildir.Path, id))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	if err := skipMaildirEnvelope(reader); err != nil {
		return nil, err
	}

	headers, err := textproto.NewReader(reader).ReadMIMEHeader()
	if err != nil {
		return nil, err
	}

	chunk, found, err := maildirReadMIMEBodyChunk(id, reader, headers, offset, limit, maxSize)
	if err != nil {
		return nil, err
	}
	if found {
		dropFileCache(file)
		return chunk, nil
	}

	chunk, err = maildirReadBodyChunk(id, maildirHeaderMap(headers), reader, offset, limit, maxSize, "message")
	if err != nil {
		return nil, err
	}
	dropFileCache(file)
	return chunk, nil
}

// WriteMessageTo streams the original RFC822 message without the internal
// maildir envelope lines.
func (maildir *Maildir) WriteMessageTo(id string, w io.Writer) error {
	file, err := os.Open(filepath.Join(maildir.Path, id))
	if err != nil {
		return err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	if err := skipMaildirEnvelope(reader); err != nil {
		return err
	}

	buf := make([]byte, 1024*1024)
	bytesSinceDrop := 0
	for {
		n, readErr := reader.Read(buf)
		if n > 0 {
			if _, err := w.Write(buf[:n]); err != nil {
				return err
			}
			bytesSinceDrop += n
			if bytesSinceDrop >= maildirFlushInterval {
				dropFileCache(file)
				bytesSinceDrop = 0
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return readErr
		}
	}

	dropFileCache(file)
	return nil
}

func skipMaildirEnvelope(reader *bufio.Reader) error {
	for {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return err
		}
		if strings.TrimRight(line, "\r\n") == "" || err == io.EOF {
			return nil
		}
	}
}

func maildirReadMIMEBodyChunk(id string, reader io.Reader, headers textproto.MIMEHeader, offset int64, limit int64, maxSize int64) (*MessageBodyChunk, bool, error) {
	contentType := headers.Get("Content-Type")
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		mediaType = strings.ToLower(strings.TrimSpace(contentType))
	}
	mediaType = strings.ToLower(mediaType)

	if strings.HasPrefix(mediaType, "multipart/") {
		boundary := params["boundary"]
		if boundary == "" {
			return nil, false, nil
		}

		multipartReader := multipart.NewReader(reader, boundary)
		var plainFallback *MessageBodyChunk
		for {
			part, err := multipartReader.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, false, err
			}

			partHeaders := textproto.MIMEHeader(part.Header)
			if maildirIsAttachment(partHeaders) {
				continue
			}

			chunk, found, err := maildirReadMIMEBodyChunk(id, part, partHeaders, offset, limit, maxSize)
			if err != nil {
				return nil, false, err
			}
			if !found {
				continue
			}
			if maildirIsHTML(partHeaders) {
				return chunk, true, nil
			}
			if plainFallback == nil {
				plainFallback = chunk
			}
			if mediaType != "multipart/alternative" {
				return plainFallback, true, nil
			}
		}
		if plainFallback != nil {
			return plainFallback, true, nil
		}
		return nil, false, nil
	}

	if !maildirIsDisplayableText(headers) {
		return nil, false, nil
	}

	chunk, err := maildirReadBodyChunk(id, maildirHeaderMap(headers), reader, offset, limit, maxSize, "mime")
	if err != nil {
		return nil, false, err
	}
	return chunk, true, nil
}

func maildirReadBodyChunk(id string, headers map[string][]string, reader io.Reader, offset int64, limit int64, maxSize int64, source string) (*MessageBodyChunk, error) {
	if offset >= maxSize {
		return &MessageBodyChunk{
			ID: id,
			Content: &data.Content{
				Headers: headers,
				Body:    "",
				Size:    int(maxSize),
			},
			Offset:     offset,
			NextOffset: offset,
			Limit:      limit,
			MaxSize:    maxSize,
			HasMore:    false,
			Truncated:  true,
			Source:     source,
		}, nil
	}

	if offset > 0 {
		if _, err := io.CopyN(io.Discard, reader, offset); err != nil {
			if err == io.EOF {
				return &MessageBodyChunk{
					ID: id,
					Content: &data.Content{
						Headers: headers,
						Body:    "",
						Size:    int(offset),
					},
					Offset:     offset,
					NextOffset: offset,
					Limit:      limit,
					MaxSize:    maxSize,
					HasMore:    false,
					Truncated:  false,
					Source:     source,
				}, nil
			}
			return nil, err
		}
	}

	readLimit := limit
	if offset+readLimit > maxSize {
		readLimit = maxSize - offset
	}
	bytes, err := ioutil.ReadAll(io.LimitReader(reader, readLimit+1))
	if err != nil {
		return nil, err
	}

	hasExtra := int64(len(bytes)) > readLimit
	if hasExtra {
		bytes = bytes[:readLimit]
	}
	nextOffset := offset + int64(len(bytes))
	hasMore := hasExtra && nextOffset < maxSize
	truncated := hasExtra && nextOffset >= maxSize

	return &MessageBodyChunk{
		ID: id,
		Content: &data.Content{
			Headers: headers,
			Body:    string(bytes),
			Size:    int(nextOffset),
		},
		Offset:     offset,
		NextOffset: nextOffset,
		Limit:      limit,
		MaxSize:    maxSize,
		HasMore:    hasMore,
		Truncated:  truncated,
		Source:     source,
	}, nil
}

func maildirHeaderMap(headers textproto.MIMEHeader) map[string][]string {
	values := make(map[string][]string, len(headers))
	for key, headerValues := range headers {
		values[key] = append([]string(nil), headerValues...)
	}
	return values
}

func maildirIsAttachment(headers textproto.MIMEHeader) bool {
	disposition := strings.ToLower(headers.Get("Content-Disposition"))
	return strings.HasPrefix(disposition, "attachment")
}

func maildirIsDisplayableText(headers textproto.MIMEHeader) bool {
	if maildirIsAttachment(headers) {
		return false
	}
	contentType := strings.ToLower(headers.Get("Content-Type"))
	if contentType == "" {
		return true
	}
	return strings.HasPrefix(contentType, "text/plain") || strings.HasPrefix(contentType, "text/html")
}

func maildirIsHTML(headers textproto.MIMEHeader) bool {
	return strings.HasPrefix(strings.ToLower(headers.Get("Content-Type")), "text/html")
}

func (maildir *Maildir) refreshLoop() {
	maildir.refreshCache()

	ticker := time.NewTicker(maildirRefreshInterval)
	defer ticker.Stop()
	for range ticker.C {
		maildir.refreshCache()
	}
}

func (maildir *Maildir) refreshCache() {
	files, err := ioutil.ReadDir(maildir.Path)
	if err != nil {
		log.Printf("Error reading maildir cache: %s", err)
		return
	}

	current := maildir.currentEntryMap()
	entries := make([]*maildirCacheEntry, 0, len(files))
	seen := make(map[string]struct{}, len(files))

	for _, fileinfo := range files {
		if fileinfo.IsDir() || strings.HasPrefix(fileinfo.Name(), maildirTempPrefix) {
			continue
		}

		id := fileinfo.Name()
		seen[id] = struct{}{}

		if entry, ok := current[id]; ok && entry.modTime.Equal(fileinfo.ModTime()) {
			entries = append(entries, entry)
			continue
		}

		entry, err := maildir.loadCacheEntry(fileinfo)
		if err != nil {
			log.Printf("Error loading maildir cache entry %s: %s", id, err)
			continue
		}
		entries = append(entries, entry)
	}

	maildir.mu.Lock()
	for id, entry := range maildir.entryByID {
		if _, ok := seen[id]; ok {
			continue
		}
		if _, err := os.Stat(filepath.Join(maildir.Path, id)); err == nil {
			entries = append(entries, entry)
		}
	}
	maildir.replaceCacheLocked(entries)
	maildir.mu.Unlock()
}

func (maildir *Maildir) currentEntryMap() map[string]*maildirCacheEntry {
	maildir.mu.RLock()
	defer maildir.mu.RUnlock()

	current := make(map[string]*maildirCacheEntry, len(maildir.entryByID))
	for id, entry := range maildir.entryByID {
		current[id] = entry
	}
	return current
}

func (maildir *Maildir) loadCacheEntry(fileinfo os.FileInfo) (*maildirCacheEntry, error) {
	msg, err := maildir.loadCompactMessage(fileinfo)
	if err != nil {
		return nil, err
	}

	return newMaildirCacheEntry(fileinfo.Name(), fileinfo.ModTime(), msg), nil
}

func (maildir *Maildir) loadCompactMessage(fileinfo os.FileInfo) (*data.Message, error) {
	path := filepath.Join(maildir.Path, fileinfo.Name())
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	envelopeBytes := 0
	raw := &data.SMTPMessage{}

	for {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return nil, err
		}
		envelopeBytes += len(line)
		trimmed := strings.TrimRight(line, "\r\n")
		if trimmed == "" || err == io.EOF {
			break
		}
		if strings.HasPrefix(trimmed, "HELO:<") {
			raw.Helo = strings.TrimSuffix(strings.TrimPrefix(trimmed, "HELO:<"), ">")
		} else if strings.HasPrefix(trimmed, "FROM:<") {
			raw.From = strings.TrimSuffix(strings.TrimPrefix(trimmed, "FROM:<"), ">")
		} else if strings.HasPrefix(trimmed, "TO:<") {
			raw.To = append(raw.To, strings.TrimSuffix(strings.TrimPrefix(trimmed, "TO:<"), ">"))
		}
		if err == io.EOF {
			break
		}
	}

	headers := make(map[string][]string)
	lastHeader := ""
	for {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return nil, err
		}
		trimmed := strings.TrimRight(line, "\r\n")
		if trimmed == "" || err == io.EOF {
			break
		}
		lastHeader = appendMailHeaderLine(headers, lastHeader, trimmed)
		if err == io.EOF {
			break
		}
	}

	size := int(fileinfo.Size()) - envelopeBytes
	if size < 0 {
		size = 0
	}

	return &data.Message{
		ID:      data.MessageID(fileinfo.Name()),
		From:    data.PathFromString(raw.From),
		To:      maildirPathsFromStrings(raw.To),
		Created: fileinfo.ModTime(),
		Content: &data.Content{
			Headers: headers,
			Size:    size,
		},
	}, nil
}

func (maildir *Maildir) snapshotEntries() []*maildirCacheEntry {
	maildir.mu.RLock()
	defer maildir.mu.RUnlock()

	entries := make([]*maildirCacheEntry, len(maildir.entries))
	copy(entries, maildir.entries)
	return entries
}

func (maildir *Maildir) putCacheEntry(entry *maildirCacheEntry) {
	maildir.mu.Lock()
	defer maildir.mu.Unlock()

	entries := make([]*maildirCacheEntry, 0, len(maildir.entries)+1)
	for _, existing := range maildir.entries {
		if existing.id != entry.id {
			entries = append(entries, existing)
		}
	}
	entries = append(entries, entry)
	maildir.replaceCacheLocked(entries)
}

func (maildir *Maildir) removeCacheEntry(id string) {
	maildir.mu.Lock()
	defer maildir.mu.Unlock()

	entries := make([]*maildirCacheEntry, 0, len(maildir.entries))
	for _, entry := range maildir.entries {
		if entry.id != id {
			entries = append(entries, entry)
		}
	}
	maildir.replaceCacheLocked(entries)
}

func (maildir *Maildir) clearCache() {
	maildir.mu.Lock()
	defer maildir.mu.Unlock()

	maildir.entries = make([]*maildirCacheEntry, 0)
	maildir.entryByID = make(map[string]*maildirCacheEntry)
}

func (maildir *Maildir) replaceCacheLocked(entries []*maildirCacheEntry) {
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].modTime.Equal(entries[j].modTime) {
			return entries[i].id > entries[j].id
		}
		return entries[i].modTime.After(entries[j].modTime)
	})

	maildir.entries = entries
	maildir.entryByID = make(map[string]*maildirCacheEntry, len(entries))
	for _, entry := range entries {
		maildir.entryByID[entry.id] = entry
	}
}

func newMaildirCacheEntry(id string, modTime time.Time, msg *data.Message) *maildirCacheEntry {
	msg.ID = data.MessageID(id)
	msg.Created = modTime
	return &maildirCacheEntry{
		id:         id,
		modTime:    modTime,
		message:    compactMaildirMessage(msg),
		searchTo:   maildirSearchTo(msg),
		searchFrom: maildirSearchFrom(msg),
	}
}

func (entry *maildirCacheEntry) matches(kind, query string) bool {
	switch kind {
	case "to":
		return containsSearchValue(entry.searchTo, query)
	case "from":
		return containsSearchValue(entry.searchFrom, query)
	default:
		return false
	}
}

func maildirSearchTo(msg *data.Message) []string {
	values := make([]string, 0)
	for _, to := range msg.To {
		values = appendPathSearchValue(values, to)
	}
	if msg.Content != nil {
		values = appendHeaderSearchValues(values, msg.Content.Headers, "To")
	}
	return values
}

func maildirSearchFrom(msg *data.Message) []string {
	values := make([]string, 0)
	values = appendPathSearchValue(values, msg.From)
	if msg.Content != nil {
		values = appendHeaderSearchValues(values, msg.Content.Headers, "From")
	}
	return values
}

func appendPathSearchValue(values []string, path *data.Path) []string {
	if path == nil {
		return values
	}
	return append(values, strings.ToLower(path.Mailbox+"@"+path.Domain))
}

func appendHeaderSearchValues(values []string, headers map[string][]string, key string) []string {
	for header, headerValues := range headers {
		if strings.ToLower(header) != strings.ToLower(key) {
			continue
		}
		for _, value := range headerValues {
			values = append(values, strings.ToLower(value))
		}
	}
	return values
}

func containsSearchValue(values []string, query string) bool {
	for _, value := range values {
		if strings.Contains(value, query) {
			return true
		}
	}
	return false
}

func compactMaildirMessage(msg *data.Message) data.Message {
	compact := data.Message{
		ID:      msg.ID,
		From:    cloneMaildirPath(msg.From),
		To:      cloneMaildirPaths(msg.To),
		Created: msg.Created,
	}
	if msg.Content != nil {
		compact.Content = &data.Content{
			Headers: cloneMaildirHeaders(msg.Content.Headers),
			Size:    msg.Content.Size,
		}
	}
	return compact
}

func cloneMaildirPath(path *data.Path) *data.Path {
	if path == nil {
		return nil
	}
	clone := *path
	if path.Relays != nil {
		clone.Relays = append([]string(nil), path.Relays...)
	}
	return &clone
}

func cloneMaildirPaths(paths []*data.Path) []*data.Path {
	if paths == nil {
		return nil
	}
	clones := make([]*data.Path, 0, len(paths))
	for _, path := range paths {
		clones = append(clones, cloneMaildirPath(path))
	}
	return clones
}

func cloneMaildirHeaders(headers map[string][]string) map[string][]string {
	if headers == nil {
		return nil
	}
	clone := make(map[string][]string, len(headers))
	for key, values := range headers {
		clone[key] = append([]string(nil), values...)
	}
	return clone
}
