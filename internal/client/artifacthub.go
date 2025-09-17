package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"tekton-hub-proxy/internal/config"
	"tekton-hub-proxy/internal/models"
)

type ArtifactHubClient struct {
	baseURL    string
	httpClient *http.Client
	maxRetries int
}

func NewArtifactHubClient(cfg config.ArtifactHubConfig) *ArtifactHubClient {
	return &ArtifactHubClient{
		baseURL: strings.TrimSuffix(cfg.BaseURL, "/"),
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		maxRetries: cfg.MaxRetries,
	}
}

func (c *ArtifactHubClient) GetPackage(repoKind, catalog, name, version string) (*models.ArtifactHubPackage, error) {
	path := fmt.Sprintf("/api/v1/packages/%s/%s/%s/%s", repoKind, catalog, name, version)
	url := c.baseURL + path

	logrus.WithFields(logrus.Fields{
		"api_call":   "GetPackage",
		"repo_kind":  repoKind,
		"catalog":    catalog,
		"name":       name,
		"version":    version,
		"url":        url,
	}).Debug("🌐 Making Artifact Hub API call")

	var response models.ArtifactHubPackage
	if err := c.makeRequest("GET", url, &response); err != nil {
		return nil, fmt.Errorf("failed to get package: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"api_call": "GetPackage",
		"status":   "success",
		"package":  response.Name,
	}).Debug("✅ Artifact Hub API call successful")

	return &response, nil
}

func (c *ArtifactHubClient) GetPackageLatest(repoKind, catalog, name string) (*models.ArtifactHubPackage, error) {
	path := fmt.Sprintf("/api/v1/packages/%s/%s/%s", repoKind, catalog, name)
	url := c.baseURL + path

	logrus.WithFields(logrus.Fields{
		"api_call":  "GetPackageLatest",
		"repo_kind": repoKind,
		"catalog":   catalog,
		"name":      name,
		"url":       url,
	}).Debug("🌐 Making Artifact Hub API call")

	var response models.ArtifactHubPackage
	if err := c.makeRequest("GET", url, &response); err != nil {
		return nil, fmt.Errorf("failed to get latest package: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"api_call": "GetPackageLatest",
		"status":   "success",
		"package":  response.Name,
		"version":  response.Version,
	}).Debug("✅ Artifact Hub API call successful")

	return &response, nil
}

func (c *ArtifactHubClient) SearchPackages(params SearchParams) (*models.ArtifactHubSearchResponse, error) {
	path := "/api/v1/packages/search"

	// Build query parameters
	queryParams := url.Values{}

	if params.Query != "" {
		queryParams.Set("ts_query_web", params.Query)
	}

	if len(params.Kinds) > 0 {
		for _, kind := range params.Kinds {
			queryParams.Add("kind", strconv.Itoa(kind))
		}
	}

	if len(params.Categories) > 0 {
		for _, category := range params.Categories {
			queryParams.Add("category", strconv.Itoa(category))
		}
	}

	if len(params.Repositories) > 0 {
		for _, repo := range params.Repositories {
			queryParams.Add("repo", repo)
		}
	}

	if params.Limit > 0 {
		queryParams.Set("limit", strconv.Itoa(params.Limit))
	}

	if params.Offset > 0 {
		queryParams.Set("offset", strconv.Itoa(params.Offset))
	}

	if params.Facets {
		queryParams.Set("facets", "true")
	}

	url := c.baseURL + path + "?" + queryParams.Encode()

	var response models.ArtifactHubSearchResponse
	if err := c.makeRequest("GET", url, &response); err != nil {
		return nil, fmt.Errorf("failed to search packages: %w", err)
	}

	return &response, nil
}

func (c *ArtifactHubClient) makeRequest(method, url string, result interface{}) error {
	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			time.Sleep(time.Duration(attempt) * time.Second)
			logrus.WithField("attempt", attempt).Debug("Retrying request")
		}

		req, err := http.NewRequest(method, url, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("User-Agent", "tekton-hub-proxy/1.0")
		req.Header.Set("Accept", "application/json")

		logrus.WithFields(logrus.Fields{
			"method":  method,
			"url":     url,
			"attempt": attempt + 1,
		}).Debug("Making HTTP request")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue
		}

		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = fmt.Errorf("failed to read response body: %w", err)
			continue
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			lastErr = fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))

			// Don't retry on client errors (4xx)
			if resp.StatusCode >= 400 && resp.StatusCode < 500 {
				break
			}
			continue
		}

		if err := json.Unmarshal(body, result); err != nil {
			lastErr = fmt.Errorf("failed to unmarshal response: %w", err)
			continue
		}

		logrus.WithField("status_code", resp.StatusCode).Debug("Request successful")
		return nil
	}

	return lastErr
}

type SearchParams struct {
	Query        string
	Kinds        []int
	Categories   []int
	Repositories []string
	Limit        int
	Offset       int
	Facets       bool
}