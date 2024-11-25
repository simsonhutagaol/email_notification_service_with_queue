# Email Notification Service with Queue - Documentation

- This document outlines the technical details, requirements, and functionality of the Email Notification Service built using Go, GORM, Asynq, Redis, Gomail, and MySQL. The service is responsible for processing user registrations and sending welcome emails asynchronously through a queue system.

## Technologies Used

- Go (1.21 or later)
- GORM (ORM for Go to interact with MySQL)
- Redis (Queue storage system)
- Asynq (Queue and background task processing library)
- Gomail (Email service for sending emails)
- MySQL (Database for storing user data)
- Mailtrap (For email testing and debugging in development)
- bcrypt (Library for hashing passwords with the bcrypt algorithm, which is used to store passwords securely)

## API Endpoints

### POST /api/users/register

- Request Body

```json
{
  "email": "john.doe@example.com",
  "name": "John Doe",
  "password": "securePass123"
}
```

- Success Response (201 Created)

```json
{
  "id": 1,
  "email": "john.doe@example.com",
  "name": "John Doe",
  "created_at": "2024-11-12T10:30:00Z"
}
```

### Error Responses

- 400 Bad Request: If the email is already registered or the input is invalid.
- 500 Internal Server Error: If the system encounters an unexpected error during registration.
