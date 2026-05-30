package adapters

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type apiResponse struct {
	Data  any            `json:"data,omitempty"`
	Error *errorResponse `json:"error,omitempty"`
}

const (
	actorUserIDHeader  = "X-Actor-User-Id"
	actorGroupIDHeader = "X-Actor-Group-Id"
	requestLoggerKey   = "request_logger"
)

func logInternalRequestError(c *gin.Context, err error) {
	if c == nil || err == nil {
		return
	}

	loggerValue, exists := c.Get(requestLoggerKey)
	if !exists {
		return
	}

	logger, ok := loggerValue.(*zap.Logger)
	if !ok || logger == nil {
		return
	}

	logger.Error(
		"http request failed",
		zap.Error(err),
		zap.String("method", c.Request.Method),
		zap.String("path", c.FullPath()),
	)
}

func decodeJSONBody[T any](body io.ReadCloser) (T, error) {
	defer body.Close()

	var payload T
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&payload); err != nil {
		return payload, fmt.Errorf("decoding request body: %w", err)
	}

	if err := decoder.Decode(new(struct{})); err != io.EOF {
		if err == nil {
			return payload, errors.New("request body must contain a single JSON object")
		}

		return payload, fmt.Errorf("validating request body: %w", err)
	}

	return payload, nil
}

func readMultipartFile(fileHeader *multipart.FileHeader) ([]byte, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("opening uploaded file: %w", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("reading uploaded file: %w", err)
	}

	return content, nil
}

func parseOptionalTime(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}

	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, err
	}

	return &parsed, nil
}
