package client

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"tekton-hub-proxy/internal/config"
	"tekton-hub-proxy/internal/models"
	"time"

	"github.com/sirupsen/logrus"
)

type cacheEntry struct {
	data      interface{}
	timestamp time.Time
}

type memoryCache struct {
	entries map[string]*cacheEntry
	mutex   sync.RWMutex
	ttl     time.Duration
	maxSize int
}

func newMemoryCache(ttl time.Duration, maxSize int) *memoryCache {
	cache := &memoryCache{
		entries: make(map[string]*cacheEntry),
		ttl:     ttl,
		maxSize: maxSize,
	}

	go cache.cleanup()
	return cache
}

func (mc *memoryCache) get(key string) (interface{}, bool) {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	entry, exists := mc.entries[key]
	if !exists {
		return nil, false
	}

	if time.Since(entry.timestamp) > mc.ttl {
		go mc.delete(key)
		return nil, false
	}

	return entry.data, true
}

func (mc *memoryCache) set(key string, data interface{}) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	if len(mc.entries) >= mc.maxSize {
		mc.evictLRU()
	}

	mc.entries[key] = &cacheEntry{
		data:      data,
		timestamp: time.Now(),
	}
}

func (mc *memoryCache) delete(key string) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	delete(mc.entries, key)
}

func (mc *memoryCache) evictLRU() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range mc.entries {
		if oldestKey == "" || entry.timestamp.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.timestamp
		}
	}

	if oldestKey != "" {
		delete(mc.entries, oldestKey)
		logrus.WithField("cache_key", oldestKey).Debug("Cache entry evicted (LRU)")
	}
}

func (mc *memoryCache) cleanup() {
	ticker := time.NewTicker(mc.ttl / 2)
	defer ticker.Stop()

	for range ticker.C {
		mc.mutex.Lock()
		now := time.Now()
		expiredCount := 0
		for key, entry := range mc.entries {
			if now.Sub(entry.timestamp) > mc.ttl {
				delete(mc.entries, key)
				expiredCount++
			}
		}
		if expiredCount > 0 {
			logrus.WithFields(logrus.Fields{
				"expired_entries": expiredCount,
				"cache_size":      len(mc.entries),
				"max_size":        mc.maxSize,
			}).Debug("Cache cleanup completed")
		}
		mc.mutex.Unlock()
	}
}

func (mc *memoryCache) size() int {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()
	return len(mc.entries)
}

type ArtifactHubClient struct {
	baseURL    string
	httpClient *http.Client
	maxRetries int
	cache      *memoryCache
}

func NewArtifactHubClient(cfg config.ArtifactHubConfig) *ArtifactHubClient {
	client := &ArtifactHubClient{
		baseURL: strings.TrimSuffix(cfg.BaseURL, "/"),
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		maxRetries: cfg.MaxRetries,
	}

	if cfg.Cache.Enabled {
		client.cache = newMemoryCache(cfg.Cache.TTL, cfg.Cache.MaxSize)
		logrus.WithFields(logrus.Fields{
			"cache_enabled": true,
			"cache_ttl":     cfg.Cache.TTL,
			"cache_max_size": cfg.Cache.MaxSize,
		}).Info("Cache enabled for Artifact Hub client")
	} else {
		logrus.Info("Cache disabled for Artifact Hub client")
	}

	return client
}

func (c *ArtifactHubClient) generateCacheKey(prefix string, params ...string) string {
	key := prefix
	for _, param := range params {
		key += ":" + param
	}

	hash := sha256.Sum256([]byte(key))
	return fmt.Sprintf("%x", hash)[:16]
}

