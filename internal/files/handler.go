package files

import (
	"context"
	"html/template"
	"log/slog"
	"net/http"
	"sort"
	"time"

	"github.com/shaolim/wa-integration/pkg/storage"
	"github.com/shaolim/wa-integration/web"
)

const presignTTL = 15 * time.Minute

var filesTmpl = template.Must(template.New("files").Parse(web.FilesHTML))
var fileDetailTmpl = template.Must(template.New("file_detail").Parse(web.FileDetailHTML))

type Lister interface {
	ListObjects(ctx context.Context) ([]storage.Object, error)
	GetPresignedURL(ctx context.Context, key string, expires time.Duration) (string, error)
	GetContentType(ctx context.Context, key string) (string, error)
}

type Handler struct {
	lister Lister
}

func NewHandler(lister Lister) *Handler {
	return &Handler{lister: lister}
}

func (h *Handler) ServeFiles(rw http.ResponseWriter, r *http.Request) {
	objects, err := h.lister.ListObjects(r.Context())
	if err != nil {
		slog.Error("files: list objects", slog.Any("error", err))
		http.Error(rw, "Failed to list files", http.StatusInternalServerError)
		return
	}

	sort.Slice(objects, func(i, j int) bool {
		return objects[i].LastModified.After(objects[j].LastModified)
	})

	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	filesTmpl.Execute(rw, map[string]any{"Files": objects})
}

func (h *Handler) ServeFileDetail(rw http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(rw, "Missing key", http.StatusBadRequest)
		return
	}

	objects, err := h.lister.ListObjects(r.Context())
	if err != nil {
		slog.Error("file detail: list objects", slog.Any("error", err))
		http.Error(rw, "Failed to fetch file info", http.StatusInternalServerError)
		return
	}

	var obj *storage.Object
	for i := range objects {
		if objects[i].Key == key {
			obj = &objects[i]
			break
		}
	}
	if obj == nil {
		http.Error(rw, "File not found", http.StatusNotFound)
		return
	}

	url, err := h.lister.GetPresignedURL(r.Context(), key, presignTTL)
	if err != nil {
		slog.Error("file detail: presign url", slog.String("key", key), slog.Any("error", err))
		http.Error(rw, "Failed to generate file URL", http.StatusInternalServerError)
		return
	}

	contentType, err := h.lister.GetContentType(r.Context(), key)
	if err != nil {
		slog.Error("file detail: get content type", slog.String("key", key), slog.Any("error", err))
		http.Error(rw, "Failed to get file type", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	fileDetailTmpl.Execute(rw, map[string]any{
		"Key":          obj.Key,
		"Size":         obj.Size,
		"LastModified": obj.LastModified,
		"URL":          url,
		"ContentType":  contentType,
	})
}
