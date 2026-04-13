package config

import (
	"fmt"
	"os"
	"sync"
)

var (
	defaultConfig *Config
	once          sync.Once
	loadErr       error
)

type Config struct {
	WAAccessToken     string
	WAVerifyToken     string
	WAAppSecret       string
	S3SecretAccessKey string
	S3AccessKey       string
	S3Region          string
}

func Load() (*Config, error) {
	once.Do(func() {
		verifyToken := os.Getenv("WHATSAPP_VERIFY_TOKEN")
		if verifyToken == "" {
			loadErr = fmt.Errorf("WHATSAPP_VERIFY_TOKEN env can't be empty")
			return
		}

		accessToken := os.Getenv("WHATSAPP_ACCESS_TOKEN")
		if accessToken == "" {
			loadErr = fmt.Errorf("WHATSAPP_ACCESS_TOKEN env can't be empty")
			return
		}

		appSecret := os.Getenv("WHATSAPP_APP_SECRET")
		if appSecret == "" {
			loadErr = fmt.Errorf("WHATSAPP_APP_SECRET env can't be empty")
			return
		}

		s3SecretAccessToken := os.Getenv("S3_SECRET_ACCESS_KEY")
		if s3SecretAccessToken == "" {
			loadErr = fmt.Errorf("S3_SECRET_ACCESS_KEY env can't be empty")
			return
		}

		s3AccessKey := os.Getenv("S3_ACCESS_KEY")
		if s3AccessKey == "" {
			loadErr = fmt.Errorf("S3_ACCESS_KEY env can't be empty")
			return
		}

		s3Region := os.Getenv("S3_REGION")
		if s3Region == "" {
			loadErr = fmt.Errorf("S3_REGION env can't be empty")
			return
		}

		defaultConfig = &Config{
			WAVerifyToken:     verifyToken,
			WAAccessToken:     accessToken,
			WAAppSecret:       appSecret,
			S3SecretAccessKey: s3SecretAccessToken,
			S3AccessKey:       s3AccessKey,
			S3Region:          s3Region,
		}
	})

	return defaultConfig, loadErr
}