func (c *ArtifactHubClient) GetPackage(repoKind, catalog, name, version string) (*models.ArtifactHubPackage, error) {
	if c.cache != nil {
		cacheKey := c.generateCacheKey("package", repoKind, catalog, name, version)

		if cached, found := c.cache.get(cacheKey); found {
			logrus.WithFields(logrus.Fields{
				"api_call":  "GetPackage",
				"repo_kind": repoKind,
				"catalog":   catalog,
				"name":      name,
				"version":   version,
			}).Info("🚀 CACHE HIT - GetPackage")
			return cached.(*models.ArtifactHubPackage), nil
		}
	}

	path := fmt.Sprintf("/api/v1/packages/%s/%s/%s/%s", repoKind, catalog, name, version)
	url := c.baseURL + path

	logrus.WithFields(logrus.Fields{
		"api_call":  "GetPackage",
		"repo_kind": repoKind,
		"catalog":   catalog,
		"name":      name,
		"version":   version,
		"url":       url,
	}).Debug("🌐 Making Artifact Hub API call")

	var response models.ArtifactHubPackage
	if err := c.makeRequest("GET", url, &response); err != nil {
		return nil, fmt.Errorf("failed to get package: %w", err)
	}

	if c.cache != nil {
		cacheKey := c.generateCacheKey("package", repoKind, catalog, name, version)
		c.cache.set(cacheKey, &response)
		logrus.WithFields(logrus.Fields{
			"api_call":   "GetPackage",
			"status":     "success",
			"package":    response.Name,
			"cache_size": c.cache.size(),
		}).Info("📦 API CALL CACHED - GetPackage")
	} else {
		logrus.WithFields(logrus.Fields{
			"api_call": "GetPackage",
			"status":   "success",
			"package":  response.Name,
		}).Info("🌐 API CALL NO CACHE - GetPackage")
	}

	return &response, nil
}

func (c *ArtifactHubClient) GetPackageLatest(repoKind, catalog, name string) (*models.ArtifactHubPackage, error) {
	if c.cache != nil {
		cacheKey := c.generateCacheKey("package-latest", repoKind, catalog, name)

		if cached, found := c.cache.get(cacheKey); found {
			logrus.WithFields(logrus.Fields{
				"api_call":  "GetPackageLatest",
				"repo_kind": repoKind,
				"catalog":   catalog,
				"name":      name,
			}).Info("🚀 CACHE HIT - GetPackageLatest")
			return cached.(*models.ArtifactHubPackage), nil
		}
	}

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

	if c.cache != nil {
		cacheKey := c.generateCacheKey("package-latest", repoKind, catalog, name)
		c.cache.set(cacheKey, &response)
		logrus.WithFields(logrus.Fields{
			"api_call":   "GetPackageLatest",
			"status":     "success",
			"package":    response.Name,
			"version":    response.Version,
			"cache_size": c.cache.size(),
		}).Info("📦 API CALL CACHED - GetPackageLatest")
	} else {
		logrus.WithFields(logrus.Fields{
			"api_call": "GetPackageLatest",
			"status":   "success",
			"package":  response.Name,
			"version":  response.Version,
		}).Info("🌐 API CALL NO CACHE - GetPackageLatest")
	}

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

	queryString := queryParams.Encode()

	if c.cache != nil {
		cacheKey := c.generateCacheKey("search", queryString)

		if cached, found := c.cache.get(cacheKey); found {
			logrus.WithFields(logrus.Fields{
				"api_call": "SearchPackages",
				"query":    params.Query,
			}).Info("🚀 CACHE HIT - SearchPackages")
			return cached.(*models.ArtifactHubSearchResponse), nil
		}
	}

	url := c.baseURL + path + "?" + queryString

	logrus.WithFields(logrus.Fields{
		"api_call": "SearchPackages",
		"query":    params.Query,
		"url":      url,
	}).Debug("🌐 Making Artifact Hub search API call")

	var response models.ArtifactHubSearchResponse
	if err := c.makeRequest("GET", url, &response); err != nil {
		return nil, fmt.Errorf("failed to search packages: %w", err)
	}

	if c.cache != nil {
		cacheKey := c.generateCacheKey("search", queryString)
		c.cache.set(cacheKey, &response)
		logrus.WithFields(logrus.Fields{
			"api_call":    "SearchPackages",
			"status":      "success",
			"results":     len(response.Packages),
			"cache_size":  c.cache.size(),
		}).Info("📦 API CALL CACHED - SearchPackages")
	} else {
		logrus.WithFields(logrus.Fields{
			"api_call": "SearchPackages",
			"status":   "success",
			"results":  len(response.Packages),
		}).Info("🌐 API CALL NO CACHE - SearchPackages")
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

		defer resp.Body.Close() //nolint:errcheck

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

