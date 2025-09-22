package githubclient

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-jose/go-jose/v3"
	"github.com/go-jose/go-jose/v3/jwt"
	"github.com/google/go-github/v74/github"
)

type Client struct {
	*github.Client
}

func NewClient(token, baseURL string) *Client {
	var tc *http.Client
	if token != "" {
		tc = github.NewClient(nil).WithAuthToken(token).Client()
	}

	client, _ := github.NewClient(tc).WithEnterpriseURLs(baseURL, baseURL)
	return &Client{client}
}

func NewClientWithApp(appID, installationID, pemFile, baseURL string) *Client {
	token, err := GenerateOAuthTokenFromApp(baseURL, appID, installationID, pemFile)
	if err != nil {
		return nil
	}

	tc := github.NewClient(nil).WithAuthToken(token).Client()
	client, _ := github.NewClient(tc).WithEnterpriseURLs(baseURL, baseURL)
	return &Client{client}
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

	claims := jwt.Claims{
		Issuer:   strconv.Itoa(appIDInt),
		IssuedAt: jwt.NewNumericDate(issuedAt),
		Expiry:   jwt.NewNumericDate(issuedAt.Add(10 * time.Minute)),
	}

	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: privateKey}, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create signer: %v", err)
	}

	token, err := jwt.Signed(signer).Claims(claims).CompactSerialize()
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %v", err)
	}

	return token, nil
}

func getInstallationAccessToken(baseURL, appJWT, installationID string) (string, error) {
	url := fmt.Sprintf("%s/app/installations/%s/access_tokens", baseURL, installationID)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+appJWT)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "terraform-provider-kw-github")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to get access token: %s - %s", resp.Status, string(body))
	}

	var tokenResp struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	return tokenResp.Token, nil
}
