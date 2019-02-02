package main

import (
	"errors"
	"os"

	"github.com/cstdev/lambdahelpers/pkg/bucket"
	"github.com/cstdev/lambdahelpers/pkg/mail"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	log "github.com/sirupsen/logrus"
)

func handleRequest() (string, error) {

	recipient := os.Getenv("RECIPIENT")
	sender := os.Getenv("SENDER")
	bucketName := os.Getenv("BUCKET")
	region := os.Getenv("REGION")

	if recipient == "" {
		return "", errors.New("RECIPIENT environment variable must be set")
	}

	if sender == "" {
		return "", errors.New("SENDER environment variable must be set")
	}

	if region == "" {
		return "", errors.New("REGION environment variable must be set")
	}

	if bucketName == "" {
		return "", errors.New("BUCKET environment variable must be set")
	}

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})

	if err != nil {
		return "", err
	}

	messageBody, objectKey, err := bucket.ReadFile(sess, bucketName)
	if err != nil {
		return "", err
	}

	err = mail.SendMail(sess, recipient, sender, messageBody)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Failed to send email")
		return "", err
	}

	err = bucket.DeleteObject(sess, bucketName, objectKey)

	return "", nil
}

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	logLevel := os.Getenv("LOG_LEVEL")
	switch logLevel {
	case "DEBUG":
		log.SetLevel(log.DebugLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}
	log.Info("Readsend")
	//lambda.Start(handleRequest)
	//handleRequest()

}
