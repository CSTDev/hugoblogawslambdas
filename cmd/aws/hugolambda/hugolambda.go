package main

import (
	"errors"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/cstdev/hugoblogawslambdas/pkg/hugo"
	"github.com/cstdev/lambdahelpers/pkg/bucket"
	"github.com/cstdev/lambdahelpers/pkg/s3/manager"
	log "github.com/sirupsen/logrus"
)

const tempDir = "/tmp/site/"

func handleRequest() (string, error) {
	srcBucket := os.Getenv("SRC_BUCKET")
	destBucket := os.Getenv("DEST_BUCKET")
	region := os.Getenv("REGION")
	host := os.Getenv("HOST")

	if srcBucket == "" {
		return "", errors.New("SRC_BUCKET environment variable must be set")
	}

	if destBucket == "" {
		return "", errors.New("DEST_BUCKET environment variable must be set")
	}

	if region == "" {
		return "", errors.New("REGION environment variable must be set")
	}

	if host == "" {
		return "", errors.New("HOST environment variable must be set to the full URL of dest bucket")
	}

	log.WithFields(log.Fields{
		"srcBucket":  srcBucket,
		"destBucket": destBucket,
		"host":       host,
	}).Info("Env variables")

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})

	if err != nil {
		log.Error("Failed to create session")
		return "", err
	}

	b := bucket.Bucket{
		Name:   srcBucket,
		Client: s3.New(sess),
		Manager: &manager.BucketManager{
			Uploader:   *s3manager.NewUploader(sess),
			Downloader: *s3manager.NewDownloader(sess),
		},
	}

	err = b.DownloadAllObjectsInBucket(tempDir, "public")
	if err != nil {
		return "", err
	}
	log.Info("Downloaded items from bucket.")

	err = hugo.Compile(host)
	if err != nil {
		return "", err
	}
	log.Info("Compiled Hugo site.")

	b.Name = destBucket
	err = b.Upload(tempDir + "public")
	if err != nil {
		return "", err
	}
	log.Info("Uploaded to site bucket.")

	return "Completed", nil

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

	lambda.Start(handleRequest)

}
