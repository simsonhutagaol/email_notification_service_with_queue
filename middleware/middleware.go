package middleware

import (
	"context"
	"email-notification-service/utils"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt"
	"github.com/joho/godotenv"
)

var jwtKey []byte

func init() {
	// Memuat variabel dari file .env
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	// Ambil secret dari variabel lingkungan
	jwtKey = []byte(os.Getenv("JWT_SECRET"))
}

// AuthMiddleware memverifikasi token JWT
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Ambil token dari header Authorization
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			utils.JsonResponse(w, http.StatusUnauthorized, map[string]string{"message": "Authorization header missing"})
			return
		}

		// Hapus prefix "Bearer "
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")

		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) { //
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			utils.JsonResponse(w, http.StatusUnauthorized, map[string]string{"message": "Unauthorized"})
			return
		}

		fmt.Println(claims)

		// Ambil data dari claims dan pastikan ada
		id, idOk := claims["id"].(float64) // JWT MapClaims menyimpan angka dalam float64, bukan int
		email, emailOk := claims["email"].(string)
		name, nameOk := claims["name"].(string)

		if !idOk || !emailOk || !nameOk {
			utils.JsonResponse(w, http.StatusUnauthorized, map[string]string{"message": "Invalid token claims"})
			return
		}

		// Konversi ID menjadi tipe yang sesuai (misalnya uint)
		idUint := uint(id) // Mengonversi dari float64 ke uint jika perlu

		// Tambahkan data ke context
		ctx := context.WithValue(r.Context(), "id", idUint)
		ctx = context.WithValue(ctx, "email", email)
		ctx = context.WithValue(ctx, "name", name)

		// Teruskan ke handler berikutnya dengan context yang diperbarui
		next.ServeHTTP(w, r.WithContext(ctx))
		// next.ServeHTTP(w, r)
	})
}
