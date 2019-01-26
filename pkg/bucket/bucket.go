package bucket

import (
	"errors"
	"io/ioutil"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	log "github.com/sirupsen/logrus"
)

func ReadFile(sess *session.Session, bucket string) (string, string, error) {
	log.WithFields(log.Fields{
		"bucket": bucket,
	}).Debug("Reading bucket")

	query := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
	}

	svc := s3.New(sess)
	resp, err := svc.ListObjectsV2(query)

	if err != nil {
		log.Error("Unable to query bucket")
		return "", "", err
	}

	if len(resp.Contents) < 1 {
		return "", "", errors.New("No files in bucket")
	}

	for _, key := range resp.Contents {

		log.WithFields(log.Fields{
			"file": key.Key,
		}).Debug("Reading File...")

		input := &s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    key.Key,
		}

		result, err := svc.GetObject(input)

		if err != nil {
			log.Error("Failed to get the file")
			return "", "", err
		}

		log.Info(result)

		body, err := ioutil.ReadAll(result.Body)
		if err != nil {
			log.Error("Unable to read bytes")
			return "", "", err
		}

		return string(body[:]), *key.Key, nil
		break
	}
	return "", "", nil
}

func DeleteObject(sess *session.Session, bucket string, key string) error {
	svc := s3.New(sess)
	_, err := svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		log.WithFields(log.Fields{
			"bucket": bucket,
			"key":    key,
		}).Error("Failed to delete")
		return err
	}

	err = svc.WaitUntilObjectNotExists(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		log.WithFields(log.Fields{
			"bucket": bucket,
			"key":    key,
		}).Error("Failed to delete")
		return err
	}

	log.WithFields(log.Fields{
		"bucket": bucket,
		"key":    key,
	}).Info("Successfully deleted")
	return nil
}

func UploadFile(sess *session.Session, bucket string, fileName string, body string) error {
	uploader := s3manager.NewUploader(sess)

	objectPath := "/content/post/" + fileName + ".md"

	fileReader := strings.NewReader(body)

	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(objectPath),
		Body:   fileReader,
	})

	if err != nil {
		log.Error("Failed to upload")
		return err
	}

	return nil
}
