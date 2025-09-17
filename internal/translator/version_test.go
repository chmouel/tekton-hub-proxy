package translator

import (
	"testing"
)

func TestVersionTranslator_TektonToArtifactHub(t *testing.T) {
	translator := NewVersionTranslator()

	tests := []struct {
		name     string
		input    string
		expected string
		hasError bool
	}{
		{
			name:     "empty version",
			input:    "",
			expected: "",
			hasError: false,
		},
		{
			name:     "simplified semver",
			input:    "0.1",
			expected: "0.1.0",
			hasError: false,
		},
		{
			name:     "full semver",
			input:    "0.1.0",
			expected: "0.1.0",
			hasError: false,
		},
		{
			name:     "major.minor.patch",
			input:    "1.2.3",
			expected: "1.2.3",
			hasError: false,
		},
		{
			name:     "complex version",
			input:    "1.0.0-alpha.1",
			expected: "1.0.0-alpha.1",
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := translator.TektonToArtifactHub(tt.input)

			if tt.hasError && err == nil {
				t.Errorf("expected error but got none")
			}

			if !tt.hasError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestVersionTranslator_ArtifactHubToTekton(t *testing.T) {
	translator := NewVersionTranslator()

	tests := []struct {
		name     string
		input    string
		expected string
		hasError bool
	}{
		{
			name:     "empty version",
			input:    "",
			expected: "",
			hasError: false,
		},
		{
			name:     "patch zero version",
			input:    "0.1.0",
			expected: "0.1",
			hasError: false,
		},
		{
			name:     "patch non-zero version",
			input:    "0.1.1",
			expected: "0.1.1",
			hasError: false,
		},
		{
			name:     "major version",
			input:    "1.0.0",
			expected: "1.0",
			hasError: false,
		},
		{
			name:     "complex version",
			input:    "1.0.0-alpha.1",
			expected: "1.0.0-alpha.1",
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := translator.ArtifactHubToTekton(tt.input)

			if tt.hasError && err == nil {
				t.Errorf("expected error but got none")
			}

			if !tt.hasError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestVersionTranslator_ValidateVersion(t *testing.T) {
	translator := NewVersionTranslator()

	tests := []struct {
		name     string
		input    string
		hasError bool
	}{
		{
			name:     "empty version",
			input:    "",
			hasError: false,
		},
		{
			name:     "valid semver",
			input:    "1.0.0",
			hasError: false,
		},
		{
			name:     "valid simplified semver",
			input:    "1.0",
			hasError: false,
		},
		{
			name:     "invalid version",
			input:    "invalid",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := translator.ValidateVersion(tt.input)

			if tt.hasError && err == nil {
				t.Errorf("expected error but got none")
			}

			if !tt.hasError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestVersionTranslator_CompareVersions(t *testing.T) {
	translator := NewVersionTranslator()

	tests := []struct {
		name     string
		v1       string
		v2       string
		expected int
		hasError bool
	}{
		{
			name:     "equal versions",
			v1:       "1.0.0",
			v2:       "1.0.0",
			expected: 0,
			hasError: false,
		},
		{
			name:     "v1 greater",
			v1:       "1.1.0",
			v2:       "1.0.0",
			expected: 1,
			hasError: false,
		},
		{
			name:     "v1 lesser",
			v1:       "1.0.0",
			v2:       "1.1.0",
			expected: -1,
			hasError: false,
		},
		{
			name:     "invalid v1",
			v1:       "invalid",
			v2:       "1.0.0",
			expected: 0,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := translator.CompareVersions(tt.v1, tt.v2)

			if tt.hasError && err == nil {
				t.Errorf("expected error but got none")
			}

			if !tt.hasError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.hasError && result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}