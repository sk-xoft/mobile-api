package requestitem

import (
	"errors"
	"io"

	"github.com/gofiber/fiber/v2"
	"github.com/smartseaman/mobile-api/internal/constants"
	"github.com/smartseaman/mobile-api/internal/transport/httpx"
	"go.uber.org/zap"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}
func (h *Handler) List(c *fiber.Ctx) error {
	v, e := h.service.List(c.UserContext(), uid(c))
	return h.result(c, v, e)
}
func (h *Handler) Preview(c *fiber.Ctx) error {
	v, e := h.service.Preview(c.UserContext(), uid(c), c.Params("itemCode"), "")
	return h.result(c, v, e)
}
func (h *Handler) RenewalPreview(c *fiber.Ctx) error {
	item := c.Params("documentRequestItemCode")
	no := c.Params("requestNo")
	if no == "" {
		no = c.Query("request_no")
	}
	requestID, _, _, _, e := h.service.repo.Renewal(c.UserContext(), uid(c), no, item)
	if e != nil {
		return h.fail(c, e)
	}
	v, e := h.service.Preview(c.UserContext(), uid(c), item, requestID)
	return h.result(c, v, e)
}
func (h *Handler) Upload(c *fiber.Ctx) error        { return h.upload(c, "") }
func (h *Handler) RenewalUpload(c *fiber.Ctx) error { return h.upload(c, c.Params("requestNo")) }
func (h *Handler) upload(c *fiber.Ctx, no string) error {
	file, e := c.FormFile("file")
	if e != nil {
		return h.fail(c, &BusinessError{Code: constants.CodeInvalidInput, Detail: "file"})
	}
	if file.Size <= 0 || file.Size > 10*1024*1024 {
		return h.fail(c, &BusinessError{Code: constants.CodeInvalidInput, Detail: "file"})
	}
	content, e := file.Open()
	if e != nil {
		return h.fail(c, e)
	}
	defer content.Close()
	bytes := make([]byte, file.Size)
	if _, e = io.ReadFull(content, bytes); e != nil {
		return h.fail(c, e)
	}
	itemCode := c.Params("itemCode")
	if itemCode == "" {
		itemCode = c.Params("documentRequestItemCode")
	}
	v, e := h.service.Upload(c.UserContext(), uid(c), itemCode, c.FormValue("documentType", constants.DocumentTypeGeneral), c.FormValue("slotCode", constants.DocumentSlotMain), file.Filename, bytes, no)
	return h.result(c, v, e)
}
func (h *Handler) result(c *fiber.Ctx, v any, e error) error {
	if e != nil {
		return h.fail(c, e)
	}
	return httpx.Success(c, v)
}
func (h *Handler) fail(c *fiber.Ctx, e error) error {
	if e != nil {
		zap.L().Error("request item operation failed",
			zap.String("request_id", c.Get("X-Request-ID")),
			zap.Any("trace_id", c.Locals("trace.id")),
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.Error(e),
		)
	}
	if errors.Is(e, ErrNotFound) {
		return httpx.Error(c, 400, constants.CodeNotFound, "data")
	}
	var b *BusinessError
	if errors.As(e, &b) {
		return httpx.Error(c, 400, b.Code, b.Detail)
	}
	return httpx.Error(c, 500, constants.CodeTechnical, "Technical exception")
}
func uid(c *fiber.Ctx) string { v, _ := c.Locals("user.mobile_uuid").(string); return v }
