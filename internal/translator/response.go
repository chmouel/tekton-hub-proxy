package translator

import (
	"fmt"
	"tekton-hub-proxy/internal/models"
	"time"

	"github.com/sirupsen/logrus"
)

type ResponseTranslator struct {
	versionTranslator *VersionTranslator
}

func NewResponseTranslator() *ResponseTranslator {
	return &ResponseTranslator{
		versionTranslator: NewVersionTranslator(),
	}
}

func (r *ResponseTranslator) ArtifactHubPackageToTektonResource(pkg *models.ArtifactHubPackage, catalogTranslator *CatalogTranslator) (*models.TektonHubResource, error) {
	logrus.WithField("package_name", pkg.Name).Debug("Converting Artifact Hub package to Tekton Hub resource")

	// Convert catalog name
	tektonCatalog, err := catalogTranslator.ArtifactHubToTekton(pkg.Repository.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to convert catalog name: %w", err)
	}

	// Extract kind from repository kind (e.g., "tekton-task" -> "task")
	kind := r.extractKindFromRepoKind(pkg.Repository.Kind)

	// Convert latest version
	latestVersion, err := r.versionTranslator.ArtifactHubToTekton(pkg.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to convert latest version: %w", err)
	}

	// Convert all versions
	var versions []models.TektonHubVersionSummary
	for i, version := range pkg.AvailableVersions {
		tektonVersion, err := r.versionTranslator.ArtifactHubToTekton(version.Version)
		if err != nil {
			logrus.WithField("version", version.Version).Warn("Failed to convert version, skipping")
			continue
		}

		versions = append(versions, models.TektonHubVersionSummary{
			ID:      i + 1, // Generate sequential ID
			Version: tektonVersion,
		})
	}

	// Build the resource
	resource := &models.TektonHubResource{
		ID:   r.generateResourceID(pkg.PackageID),
		Name: pkg.Name,
		Kind: kind,
		Catalog: models.TektonHubCatalog{
			ID:       r.generateCatalogID(pkg.Repository.Name),
			Name:     tektonCatalog,
			Provider: "github", // Default to github
			Type:     r.getCatalogType(pkg.Repository.Official),
			URL:      pkg.Repository.URL,
		},
		Categories:    r.convertKeywordsToCategories(pkg.Keywords),
		Tags:          r.convertKeywordsToTags(pkg.Keywords),
		Platforms:     []models.TektonHubPlatform{{ID: 1, Name: "linux/amd64"}}, // Default platform
		Rating:        4.0,                                                      // Default rating since Artifact Hub doesn't provide this
		HubURLPath:    fmt.Sprintf("%s/%s/%s", tektonCatalog, kind, pkg.Name),
		HubRawURLPath: fmt.Sprintf("/%s/%s/%s/raw", tektonCatalog, kind, pkg.Name),
		Versions:      versions,
	}

	// Set latest version details
	resource.LatestVersion = models.TektonHubResourceVersion{
		ID:                  1,
		Version:             latestVersion,
		DisplayName:         pkg.DisplayName,
		Description:         pkg.Description,
		MinPipelinesVersion: pkg.Data.PipelinesMinVersion,
		RawURL:              pkg.ContentURL,
		WebURL:              pkg.ContentURL,
		UpdatedAt:           time.Unix(pkg.TS, 0),
		Platforms:           resource.Platforms,
		HubURLPath:          resource.HubURLPath,
		HubRawURLPath:       resource.HubRawURLPath,
		Deprecated:          pkg.Deprecated,
	}

	return resource, nil
}

func (r *ResponseTranslator) ArtifactHubPackageToTektonYAML(pkg *models.ArtifactHubPackage) (*models.TektonHubYamlResponse, error) {
	return &models.TektonHubYamlResponse{
		Data: models.TektonHubYamlData{
			YAML: pkg.Data.ManifestRaw,
		},
	}, nil
}

