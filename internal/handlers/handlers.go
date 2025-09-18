package handlers

import (
	"encoding/json"
	"html/template"
	"net/http"
	"tekton-hub-proxy/internal/client"
	"tekton-hub-proxy/internal/config"
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
	config             *config.Config
}

func NewHandlers(
	artifactHubClient *client.ArtifactHubClient,
	catalogTranslator *translator.CatalogTranslator,
	responseTranslator *translator.ResponseTranslator,
	versionTranslator *translator.VersionTranslator,
	config *config.Config,
) *Handlers {
	return &Handlers{
		artifactHubClient:  artifactHubClient,
		catalogTranslator:  catalogTranslator,
		responseTranslator: responseTranslator,
		versionTranslator:  versionTranslator,
		config:             config,
	}
}

func (h *Handlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func (h *Handlers) LandingPage(w http.ResponseWriter, r *http.Request) {
	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Tekton Hub to Artifact Hub Proxy</title>
    <link rel="icon" type="image/svg+xml" href="data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMzIiIGhlaWdodD0iMzIiIHZpZXdCb3g9IjAgMCAzMiAzMiIgZmlsbD0ibm9uZSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KICA8IS0tIEJhY2tncm91bmQgY2lyY2xlIC0tPgogIDxjaXJjbGUgY3g9IjE2IiBjeT0iMTYiIHI9IjE2IiBmaWxsPSIjNjY3ZWVhIi8+CiAgCiAgPCEtLSBUZWt0b24gbm9kZSAobGVmdCkgLS0+CiAgPGNpcmNsZSBjeD0iOCIgY3k9IjE2IiByPSI0IiBmaWxsPSIjZmZmZmZmIiBzdHJva2U9IiM0NTUyOGQiIHN0cm9rZS13aWR0aD0iMSIvPgogIDx0ZXh0IHg9IjgiIHk9IjE4LjUiIGZvbnQtZmFtaWx5PSJBcmlhbCIgZm9udC1zaXplPSI2IiBmaWxsPSIjNDU1MjhkIiB0ZXh0LWFuY2hvcj0ibWlkZGxlIiBmb250LXdlaWdodD0iYm9sZCI+VDwvdGV4dD4KICA8IS0tIEFycm93IC0tPgogIDxwYXRoIGQ9Im0xMi41IDE0IDMgMi0zIDJ2LTFoLTF2LTJ6IiBmaWxsPSIjZmZmZmZmIi8+CiAgCiAgPCEtLSBQcm94eSBub2RlIChjZW50ZXIpIC0tPgogIDxyZWN0IHg9IjEzIiB5PSIxMyIgd2lkdGg9IjYiIGhlaWdodD0iNiIgcng9IjEiIGZpbGw9IiNmZmZmZmYiIHN0cm9rZT0iIzQ1NTI4ZCIgc3Ryb2tlLXdpZHRoPSIxIi8+CiAgPHRleHQgeD0iMTYiIHk9IjE3LjUiIGZvbnQtZmFtaWx5PSJBcmlhbCIgZm9udC1zaXplPSI1IiBmaWxsPSIjNDU1MjhkIiB0ZXh0LWFuY2hvcj0ibWlkZGxlIiBmb250LXdlaWdodD0iYm9sZCI+UDwvdGV4dD4KICA8IS0tIEFycm93IDIgLS0+CiAgPHBhdGggZD0ibTE5LjUgMTQgMyAyLTMgMnYtMWgtMXYtMnoiIGZpbGw9IiNmZmZmZmYiLz4KICA8IS0tIEFydGlmYWN0IEh1YiBub2RlIChyaWdodCkgLS0+CiAgPGNpcmNsZSBjeD0iMjQiIGN5PSIxNiIgcj0iNCIgZmlsbD0iI2ZmZmZmZiIgc3Ryb2tlPSIjNDU1MjhkIiBzdHJva2Utd2lkdGg9IjEiLz4KICA8dGV4dCB4PSIyNCIgeT0iMTguNSIgZm9udC1mYW1pbHk9IkFyaWFsIiBmb250LXNpemU9IjYiIGZpbGw9IiM0NTUyOGQiIHRleHQtYW5jaG9yPSJtaWRkbGUiIGZvbnQtd2VpZ2h0PSJib2xkIj5BPC90ZXh0Pgo8L3N2Zz4K">
    <link rel="shortcut icon" href="data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMzIiIGhlaWdodD0iMzIiIHZpZXdCb3g9IjAgMCAzMiAzMiIgZmlsbD0ibm9uZSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KICA8IS0tIEJhY2tncm91bmQgY2lyY2xlIC0tPgogIDxjaXJjbGUgY3g9IjE2IiBjeT0iMTYiIHI9IjE2IiBmaWxsPSIjNjY3ZWVhIi8+CiAgCiAgPCEtLSBUZWt0b24gbm9kZSAobGVmdCkgLS0+CiAgPGNpcmNsZSBjeD0iOCIgY3k9IjE2IiByPSI0IiBmaWxsPSIjZmZmZmZmIiBzdHJva2U9IiM0NTUyOGQiIHN0cm9rZS13aWR0aD0iMSIvPgogIDx0ZXh0IHg9IjgiIHk9IjE4LjUiIGZvbnQtZmFtaWx5PSJBcmlhbCIgZm9udC1zaXplPSI2IiBmaWxsPSIjNDU1MjhkIiB0ZXh0LWFuY2hvcj0ibWlkZGxlIiBmb250LXdlaWdodD0iYm9sZCI+VDwvdGV4dD4KICA8IS0tIEFycm93IC0tPgogIDxwYXRoIGQ9Im0xMi41IDE0IDMgMi0zIDJ2LTFoLTF2LTJ6IiBmaWxsPSIjZmZmZmZmIi8+CiAgCiAgPCEtLSBQcm94eSBub2RlIChjZW50ZXIpIC0tPgogIDxyZWN0IHg9IjEzIiB5PSIxMyIgd2lkdGg9IjYiIGhlaWdodD0iNiIgcng9IjEiIGZpbGw9IiNmZmZmZmYiIHN0cm9rZT0iIzQ1NTI4ZCIgc3Ryb2tlLXdpZHRoPSIxIi8+CiAgPHRleHQgeD0iMTYiIHk9IjE3LjUiIGZvbnQtZmFtaWx5PSJBcmlhbCIgZm9udC1zaXplPSI1IiBmaWxsPSIjNDU1MjhkIiB0ZXh0LWFuY2hvcj0ibWlkZGxlIiBmb250LXdlaWdodD0iYm9sZCI+UDwvdGV4dD4KICA8IS0tIEFycm93IDIgLS0+CiAgPHBhdGggZD0ibTE5LjUgMTQgMyAyLTMgMnYtMWgtMXYtMnoiIGZpbGw9IiNmZmZmZmYiLz4KICA8IS0tIEFydGlmYWN0IEh1YiBub2RlIChyaWdodCkgLS0+CiAgPGNpcmNsZSBjeD0iMjQiIGN5PSIxNiIgcj0iNCIgZmlsbD0iI2ZmZmZmZiIgc3Ryb2tlPSIjNDU1MjhkIiBzdHJva2Utd2lkdGg9IjEiLz4KICA8dGV4dCB4PSIyNCIgeT0iMTguNSIgZm9udC1mYW1pbHk9IkFyaWFsIiBmb250LXNpemU9IjYiIGZpbGw9IiM0NTUyOGQiIHRleHQtYW5jaG9yPSJtaWRkbGUiIGZvbnQtd2VpZ2h0PSJib2xkIj5BPC90ZXh0Pgo8L3N2Zz4K">
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            line-height: 1.6;
            background: linear-gradient(135deg, #f8fafc 0%, #e2e8f0 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            color: #333;
        }

        .container {
            background: white;
            border-radius: 20px;
            box-shadow: 0 20px 40px rgba(0, 0, 0, 0.1);
            max-width: 800px;
            width: 90%;
            padding: 3rem;
            text-align: center;
            position: relative;
            overflow: hidden;
        }

        .container::before {
            content: '';
            position: absolute;
            top: 0;
            left: 0;
            right: 0;
            height: 5px;
            background: linear-gradient(90deg, #667eea, #764ba2);
        }

        .logo-section {
            display: flex;
            justify-content: center;
            align-items: center;
            gap: 2rem;
            margin-bottom: 2rem;
            flex-wrap: wrap;
        }

        .logo {
            width: 120px;
            height: 120px;
            object-fit: contain;
            filter: drop-shadow(0 4px 8px rgba(0, 0, 0, 0.1));
            transition: transform 0.2s;
        }

        .logo:hover {
            transform: scale(1.05);
        }

        .arrow {
            font-size: 2rem;
            color: #667eea;
            font-weight: bold;
        }

        h1 {
            color: #2d3748;
            font-size: 2.5rem;
            font-weight: 700;
            margin-bottom: 1rem;
            background: linear-gradient(135deg, #667eea, #764ba2);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
        }

        .subtitle {
            color: #718096;
            font-size: 1.2rem;
            margin-bottom: 2rem;
            max-width: 600px;
            margin-left: auto;
            margin-right: auto;
        }

        .features {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 1.5rem;
            margin: 2rem 0;
        }

        .feature {
            background: #f7fafc;
            padding: 1.5rem;
            border-radius: 12px;
            border: 1px solid #e2e8f0;
            transition: transform 0.2s, box-shadow 0.2s;
        }

        .feature:hover {
            transform: translateY(-2px);
            box-shadow: 0 8px 25px rgba(0, 0, 0, 0.1);
        }

        .feature-icon {
            font-size: 2rem;
            margin-bottom: 1rem;
        }

        .feature h3 {
            color: #2d3748;
            font-size: 1.1rem;
            margin-bottom: 0.5rem;
        }

        .feature p {
            color: #718096;
            font-size: 0.9rem;
        }

        .api-endpoints {
            background: #f8f9fa;
            border-radius: 12px;
            padding: 1.5rem;
            margin: 2rem 0;
            text-align: left;
        }

        .api-endpoints h3 {
            color: #2d3748;
            margin-bottom: 1rem;
            text-align: center;
        }

        .endpoint {
            background: white;
            padding: 0.75rem 1rem;
            margin: 0.5rem 0;
            border-radius: 8px;
            border-left: 4px solid #667eea;
            font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
            font-size: 0.9rem;
        }

        .method {
            color: #38a169;
            font-weight: bold;
            margin-right: 0.5rem;
        }

        .footer {
            margin-top: 2rem;
            padding-top: 1.5rem;
            border-top: 1px solid #e2e8f0;
            color: #718096;
            font-size: 0.9rem;
        }

        .disclaimer {
            margin-top: 2rem;
            padding-top: 1rem;
            border-top: 1px solid #f1f5f9;
            color: #94a3b8;
            font-size: 0.75rem;
            text-align: center;
            line-height: 1.4;
        }

        .disclaimer a {
            color: #667eea;
            text-decoration: none;
        }

        .disclaimer a:hover {
            text-decoration: underline;
        }

        .status-badge {
            display: inline-flex;
            align-items: center;
            background: #48bb78;
            color: white;
            padding: 0.25rem 0.75rem;
            border-radius: 20px;
            font-size: 0.8rem;
            font-weight: 500;
            margin-bottom: 1rem;
        }

        .status-dot {
            width: 8px;
            height: 8px;
            background: white;
            border-radius: 50%;
            margin-right: 0.5rem;
            animation: pulse 2s infinite;
        }

        @keyframes pulse {
            0%, 100% { opacity: 1; }
            50% { opacity: 0.5; }
        }

        @media (max-width: 768px) {
            .container {
                padding: 2rem;
            }

            h1 {
                font-size: 2rem;
            }

            .logo-section {
                gap: 1rem;
            }

            .logo {
                width: 100px;
                height: 100px;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="status-badge">
            <div class="status-dot"></div>
            Service Running
        </div>

        <div class="logo-section">
            <a href="https://tekton.dev" target="_blank" rel="noopener noreferrer">
                <img src="https://tekton.dev/images/tekton-horizontal-color.png" alt="Tekton Logo" class="logo" style="width: 160px;">
            </a>
            <div class="arrow">→</div>
            <a href="https://artifacthub.io" target="_blank" rel="noopener noreferrer">
                <img src="https://www.cncf.io/wp-content/uploads/2023/04/artifacthub-horizontal-color.svg" alt="Artifact Hub Logo" class="logo">
            </a>
        </div>

        <h1>Tekton Hub to Artifact Hub Proxy</h1>
        <p class="subtitle">
            A transition proxy that bridges Tekton Hub API calls to Artifact Hub.
            <strong>For migration assistance only</strong> - users should transition to using Artifact Hub directly.
        </p>

        <div class="features">
            <div class="feature">
                <div class="feature-icon">🔄</div>
                <h3>API Translation</h3>
                <p>Converts Tekton Hub API endpoints to Artifact Hub format automatically</p>
            </div>
            <div class="feature">
                <div class="feature-icon">🗂️</div>
                <h3>Catalog Mapping</h3>
                <p>Configurable mapping between Tekton Hub and Artifact Hub catalog names</p>
            </div>
            <div class="feature">
                <div class="feature-icon">📦</div>
                <h3>Version Conversion</h3>
                <p>Handles conversion between simplified semver (0.1) and full semver (0.1.0)</p>
            </div>
            <div class="feature">
                <div class="feature-icon">⚡</div>
                <h3>High Performance</h3>
                <p>Built with Go for optimal performance and includes intelligent caching with {{.CacheTTL}} TTL to Artifact Hub</p>
            </div>
        </div>

        <div class="api-endpoints">
            <h3>Available API Endpoints</h3>
            <div class="endpoint"><span class="method">GET</span>/v1/catalogs</div>
            <div class="endpoint"><span class="method">GET</span>/v1/resources</div>
            <div class="endpoint"><span class="method">GET</span>/v1/resource/{catalog}/{kind}/{name}</div>
            <div class="endpoint"><span class="method">GET</span>/v1/resource/{catalog}/{kind}/{name}/{version}</div>
            <div class="endpoint"><span class="method">GET</span>/health</div>
        </div>

        <div class="api-endpoints">
            <h3>🧪 Testing with <a href="https://tekton.dev/docs/pipelines/hub-resolver/">Tekton Hub Resolver</a></h3>
            <p style="margin-bottom: 1rem; color: #4a5568; font-size: 0.9rem;">Configure your Tekton Hub Resolver to use this proxy, then test with a simple PipelineRun:</p>

            <div style="background: #f8f9fa; padding: 1rem; border-radius: 8px; margin-bottom: 1rem;">
                <h4 style="color: #2d3748; margin-bottom: 0.5rem; font-size: 0.9rem;">1. Configure Tekton Hub Resolver:</h4>
                <div class="endpoint" style="font-size: 0.8rem; margin: 0;">
                    kubectl set env -n tekton-pipelines-resolvers deployments.app/tekton-pipelines-remote-resolvers TEKTON_HUB_API=https://tknhub.pipelinesascode.com/
                </div>
            </div>

            <div style="background: #f8f9fa; padding: 1rem; border-radius: 8px;">
                <h4 style="color: #2d3748; margin-bottom: 0.5rem; font-size: 0.9rem;">2. Test with Example PipelineRun:</h4>
                <pre style="background: white; padding: 0.75rem; border-radius: 4px; font-size: 0.7rem; color: #4a5568; margin: 0; overflow-x: auto; border-left: 3px solid #667eea;">apiVersion: tekton.dev/v1
kind: PipelineRun
metadata:
  generateName: hub-test-
spec:
  pipelineSpec:
    tasks:
      - name: fetch-repo
        taskRef:
          resolver: hub
          params:
            - name: kind
              value: task
            - name: name
              value: tkn
            - name: version
              value: "0.4"
            - name: type
              value: tekton
            - name: catalog
              value: tekton</pre>
            </div>
        </div>

        <div class="api-endpoints">
						<h3>🚀 Testing with <a href="https://pipelinesascode.com/">Pipelines-as-Code</a></h3>
            <p style="margin-bottom: 1rem; color: #4a5568; font-size: 0.9rem;">
                <strong>Note:</strong> Latest Pipelines-as-Code versions automatically use Artifact Hub.
                This configuration is only needed for older versions that cannot be upgraded:
            </p>

            <div style="background: #f8f9fa; padding: 1rem; border-radius: 8px; margin-bottom: 1rem;">
                <h4 style="color: #2d3748; margin-bottom: 0.5rem; font-size: 0.9rem;">Add to PaC ConfigMap:</h4>
                <pre style="background: white; padding: 0.75rem; border-radius: 4px; font-size: 0.7rem; color: #4a5568; margin: 0; overflow-x: auto; border-left: 3px solid #667eea;">hub-url: https://tknhub.pipelinesascode.com/
hub-catalog-type: tektonhub</pre>
            </div>

            <div style="background: #f8f9fa; padding: 1rem; border-radius: 8px;">
                <p style="color: #4a5568; font-size: 0.8rem; margin: 0;">
                    See <a href="https://pipelinesascode.com/docs/install/settings/" target="_blank" rel="noopener noreferrer" style="color: #667eea;">PaC Settings Documentation</a> for complete configuration details.
                </p>
            </div>
        </div>

        <div class="cache-warning">
            <h3 style="color: #f56565; margin-bottom: 1rem; display: flex; align-items: center;">
                <span style="margin-right: 0.5rem;">⚠️</span>
                Cache Notice
            </h3>
            <p style="background: #fed7d7; color: #742a2a; padding: 1rem; border-radius: 8px; margin-bottom: 1rem; line-height: 1.5;">
                <strong>Important:</strong> This proxy caches Artifact Hub responses for {{.CacheTTL}} to improve performance.
                If tasks or pipelines are updated in Artifact Hub, you may see stale data until the cache expires.
								For critical updates, switch to use https://artifacthub.io/ directly.
            </p>
        </div>

        <div class="footer">
            <p>🚀 Ready to serve Tekton Hub API requests backed by Artifact Hub</p>
            <p>Visit <code>/health</code> for service status • <code>/v1/catalogs</code> to explore available catalogs</p>
        </div>

        <div class="disclaimer">
            <p>Service gratuitously provided by <a href="https://pipelinesascode.com" target="_blank" rel="noopener noreferrer">pipelinesascode.com</a></p>
						<p>Not affiliated with Tekton or Artifact Hub projects</p>
						<p>No guarantee of uptime or SLAs • Use at your own risk</p>
            <p>Source code: <a href="https://github.com/chmouel/tekton-hub-proxy" target="_blank" rel="noopener noreferrer">chmouel/tekton-hub-proxy</a> • For a best effort support contact <a href="https://x.com/chmouel" target="_blank" rel="noopener noreferrer">@chmouel</a> on X</p>
        </div>
    </div>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t, err := template.New("landing").Parse(tmpl)
	if err != nil {
		logrus.WithError(err).Error("Failed to parse landing page template")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := struct {
		CacheTTL string
	}{
		CacheTTL: h.config.ArtifactHub.Cache.TTL.String(),
	}

	if err := t.Execute(w, data); err != nil {
		logrus.WithError(err).Error("Failed to execute landing page template")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
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
		"original_catalog":     catalog,
		"translated_catalog":   artifactHubCatalog,
		"original_kind":        kind,
		"translated_repo_kind": repoKind,
		"name":                 name,
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
	_, _ = w.Write([]byte(pkg.Data.ManifestRaw))
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
	_, _ = w.Write([]byte(pkg.Data.ManifestRaw))
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

func (h *Handlers) writeJSONResponse(w http.ResponseWriter, statusCode int, data any) {
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
