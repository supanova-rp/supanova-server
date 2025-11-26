package objectstorage

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsCfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/cloudfront/sign"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/supanova-rp/supanova-server/internal/config"
)

const (
	URLExpiry = time.Hour * 6
	CDNExpiry = time.Hour * 2
)

type Store struct {
	client     *s3.Client
	bucketName string
	CDN        *CDN
}

type CDN struct {
	signer *sign.URLSigner
	domain string
}

func New(ctx context.Context, c *config.AWS) (*Store, error) {
	cfg, err := awsCfg.LoadDefaultConfig(
		ctx,
		awsCfg.WithRegion(c.Region),
		awsCfg.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				c.AccessKey,
				c.SecretKey,
				"",
			),
		))
	if err != nil {
		return nil, fmt.Errorf("failed to load aws config: %v", err)
	}

	cfKey, err := parseCDNKey()
	if err != nil {
		return nil, err
	}
	cfSigner := sign.NewURLSigner(c.CDNKeyPairID, cfKey)

	CDN := &CDN{
		domain: c.CDNDomain,
		signer: cfSigner,
	}

	return &Store{
		client:     s3.NewFromConfig(cfg),
		bucketName: c.BucketName,
		CDN:        CDN,
	}, nil
}

func (s *Store) GenerateUploadURL(ctx context.Context, key string, contentType *string) (string, error) {
	input := &s3.PutObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	}
	if contentType != nil {
		input.ContentType = contentType
	}

	presignClient := s3.NewPresignClient(s.client)

	req, err := presignClient.PresignPutObject(ctx, input, s3.WithPresignExpires(URLExpiry))
	if err != nil {
		return "", fmt.Errorf("failed to generate signed s3 URL: %v", err)
	}

	return req.URL, nil
}

func parseCDNKey() (*rsa.PrivateKey, error) {
	// TODO: remove this once AWS Secrets Manager logic is implemented
	if os.Getenv("ENVIRONMENT") == string(config.EnvironmentTest) {
		return nil, nil
	}

	const cfKeyPath = "./cloudfront_private_key.pem"
	cfKeyBytes, err := os.ReadFile(cfKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read CDN private key from %s: %v", cfKeyPath, err)
	}

	block, _ := pem.Decode(cfKeyBytes)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	cfKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return cfKey, nil
}

func (s *Store) GetCDNURL(ctx context.Context, key string) (string, error) {
	URL := fmt.Sprintf("https://%s/%s", s.CDN.domain, key)

	return s.CDN.signer.Sign(URL, time.Now().Add(CDNExpiry))
}
