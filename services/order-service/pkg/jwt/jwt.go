package jwt

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

// Claims representa as claims do JWT token
// Deve ser idêntico ao User Service para compatibilidade
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// ValidateToken valida um token JWT e retorna as claims
// secret: o mesmo JWT_SECRET usado pelo User Service
// tokenString: o token JWT a ser validado (sem o prefixo "Bearer ")
func ValidateToken(tokenString, secret string) (*Claims, error) {
	// Parse o token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validar algoritmo de assinatura
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("método de assinatura inválido: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("erro ao fazer parse do token: %w", err)
	}

	// Verificar se o token é válido
	if !token.Valid {
		return nil, fmt.Errorf("token inválido")
	}

	// Extrair claims
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("erro ao extrair claims do token")
	}

	return claims, nil
}

// ExtractUserID é uma função auxiliar que extrai apenas o UserID do token
func ExtractUserID(tokenString, secret string) (string, error) {
	claims, err := ValidateToken(tokenString, secret)
	if err != nil {
		return "", err
	}
	return claims.UserID, nil
}
