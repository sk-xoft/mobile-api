package routes

import (
	"database/sql"

	"github.com/gofiber/fiber/v2"
	"github.com/smartseaman/mobile-api/internal/domain/auth"
	"github.com/smartseaman/mobile-api/internal/domain/catalog"
	"github.com/smartseaman/mobile-api/internal/domain/content"
	"github.com/smartseaman/mobile-api/internal/domain/delivery"
	"github.com/smartseaman/mobile-api/internal/domain/document"
	"github.com/smartseaman/mobile-api/internal/domain/fcm"
	"github.com/smartseaman/mobile-api/internal/domain/master"
	"github.com/smartseaman/mobile-api/internal/domain/notification"
	"github.com/smartseaman/mobile-api/internal/domain/profile"
	"github.com/smartseaman/mobile-api/internal/domain/renewal"
	"github.com/smartseaman/mobile-api/internal/domain/requestitem"
	"github.com/smartseaman/mobile-api/internal/middleware"
	"github.com/smartseaman/mobile-api/internal/transport/httpx"
)

func Register(app *fiber.App, db *sql.DB, authHandler *auth.Handler, profileHandler *profile.Handler, masterHandler *master.Handler, documentHandler *document.Handler, deliveryHandler *delivery.Handler, fcmHandler *fcm.Handler, renewalHandler *renewal.Handler, contentHandler *content.Handler, catalogHandler *catalog.Handler, notificationHandler *notification.Handler, requestItemHandler *requestitem.Handler, jwtSecret string) {
	app.Get("/", func(c *fiber.Ctx) error { return c.SendString("Smart Seaman Mobile API") })
	app.Get("/health", func(c *fiber.Ctx) error { return c.JSON(fiber.Map{"status": "UP"}) })
	app.Get("/privacy_policy", func(c *fiber.Ctx) error { return httpx.Success(c, "success") })
	app.Post("/v1/payments/omise/webhook", renewalHandler.Webhook)

	v1 := app.Group("/v1", middleware.APIContext())
	registerPublic(v1, authHandler, profileHandler, masterHandler)
	registerProtected(v1, middleware.Authenticate(jwtSecret, db), authHandler, profileHandler, masterHandler, documentHandler, deliveryHandler, fcmHandler, renewalHandler, contentHandler, catalogHandler, notificationHandler, requestItemHandler)
}

func registerPublic(router fiber.Router, authHandler *auth.Handler, profileHandler *profile.Handler, masterHandler *master.Handler) {
	router.Post("/login", authHandler.Login)
	router.Post("/register", authHandler.Register)
	router.Get("/refresh-token", authHandler.Refresh)
	router.Get("/activate-user", authHandler.Activate)
	router.Post("/reset-password", authHandler.RequestPasswordReset)
	router.Post("/activate-forgot-password", authHandler.CompletePasswordReset)
	router.Post("/profile/active", authHandler.ActivateBody)
	router.Get("/master", masterHandler.List)
}

