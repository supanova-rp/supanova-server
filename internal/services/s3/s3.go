package s3

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	awsS3 "github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Cfg struct {
	BucketName          string
	Region              string
	AccessKeyId         string
	SecretAccessKey     string
	CloudfrontDomain    string
	CloudfrontKeyPairID string
}

type S3 struct {
	s3Client   *awsS3.Client
	bucketName string
	cfClient   *cloudfront.Client
	cfSigner   *cloudfront.Signer
}

func New(ctx context.Context, cfg *S3Cfg) (*S3, error) {
	// TODO: pass config with the options pattern e.g. with xyz
	cfg, err := config.LoadDefaultConfig(ctx, cfg)
	if err != nil {
		slog.Error("failed to load aws config", slog.Any("err", err))
		return nil, err
	}

	cfFile, err := os.Open("./cloudfront_private_key.pem")
	if err != nil {
		slog.Error("failed to load cloudfront private key file", slog.Any("err", err))
	}

	cfKey, err := io.ReadAll(cfFile)
	if err != nil {
		slog.Error("failed to parse cloudfront private key", slog.Any("err", err))
	}

	return &S3{
		s3Client: awsS3.NewFromConfig(cfg),
		cfClient: cloudfront.NewFromConfig(cfg),
		//TODO: create a signer and pass the cfKey
	}, nil
}

func (s *S3) GenerateUploadURL(ctx context.Context, key string, contentType *string) (string, error) {
	input := &s3.PutObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	}

	if contentType != nil {
		input.ContentType = contentType
	}

	_, err := s.s3Client.PutObject(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to upload file key to s3: %v", err)
	}

	presignClient := s3.NewPresignClient(s.s3Client)

	// TODO: can you do this without repeating bucket: aws.string() etc?
	req, err := presignClient.PresignGetObject(ctx, &awsS3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(time.Hour*6))
	if err != nil {
		return "", fmt.Errorf("failed to generate signed s3 URL: %v", err)
	}

	return req.URL, nil
}

// TODO: GetCloudfrontURL
