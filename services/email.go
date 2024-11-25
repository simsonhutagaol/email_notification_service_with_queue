package services

import (
	"fmt"
	"gopkg.in/gomail.v2"
	"log"
	"os"
	"strings"
	"io/ioutil"
	"html/template"
)

func SendWelcomeEmail(email string) error {
	log.Println("Sending email to:", email)

	// Ambil variabel dari environment
	smtpHost := os.Getenv("SMTP_HOST")
	smtpUserId := os.Getenv("SMTP_USER_ID")
	smtpUserEmail := os.Getenv("SMTP_USER_EMAIL")
	smtpPassword := os.Getenv("SMTP_PASSWORD")

	// Baca template HTML
	templatePath := "templates/welcome_email.html"
	templateContent, err := ioutil.ReadFile(templatePath)
	if err != nil {
		log.Printf("Failed to read template file: %v", err)
		return fmt.Errorf("failed to read template: %v", err)
	}

	tmpl, err := template.New("welcome").Parse(string(templateContent))
	if err != nil {
		log.Printf("Failed to parse template: %v", err)
		return fmt.Errorf("failed to parse template: %v", err)
	}

	// Data untuk template 
	data := struct {
		Name string
	}{
		Name: email,
	}

	var body strings.Builder
	err = tmpl.Execute(&body, data)
	if err != nil {
		log.Printf("Failed to execute template: %v", err)
		return fmt.Errorf("failed to execute template: %v", err)
	}

	// Kirim email menggunakan gomail
	msg := gomail.NewMessage()
	msg.SetHeader("From", smtpUserEmail)
	msg.SetHeader("To", email)
	msg.SetHeader("Subject", "Welcome to Our Service")
	msg.SetBody("text/html", body.String())

	// Mengirim email
	dialer := gomail.NewDialer(smtpHost, 2525, smtpUserId, smtpPassword)
	if err := dialer.DialAndSend(msg); err != nil {
		log.Printf("Failed to send email: %v", err)
		return fmt.Errorf("failed to send email: %v", err)
	}

	log.Printf("Email sent to: %s", email)
	return nil
}
