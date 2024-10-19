package client

import (
	"bufio"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"

	"github.com/nikolalohinski/free-go/types"
)

const (
	pathNotFoundCode        = "path_not_found"
	destinationConflictCode = "destination_conflict"
)

func (c *client) GetFileInfo(ctx context.Context, path string) (types.FileInfo, error) {
	base64Path := base64.StdEncoding.EncodeToString([]byte(path))

	response, err := c.get(ctx, fmt.Sprintf("fs/info/%s", base64Path), c.withSession(ctx))
	if err != nil {
		if response != nil && response.ErrorCode == pathNotFoundCode {
			return types.FileInfo{}, ErrPathNotFound
		}
		return types.FileInfo{}, fmt.Errorf("failed to GET fs/info/%s endpoint: %w", base64Path, err)
	}

	result := types.FileInfo{}
	if response.Result != nil {
		if err = c.fromGenericResponse(response, &result); err != nil {
			return types.FileInfo{}, fmt.Errorf("failed to get file info from generic response: %w", err)
		}
	}

	return result, nil
}

func (c *client) RemoveFiles(ctx context.Context, paths []string) (task types.FileSystemTask, err error) {
	files := make([]types.Base64Path, len(paths))
	for i, p := range paths {
		files[i] = types.Base64Path(p)
	}

	response, err := c.post(ctx, "fs/rm/", map[string]interface{}{"files": files}, c.withSession(ctx))
	if err != nil {
		return task, fmt.Errorf("failed to POST to fs/rm/ endpoint: %w", err)
	}

	if err = c.fromGenericResponse(response, &task); err != nil {
		return task, fmt.Errorf("failed to get a filesystem task from a generic response: %w", err)
	}

	return task, nil
}

func (c *client) GetFileSystemTask(ctx context.Context, identifier int64) (task types.FileSystemTask, err error) {
	response, err := c.get(ctx, fmt.Sprintf("fs/tasks/%d", identifier), c.withSession(ctx))
	if err != nil {
		return task, fmt.Errorf("failed to GET fs/tasks/%d endpoint: %w", identifier, err)
	}

	if err = c.fromGenericResponse(response, &task); err != nil {
		return task, fmt.Errorf("failed to get a filesystem task from a generic response: %w", err)
	}

	return task, nil
}

func (c *client) DeleteFileSystemTask(ctx context.Context, identifier int64) error {
	response, err := c.delete(ctx, fmt.Sprintf("fs/tasks/%d", identifier), c.withSession(ctx))
	if err != nil {
		if response != nil && response.ErrorCode == codeTaskNotFound {
			return ErrTaskNotFound
		}
		return fmt.Errorf("failed to DELETE fs/tasks/%d endpoint: %w", identifier, err)
	}

	return nil
}

func (c *client) CreateDirectory(ctx context.Context, parent, name string) (string, error) {
	response, err := c.post(ctx, "fs/mkdir/", map[string]interface{}{
		"parent":  types.Base64Path(parent),
		"dirname": name,
	}, c.withSession(ctx))
	if err != nil {
		if response != nil && response.ErrorCode == destinationConflictCode {
			return "", ErrDestinationConflict
		}
		return "", fmt.Errorf("failed to POST to fs/mkdir/ endpoint: %w", err)
	}

	var result types.Base64Path
	if err = c.fromGenericResponse(response, &result); err != nil {
		return "", fmt.Errorf("failed to get a base64 string from a generic response: %w", err)
	}

	return string(result), nil
}

func (c *client) AddHashFileTask(ctx context.Context, payload types.HashPayload) (task types.FileSystemTask, err error) {
	response, err := c.post(ctx, "fs/hash/", payload, c.withSession(ctx))
	if err != nil {
		return task, fmt.Errorf("failed to POST to fs/hash/ endpoint: %w", err)
	}

	if err = c.fromGenericResponse(response, &task); err != nil {
		return task, fmt.Errorf("failed to get a filesystem task from a generic response: %w", err)
	}

	return task, nil
}

func (c *client) GetHashResult(ctx context.Context, identifier int64) (result string, err error) {
	response, err := c.get(ctx, fmt.Sprintf("fs/tasks/%d/hash/", identifier), c.withSession(ctx))
	if err != nil {
		return result, fmt.Errorf("failed to GET fs/tasks/%d/hash endpoint: %w", identifier, err)
	}

	if err = c.fromGenericResponse(response, &result); err != nil {
		return result, fmt.Errorf("failed to get a hash result from a generic response: %w", err)
	}

	return result, nil
}

func (c *client) GetFile(ctx context.Context, path string) (result types.File, err error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/dl/%s", c.base, base64.StdEncoding.EncodeToString([]byte(path))), nil)
	if err != nil {
		return result, fmt.Errorf("failed to forge new request: %w", err)
	}

	if err := c.withSession(ctx)(request); err != nil {
		return result, fmt.Errorf("failed to apply option to request: %w", err)
	}

	httpResponse, err := c.httpClient.Do(request)
	if err != nil {
		return result, fmt.Errorf("failed to perform request: %w", err)
	}

	if httpResponse.StatusCode != http.StatusOK {
		content, err := io.ReadAll(httpResponse.Body)
		if err != nil {
			return result, errors.Join(
				fmt.Errorf("failed with status '%d'", httpResponse.StatusCode),
				fmt.Errorf("failed to read response body: %w", err),
			)
		}

		return result, fmt.Errorf("failed with status '%d': server returned '%s'", httpResponse.StatusCode, content)
	}

	mediatype := ""
	if contentType := httpResponse.Header.Get("Content-Type"); contentType != "" {
		mediatype, _, err = mime.ParseMediaType(contentType)
		if err != nil {
			return result, fmt.Errorf("failed to parse media type: %w", err)
		}
	}

	var filename string

	if contentDisposition := httpResponse.Header.Get("Content-Disposition"); contentDisposition != "" {
		_, params, err := mime.ParseMediaType(contentDisposition)
		if err != nil {
			return result, fmt.Errorf("failed to parse media type: %w", err)
		}

		filename, _ = params["filename"]
	}

	return types.File{
		ContentType: mediatype,
		FileName:    filename,
		Content:     bufio.NewReader(httpResponse.Body),
	}, nil
}
