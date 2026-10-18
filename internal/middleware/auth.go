package middleware

import (
	"database/sql"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/smartseaman/mobile-api/internal/transport/httpx"
)

const (
	JWTSubject = "jwt.subject"
	JWTID      = "jwt.id"
)

func Authenticate(secret string, db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authorization := c.Get(fiber.HeaderAuthorization)
		if authorization == "" {
			return httpx.Error(c, fiber.StatusUnauthorized, "MA00004", "Authorization")
		}
		if !strings.HasPrefix(authorization, "Bearer ") {
			return httpx.Error(c, fiber.StatusUnauthorized, "MA00011", "Authorization")
		}

		token, err := jwt.Parse(strings.TrimPrefix(authorization, "Bearer "), func(token *jwt.Token) (any, error) {
			return []byte(secret), nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
		if err != nil || !token.Valid {
			return httpx.Error(c, fiber.StatusUnauthorized, "MA00007", "Invalid token")
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return httpx.Error(c, fiber.StatusUnauthorized, "MA00007", "Invalid token")
		}
		subject, err := claims.GetSubject()
		if err != nil || subject == "" {
			return httpx.Error(c, fiber.StatusUnauthorized, "MA00007", "Invalid token")
		}
		c.Locals(JWTSubject, subject)
		if id, ok := claims["jti"].(string); ok {
			c.Locals(JWTID, id)
			var sessionUserID, online string
			err := db.QueryRowContext(c.UserContext(),
				"SELECT USER_ID, IS_ONLINE FROM t_session WHERE CLIENT_SESSION_ID = ? LIMIT 1", id,
			).Scan(&sessionUserID, &online)
			if err != nil || online != "YES" {
				return httpx.Error(c, fiber.StatusUnauthorized, "MA00026", "Incorrect client session id.")
			}
			var databaseUsername string
			err = db.QueryRowContext(c.UserContext(),
				"SELECT USERNAME FROM m_mobile_users WHERE MOBILE_UUID = ? LIMIT 1", sessionUserID,
			).Scan(&databaseUsername)
			if err != nil || databaseUsername != subject {
				return httpx.Error(c, fiber.StatusUnauthorized, "MA00016", "Email is not found.")
			}
			c.Locals("user.mobile_uuid", sessionUserID)
			c.Locals("user.username", databaseUsername)
		} else {
			return httpx.Error(c, fiber.StatusUnauthorized, "MA00018", "Authorization.")
		}
		return c.Next()
	}
}
