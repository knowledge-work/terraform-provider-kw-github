package githubclient

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/go-github/v74/github"
)

type Client struct {
	*github.Client
	Owner string
}

func NewClient(token, baseURL, owner string) *Client {
	var tc *http.Client
	if token != "" {
		tc = github.NewClient(nil).WithAuthToken(token).Client()
	}

	client, _ := github.NewClient(tc).WithEnterpriseURLs(baseURL, baseURL)
	return &Client{client, owner}
}

func NewClientWithApp(appID, installationID, pemFile, baseURL, owner string) *Client {
	token, err := GenerateOAuthTokenFromApp(baseURL, appID, installationID, pemFile)
	if err != nil {
		return nil
	}

	tc := github.NewClient(nil).WithAuthToken(token).Client()
	client, _ := github.NewClient(tc).WithEnterpriseURLs(baseURL, baseURL)
	return &Client{client, owner}
}

func GenerateOAuthTokenFromApp(baseURL, appID, appInstallationID, pemData string) (string, error) {
	appJWT, err := generateAppJWT(appID, time.Now(), []byte(pemData))
	if err != nil {
		return "", err
	}

	token, err := getInstallationAccessToken(baseURL, appJWT, appInstallationID)
	if err != nil {
		return "", err
	}

	return token, nil
}

func generateAppJWT(appID string, issuedAt time.Time, privateKeyPEM []byte) (string, error) {
	block, _ := pem.Decode(privateKeyPEM)
	if block == nil {
		return "", errors.New("failed to parse PEM block containing the key")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		parsedKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return "", fmt.Errorf("failed to parse private key: %v", err)
		}
		var ok bool
		privateKey, ok = parsedKey.(*rsa.PrivateKey)
		if !ok {
			return "", errors.New("private key is not RSA")
		}
	}

	appIDInt, err := strconv.Atoi(appID)
	if err != nil {
		return "", fmt.Errorf("invalid app ID: %v", err)
	}

	claims := jwt.RegisteredClaims{
		Issuer:    strconv.Itoa(appIDInt),
		IssuedAt:  jwt.NewNumericDate(issuedAt),
		ExpiresAt: jwt.NewNumericDate(issuedAt.Add(10 * time.Minute)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %v", err)
	}

	return tokenString, nil
}

func getInstallationAccessToken(baseURL, appJWT, installationID string) (string, error) {
	tc := &http.Client{
		Transport: &jwtTransport{
			token: appJWT,
			rt:    http.DefaultTransport,
		},
		Timeout: 30 * time.Second,
	}

	client, err := github.NewClient(tc).WithEnterpriseURLs(baseURL, baseURL)
	if err != nil {
		return "", fmt.Errorf("failed to create github client: %v", err)
	}

	installationIDInt, err := strconv.ParseInt(installationID, 10, 64)
	if err != nil {
		return "", fmt.Errorf("invalid installation ID: %v", err)
	}

	token, _, err := client.Apps.CreateInstallationToken(
		context.Background(),
		installationIDInt,
		&github.InstallationTokenOptions{},
	)
	if err != nil {
		return "", fmt.Errorf("failed to create installation token: %v", err)
	}

	return token.GetToken(), nil
}

type jwtTransport struct {
	token string
	rt    http.RoundTripper
}

func (t *jwtTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+t.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "terraform-provider-kw-github")
	return t.rt.RoundTrip(req)
}
