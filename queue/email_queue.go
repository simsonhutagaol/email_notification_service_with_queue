package queue

import (
    "context"
    "fmt"
    "log"
    "email-notification-service/services"
    "email-notification-service/config"
    "github.com/hibiken/asynq"
    "encoding/json"
)

const (
    // Task name 
    TaskSendWelcomeEmail = "send:welcome_email"
)

type EmailPayload struct {
    Email string `json:"email"`
}

func ProcessSendWelcomeEmail(ctx context.Context, task *asynq.Task) error {
    var payload EmailPayload
    err := json.Unmarshal(task.Payload(), &payload)
    if err != nil {
        return fmt.Errorf("could not unmarshal payload: %v", err)
    }
    
    err = services.SendWelcomeEmail(payload.Email)
    if err != nil {
        log.Printf("Failed to send welcome email: %v", err)
        return err
    }

    log.Printf("Welcome email sent to: %s", payload.Email)
    return nil
}

//antrean email
func CreateEmailQueue() (*asynq.Client, *asynq.ServeMux) {
    redisClient := config.ConnectToRedis()
    client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisClient.Options().Addr})

    mux := asynq.NewServeMux()
    mux.HandleFunc(TaskSendWelcomeEmail, ProcessSendWelcomeEmail)

    return client, mux
}