func (r *ResponseTranslator) ArtifactHubPackageToTektonReadme(pkg *models.ArtifactHubPackage) (*models.TektonHubReadmeResponse, error) {
	return &models.TektonHubReadmeResponse{
		Data: models.TektonHubReadmeData{
			README: pkg.README,
			YAML:   pkg.Data.ManifestRaw,
		},
	}, nil
}

func (r *ResponseTranslator) ArtifactHubSearchToTektonResources(search *models.ArtifactHubSearchResponse, catalogTranslator *CatalogTranslator) (*models.TektonHubResourcesResponse, error) {
	var resources []models.TektonHubResource

	for _, pkg := range search.Packages {
		// Convert package summary to full package for conversion
		fullPkg := r.packageSummaryToFullPackage(&pkg)

		resource, err := r.ArtifactHubPackageToTektonResource(fullPkg, catalogTranslator)
		if err != nil {
			logrus.WithField("package", pkg.Name).Warn("Failed to convert package, skipping")
			continue
		}

		resources = append(resources, *resource)
	}

	return &models.TektonHubResourcesResponse{
		Data: resources,
	}, nil
}

func (r *ResponseTranslator) extractKindFromRepoKind(kind int) string {
	// Map Artifact Hub repository kinds to Tekton kinds
	// This is a simplified mapping - in reality, you'd need to query Artifact Hub for kind mappings
	switch kind {
	case 12: // Tekton task
		return "task"
	case 13: // Tekton pipeline
		return "pipeline"
	default:
		return "task" // Default to task
	}
}

func (r *ResponseTranslator) generateResourceID(packageID string) int {
	// Simple hash-based ID generation
	hash := 0
	for _, c := range packageID {
		hash = hash*31 + int(c)
	}
	if hash < 0 {
		hash = -hash
	}
	return hash % 1000000 // Keep it reasonable
}

func (r *ResponseTranslator) generateCatalogID(catalogName string) int {
	// Simple hash-based ID generation for catalog
	hash := 0
	for _, c := range catalogName {
		hash = hash*31 + int(c)
	}
	if hash < 0 {
		hash = -hash
	}
	return hash % 1000 // Keep it reasonable
}

func (r *ResponseTranslator) getCatalogType(official bool) string {
	if official {
		return "official"
	}
	return "community"
}

func (r *ResponseTranslator) convertKeywordsToCategories(keywords []string) []models.TektonHubCategory {
	var categories []models.TektonHubCategory
	for i, keyword := range keywords {
		if i >= 5 { // Limit to 5 categories
			break
		}
		categories = append(categories, models.TektonHubCategory{
			ID:   i + 1,
			Name: keyword,
		})
	}
	return categories
}

func (r *ResponseTranslator) convertKeywordsToTags(keywords []string) []models.TektonHubTag {
	var tags []models.TektonHubTag
	for i, keyword := range keywords {
		tags = append(tags, models.TektonHubTag{
			ID:   i + 1,
			Name: keyword,
		})
	}
	return tags
}

func (r *ResponseTranslator) packageSummaryToFullPackage(summary *models.ArtifactHubPackageSummary) *models.ArtifactHubPackage {
	return &models.ArtifactHubPackage{
		PackageID:      summary.PackageID,
		Name:           summary.Name,
		NormalizedName: summary.NormalizedName,
		LogoImageID:    summary.LogoImageID,
		DisplayName:    summary.DisplayName,
		Description:    summary.Description,
		Version:        summary.Version,
		AppVersion:     summary.AppVersion,
		Deprecated:     summary.Deprecated,
		Signed:         summary.Signed,
		Official:       summary.Official,
		CNCF:           summary.CNCF,
		TS:             summary.TS,
		Repository:     summary.Repository,
		Keywords:       []string{}, // Keywords not available in summary
		Data: models.ArtifactHubPackageData{
			ManifestRaw: "", // Not available in summary
		},
	}
}

