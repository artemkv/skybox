package main

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
)

var (
	PART_SIZE             int64 = 5 * 1024 * 1024
	EXISTS_WAITER_TIMEOUT       = time.Minute
)

// Uploads file with metadata
// Waits until available
func UploadFile(bucket string, objectKey string, reader io.Reader, meta map[string]string) error {
	// Setup client
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return err
	}
	s3client := s3.NewFromConfig(cfg)

	// Setup uploader
	manager := transfermanager.New(s3client, func(u *transfermanager.Options) {
		u.PartSizeBytes = PART_SIZE
	})

	// Initialize input
	input := &transfermanager.UploadObjectInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(objectKey),
		Body:     reader,
		Metadata: meta,
	}

	// Upload file
	_, err = manager.UploadObject(context.TODO(), input)
	if err != nil {
		return err
	}

	// Wait for finished
	waitInput := &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(objectKey),
	}
	err = s3.NewObjectExistsWaiter(s3client).Wait(
		context.TODO(), waitInput, EXISTS_WAITER_TIMEOUT)
	if err != nil {
		return err
	}

	return nil
}

// If no file found, returns nil
func GetFileContent(bucket string, objectKey string) ([]byte, error) {
	// Setup client
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}
	s3client := s3.NewFromConfig(cfg)

	// Initialize input
	input := &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &objectKey,
	}

	// Fetch the content
	output, err := s3client.GetObject(context.TODO(), input)
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			if apiErr.ErrorCode() == "NoSuchKey" {
				return nil, nil
			}
		}
		return nil, err
	}

	// Process the output
	defer output.Body.Close()
	bytes, err := io.ReadAll(output.Body)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func DeleteFile(bucket string, objectKey string) error {
	// Setup client
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return err
	}
	s3client := s3.NewFromConfig(cfg)

	// Initialize input
	input := &s3.DeleteObjectInput{
		Bucket: &bucket,
		Key:    &objectKey,
	}

	// Delete the file
	_, err = s3client.DeleteObject(context.TODO(), input)
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			if apiErr.ErrorCode() == "NoSuchKey" {
				return nil
			}
		}
		return err
	}

	return nil
}
