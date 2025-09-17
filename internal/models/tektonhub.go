package models

import "time"

// Tekton Hub API response models

type TektonHubCatalogResponse struct {
	Data []TektonHubCatalog `json:"data"`
}

type TektonHubCatalog struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Provider string `json:"provider"`
	Type     string `json:"type"`
	URL      string `json:"url"`
}

type TektonHubResourceResponse struct {
	Data TektonHubResource `json:"data"`
}

type TektonHubResource struct {
	ID             int                       `json:"id"`
	Name           string                    `json:"name"`
	Kind           string                    `json:"kind"`
	Catalog        TektonHubCatalog          `json:"catalog"`
	Categories     []TektonHubCategory       `json:"categories"`
	Tags           []TektonHubTag            `json:"tags"`
	Platforms      []TektonHubPlatform       `json:"platforms"`
	Rating         float64                   `json:"rating"`
	LatestVersion  TektonHubResourceVersion  `json:"latestVersion"`
	Versions       []TektonHubVersionSummary `json:"versions"`
	HubURLPath     string                    `json:"hubURLPath"`
	HubRawURLPath  string                    `json:"hubRawURLPath"`
}

type TektonHubResourceVersion struct {
	ID                  int                 `json:"id"`
	Version             string              `json:"version"`
	DisplayName         string              `json:"displayName"`
	Description         string              `json:"description"`
	MinPipelinesVersion string              `json:"minPipelinesVersion"`
	RawURL              string              `json:"rawURL"`
	WebURL              string              `json:"webURL"`
	UpdatedAt           time.Time           `json:"updatedAt"`
	Platforms           []TektonHubPlatform `json:"platforms"`
	HubURLPath          string              `json:"hubURLPath"`
	HubRawURLPath       string              `json:"hubRawURLPath"`
	Resource            *TektonHubResource  `json:"resource,omitempty"`
	Deprecated          bool                `json:"deprecated"`
}

type TektonHubVersionSummary struct {
	ID      int    `json:"id"`
	Version string `json:"version"`
}

type TektonHubCategory struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type TektonHubTag struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type TektonHubPlatform struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type TektonHubYamlResponse struct {
	Data TektonHubYamlData `json:"data"`
}

type TektonHubYamlData struct {
	YAML string `json:"yaml"`
}

type TektonHubResourcesResponse struct {
	Data []TektonHubResource `json:"data"`
}

type TektonHubVersionsResponse struct {
	Data TektonHubVersionsData `json:"data"`
}

type TektonHubVersionsData struct {
	Latest   TektonHubResourceVersion  `json:"latest"`
	Versions []TektonHubResourceVersion `json:"versions"`
}

type TektonHubReadmeResponse struct {
	Data TektonHubReadmeData `json:"data"`
}

type TektonHubReadmeData struct {
	README string `json:"readme"`
	YAML   string `json:"yaml"`
}