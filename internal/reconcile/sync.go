package reconcile

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/gregsharpe-bot/readest-to-obsidian-sync/internal/events"
)

type ObjectStore interface {
	GetObject(context.Context, *awss3.GetObjectInput, ...func(*awss3.Options)) (*awss3.GetObjectOutput, error)
}

type Config struct {
	BookHash  string `json:"bookHash"`
	MetaHash  string `json:"metaHash"`
	Progress  []int  `json:"progress"`
	UpdatedAt int64  `json:"updatedAt"`
	Booknotes []Note `json:"booknotes"`
}

type Note struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Page      int    `json:"page"`
	Text      string `json:"text"`
	Note      string `json:"note"`
	Color     string `json:"color"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt int64  `json:"updatedAt"`
	DeletedAt *int64 `json:"deletedAt"`
}

type Book struct {
	Hash      string         `json:"hash"`
	MetaHash  string         `json:"metaHash"`
	Title     string         `json:"title"`
	Author    string         `json:"author"`
	Format    string         `json:"format"`
	Tags      []string       `json:"tags"`
	Metadata  map[string]any `json:"metadata"`
	DeletedAt *int64         `json:"deletedAt"`
}

type libraryEnvelope struct {
	Books []Book `json:"books"`
}

type CommandRunner func(context.Context, string, ...string) error

type Syncer struct {
	Store       ObjectStore
	Bucket      string
	Prefix      string
	Vault       string
	NotesFolder string
	RunCommand  CommandRunner
}

func (s Syncer) Process(ctx context.Context, record events.Record) error {
	if record.EventName == "ObjectRemoved:Delete" || strings.HasPrefix(record.EventName, "ObjectRemoved:") {
		return nil
	}
	return s.Sync(ctx, record)
}

func (s Syncer) Sync(ctx context.Context, record events.Record) error {
	hash, err := bookHashFromKey(record.S3.Object.Key, s.Prefix)
	if err != nil {
		return err
	}
	var library []Book
	var envelope libraryEnvelope
	if err := s.readJSON(ctx, s.objectKey("library.json"), &envelope); err == nil {
		library = envelope.Books
	} else {
		if err := s.readJSON(ctx, s.objectKey("library.json"), &library); err != nil {
			return fmt.Errorf("read library.json: %w", err)
		}
	}
	var book *Book
	for index := range library {
		if library[index].Hash == hash {
			book = &library[index]
			break
		}
	}
	if book == nil {
		return fmt.Errorf("book %q is not present in library.json", hash)
	}
	if book.DeletedAt != nil {
		return nil
	}
	var config Config
	if err := s.readJSON(ctx, s.bookObjectKey(hash), &config); err != nil {
		return fmt.Errorf("read %s/config.json: %w", hash, err)
	}
	if config.BookHash != "" && config.BookHash != hash {
		return fmt.Errorf("config bookHash %q does not match object hash %q", config.BookHash, hash)
	}
	content := Render(*book, config)
	if err := writeNote(s.Vault, s.NotesFolder, *book, content); err != nil {
		return err
	}
	if s.RunCommand != nil {
		if err := s.RunCommand(ctx, "sync", "--path", s.Vault); err != nil {
			return fmt.Errorf("sync Obsidian vault: %w", err)
		}
	}
	return nil
}

func (s Syncer) objectKey(key string) string {
	if s.Prefix == "" {
		return key
	}
	return strings.Trim(s.Prefix, "/") + "/" + key
}

func (s Syncer) bookObjectKey(hash string) string {
	key := hash + "/config.json"
	if s.Prefix != "" {
		key = "books/" + key
	}
	return s.objectKey(key)
}

func (s Syncer) readJSON(ctx context.Context, key string, target any) error {
	output, err := s.Store.GetObject(ctx, &awss3.GetObjectInput{Bucket: &s.Bucket, Key: &key})
	if err != nil {
		return err
	}
	defer output.Body.Close()
	data, err := io.ReadAll(output.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

var hashPattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

func bookHashFromKey(key, prefix string) (string, error) {
	key = strings.TrimPrefix(key, "/")
	prefix = strings.Trim(prefix, "/")
	if prefix != "" {
		prefixWithSeparator := prefix + "/"
		if !strings.HasPrefix(key, prefixWithSeparator) {
			return "", fmt.Errorf("unsupported Readest object key %q", key)
		}
		key = strings.TrimPrefix(key, prefixWithSeparator)
	}
	key = strings.TrimPrefix(key, "books/")
	parts := strings.Split(key, "/")
	if len(parts) < 2 || parts[1] != "config.json" || !hashPattern.MatchString(parts[0]) {
		return "", fmt.Errorf("unsupported Readest object key %q", key)
	}
	return parts[0], nil
}

func Render(book Book, config Config) string {
	var builder strings.Builder
	title := book.Title
	if title == "" {
		title = book.Hash
	}
	builder.WriteString("---\n")
	fm(&builder, "title", title)
	fm(&builder, "author", book.Author)
	fm(&builder, "readest_hash", book.Hash)
	fm(&builder, "readest_meta_hash", book.MetaHash)
	fm(&builder, "format", book.Format)
	fm(&builder, "source_updated_at", time.UnixMilli(config.UpdatedAt).UTC().Format(time.RFC3339))
	builder.WriteString("---\n\n# ")
	builder.WriteString(title)
	builder.WriteString("\n\n")
	builder.WriteString("## Highlights\n\n")
	for _, note := range config.Booknotes {
		if note.DeletedAt != nil || note.Text == "" {
			continue
		}
		builder.WriteString("> ")
		builder.WriteString(strings.ReplaceAll(note.Text, "\n", "\n> "))
		builder.WriteString("\n\n")
		if note.Note != "" {
			builder.WriteString(note.Note)
			builder.WriteString("\n\n")
		}
		builder.WriteString("<!-- readest-note: ")
		builder.WriteString(note.ID)
		if note.Page > 0 {
			builder.WriteString(fmt.Sprintf("; page: %d", note.Page))
		}
		if note.Color != "" {
			builder.WriteString("; color: ")
			builder.WriteString(note.Color)
		}
		builder.WriteString(" -->\n\n")
	}
	return builder.String()
}

func fm(builder *strings.Builder, key, value string) {
	value = strings.ReplaceAll(strings.ReplaceAll(value, "\\", "\\\\"), "\"", "\\\"")
	builder.WriteString(key + ": \"" + value + "\"\n")
}

func writeNote(vault, folder string, book Book, content string) error {
	filename := safeFilename(book.Title)
	if filename == "" {
		filename = book.Hash
	}
	filename += " [" + book.Hash + "].md"
	folder = filepath.ToSlash(filepath.Clean(folder))
	if folder == "." || folder == ".." || strings.HasPrefix(folder, "../") || filepath.IsAbs(folder) {
		return fmt.Errorf("notes folder must stay within the vault: %q", folder)
	}
	directory := filepath.Join(vault, filepath.FromSlash(folder))
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return fmt.Errorf("create note directory: %w", err)
	}
	temporary, err := os.CreateTemp(directory, ".readest-*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary note: %w", err)
	}
	temporaryName := temporary.Name()
	defer os.Remove(temporaryName)
	if err := temporary.Chmod(0o644); err != nil {
		return err
	}
	if _, err := temporary.WriteString(content); err != nil {
		temporary.Close()
		return fmt.Errorf("write note: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := os.Rename(temporaryName, filepath.Join(directory, filename)); err != nil {
		return fmt.Errorf("replace note: %w", err)
	}
	return nil
}

func safeFilename(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Map(func(r rune) rune {
		switch r {
		case '/', '\\', ':', '*', '?', '"', '<', '>', '|':
			return '-'
		default:
			return r
		}
	}, value)
	return strings.Trim(strings.Join(strings.Fields(value), " "), ".")
}

func ExecCommand(ctx context.Context, command string, args ...string) error {
	return exec.CommandContext(ctx, "ob", append([]string{command}, args...)...).Run()
}
