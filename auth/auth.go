package auth

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"net/http"
	"os"
	"time"
)

type Claims struct {
	jwt.RegisteredClaims
}

func DefaultCookie(c *http.Cookie) {
	c.Value = ""
	c.Expires = time.Now().Add(time.Hour * (-1))

}
func CreateJWTCookieUser(w http.ResponseWriter, r *http.Request, userID uuid.UUID) error {
	claims := &Claims{jwt.RegisteredClaims{
		Issuer:    r.Host,
		Subject:   userID.String(),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}}

	/*	err := godotenv.Load()
		if err != nil {
			return err
		}*/
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt-token",
		Value:    tokenString,
		Expires:  time.Now().Add(time.Hour * 24),
		HttpOnly: true,
	})
	return nil
}
