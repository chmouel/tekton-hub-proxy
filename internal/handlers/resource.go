package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"tekton-hub-proxy/internal/client"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func (h *Handlers) GetResourceByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid resource ID")
		return
	}

	logrus.WithField("id", id).Debug("Getting resource by ID")

	// Since we don't have direct ID mapping, we'll return a placeholder error
	// In a real implementation, you'd need to maintain an ID mapping or search by ID
	h.writeErrorResponse(w, http.StatusNotImplemented, "resource lookup by ID not implemented")
}

func (h *Handlers) GetResourceVersionsByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid resource ID")
		return
	}

	logrus.WithField("id", id).Debug("Getting resource versions by ID")

	// Since we don't have direct ID mapping, we'll return a placeholder error
	h.writeErrorResponse(w, http.StatusNotImplemented, "resource versions lookup by ID not implemented")
}

func (h *Handlers) GetResourceByVersionID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	versionIDStr := vars["versionID"]

	versionID, err := strconv.Atoi(versionIDStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid version ID")
		return
	}

	logrus.WithField("version_id", versionID).Debug("Getting resource by version ID")

	// Since we don't have direct version ID mapping, we'll return a placeholder error
	h.writeErrorResponse(w, http.StatusNotImplemented, "resource lookup by version ID not implemented")
}

func (h *Handlers) ListResources(w http.ResponseWriter, r *http.Request) {
	logrus.Debug("Listing resources")

	// Parse query parameters
	limit := 1000
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// Search for Tekton packages across all catalogs
	searchParams := client.SearchParams{
		Query:  "",
		Kinds:  []int{12, 13}, // Tekton task and pipeline kinds
		Limit:  limit,
		Facets: false,
	}

	// Add repositories based on our catalog mappings
	mappings := h.catalogTranslator.GetAvailableMappings()
	for _, artifactHubCatalog := range mappings {
		searchParams.Repositories = append(searchParams.Repositories, artifactHubCatalog)
	}

	searchResult, err := h.artifactHubClient.SearchPackages(searchParams)
	if err != nil {
		logrus.WithError(err).Error("Failed to search packages")
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to list resources")
		return
	}

	// Convert to Tekton Hub format
	response, err := h.responseTranslator.ArtifactHubSearchToTektonResources(searchResult, h.catalogTranslator)
	if err != nil {
		logrus.WithError(err).Error("Failed to convert search results")
		h.writeErrorResponse(w, http.StatusInternalServerError, "conversion error")
		return
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

func (h *Handlers) QueryResources(w http.ResponseWriter, r *http.Request) {
	logrus.Debug("Querying resources")

	query := r.URL.Query()

	// Parse query parameters
	searchParams := client.SearchParams{
		Query:  query.Get("name"),
		Limit:  1000,
		Facets: false,
	}

	// Parse limit
	if limitStr := query.Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			searchParams.Limit = limit
		}
	}

	// Parse catalogs
	if catalogs := query["catalogs"]; len(catalogs) > 0 {
		for _, catalog := range catalogs {
			// Convert catalog names to Artifact Hub format
			if artifactHubCatalog, err := h.catalogTranslator.TektonToArtifactHub(catalog); err == nil {
				searchParams.Repositories = append(searchParams.Repositories, artifactHubCatalog)
			}
		}
	} else {
		// If no catalogs specified, search all mapped catalogs
		mappings := h.catalogTranslator.GetAvailableMappings()
		for _, artifactHubCatalog := range mappings {
			searchParams.Repositories = append(searchParams.Repositories, artifactHubCatalog)
		}
	}

	// Parse kinds
	if kinds := query["kinds"]; len(kinds) > 0 {
		for _, kind := range kinds {
			switch strings.ToLower(kind) {
			case "task":
				searchParams.Kinds = append(searchParams.Kinds, 12) // Tekton task kind
			case "pipeline":
				searchParams.Kinds = append(searchParams.Kinds, 13) // Tekton pipeline kind
			}
		}
	} else {
		// Default to both tasks and pipelines
		searchParams.Kinds = []int{12, 13}
	}

	// Parse categories and tags - these would need more sophisticated mapping
	// For now, we'll include them in the general query
	if categories := query["categories"]; len(categories) > 0 {
		if searchParams.Query != "" {
			searchParams.Query += " "
		}
		searchParams.Query += strings.Join(categories, " ")
	}

	if tags := query["tags"]; len(tags) > 0 {
		if searchParams.Query != "" {
			searchParams.Query += " "
		}
		searchParams.Query += strings.Join(tags, " ")
	}

	logrus.WithFields(logrus.Fields{
		"query":        searchParams.Query,
		"kinds":        searchParams.Kinds,
		"repositories": searchParams.Repositories,
		"limit":        searchParams.Limit,
	}).Debug("Search parameters")

	// Search packages
	searchResult, err := h.artifactHubClient.SearchPackages(searchParams)
	if err != nil {
		logrus.WithError(err).Error("Failed to search packages")
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to query resources")
		return
	}

	// Convert to Tekton Hub format
	response, err := h.responseTranslator.ArtifactHubSearchToTektonResources(searchResult, h.catalogTranslator)
	if err != nil {
		logrus.WithError(err).Error("Failed to convert search results")
		h.writeErrorResponse(w, http.StatusInternalServerError, "conversion error")
		return
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

