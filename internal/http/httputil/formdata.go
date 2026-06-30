package httputil

import (
	"log/slog"
	"mime/multipart"
	"net/http"
)

func File(r *http.Request, key string) (file multipart.File, fileHeader *multipart.FileHeader, err error) {
	err = r.ParseMultipartForm(32 << 20)
	if err != nil {
		slog.Error("failed to parse multipart form", "err", err)
		return nil, nil, err
	}

	// check if "file" exists in the form first
	if r.MultipartForm.File["file"] != nil {
		file, fileHeader, err = r.FormFile("file")
		return file, fileHeader, err
	}

	file, fileHeader, err = r.FormFile(key)
	return file, fileHeader, err
}

// FileAny returns the first multipart file found under any of the provided keys.
// It is used by the V3 upload route so that clients which send the file under a
// media-typed field name (e.g. FiveM's screenshot-basic, which defaults to
// "image") work without changing the client. "file" is always tried first.
func FileAny(r *http.Request, keys ...string) (file multipart.File, fileHeader *multipart.FileHeader, err error) {
	err = r.ParseMultipartForm(32 << 20)
	if err != nil {
		slog.Error("failed to parse multipart form", "err", err)
		return nil, nil, err
	}

	for _, key := range append([]string{"file"}, keys...) {
		if r.MultipartForm.File[key] != nil {
			return r.FormFile(key)
		}
	}

	// Fall back to "file" so the returned error matches the documented field.
	return r.FormFile("file")
}

func Metadata(r *http.Request) string {
	return r.FormValue("metadata")
}

func Resource(r *http.Request) string {
	return r.FormValue("resource")
}

func Filename(r *http.Request) string {
	return r.FormValue("filename")
}
