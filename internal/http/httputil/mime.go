package httputil

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"strings"

	"github.com/gabriel-vasile/mimetype"
)

const mimeSniffBytes = 3072

func DetectMime(file multipart.File) (*mimetype.MIME, error) {
	buf := make([]byte, mimeSniffBytes)
	n, err := io.ReadFull(file, buf)
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	if seeker, ok := file.(io.Seeker); ok {
		if _, err := seeker.Seek(0, io.SeekStart); err != nil {
			return nil, fmt.Errorf("error resetting file: %w", err)
		}
	}

	return mimetype.Detect(buf[:n]), nil
}

func GetMimeDetails(fileHeader *multipart.FileHeader, file multipart.File) (string, string, string, error) {
	mime, err := DetectMime(file)
	if err != nil {
		return "", "", "", err
	}

	mimeType := mime.String()
	extension := mime.Extension()
	fileType := strings.Split(mimeType, "/")[0]

	return mimeType, extension, fileType, nil
}
