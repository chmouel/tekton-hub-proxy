package handlers

import (
	"encoding/json"
	"net/http"
	"tekton-hub-proxy/internal/client"
	"tekton-hub-proxy/internal/models"
	"tekton-hub-proxy/internal/translator"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type Handlers struct {
	artifactHubClient  *client.ArtifactHubClient
	catalogTranslator  *translator.CatalogTranslator
	responseTranslator *translator.ResponseTranslator
	versionTranslator  *translator.VersionTranslator
}

func NewHandlers(
	artifactHubClient *client.ArtifactHubClient,
	catalogTranslator *translator.CatalogTranslator,
	responseTranslator *translator.ResponseTranslator,
	versionTranslator *translator.VersionTranslator,
) *Handlers {
	return &Handlers{
		artifactHubClient:  artifactHubClient,
		catalogTranslator:  catalogTranslator,
		responseTranslator: responseTranslator,
		versionTranslator:  versionTranslator,
	}
}

func (h *Handlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func (h *Handlers) ListCatalogs(w http.ResponseWriter, r *http.Request) {
	logrus.Debug("Listing catalogs")

	// Create mock catalogs based on our mappings
	var catalogs []models.TektonHubCatalog
	mappings := h.catalogTranslator.GetAvailableMappings()

	id := 1
	for tektonCatalog := range mappings {
		catalogs = append(catalogs, models.TektonHubCatalog{
			ID:       id,
			Name:     tektonCatalog,
			Provider: "github",
			Type:     "community",
			URL:      "https://github.com/tektoncd/catalog",
		})
		id++
	}

	response := models.TektonHubCatalogResponse{
		Data: catalogs,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

func (h *Handlers) GetResource(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	catalog := vars["catalog"]
	kind := vars["kind"]
	name := vars["name"]

	logrus.WithFields(logrus.Fields{
		"catalog": catalog,
		"kind":    kind,
		"name":    name,
	}).Debug("Getting resource")

	// Convert catalog name
	artifactHubCatalog, err := h.catalogTranslator.TektonToArtifactHub(catalog)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid catalog name")
		return
	}

	// Convert kind to repo kind
	repoKind := h.catalogTranslator.KindToRepoKind(kind)

	// Log the translation for debugging
	logrus.WithFields(logrus.Fields{
		"original_catalog":    catalog,
		"translated_catalog":  artifactHubCatalog,
		"original_kind":       kind,
		"translated_repo_kind": repoKind,
		"name":                name,
	}).Info("🔍 Translation details")

	// Get latest package from Artifact Hub
	pkg, err := h.artifactHubClient.GetPackageLatest(repoKind, artifactHubCatalog, name)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"repo_kind": repoKind,
			"catalog":   artifactHubCatalog,
			"name":      name,
			"error":     err.Error(),
		}).Error("Failed to get package from Artifact Hub")
		h.writeErrorResponse(w, http.StatusNotFound, "resource not found")
		return
	}

	// Convert to Tekton Hub format
	resource, err := h.responseTranslator.ArtifactHubPackageToTektonResource(pkg, h.catalogTranslator)
	if err != nil {
		logrus.WithError(err).Error("Failed to convert package to resource")
		h.writeErrorResponse(w, http.StatusInternalServerError, "conversion error")
		return
	}

	response := models.TektonHubResourceResponse{
		Data: *resource,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

func (h *Handlers) GetResourceVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	catalog := vars["catalog"]
	kind := vars["kind"]
	name := vars["name"]
	version := vars["version"]

	logrus.WithFields(logrus.Fields{
		"catalog": catalog,
		"kind":    kind,
		"name":    name,
		"version": version,
	}).Debug("Getting resource version")

	// Convert catalog name
	artifactHubCatalog, err := h.catalogTranslator.TektonToArtifactHub(catalog)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid catalog name")
		return
	}

	// Convert version
	artifactHubVersion, err := h.versionTranslator.TektonToArtifactHub(version)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid version format")
		return
	}

	// Convert kind to repo kind
	repoKind := h.catalogTranslator.KindToRepoKind(kind)

	// Get package from Artifact Hub
	pkg, err := h.artifactHubClient.GetPackage(repoKind, artifactHubCatalog, name, artifactHubVersion)
	if err != nil {
		logrus.WithError(err).Error("Failed to get package from Artifact Hub")
		h.writeErrorResponse(w, http.StatusNotFound, "resource version not found")
		return
	}

	// Convert to Tekton Hub format
	resource, err := h.responseTranslator.ArtifactHubPackageToTektonResource(pkg, h.catalogTranslator)
	if err != nil {
		logrus.WithError(err).Error("Failed to convert package to resource")
		h.writeErrorResponse(w, http.StatusInternalServerError, "conversion error")
		return
	}

	// Return the latest version details as resource version
	response := models.TektonHubResourceVersion{
		ID:                  resource.LatestVersion.ID,
		Version:             version, // Use original Tekton version
		DisplayName:         resource.LatestVersion.DisplayName,
		Description:         resource.LatestVersion.Description,
		MinPipelinesVersion: resource.LatestVersion.MinPipelinesVersion,
		RawURL:              resource.LatestVersion.RawURL,
		WebURL:              resource.LatestVersion.WebURL,
		UpdatedAt:           resource.LatestVersion.UpdatedAt,
		Platforms:           resource.LatestVersion.Platforms,
		HubURLPath:          resource.LatestVersion.HubURLPath,
		HubRawURLPath:       resource.LatestVersion.HubRawURLPath,
		Resource:            resource,
		Deprecated:          resource.LatestVersion.Deprecated,
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{"data": response})
}

