package translator

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/sirupsen/logrus"
)

type VersionTranslator struct {
	semverRegex *regexp.Regexp
}

func NewVersionTranslator() *VersionTranslator {
	return &VersionTranslator{
		semverRegex: regexp.MustCompile(`^(\d+)\.(\d+)(?:\.(\d+))?`),
	}
}

func (v *VersionTranslator) TektonToArtifactHub(tektonVersion string) (string, error) {
	if tektonVersion == "" {
		return "", nil
	}

	logrus.WithFields(logrus.Fields{
		"translation_type": "version",
		"direction":        "tekton_to_artifacthub",
		"input":            tektonVersion,
	}).Debug("🔄 Translation request")

	// If it's already a full semver (x.y.z), return as-is
	if v.isFullSemver(tektonVersion) {
		logrus.WithFields(logrus.Fields{
			"translation_type": "version",
			"direction":        "tekton_to_artifacthub",
			"input":            tektonVersion,
			"output":           tektonVersion,
			"status":           "unchanged_full_semver",
		}).Debug("✅ Already full semver, no conversion needed")
		return tektonVersion, nil
	}

	// If it's simplified semver (x.y), append .0
	if v.isSimplifiedSemver(tektonVersion) {
		artifactHubVersion := tektonVersion + ".0"
		logrus.WithFields(logrus.Fields{
			"translation_type":    "version",
			"direction":           "tekton_to_artifacthub",
			"input":               tektonVersion,
			"output":              artifactHubVersion,
			"status":              "converted_simplified_to_full",
		}).Debug("✅ Converted simplified semver to full semver")
		return artifactHubVersion, nil
	}

	// For other formats, try to parse and normalize
	ver, err := version.NewVersion(tektonVersion)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"translation_type": "version",
			"direction":        "tekton_to_artifacthub",
			"input":            tektonVersion,
			"output":           tektonVersion,
			"status":           "invalid_passthrough",
			"error":            err.Error(),
		}).Debug("⚠️  Invalid version format, using as-is")
		return tektonVersion, nil
	}

	normalized := ver.String()
	logrus.WithFields(logrus.Fields{
		"translation_type": "version",
		"direction":        "tekton_to_artifacthub",
		"input":            tektonVersion,
		"output":           normalized,
		"status":           "normalized",
	}).Debug("✅ Normalized version")

	return normalized, nil
}

func (v *VersionTranslator) ArtifactHubToTekton(artifactHubVersion string) (string, error) {
	if artifactHubVersion == "" {
		return "", nil
	}

	logrus.WithFields(logrus.Fields{
		"translation_type": "version",
		"direction":        "artifacthub_to_tekton",
		"input":            artifactHubVersion,
	}).Debug("🔄 Translation request")

	// Parse the version
	ver, err := version.NewVersion(artifactHubVersion)
	if err != nil {
		logrus.WithField("version", artifactHubVersion).Warn("Invalid version format, using as-is")
		return artifactHubVersion, nil
	}

	// If it has a pre-release, don't simplify it.
	if ver.Prerelease() != "" {
		return artifactHubVersion, nil
	}

	segments := ver.Segments()
	if len(segments) >= 2 {
		// For versions like 1.2.0, convert to 1.2 (simplified semver)
		if len(segments) >= 3 && segments[2] == 0 {
			tektonVersion := fmt.Sprintf("%d.%d", segments[0], segments[1])
			logrus.WithFields(logrus.Fields{
				"artifacthub_version": artifactHubVersion,
				"tekton_version":      tektonVersion,
			}).Debug("Converted full semver to simplified semver")
			return tektonVersion, nil
		}
	}

	return artifactHubVersion, nil
}

func (v *VersionTranslator) isFullSemver(versionStr string) bool {
	matches := v.semverRegex.FindStringSubmatch(versionStr)
	return len(matches) >= 4 && matches[3] != ""
}

func (v *VersionTranslator) isSimplifiedSemver(versionStr string) bool {
	matches := v.semverRegex.FindStringSubmatch(versionStr)
	return len(matches) >= 3 && matches[3] == "" && strings.Count(versionStr, ".") == 1
}

func (v *VersionTranslator) ValidateVersion(versionStr string) error {
	if versionStr == "" {
		return nil
	}

	_, err := version.NewVersion(versionStr)
	if err != nil {
		return fmt.Errorf("invalid version format: %s", versionStr)
	}

	return nil
}

func (v *VersionTranslator) CompareVersions(v1, v2 string) (int, error) {
	ver1, err := version.NewVersion(v1)
	if err != nil {
		return 0, fmt.Errorf("invalid version v1: %s", v1)
	}

	ver2, err := version.NewVersion(v2)
	if err != nil {
		return 0, fmt.Errorf("invalid version v2: %s", v2)
	}

	return ver1.Compare(ver2), nil
}