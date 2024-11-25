package handlers

import (
	"email-notification-service/config"
	"email-notification-service/models"
	"email-notification-service/queue"
	"email-notification-service/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

type CustomClaims struct {
	ID    uint   `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	jwt.StandardClaims
}

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

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, map[string]string{"message": "Error hashing password"})
		return
	}
	user.Password = string(hashedPassword)

	db := config.ConnectToDB()
	defer db.Close() //artinya akan dijalankan di akhir fungsi

	// Cek email sudah ada
	var existingUser models.User
	db.Where("email = ?", user.Email).First(&existingUser) // SELECT * FROM users WHERE email = user.Email LIMIT 1
	if existingUser.ID != 0 {                              // jika user ditemukan di database
		utils.JsonResponse(w, http.StatusBadRequest, map[string]string{"message": "Email already exists"})
		return
	}

	if err := db.Create(&user).Error; err != nil { // jika terjadi error saat menyimpan user ke database
		utils.JsonResponse(w, http.StatusInternalServerError, map[string]string{"message": "Failed to register user"})
		return
	}

	// Kirim task untuk antrean email
	emailPayload, err := json.Marshal(user) // mengubah struct user menjadi JSON
	if err != nil {                         // jika gagal mengubah struct user menjadi JSON
		http.Error(w, "Failed to marshal email payload", http.StatusInternalServerError)
		return
	}

	redisClient := config.ConnectToRedis() //menghubungkan ke Redis dan di assign ke redisClient
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisClient.Options().Addr})
	task := asynq.NewTask(queue.TaskSendWelcomeEmail, emailPayload)

	//add task ke antrean
	_, err = client.Enqueue(task, asynq.MaxRetry(3)) // artinya akan diulang 3 kali jika gagal
	if err != nil {                                  // nil artinya tidak ada error
		http.Error(w, "Failed to enqueue email task", http.StatusInternalServerError)
		return
	}

	//hapus key password dari response
	response := map[string]string{
		"id":      fmt.Sprintf("%d", user.ID),
		"email":   user.Email,
		"name":    user.Name,
		"message": "Register success",
	}

	utils.JsonResponse(w, http.StatusCreated, response)

}

func LoginUser(w http.ResponseWriter, r *http.Request) { //w http.ResponseWriter adalah untuk menulis response ke client, dan r *http.Request adalah untuk membaca request dari client
	var user models.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	db := config.ConnectToDB()
	defer db.Close()

	var existingUser models.User
	db.Where("email = ?", user.Email).First(&existingUser)
	if existingUser.ID == 0 {
		utils.JsonResponse(w, http.StatusNotFound, map[string]string{"message": "User not found"})
		return
	}
	fmt.Println(existingUser.Password)
	fmt.Println(user.Password)

	// Bandingkan password
	err = bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(user.Password))
	if err != nil {
		utils.JsonResponse(w, http.StatusUnauthorized, map[string]string{"message": "Invalid password/email"})
		return
	}
	// Buat klaim JWT
	claims := &CustomClaims{
		ID:    existingUser.ID,
		Email: existingUser.Email,
		Name:  existingUser.Name,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(), // Token berlaku selama 24 jam
			Subject:   fmt.Sprintf("%d", existingUser.ID),    // Subject adalah ID pengguna
		},
	}

	// Buat token JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		utils.JsonResponse(w, http.StatusInternalServerError, map[string]string{"message": "Could not generate token"})
		return
	}

	// Kirim token sebagai respons
	respon := map[string]string{
		"name":    existingUser.Name,
		"token":   tokenString,
		"message": "Login success",
	}

	utils.JsonResponse(w, http.StatusOK, respon)
}

func Protec(w http.ResponseWriter, r *http.Request) {
	// Ambil data dari context
	id, idOk := r.Context().Value("id").(uint) // Ambil id sebagai uint sesuai dengan yang disimpan di middleware
	email, emailOk := r.Context().Value("email").(string)
	name, nameOk := r.Context().Value("name").(string)

	// Pastikan semua data ada di context
	if !idOk || !emailOk || !nameOk {
		utils.JsonResponse(w, http.StatusUnauthorized, map[string]string{"message": "Unauthorized"})
		return
	}

	utils.JsonResponse(w, http.StatusOK, map[string]string{
		"message": "Protected Profile",
		"id":      fmt.Sprintf("%d", id), // Convert id ke string jika perlu
		"email":   email,
		"name":    name,
	})
}