func (h *Handlers) GetResourceYAML(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	catalog := vars["catalog"]
	kind := vars["kind"]
	name := vars["name"]
	version := vars["version"]

	logrus.WithFields(logrus.Fields{
		"catalog": catalog,
		"kind":    kind,
		"name":    name,
		"version": version,
	}).Debug("Getting resource YAML")

	pkg, err := h.getPackageFromArtifactHub(catalog, kind, name, version)
	if err != nil {
		logrus.WithError(err).Error("Failed to get package from Artifact Hub")
		h.writeErrorResponse(w, http.StatusNotFound, "resource not found")
		return
	}

	// Convert to Tekton Hub YAML format
	response, err := h.responseTranslator.ArtifactHubPackageToTektonYAML(pkg)
	if err != nil {
		logrus.WithError(err).Error("Failed to convert package to YAML")
		h.writeErrorResponse(w, http.StatusInternalServerError, "conversion error")
		return
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

func (h *Handlers) GetResourceYAMLRaw(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	catalog := vars["catalog"]
	kind := vars["kind"]
	name := vars["name"]
	version := vars["version"]

	logrus.WithFields(logrus.Fields{
		"catalog": catalog,
		"kind":    kind,
		"name":    name,
		"version": version,
	}).Debug("Getting raw resource YAML")

	pkg, err := h.getPackageFromArtifactHub(catalog, kind, name, version)
	if err != nil {
		logrus.WithError(err).Error("Failed to get package from Artifact Hub")
		h.writeErrorResponse(w, http.StatusNotFound, "resource not found")
		return
	}

	// Return raw YAML content
	w.Header().Set("Content-Type", "application/x-yaml")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(pkg.Data.ManifestRaw))
}

func (h *Handlers) GetLatestResourceYAML(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	catalog := vars["catalog"]
	kind := vars["kind"]
	name := vars["name"]

	logrus.WithFields(logrus.Fields{
		"catalog": catalog,
		"kind":    kind,
		"name":    name,
	}).Debug("Getting latest resource YAML")

	pkg, err := h.getPackageFromArtifactHub(catalog, kind, name, "")
	if err != nil {
		logrus.WithError(err).Error("Failed to get package from Artifact Hub")
		h.writeErrorResponse(w, http.StatusNotFound, "resource not found")
		return
	}

	// Return raw YAML content
	w.Header().Set("Content-Type", "application/x-yaml")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(pkg.Data.ManifestRaw))
}

func (h *Handlers) GetResourceReadme(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	catalog := vars["catalog"]
	kind := vars["kind"]
	name := vars["name"]
	version := vars["version"]

	logrus.WithFields(logrus.Fields{
		"catalog": catalog,
		"kind":    kind,
		"name":    name,
		"version": version,
	}).Debug("Getting resource README")

	pkg, err := h.getPackageFromArtifactHub(catalog, kind, name, version)
	if err != nil {
		logrus.WithError(err).Error("Failed to get package from Artifact Hub")
		h.writeErrorResponse(w, http.StatusNotFound, "resource not found")
		return
	}

	// Convert to Tekton Hub README format
	response, err := h.responseTranslator.ArtifactHubPackageToTektonReadme(pkg)
	if err != nil {
		logrus.WithError(err).Error("Failed to convert package to README")
		h.writeErrorResponse(w, http.StatusInternalServerError, "conversion error")
		return
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

func (h *Handlers) getPackageFromArtifactHub(catalog, kind, name, version string) (*models.ArtifactHubPackage, error) {
	// Convert catalog name
	artifactHubCatalog, err := h.catalogTranslator.TektonToArtifactHub(catalog)
	if err != nil {
		return nil, err
	}

	// Convert kind to repo kind
	repoKind := h.catalogTranslator.KindToRepoKind(kind)

	if version == "" {
		// Get latest version
		return h.artifactHubClient.GetPackageLatest(repoKind, artifactHubCatalog, name)
	}

	// Convert version
	artifactHubVersion, err := h.versionTranslator.TektonToArtifactHub(version)
	if err != nil {
		return nil, err
	}

	return h.artifactHubClient.GetPackage(repoKind, artifactHubCatalog, name, artifactHubVersion)
}

func (h *Handlers) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logrus.WithError(err).Error("Failed to encode JSON response")
	}
}

func (h *Handlers) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	h.writeJSONResponse(w, statusCode, map[string]string{
		"error": message,
		"name":  http.StatusText(statusCode),
	})
}

