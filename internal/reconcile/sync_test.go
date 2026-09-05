package reconcile

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/gregsharpe-bot/readest-to-obsidian-sync/internal/events"
)

type fakeStore struct {
	objects map[string]string
}

func (f fakeStore) GetObject(_ context.Context, input *awss3.GetObjectInput, _ ...func(*awss3.Options)) (*awss3.GetObjectOutput, error) {
	body := f.objects[*input.Key]
	return &awss3.GetObjectOutput{Body: io.NopCloser(strings.NewReader(body))}, nil
}

func TestSyncWritesBookNoteAndRunsObsidianSync(t *testing.T) {
	vault := t.TempDir()
	store := fakeStore{objects: map[string]string{
		"library.json":       `[{"hash":"book-1","metaHash":"meta-1","title":"A Book","author":"An Author","format":"EPUB"}]`,
		"book-1/config.json": `{"bookHash":"book-1","updatedAt":1788211499680,"booknotes":[{"id":"note-1","type":"annotation","text":"A useful quote","note":"My thought","page":12},{"id":"deleted","type":"annotation","text":"Do not show","deletedAt":1}]}`,
	}}
	var command string
	syncer := Syncer{
		Store:       store,
		Bucket:      "books",
		Vault:       vault,
		NotesFolder: "Readest",
		RunCommand: func(_ context.Context, name string, args ...string) error {
			command = name + " " + args[0] + " " + args[1]
			return nil
		},
	}
	err := syncer.Process(context.Background(), events.Record{EventName: "ObjectCreated:Put", S3: events.S3Record{Object: events.Object{Key: "book-1/config.json"}}})
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(vault, "Readest", "A Book [book-1].md"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !containsAll(content, "title: \"A Book\"", "A useful quote", "My thought", "page: 12") {
		t.Fatalf("note content missing expected fields:\n%s", content)
	}
	if strings.Contains(content, "Do not show") || command != "sync --path "+vault {
		t.Fatalf("unexpected deleted note or command: %q\n%s", command, content)
	}
}

func TestSyncSupportsS3Prefix(t *testing.T) {
	vault := t.TempDir()
	store := fakeStore{objects: map[string]string{
		"Readest/library.json":             `[{"hash":"book-1","title":"A Book"}]`,
		"Readest/books/book-1/config.json": `{"bookHash":"book-1"}`,
	}}
	syncer := Syncer{
		Store:       store,
		Prefix:      "Readest",
		Vault:       vault,
		NotesFolder: "Readest",
	}
	if err := syncer.Process(context.Background(), events.Record{EventName: "ObjectCreated:Put", S3: events.S3Record{Object: events.Object{Key: "Readest/books/book-1/config.json"}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(vault, "Readest", "A Book [book-1].md")); err != nil {
		t.Fatal(err)
	}
}

func TestRenderEscapesFrontmatterAndSkipsDeletedNotes(t *testing.T) {
	content := Render(Book{Hash: "hash", Title: `A "Book"`}, Config{Booknotes: []Note{{ID: "deleted", Text: "hidden", DeletedAt: ptr(1)}}})
	if !strings.Contains(content, `title: "A \"Book\""`) || strings.Contains(content, "hidden") {
		t.Fatalf("unexpected rendered note:\n%s", content)
	}
}

func ptr(value int64) *int64 { return &value }

func contains(value, part string) bool { return strings.Contains(value, part) }

func containsAll(value string, parts ...string) bool {
	for _, part := range parts {
		if !contains(value, part) {
			return false
		}
	}
	return true
}
