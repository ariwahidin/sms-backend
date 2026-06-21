package middleware

import (
	"github.com/gofiber/fiber/v2"

	"sms-backend/model"
	"sms-backend/service"
)

// RequireRole allows only specified roles to proceed
func RequireRole(roles ...model.Role) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims, ok := c.Locals(LocalsUserKey).(*service.JWTClaims)
		if !ok || claims == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		for _, role := range roles {
			if claims.Role == role {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error":          "Forbidden: insufficient role",
			"your_role":      string(claims.Role),
			"required_roles": rolesToStrings(roles),
		})
	}
}

func rolesToStrings(roles []model.Role) []string {
	out := make([]string, len(roles))
	for i, r := range roles {
		out[i] = string(r)
	}
	return out
}

// GetClaims helper to extract claims from context
func GetClaims(c *fiber.Ctx) *service.JWTClaims {
	claims, _ := c.Locals(LocalsUserKey).(*service.JWTClaims)
	return claims
}
