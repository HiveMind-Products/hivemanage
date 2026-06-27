package file

import (
	"fmt"
	"path"
	"strings"

	"github.com/fivemanage/lite/internal/crypt"
)

func generateFileKey(organizationID string, ext string) (string, error) {
	filename, err := crypt.GenerateFilename()
	if err != nil {
		return "", err
	}

	key := fmt.Sprintf("%s/%s%s", organizationID, filename, ext)
	return key, nil
}

// generateFileKeyV3 builds a storage key honoring the optional V3 path (folder)
// and filename. The org id is always the top-level prefix so files stay scoped
// to their organization. A random filename is used when none is supplied.
func generateFileKeyV3(organizationID, folder, filename, ext string) (string, error) {
	name := sanitizeName(filename)
	if name == "" {
		generated, err := crypt.GenerateFilename()
		if err != nil {
			return "", err
		}
		name = generated + ext
	} else if path.Ext(name) == "" && ext != "" {
		name += ext
	}

	folder = sanitizeFolder(folder)
	if folder != "" {
		return fmt.Sprintf("%s/%s/%s", organizationID, folder, name), nil
	}
	return fmt.Sprintf("%s/%s", organizationID, name), nil
}

// sanitizeFolder normalizes a user-supplied folder path and strips any traversal
// or absolute components so it can never escape the organization prefix.
func sanitizeFolder(folder string) string {
	folder = strings.TrimSpace(folder)
	if folder == "" {
		return ""
	}
	cleaned := strings.Trim(path.Clean("/"+strings.Trim(folder, "/")), "/")
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return ""
	}
	return cleaned
}

// sanitizeName strips any directory components from a supplied filename.
func sanitizeName(filename string) string {
	filename = strings.TrimSpace(filename)
	if filename == "" {
		return ""
	}
	return path.Base(strings.Trim(filename, "/"))
}

func generateWebKey(organizationID, filename string) string {
	key := fmt.Sprintf("%s/%s", organizationID, filename)
	return key
}
