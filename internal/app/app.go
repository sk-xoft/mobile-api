package app

import (
	"context"
	"database/sql"
	payments "github.com/smartseaman/mobile-api/internal/domain/payment"
	paymentmysql "github.com/smartseaman/mobile-api/internal/platform/mysql/payment"
	paymentprovider "github.com/smartseaman/mobile-api/internal/platform/payment"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/smartseaman/mobile-api/internal/config"
	"github.com/smartseaman/mobile-api/internal/domain/auth"
	"github.com/smartseaman/mobile-api/internal/domain/catalog"
	"github.com/smartseaman/mobile-api/internal/domain/content"
	"github.com/smartseaman/mobile-api/internal/domain/delivery"
	"github.com/smartseaman/mobile-api/internal/domain/document"
	"github.com/smartseaman/mobile-api/internal/domain/eventlog"
	"github.com/smartseaman/mobile-api/internal/domain/fcm"
	"github.com/smartseaman/mobile-api/internal/domain/master"
	"github.com/smartseaman/mobile-api/internal/domain/notification"
	"github.com/smartseaman/mobile-api/internal/domain/profile"
	"github.com/smartseaman/mobile-api/internal/domain/renewal"
	"github.com/smartseaman/mobile-api/internal/domain/requestitem"
	"github.com/smartseaman/mobile-api/internal/middleware"
	"github.com/smartseaman/mobile-api/internal/platform/database"
	mail "github.com/smartseaman/mobile-api/internal/platform/email"
	authmysql "github.com/smartseaman/mobile-api/internal/platform/mysql/auth"
	catalogmysql "github.com/smartseaman/mobile-api/internal/platform/mysql/catalog"
	contentmysql "github.com/smartseaman/mobile-api/internal/platform/mysql/content"
	deliverymysql "github.com/smartseaman/mobile-api/internal/platform/mysql/delivery"
	documentmysql "github.com/smartseaman/mobile-api/internal/platform/mysql/document"
	eventlogmysql "github.com/smartseaman/mobile-api/internal/platform/mysql/eventlog"
	fcmmysql "github.com/smartseaman/mobile-api/internal/platform/mysql/fcm"
	mastermysql "github.com/smartseaman/mobile-api/internal/platform/mysql/master"
	notificationmysql "github.com/smartseaman/mobile-api/internal/platform/mysql/notification"
	profilemysql "github.com/smartseaman/mobile-api/internal/platform/mysql/profile"
	renewalmysql "github.com/smartseaman/mobile-api/internal/platform/mysql/renewal"
	requestitemmysql "github.com/smartseaman/mobile-api/internal/platform/mysql/requestitem"
	"github.com/smartseaman/mobile-api/internal/platform/storage"
	authhttp "github.com/smartseaman/mobile-api/internal/transport/http/auth"
	cataloghttp "github.com/smartseaman/mobile-api/internal/transport/http/catalog"
	contenthttp "github.com/smartseaman/mobile-api/internal/transport/http/content"
	deliveryhttp "github.com/smartseaman/mobile-api/internal/transport/http/delivery"
	documenthttp "github.com/smartseaman/mobile-api/internal/transport/http/document"
	fcmhttp "github.com/smartseaman/mobile-api/internal/transport/http/fcm"
	masterhttp "github.com/smartseaman/mobile-api/internal/transport/http/master"
	notificationhttp "github.com/smartseaman/mobile-api/internal/transport/http/notification"
	profilehttp "github.com/smartseaman/mobile-api/internal/transport/http/profile"
	renewalhttp "github.com/smartseaman/mobile-api/internal/transport/http/renewal"
	requestitemhttp "github.com/smartseaman/mobile-api/internal/transport/http/requestitem"
	"github.com/smartseaman/mobile-api/internal/transport/routes"
)

type Application struct {
	server       *fiber.App
	db           *sql.DB
	address      string
	stopPayments context.CancelFunc
	paymentsDone chan struct{}
}

