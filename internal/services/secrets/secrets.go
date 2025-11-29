package secrets

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type SecretsManager struct {
	client *secretsmanager.Client
}

func New(ctx context.Context, cfg *aws.Config) *SecretsManager {
	return &SecretsManager{
		client: secretsmanager.NewFromConfig(*cfg),
	}
}

func (s *SecretsManager) Get(ctx context.Context, name string) (string, error) {
	result, err := s.client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: &name,
	})
	if err != nil {
		return "", err
	}

	return *result.SecretString, nil
}
