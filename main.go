package main

import (
	"email-notification-service/config"
	"email-notification-service/handlers"
	"email-notification-service/middleware"
	"email-notification-service/queue"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/hibiken/asynq"
)

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/api/users/register", handlers.RegisterUser).Methods("POST")
	router.HandleFunc("/api/users/login", handlers.LoginUser).Methods("POST")
	router.Handle("/api/users/protec", middleware.AuthMiddleware(http.HandlerFunc(handlers.Protec))).Methods("GET")

	//antrean dan worker
	redisClient := config.ConnectToRedis()
	_, mux := queue.CreateEmailQueue()

	//server Asynq untuk memproses antrean
	server := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisClient.Options().Addr},
		asynq.Config{
			Concurrency: 10,
			RetryDelayFunc: func(n int, e error, t *asynq.Task) time.Duration {
				return 3 * time.Second
			},
		},
	)

	// Jalankan worker
	go func() {
		fmt.Println("Starting Asynq worker...")
		if err := server.Start(mux); err != nil {
			log.Fatalf("could not start worker: %v", err)
		}
	}()

	log.Println("Starting HTTP server on :8080...")
	log.Fatal(http.ListenAndServe(":8080", router))
}
