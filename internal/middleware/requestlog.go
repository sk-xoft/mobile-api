package middleware

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"mime"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/smartseaman/mobile-api/internal/domain/eventlog"
	"go.uber.org/zap"
)

const (
	requestLogMaxBodyBytes = 32 * 1024
	fileBase64Marker       = "[FILE_BASE64_OMITTED]"
	redactedMarker         = "[REDACTED]"
)

var sensitiveFieldNames = map[string]struct{}{
	"password": {}, "confirmpassword": {}, "oldpassword": {}, "newpassword": {},
	"token": {}, "tokenfcm": {}, "accesstoken": {}, "refreshtoken": {}, "reftoken": {}, "uid": {}, "secret": {}, "jwt": {},
}

var fileFieldNames = map[string]struct{}{
	"file": {}, "filebase64": {}, "filecert": {}, "filedata": {}, "image": {},
	"imagebase64": {}, "imageprofile": {}, "profileimage": {}, "documentfile": {}, "base64": {},
}

type EventRecorder interface {
	Record(context.Context, eventlog.Event) error
}

func RequestLogger(recorders ...EventRecorder) fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := strings.TrimSpace(c.Get("X-Request-ID"))
		if requestID == "" {
			requestID = uuid.NewString()
		}
		c.Set("X-Request-ID", requestID)
		startedAt := time.Now()

		err := c.Next()
		responseBody := c.Response().Body()
		status := c.Response().StatusCode()
		if err != nil && status < fiber.StatusBadRequest {
			status = fiber.StatusInternalServerError
		}

		zap.L().Info("http request completed",
			zap.String("request_id", requestID),
			zap.String("trace_id", traceID(c)),
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.Int("status", status),
			zap.Int64("duration_ms", time.Since(startedAt).Milliseconds()),
			zap.String("request_body", sanitizeRequestBody(c.Get(fiber.HeaderContentType), c.Body())),
			zap.String("response_body", sanitizeResponseBody(c.GetRespHeader(fiber.HeaderContentType), responseBody)),
		)
		if err != nil || status >= fiber.StatusBadRequest {
			level := zap.WarnLevel
			if status >= fiber.StatusInternalServerError || err != nil {
				level = zap.ErrorLevel
			}
			fields := []zap.Field{
				zap.String("request_id", requestID),
				zap.String("trace_id", traceID(c)),
				zap.String("method", c.Method()),
				zap.String("path", c.Path()),
				zap.Int("status", status),
				zap.String("error", errorMessage(err)),
			}
			if level == zap.ErrorLevel {
				zap.L().Error("http request error", fields...)
			} else {
				zap.L().Warn("http request error", fields...)
			}
		}
		if len(recorders) > 0 {
			actorUserID, _ := c.Locals("user.mobile_uuid").(string)
			eventName := "http.request.completed"
			if err != nil || status >= fiber.StatusBadRequest {
				eventName = "http.request.failed"
			}
			_ = recorders[0].Record(c.UserContext(), eventlog.Event{
				Name:        eventName,
				ActorUserID: actorUserID,
				RequestID:   requestID,
				TraceID:     traceID(c),
				Metadata: map[string]any{
					"method":      c.Method(),
					"path":        c.Path(),
					"status":      status,
					"duration_ms": time.Since(startedAt).Milliseconds(),
				},
			})
		}

		return err
	}
}

func errorMessage(err error) string {
	if err == nil {
		return ""
	}
	message := err.Error()
	if len(message) > 2000 {
		return message[:2000] + "...[ERROR_TRUNCATED]"
	}
	return message
}

func traceID(c *fiber.Ctx) string {
	if value, ok := c.Locals(TraceID).(string); ok {
		return value
	}
	return ""
}

func sanitizeRequestBody(contentType string, body []byte) string {
	mediaType, _, _ := mime.ParseMediaType(contentType)
	if mediaType == "multipart/form-data" {
		return "[MULTIPART_FILE_OMITTED]"
	}
	if len(body) == 0 {
		return ""
	}
	return sanitizeBody(mediaType, body)
}

func sanitizeResponseBody(contentType string, body []byte) string {
	if len(body) == 0 {
		return ""
	}
	mediaType, _, _ := mime.ParseMediaType(contentType)
	if strings.HasPrefix(mediaType, "image/") || mediaType == "application/pdf" || mediaType == "application/octet-stream" {
		return "[BINARY_RESPONSE_OMITTED]"
	}
	return sanitizeBody(mediaType, body)
}

func sanitizeBody(mediaType string, body []byte) string {
	if mediaType == "application/json" || json.Valid(body) {
		var payload any
		if err := json.Unmarshal(body, &payload); err == nil {
			payload = sanitizeValue(payload, "")
			encoded, err := json.Marshal(payload)
			if err == nil {
				return capLogBody(encoded)
			}
		}
	}
	if looksLikeBase64File(string(body)) {
		return fileBase64Marker
	}
	return capLogBody(body)
}

func sanitizeValue(value any, fieldName string) any {
	normalizedFieldName := strings.ToLower(strings.TrimSpace(fieldName))
	if _, ok := sensitiveFieldNames[normalizedFieldName]; ok {
		return redactedMarker
	}
	if _, ok := fileFieldNames[normalizedFieldName]; ok {
		return fileBase64Marker
	}

	switch typedValue := value.(type) {
	case map[string]any:
		for key, child := range typedValue {
			typedValue[key] = sanitizeValue(child, key)
		}
	case []any:
		for index, child := range typedValue {
			typedValue[index] = sanitizeValue(child, fieldName)
		}
	case string:
		if looksLikeBase64File(typedValue) {
			return fileBase64Marker
		}
	}
	return value
}

func looksLikeBase64File(value string) bool {
	value = strings.TrimSpace(value)
	lowerValue := strings.ToLower(value)
	if strings.HasPrefix(lowerValue, "data:") && strings.Contains(lowerValue, ";base64,") {
		return true
	}
	if len(value) < 256 || strings.ContainsAny(value, " /\\\r\n") {
		return false
	}
	decoded, err := base64.StdEncoding.DecodeString(value)
	return err == nil && len(decoded) >= 128
}

func capLogBody(body []byte) string {
	body = bytes.TrimSpace(body)
	if len(body) <= requestLogMaxBodyBytes {
		return string(body)
	}
	return string(body[:requestLogMaxBodyBytes]) + "...[BODY_TRUNCATED]"
}
