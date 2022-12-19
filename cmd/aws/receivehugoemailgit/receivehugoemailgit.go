package main

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/cstdev/lambdahelpers/pkg/mail"
	"github.com/cstdev/lambdahelpers/pkg/storage"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	log "github.com/sirupsen/logrus"
)

const tempDir = "/tmp/site"
const postDir = "content/post/"

func handleRequest() (string, error) {
	bucketName := os.Getenv("BUCKET")
	region := os.Getenv("REGION")
	gitRepository := os.Getenv("GIT_REPO")
	gitToken := os.Getenv("GIT_TOKEN")
	gitAuthor := os.Getenv("GIT_AUTHOR")
	gitEmail := os.Getenv("GIT_EMAIL")

	if region == "" {
		return "", errors.New("REGION environment variable must be set")
	}

	if bucketName == "" {
		return "", errors.New("BUCKET environment variable must be set")
	}

	if gitRepository == "" {
		return "", errors.New("GIT_REPO environment variable must be set")
	}

	if gitToken == "" {
		return "", errors.New("GIT_TOKEN environment variable must be set")
	}

	if gitAuthor == "" {
		return "", errors.New("GIT_AUTHOR environment variable must be set")
	}

	if gitEmail == "" {
		return "", errors.New("GIT_EMAIL environment variable must be set")
	}

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})

	if err != nil {
		log.Error("failed to connect to s3")
		return "", err
	}

	b := storage.Bucket{
		Client: s3.New(sess),
		Name:   bucketName,
	}

	messageBody, objectKey, err := b.ReadFile()
	if err != nil {
		log.Error("failed to read file")
		return "", err
	}

	parsedBody := mail.ParseBody(messageBody)

	log.WithFields(log.Fields{
		"key":  objectKey,
		"body": parsedBody,
	}).Debug("Message")

	// Checkout repo
	
	r, err := git.PlainClone(tempDir, false, &git.CloneOptions{
		URL: gitRepository,
		Progress: os.Stdout,
		Auth: &http.BasicAuth {
			Username: "x-token-auth",
			Password: gitToken,
		},
	})

	if err != nil {
		log.WithField("repository", gitRepository).Error("failed to clone repository")
		return "", err
	}

	w, err := r.Worktree()
	if err != nil {
		return "", err
	}

	// Write file to /content/posts
	fileName := parsedBody.Subject + ".md"
	filePath := filepath.Join(tempDir, postDir, fileName)
	fileContent := parsedBody.Body
	err = os.WriteFile(filePath, []byte(fileContent), 0644)
	if err != nil {
		log.Error("failed to write file to commit")
		return "", err
	}

	// Commit
	_, err = w.Add(filepath.Join(postDir,fileName))
	if err != nil {
		log.Error("failed to add file to work tree")
		return "", err
	}

	_, err = w.Commit("Added post " + parsedBody.Subject, &git.CommitOptions{
		Author: &object.Signature {
			Name: gitAuthor,
			Email: gitEmail,
			When: time.Now(),
		},
	})
	if err != nil {
		log.Error("failed to commit file")
		return "", err
	}
	// Push
	err = r.Push(&git.PushOptions{
		Auth: &http.BasicAuth {
			Username: "x-token-auth",
			Password: gitToken,
		},
	})
	if err != nil {
		log.Error("failed to push commit")
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
