package objectstorage

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/cloudfront/sign"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/supanova-rp/supanova-server/internal/config"
)

const (
	URLExpiry = time.Hour * 12
	CDNExpiry = time.Hour * 12
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

func New(ctx context.Context, cfg *config.AWS, awsCfg *aws.Config, cdnKey string) (*Store, error) {
	parsedCDNKey, err := parseCDNKey(cdnKey)
	if err != nil {
		return nil, err
	}

	CDNSigner := sign.NewURLSigner(cfg.CDNKeyPairID, parsedCDNKey)
	CDN := &CDN{
		domain: cfg.CDNDomain,
		signer: CDNSigner,
	}

	return &Store{
		client:     s3.NewFromConfig(*awsCfg),
		bucketName: cfg.BucketName,
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

// Copy duplicates an object within the bucket. S3 does the copy itself, so nothing is downloaded.
func (s *Store) Copy(ctx context.Context, sourceKey, destinationKey string) error {
	_, err := s.client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(s.bucketName),
		CopySource: aws.String(s.bucketName + "/" + url.PathEscape(sourceKey)),
		Key:        aws.String(destinationKey),
	})
	if err != nil {
		return fmt.Errorf("failed to copy s3 object %s to %s: %w", sourceKey, destinationKey, err)
	}

	return nil
}

func parseCDNKey(pemKey string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemKey))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return key, nil
}

func (s *Store) GetCDNURL(ctx context.Context, key string) (string, error) {
	URL := fmt.Sprintf("https://%s/%s", s.CDN.domain, key)

	return s.CDN.signer.Sign(URL, time.Now().Add(CDNExpiry))
}
