package utils

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
)

var (
	S3           *s3.Client
	MediaBucket  string
	CoversBucket string
)

func newS3Client(ctx context.Context) (*s3.Client, error) {
	region := os.Getenv("S3_REGION")
	if region == "" {
		region = "us-east-1"
	}

	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			os.Getenv("S3_ACCESS_KEY_ID"),
			os.Getenv("S3_SECRET_ACCESS_KEY"),
			"",
		)),
	)
	if err != nil {
		return nil, err
	}

	return s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(os.Getenv("S3_ENDPOINT"))
		o.UsePathStyle = true
	}), nil
}

func ensureBucket(ctx context.Context, client *s3.Client, bucket string) error {
	_, err := client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(bucket)})
	if err == nil {
		return nil
	}

	var apiErr smithy.APIError
	if !errors.As(err, &apiErr) {
		return err
	}
	if apiErr.ErrorCode() != "NotFound" && apiErr.ErrorCode() != "NoSuchBucket" {
		return err
	}

	_, err = client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(bucket)})
	return err
}

func InitStorage() {
	ctx := context.Background()

	MediaBucket = os.Getenv("S3_BUCKET_MEDIA")
	CoversBucket = os.Getenv("S3_BUCKET_COVERS")

	client, err := newS3Client(ctx)
	CheckError(err)

	CheckError(ensureBucket(ctx, client, MediaBucket))
	CheckError(ensureBucket(ctx, client, CoversBucket))

	S3 = client
	log.Println("Initialized S3 storage")
}

func PutObject(ctx context.Context, bucket, key string, body []byte, contentType string) error {
	_, err := S3.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(body),
		ContentType: aws.String(contentType),
	})
	return err
}

func GetObject(ctx context.Context, bucket, key string) ([]byte, error) {
	out, err := S3.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	defer out.Body.Close()

	return io.ReadAll(out.Body)
}

func DeleteObject(ctx context.Context, bucket, key string) error {
	_, err := S3.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	return err
}

func IsNotFound(err error) bool {
	var apiErr smithy.APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	code := apiErr.ErrorCode()
	return code == "NoSuchKey" || code == "NotFound"
}
