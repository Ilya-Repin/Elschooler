package fetcher

import (
	"Elschool-API/internal/domain/models"
	"Elschool-API/internal/infra/parser"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var (
	ErrCantFetch = errors.New("fetching failed")
)

const (
	HttpsPrefix   = "https://"
	Diaries       = "/users/diaries"
	Grades        = "/users/diaries/grades"
	Results       = "/users/diaries/results"
	JwtCookieName = "JWToken"
)

type Fetcher struct {
	httpClient *http.Client
	parser     parser.Parser
	url        string
}

func New(url string) *Fetcher {
	return &Fetcher{
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		url: url, parser: *parser.New(),
	}
}

func (f *Fetcher) FetchDayMarks(ctx context.Context, jwt, date string) (marks models.DayMarks, err error) {
	const op = "infra.fetcher.FetchDayMarks"

	headers, err := f.getUrlHeaders(ctx, jwt)
	if err != nil {
		return models.DayMarks{}, fmt.Errorf("%s:  %w", op, err)
	}

	url := HttpsPrefix + f.url + Grades + headers
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return models.DayMarks{}, fmt.Errorf("%s:  %w", op, err)
	}

	req.AddCookie(&http.Cookie{
		Name:  JwtCookieName,
		Value: jwt,
	})

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return models.DayMarks{}, fmt.Errorf("%s: %w", op, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return models.DayMarks{}, fmt.Errorf("%s: %w", op, resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return models.DayMarks{}, fmt.Errorf("%s: %w", op, err)
	}

	marks, err = f.parser.ParseDayMarks(date, string(bodyBytes))
	if err != nil {
		return models.DayMarks{}, fmt.Errorf("%s: %w", op, err)
	}

	return marks, nil
}

func (f *Fetcher) FetchAverageMarks(ctx context.Context, jwt string, period int32) (marks models.AverageMarks, err error) {
	const op = "infra.fetcher.FetchAverageMarks"

	headers, err := f.getUrlHeaders(ctx, jwt)
	if err != nil {
		return models.AverageMarks{}, fmt.Errorf("%s: %w", op, err)
	}

	url := HttpsPrefix + f.url + Grades + headers
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return models.AverageMarks{}, fmt.Errorf("%s: %w", op, err)
	}

	req.AddCookie(&http.Cookie{
		Name:  JwtCookieName,
		Value: jwt,
	})

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return models.AverageMarks{}, fmt.Errorf("%s: %w", op, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return models.AverageMarks{}, fmt.Errorf("%s: %w", op, resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return models.AverageMarks{}, fmt.Errorf("%s: %w", op, err)
	}

	marks, err = f.parser.ParseAverageMarks(period, string(bodyBytes))
	if err != nil {
		return models.AverageMarks{}, fmt.Errorf("%s: %w", op, err)
	}

	return marks, nil
}

func (f *Fetcher) FetchFinalMarks(ctx context.Context, jwt string) (marks models.FinalMarks, err error) {
	const op = "infra.fetcher.FetchFinalMarks"

	headers, err := f.getUrlHeaders(ctx, jwt)
	if err != nil {
		return models.FinalMarks{}, fmt.Errorf("%s: %w", op, err)
	}

	url := HttpsPrefix + f.url + Results + headers

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return models.FinalMarks{}, fmt.Errorf("%s: %w", op, err)
	}

	req.AddCookie(&http.Cookie{
		Name:  JwtCookieName,
		Value: jwt,
	})

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return models.FinalMarks{}, fmt.Errorf("%s: %w", op, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return models.FinalMarks{}, fmt.Errorf("%s: %w", op, resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return models.FinalMarks{}, fmt.Errorf("%s: %w", op, err)
	}

	marks, err = f.parser.ParseFinalMarks(string(bodyBytes))
	if err != nil {
		return models.FinalMarks{}, fmt.Errorf("%s: %w", op, err)
	}

	return marks, nil
}

func (f *Fetcher) getUrlHeaders(ctx context.Context, jwt string) (headers string, err error) {
	const op = "infra.fetcher.getUrlHeaders"
	url := HttpsPrefix + f.url + Diaries

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	req.AddCookie(&http.Cookie{
		Name:  JwtCookieName,
		Value: jwt,
	})

	httpClient := &http.Client{
		Timeout: 15 * time.Second, CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	bodyStr := string(bodyBytes)

	startIdx := strings.Index(bodyStr, "?")
	endIdx := strings.Index(bodyStr, "here")

	if startIdx == -1 || endIdx == -1 {
		return "", fmt.Errorf("%s: %w", op, ErrCantFetch)
	}

	headers = bodyStr[startIdx : endIdx-2]

	toDelete := "amp;"
	for {
		start := strings.Index(headers, toDelete)
		if start == -1 {
			break
		}
		headers = headers[:start] + headers[start+len(toDelete):]
	}

	return headers, nil
}
