package main

import (
	"errors"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/cstdev/recieve-hugo-email/pkg/bucket"
	"github.com/cstdev/recieve-hugo-email/pkg/mail"
	log "github.com/sirupsen/logrus"
)

func handleRequest() (string, error) {
	bucketName := os.Getenv("BUCKET")
	region := os.Getenv("REGION")
	siteBucket := os.Getenv("SITE_BUCKET")

	if region == "" {
		return "", errors.New("REGION environment variable must be set")
	}

	if bucketName == "" {
		return "", errors.New("BUCKET environment variable must be set")
	}

	if siteBucket == "" {
		return "", errors.New("SITE_BUCKET environment variable must be set")
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

	parsedBody := mail.ParseBody(messageBody)

	log.WithFields(log.Fields{
		"key":  objectKey,
		"body": parsedBody,
	}).Debug("Message")

	err = bucket.UploadFile(sess, siteBucket, parsedBody.Subject, parsedBody.Body)
	if err != nil {
		return "", err
	}

	err = bucket.DeleteObject(sess, bucketName, objectKey)
	if err != nil {
		log.Error("Unable to delete")
		return "", nil
	}

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

	//lambda.Start(handleRequest)
	_, err := handleRequest()
	if err != nil {
		log.Error(err)
	}
}