func registerProtected(router fiber.Router, authMiddleware fiber.Handler, authHandler *auth.Handler, profileHandler *profile.Handler, masterHandler *master.Handler, documentHandler *document.Handler, deliveryHandler *delivery.Handler, fcmHandler *fcm.Handler, renewalHandler *renewal.Handler, contentHandler *content.Handler, catalogHandler *catalog.Handler, notificationHandler *notification.Handler, requestItemHandler *requestitem.Handler) {
	secured := router.Group("", authMiddleware)
	secured.Post("/change-password", authHandler.ChangePassword)
	secured.Get("/profile", profileHandler.Get)
	secured.Post("/profile-update", profileHandler.Update)
	secured.Get("/profile-image", profileHandler.Image)
	secured.Delete("/profile/inactive", profileHandler.Deactivate)
	secured.Get("/master/documents", masterHandler.Documents)
	secured.Get("/master/courses", masterHandler.Courses)
	secured.Get("/master/provinces", masterHandler.Provinces)
	secured.Get("/master/districts", masterHandler.Districts)
	secured.Get("/master/subdistricts", masterHandler.Subdistricts)
	secured.Get("/delivery-addresses", deliveryHandler.Default)
	secured.Post("/delivery-addresses", deliveryHandler.Create)
	secured.Put("/delivery-addresses/:addressId", deliveryHandler.Update)
	secured.Post("/delivery-addresses/:requestNo", deliveryHandler.CreateForRenewal)
	secured.Post("/fcm-update", fcmHandler.Update)
	secured.Get("/document-renewals/statuses", renewalHandler.Statuses)
	secured.Get("/document-renewals/prices", renewalHandler.Price)
	secured.Get("/document-renewals", renewalHandler.Documents)
	secured.Post("/document-renewals", renewalHandler.CreateRequest)
	secured.Post("/document-renewals/requests/validate-and-create", renewalHandler.Validate)
	secured.Get("/document-renewals/my", renewalHandler.ListRequests)
	secured.Put("/document-renewals/requests/:requestNo/mobile", renewalHandler.UpdateMobile)
	secured.Delete("/document-renewals/requests/:requestNo", renewalHandler.DeleteRequest)
	secured.Delete("/document-renewals/requests/:requestNo/hard", renewalHandler.HardDelete)
	secured.Post("/document-renewals/:requestNo/resubmit", renewalHandler.Resubmit)
	secured.Get("/document-renewals/:requestNo/timeline", renewalHandler.Timeline)
	secured.Get("/document-renewals/:requestNo/stage", renewalHandler.Stage)
	secured.Get("/document-renewals/:requestNo", renewalHandler.Detail)
	secured.Post("/document-renewals/:requestId/payments", renewalHandler.Payment)
	secured.Get("/document-renewals/:requestId/payments", renewalHandler.CurrentPayment)
	secured.Get("/document-renewals/:requestId/payments/:transactionId", renewalHandler.PaymentStatus)
	secured.Get("/forms", contentHandler.Forms)
	secured.Get("/form", contentHandler.FormFile)
	secured.Get("/news", contentHandler.News)
	secured.Get("/news/detail", contentHandler.NewsDetail)
	secured.Get("/news/preview", contentHandler.NewsImage)
	secured.Get("/banners", catalogHandler.Banners)
	secured.Get("/banners/preview", catalogHandler.BannerImage)
	secured.Get("/vouchers", catalogHandler.Vouchers)
	secured.Get("/vouchers/detail", catalogHandler.VoucherDetail)
	secured.Get("/vouchers/preview", catalogHandler.VoucherImage)
	secured.Get("/vouchers/qr", catalogHandler.VoucherQR)
	secured.Get("/school-trainings", catalogHandler.Schools)
	secured.Get("/school-trainings/detail", catalogHandler.SchoolDetail)
	secured.Post("/documents/request-items/:itemCode/files", requestItemHandler.Upload)
	secured.Get("/documents/profile-items", requestItemHandler.List)
	secured.Get("/documents/profile-items/:itemCode/preview", requestItemHandler.Preview)
	secured.Get("/document-renewals/items/preview", requestItemHandler.RenewalPreview)
	secured.Get("/document-renewals/:requestNo/items/:documentRequestItemCode/preview", requestItemHandler.RenewalPreview)
	secured.Post("/document-renewals/:requestNo/items/:documentRequestItemCode/file", requestItemHandler.RenewalUpload)
	secured.Get("/send-noti", notificationHandler.Manual)
	secured.Get("/notifications", notificationHandler.List)
	secured.Post("/notifications/update", notificationHandler.MarkID)
	secured.Post("/notifications/update/valueId", notificationHandler.MarkValue)
	secured.Get("/notifications/update/all", notificationHandler.MarkAll)
	secured.Get("/documents/certification/COT", documentHandler.List("COT"))
	secured.Get("/documents/certification/DOC", documentHandler.List("Document"))
	secured.Get("/documents/certification/to-expiration", documentHandler.CloseToExpiration)
	secured.Post("/documents/certification/create", documentHandler.Create)
	secured.Post("/documents/certification/update", documentHandler.Update)
	secured.Delete("/documents/certification/delete", documentHandler.Delete)
	secured.Get("/documents/certification/edit", documentHandler.Edit)
	secured.Get("/documents/certification/view", documentHandler.View)
	for _, route := range protectedRoutes {
		secured.Add(route.method, route.path, pending(route.method, route.path))
	}
}

var protectedRoutes = []struct{ method, path string }{}

func pending(method, path string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return httpx.Error(c, fiber.StatusNotImplemented, "MIGRATION_PENDING", method+" "+path+" is not migrated")
	}
}
