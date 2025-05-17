package auth

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var (
	ErrUserAuthFailed = errors.New("error during auth user")
)

const (
	HttpsPrefix   = "https://"
	LogonIndex    = "/Logon/Index"
	PrivateOffice = "/privateoffice"
	JwtCookieName = "JWToken"
)

type UserAuthClient struct {
	httpClient *http.Client
	url        string
}

func New(url string) *UserAuthClient {
	return &UserAuthClient{
		httpClient: &http.Client{
			Timeout: 15 * time.Second, CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		url: url,
	}
}

func (u *UserAuthClient) AuthStudent(ctx context.Context, login, password string) (jwt string, err error) {
	const op = "infra.auth.AuthStudent"

	data := fmt.Sprintf("login=%s&password=%s&GoogleAuthCode=", login, password)
	url := HttpsPrefix + u.url + LogonIndex

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBufferString(data))
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := u.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusFound {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("%s: %w", op, err)
		}

		if strings.Contains(string(body), PrivateOffice) {
			cookies := resp.Cookies()

			if len(cookies) == 0 {
				return "", fmt.Errorf("%s: %w", op, ErrUserAuthFailed)
			}

			var token string
			for _, cookie := range cookies {
				if cookie.Name == JwtCookieName {
					token = cookie.Value
					break
				}
			}

			if token == "" {
				return "", fmt.Errorf("%s: %w", op, ErrUserAuthFailed)
			}

			return token, nil
		}
	}

	return "", fmt.Errorf("%s: %w", op, ErrUserAuthFailed)
}

func (u *UserAuthClient) CheckToken(ctx context.Context, jwt string) (status bool, err error) {
	const op = "infra.auth.CheckToken"
	url := HttpsPrefix + u.url + PrivateOffice

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	req.AddCookie(&http.Cookie{
		Name:  JwtCookieName,
		Value: jwt,
	})

	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true, nil
	}

	return false, nil
}