func New(cfg config.Config) (*Application, error) {
	dbContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	db, err := database.Open(dbContext, cfg.Database)
	if err != nil {
		return nil, err
	}

	server := fiber.New(fiber.Config{DisableStartupMessage: true})
	server.Use(recover.New())
	eventLogService := eventlog.NewService(eventlogmysql.NewRepository(db))
	server.Use(middleware.RequestLogger(eventLogService))
	server.Use(cors.New(cors.Config{
		AllowOrigins:     join(cfg.Security.AllowedOrigins),
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "*",
		AllowCredentials: true,
	}))
	server.Use(func(c *fiber.Ctx) error {
		c.Set("X-Frame-Options", "SAMEORIGIN")
		c.Set("X-Content-Security-Policy", "script-src 'self'")
		c.Set("Strict-Transport-Security", "max-age=31536000 ; includeSubDomains")
		return c.Next()
	})
	objectStorage, err := storage.New(cfg.Storage)
	if err != nil {
		return nil, err
	}
	mailer := mail.NewSender(cfg.Email)
	authRepository := authmysql.NewRepository(db)
	authService := auth.NewService(authRepository, mailer, cfg.Security.JWTSecret)
	profileService := profile.NewService(profilemysql.NewRepository(db), objectStorage, cfg.Storage.ProfilePath)
	masterService := master.NewService(mastermysql.NewRepository(db))
	documentService := document.NewService(documentmysql.NewRepository(db), objectStorage, cfg.Storage.DocumentPath)
	deliveryService := delivery.NewService(deliverymysql.NewRepository(db))
	paymentService := payments.NewService(paymentmysql.NewStore(db), paymentprovider.NewOmise(cfg.Payment.APIURL, cfg.Payment.SecretKey, cfg.Payment.WebhookSecret, cfg.Payment.QRExpire), cfg.Payment.LiveMode)
	renewalService := renewal.NewService(renewalmysql.NewRepository(db, renewal.NewOmiseClient(cfg.Payment.QRExpire)), paymentService)
	contentService := content.NewService(contentmysql.NewRepository(db), objectStorage, cfg.Storage.FormPath, cfg.Storage.NewsPath)
	catalogService := catalog.NewService(catalogmysql.NewRepository(db), objectStorage, cfg.Storage.BannerPath, cfg.Storage.VoucherPath, cfg.Storage.VoucherQRPath)
	notificationService := notification.NewService(notificationmysql.NewRepository(db))
	requestItemService := requestitem.NewService(requestitemmysql.NewRepository(db), objectStorage, cfg.Storage.RequestItemPath)
	fcmService := fcm.NewService(fcmmysql.NewRepository(db))
	routes.Register(server, db, authhttp.NewHandler(authService, authRepository), profilehttp.NewHandler(profileService), masterhttp.NewHandler(masterService), documenthttp.NewHandler(documentService), deliveryhttp.NewHandler(deliveryService), fcmhttp.NewHandler(fcmService), renewalhttp.NewHandler(renewalService), contenthttp.NewHandler(contentService), cataloghttp.NewHandler(catalogService), notificationhttp.NewHandler(notificationService), requestitemhttp.NewHandler(requestItemService), cfg.Security.JWTSecret)

	application := &Application{server: server, db: db, address: cfg.App.Address, paymentsDone: make(chan struct{})}
	if cfg.Payment.WorkerEnabled && cfg.Payment.SecretKey != "" && db != nil {
		workerContext, cancel := context.WithCancel(context.Background())
		application.stopPayments = cancel
		go func() { defer close(application.paymentsDone); paymentService.Run(workerContext) }()
	} else {
		close(application.paymentsDone)
	}
	return application, nil
}

func (a *Application) Listen() error { return a.server.Listen(a.address) }
func (a *Application) Shutdown(ctx context.Context) error {
	err := a.server.ShutdownWithContext(ctx)
	if a.stopPayments != nil {
		a.stopPayments()
	}
	if a.paymentsDone == nil {
		return err
	}
	select {
	case <-a.paymentsDone:
	case <-ctx.Done():
		if err == nil {
			err = ctx.Err()
		}
	}
	return err
}
func (a *Application) Close() {
	if a.stopPayments != nil {
		a.stopPayments()
	}
	if a.paymentsDone != nil {
		<-a.paymentsDone
	}
	if a.db != nil {
		a.db.Close()
	}
}

func join(values []string) string {
	result := ""
	for index, value := range values {
		if index > 0 {
			result += ","
		}
		result += value
	}
	return result
}
