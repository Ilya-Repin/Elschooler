package tests

import (
	"Elschool-API/internal/app"
	"Elschool-API/internal/config"
	"Elschool-API/internal/infra/auth"
	"Elschool-API/internal/infra/cache/redis"
	"Elschool-API/internal/infra/fetcher"
	"Elschool-API/internal/infra/metrics"
	"Elschool-API/internal/infra/storage/postgres"
	"fmt"
	"github.com/jarcoal/httpmock"
	"io"
	"log/slog"
	"net/http"
	"os"
	"testing"
)

const (
	headers = "?rooId=11&instituteId=111&departmentId=111111&pupilId=1111111"
)

func TestMain(m *testing.M) {
	cfg := config.MustLoadByPath("../config/local_tests.yaml")

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	log.Info("starting server", slog.String("env", cfg.Env))

	db, err := postgres.InitDB(&cfg.StorageConfig)
	if err != nil {
		panic(err)
	}

	rclient, err := redis.InitCache(&cfg.CacheConfig)
	if err != nil {
		panic(err)
	}

	metr, err := metrics.New(&cfg.MetricsConfig)
	if err != nil {
		panic(err)
	}

	application := app.New(log, db, rclient, metr, cfg.InfraConfig, cfg.GRPCConfig.Port)

	go application.GRPCsrv.MustRun()

	defer func() {
		application.GRPCsrv.Stop()
	}()

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	mockUrlLogon := auth.HttpsPrefix + cfg.InfraConfig.Url + auth.LogonIndex

	httpmock.RegisterResponder("POST", mockUrlLogon,
		func(req *http.Request) (*http.Response, error) {
			reqBody := req.FormValue("login")
			reqPassword := req.FormValue("password")

			if reqBody == "invalidLogin" && reqPassword == "invalidPassword" || reqBody == "" || reqPassword == "" {
				resp := httpmock.NewStringResponse(http.StatusForbidden, "")
				return resp, nil
			}

			body := "`<html><head><title>Object moved</title></head><body>\n<h2>Object moved to <a href=\"/privateoffice\">here</a>.</h2>\n</body></html>`"
			resp := httpmock.NewStringResponse(http.StatusFound, body)
			resp.Header.Set("Set-Cookie", fmt.Sprintf("%s=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJZCI6IjExMTIyMjMiLCJNdXN0Q2hhbmdlUGFzc3dvcmQiOiJGYWxzZSIsInJvbGUiOiIxLERlcGFydG1lbnQsLDEsMTExMSwxMTExMSwiLCJFU0lBTG9nb24iOiJGYWxzZSIsIkF2YWlsYWJsZVJvbGVzIjoiMSIsIm5iZiI6MTExMTExMTExMSwiZXhwIjoxMTExMTExMTExLCJpYXQiOjExMTExMTExMTEsImlzcyI6IjExLjExLjExLjExIiwiYXVkIjoic2Nob29sLnJ1In0.Y8aCt0hWcceQBtIqzdfXMJ8Df_r2A08Oa9wn-7r1PWI", auth.JwtCookieName))
			resp.Header.Set("Content-Type", "text/html")
			return resp, nil
		},
	)

	mockUrlCheck := auth.HttpsPrefix + cfg.InfraConfig.Url + auth.PrivateOffice

	httpmock.RegisterResponder("GET", mockUrlCheck,
		func(req *http.Request) (*http.Response, error) {
			cookie, err := req.Cookie(auth.JwtCookieName)

			if err != nil || cookie.String() == "" {
				resp := httpmock.NewStringResponse(http.StatusForbidden, "")
				return resp, nil
			}

			body := "`<html><head><title>PrivateOffice</title></head><body>PrivateOffice</body></html>`"
			resp := httpmock.NewStringResponse(http.StatusOK, body)
			resp.Header.Set("Content-Type", "text/html")
			return resp, nil
		},
	)

	mockUrlHeaders := fetcher.HttpsPrefix + cfg.InfraConfig.Url + fetcher.Diaries

	httpmock.RegisterResponder("GET", mockUrlHeaders,
		func(req *http.Request) (*http.Response, error) {
			cookie, err := req.Cookie(auth.JwtCookieName)

			if err != nil || cookie.String() == "" {
				resp := httpmock.NewStringResponse(http.StatusForbidden, "")
				return resp, nil
			}

			body := "`<html><head><title>Object moved</title></head><body>\n\t<h2>Object moved to <a href=\"/users/diaries/details?rooId=11&amp;instituteId=111&amp;departmentId=111111&amp;pupilId=1111111\">here</a>.</h2>\n</body></html>`"
			resp := httpmock.NewStringResponse(http.StatusFound, body)
			resp.Header.Set("Content-Type", "text/html")
			return resp, nil
		},
	)

	mockUrlGrades := fetcher.HttpsPrefix + cfg.InfraConfig.Url + fetcher.Grades + headers
	httpmock.RegisterResponder("GET", mockUrlGrades,
		func(req *http.Request) (*http.Response, error) {
			cookie, err := req.Cookie(auth.JwtCookieName)

			if err != nil || cookie.String() == "" {
				resp := httpmock.NewStringResponse(http.StatusForbidden, "")
				return resp, nil
			}

			file, err := os.Open("./html/test_page_grades.html")
			if err != nil {
				return nil, err
			}
			defer file.Close()

			data, err := io.ReadAll(file)
			if err != nil {
				return nil, err
			}

			body := string(data)
			resp := httpmock.NewStringResponse(http.StatusOK, body)
			resp.Header.Set("Content-Type", "text/html")
			return resp, nil
		},
	)

	mockUrlResult := fetcher.HttpsPrefix + cfg.InfraConfig.Url + fetcher.Results + headers
	httpmock.RegisterResponder("GET", mockUrlResult,
		func(req *http.Request) (*http.Response, error) {
			cookie, err := req.Cookie(auth.JwtCookieName)

			if err != nil || cookie.String() == "" {
				resp := httpmock.NewStringResponse(http.StatusForbidden, "")
				return resp, nil
			}

			file, err := os.Open("./html/test_page_result.html")
			if err != nil {
				return nil, err
			}
			defer file.Close()

			data, err := io.ReadAll(file)
			if err != nil {
				return nil, err
			}

			body := string(data)
			resp := httpmock.NewStringResponse(http.StatusOK, body)
			resp.Header.Set("Content-Type", "text/html")
			return resp, nil
		},
	)

	exitCode := m.Run()

	os.Exit(exitCode)
}
