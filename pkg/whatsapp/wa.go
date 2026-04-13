package whatsapp

type WAClient struct {
	WAAccessToken string
	WAVerifyToken string
	WAAppSecret   string
}

func New(accessToken string, verifyToken string, appSecret string) *WAClient {
	return &WAClient{
		WAAccessToken: accessToken,
		WAVerifyToken: verifyToken,
		WAAppSecret:   appSecret,
	}
}
