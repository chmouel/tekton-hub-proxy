package translator

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"tekton-hub-proxy/internal/config"
)

type CatalogTranslator struct {
	mappings            map[string]string
	reverseMappings     map[string]string
	catalogMappingArray []config.CatalogMapping
}

func NewCatalogTranslator(catalogMappings []config.CatalogMapping) *CatalogTranslator {
	// Convert array to maps for fast lookup
	mappings := make(map[string]string)
	reverseMappings := make(map[string]string)

	for _, mapping := range catalogMappings {
		mappings[mapping.TektonHub] = mapping.ArtifactHub
		reverseMappings[mapping.ArtifactHub] = mapping.TektonHub
	}

	return &CatalogTranslator{
		mappings:            mappings,
		reverseMappings:     reverseMappings,
		catalogMappingArray: catalogMappings,
	}
}

func (c *CatalogTranslator) TektonToArtifactHub(tektonCatalog string) (string, error) {
	logrus.WithFields(logrus.Fields{
		"translation_type": "catalog",
		"direction":        "tekton_to_artifacthub",
		"input":            tektonCatalog,
	}).Debug("🔄 Translation request")

	if artifactHubCatalog, exists := c.mappings[tektonCatalog]; exists {
		logrus.WithFields(logrus.Fields{
			"translation_type":    "catalog",
			"direction":           "tekton_to_artifacthub",
			"tekton_catalog":      tektonCatalog,
			"artifacthub_catalog": artifactHubCatalog,
			"status":              "mapped",
		}).Debug("✅ Catalog mapping found")
		return artifactHubCatalog, nil
	}

	logrus.WithFields(logrus.Fields{
		"translation_type": "catalog",
		"direction":        "tekton_to_artifacthub",
		"tekton_catalog":   tektonCatalog,
		"status":           "passthrough",
	}).Debug("⚠️  No catalog mapping found, using original name")
	return tektonCatalog, nil
}

func (c *CatalogTranslator) ArtifactHubToTekton(artifactHubCatalog string) (string, error) {
	logrus.WithFields(logrus.Fields{
		"translation_type": "catalog",
		"direction":        "artifacthub_to_tekton",
		"input":            artifactHubCatalog,
	}).Debug("🔄 Translation request")

	if tektonCatalog, exists := c.reverseMappings[artifactHubCatalog]; exists {
		logrus.WithFields(logrus.Fields{
			"translation_type":    "catalog",
			"direction":           "artifacthub_to_tekton",
			"artifacthub_catalog": artifactHubCatalog,
			"tekton_catalog":      tektonCatalog,
			"status":              "mapped",
		}).Debug("✅ Reverse catalog mapping found")
		return tektonCatalog, nil
	}

	logrus.WithFields(logrus.Fields{
		"translation_type":    "catalog",
		"direction":           "artifacthub_to_tekton",
		"artifacthub_catalog": artifactHubCatalog,
		"status":              "passthrough",
	}).Debug("⚠️  No reverse catalog mapping found, using original name")
	return artifactHubCatalog, nil
}

func (c *CatalogTranslator) GetAvailableMappings() map[string]string {
	return c.mappings
}

func (c *CatalogTranslator) KindToRepoKind(kind string) string {
	return fmt.Sprintf("tekton-%s", kind)
}