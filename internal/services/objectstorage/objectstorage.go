package objectstorage

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
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

func New(ctx context.Context, customCfg *config.AWS, AWSCfg *aws.Config, CDNKey string) (*Store, error) {
	parsedCDNKey, err := parseCDNKey(CDNKey)
	if err != nil {
		return nil, err
	}

	CDNSigner := sign.NewURLSigner(customCfg.CDNKeyPairID, parsedCDNKey)
	CDN := &CDN{
		domain: customCfg.CDNDomain,
		signer: CDNSigner,
	}

	return &Store{
		client:     s3.NewFromConfig(*AWSCfg),
		bucketName: customCfg.BucketName,
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
