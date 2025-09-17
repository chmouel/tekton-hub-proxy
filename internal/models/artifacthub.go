package models

// Artifact Hub API response models

type ArtifactHubPackageResponse struct {
	Data ArtifactHubPackage `json:"data"`
}

type ArtifactHubPackage struct {
	PackageID           string                     `json:"package_id"`
	Name                string                     `json:"name"`
	NormalizedName      string                     `json:"normalized_name"`
	LogoImageID         string                     `json:"logo_image_id"`
	DisplayName         string                     `json:"display_name"`
	Description         string                     `json:"description"`
	Version             string                     `json:"version"`
	AppVersion          string                     `json:"app_version"`
	License             string                     `json:"license"`
	Deprecated          bool                       `json:"deprecated"`
	Signed              bool                       `json:"signed"`
	Official            bool                       `json:"official"`
	CNCF                *bool                      `json:"cncf"`
	TS                  int64                      `json:"ts"`
	Repository          ArtifactHubRepository      `json:"repository"`
	LatestVersion       string                     `json:"latest_version"`
	AvailableVersions   []ArtifactHubVersion       `json:"available_versions"`
	Keywords            []string                   `json:"keywords"`
	HomeURL             string                     `json:"home_url"`
	README              string                     `json:"readme"`
	Links               []ArtifactHubLink          `json:"links"`
	Maintainers         []ArtifactHubMaintainer    `json:"maintainers"`
	ContainersImages    []ArtifactHubContainerImg  `json:"containers_images"`
	HasValuesSchema     bool                       `json:"has_values_schema"`
	HasChangelog        bool                       `json:"has_changelog"`
	ContentURL          string                     `json:"content_url"`
	ContainsSecurityUpdates bool                   `json:"contains_security_updates"`
	Prerelease          bool                       `json:"prerelease"`
	Data                ArtifactHubPackageData     `json:"data"`
}

type ArtifactHubPackageData struct {
	ManifestRaw         string `json:"manifestRaw"`
	PipelinesMinVersion string `json:"pipelines.minVersion"`
}

type ArtifactHubRepository struct {
	RepositoryID       string `json:"repository_id"`
	Kind               int    `json:"kind"`
	Name               string `json:"name"`
	DisplayName        string `json:"display_name"`
	URL                string `json:"url"`
	VerifiedPublisher  bool   `json:"verified_publisher"`
	Official           bool   `json:"official"`
	CNCF               *bool  `json:"cncf"`
	Private            bool   `json:"private"`
	ScannerDisabled    bool   `json:"scanner_disabled"`
	UserAlias          string `json:"user_alias"`
	OrganizationName   string `json:"organization_name"`
	OrganizationDisplayName string `json:"organization_display_name"`
}

type ArtifactHubVersion struct {
	Version                 string `json:"version"`
	ContainsSecurityUpdates bool   `json:"contains_security_updates"`
	Prerelease              bool   `json:"prerelease"`
	TS                      int64  `json:"ts"`
}

type ArtifactHubLink struct {
	URL  string `json:"url"`
	Name string `json:"name"`
}

type ArtifactHubMaintainer struct {
	MaintainerID string `json:"maintainer_id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
}

type ArtifactHubContainerImg struct {
	Image       string `json:"image"`
	Name        string `json:"name"`
	Whitelisted bool   `json:"whitelisted"`
}

type ArtifactHubSearchResponse struct {
	Packages []ArtifactHubPackageSummary `json:"packages"`
	Facets   []ArtifactHubFacet          `json:"facets"`
}

type ArtifactHubPackageSummary struct {
	PackageID      string                `json:"package_id"`
	Name           string                `json:"name"`
	NormalizedName string                `json:"normalized_name"`
	LogoImageID    string                `json:"logo_image_id"`
	DisplayName    string                `json:"display_name"`
	Description    string                `json:"description"`
	Version        string                `json:"version"`
	AppVersion     string                `json:"app_version"`
	Deprecated     bool                  `json:"deprecated"`
	Signed         bool                  `json:"signed"`
	Official       bool                  `json:"official"`
	CNCF           *bool                 `json:"cncf"`
	TS             int64                 `json:"ts"`
	Repository     ArtifactHubRepository `json:"repository"`
	Stars          int                   `json:"stars"`
	Category       int                   `json:"category"`
}

type ArtifactHubFacet struct {
	Title     string                   `json:"title"`
	FilterKey string                   `json:"filter_key"`
	Options   []ArtifactHubFacetOption `json:"options"`
}

type ArtifactHubFacetOption struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Total     int    `json:"total"`
	FilterKey string `json:"filter_key"`
}
