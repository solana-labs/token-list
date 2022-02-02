package auth

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/go-github/v40/github"
	"golang.org/x/oauth2"
	"k8s.io/klog/v2"
	"time"
)

var ErrInvalidKey = errors.New("invalid key")

// signJWTFromPEM returns a signed JWT from a PEM-encoded private key.
func signJWTFromPEM(key []byte, appId int64) (string, error) {
	// decode PEM
	block, _ := pem.Decode(key)
	if block == nil {
		return "", ErrInvalidKey
	}

	// parse key
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return "", err
	}

	// sign
	return signJWT(privateKey, appId)
}

func signJWT(privateKey *rsa.PrivateKey, appId int64) (string, error) {
	// sign
	token := jwt.NewWithClaims(jwt.SigningMethodRS256,
		jwt.MapClaims{
			"exp": time.Now().Add(10 * time.Minute).Unix(),
			"iat": time.Now().Unix(),
			"iss": fmt.Sprintf("%d", appId),
		})

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func GetInstallationToken(privateKey []byte, appId int64, org string) (string, error) {
	token, err := signJWTFromPEM(privateKey, appId)
	if err != nil {
		klog.Exitf("failed to sign JWT: %v", err)
	}

	// get installation access token for app
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)


	var installations []*github.Installation
	var page int
	for {
		klog.Infof("getting installations for org %s, page %d", org, page)
		is, resp, err := client.Apps.ListInstallations(ctx, &github.ListOptions{
			Page:    page,
			PerPage: 100,
		})
		if err != nil {
			klog.Exitf("failed to list installations: %v", err)
		}
		installations = append(installations, is...)
		if resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
	}

	for _, i := range installations {
		if i.GetAppID() == appId && i.GetAccount().GetLogin() == org && i.GetTargetType() == "Organization" {
			klog.Infof("installation id: %v", i.GetID())
			klog.Infof("installed on %s: %s", i.GetTargetType(), i.GetAccount().GetLogin())

			// Get an installation token
			it, _, err := client.Apps.CreateInstallationToken(ctx, i.GetID(), nil)
			if err != nil {
				klog.Exitf("failed to create installation token: %v", err)
			}
			return it.GetToken(), nil
		}
	}

	return "", errors.New("no installation found")
}
