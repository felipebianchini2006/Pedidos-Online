package middleware

import (
	"strings"

	customjwt "pedidos-online/order-service/pkg/jwt"

	"github.com/gofiber/fiber/v2"
)

// AuthMiddleware valida o token JWT nas requisições
func AuthMiddleware(jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Obter o header Authorization
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Token não fornecido",
			})
		}

		// Extrair o token (formato: "Bearer <token>")
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Formato de token inválido. Use: Bearer <token>",
			})
		}

		tokenString := parts[1]

		// Validar o token
		claims, err := customjwt.ValidateToken(tokenString, jwtSecret)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Token inválido ou expirado",
			})
		}

		// Armazenar as claims no contexto para uso posterior
		c.Locals("user_id", claims.UserID)
		c.Locals("email", claims.Email)

		return c.Next()
	}
}

// GetUserID extrai o user_id do contexto (após passar pelo AuthMiddleware)
func GetUserID(c *fiber.Ctx) string {
	userID, ok := c.Locals("user_id").(string)
	if !ok {
		return ""
	}
	return userID
}

// GetUserEmail extrai o email do contexto (após passar pelo AuthMiddleware)
func GetUserEmail(c *fiber.Ctx) string {
	email, ok := c.Locals("email").(string)
	if !ok {
		return ""
	}
	return email
}
