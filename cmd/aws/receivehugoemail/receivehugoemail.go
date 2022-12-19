package main

import (
	"errors"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/cstdev/lambdahelpers/pkg/mail"
	"github.com/cstdev/lambdahelpers/pkg/s3/manager"
	"github.com/cstdev/lambdahelpers/pkg/storage"
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

	b := storage.Bucket{
		Client: s3.New(sess),
		Name:   bucketName,
	}

	destBucket := storage.Bucket{
		Name: siteBucket,
		Manager: &manager.BucketManager{
			Uploader: *s3manager.NewUploader(sess),
		},
	}

	messageBody, objectKey, err := b.ReadFile()
	if err != nil {
		return "", err
	}

	parsedBody := mail.ParseBody(messageBody)

	log.WithFields(log.Fields{
		"key":  objectKey,
		"body": parsedBody,
	}).Debug("Message")

	err = destBucket.UploadFile(parsedBody.Subject, parsedBody.Body)
	if err != nil {
		return "", err
	}

	err = b.DeleteObject(objectKey)
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

	log.Info("Receive Email")
	lambda.Start(handleRequest)
}
