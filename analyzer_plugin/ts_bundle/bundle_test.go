package ts_bundle

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
)

// normalizeString normalizes line endings and removes leading/trailing whitespace.
func normalizeString(s string) string {
	return strings.TrimSpace(strings.ReplaceAll(s, "\r\n", "\n"))
}

func TestGenerateBundle(t *testing.T) {
	testCases := []struct {
		name               string
		entryFile          string
		entryType          string
		expectedOutputFile string
		projectRoot        string
	}{
		{
			name:               "Simple Dependency",
			entryFile:          "test_data/simple/index1.ts",
			entryType:          "Type1",
			expectedOutputFile: "test_data/simple/expected.ts",
			projectRoot:        "test_data/simple",
		},
		{
			name:               "Circular Dependency",
			entryFile:          "test_data/circular/circ1.ts",
			entryType:          "CircType1",
			expectedOutputFile: "test_data/circular/expected.ts",
			projectRoot:        "test_data/circular",
		},
		{
			name:               "Type Name Collision",
			entryFile:          "test_data/collision/coll1.ts",
			entryType:          "Container",
			expectedOutputFile: "test_data/collision/expected.ts",
			projectRoot:        "test_data/collision",
		},
		{
			name:               "Import Alias",
			entryFile:          "test_data/alias/alias1.ts",
			entryType:          "Container",
			expectedOutputFile: "test_data/alias/expected.ts",
			projectRoot:        "test_data/alias",
		},
		{
			name:               "Namespace Import",
			entryFile:          "test_data/namespace/ns1.ts",
			entryType:          "Container",
			expectedOutputFile: "test_data/namespace/expected.ts",
			projectRoot:        "test_data/namespace",
		},
		{
			name:               "Re-export and Default Export",
			entryFile:          "test_data/export/index.ts",
			entryType:          "Container",
			expectedOutputFile: "test_data/export/expected.ts",
			projectRoot:        "test_data/export",
		},
		{
			name:               "Complex Exports",
			entryFile:          "test_data/complex_exports/index.ts",
			entryType:          "Container",
			expectedOutputFile: "test_data/complex_exports/expected.ts",
			projectRoot:        "test_data/complex_exports",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Convert test paths to absolute paths before passing them to the function.
			absEntryFile, err := filepath.Abs(tc.entryFile)
			if err != nil {
				t.Fatalf("Could not get absolute path for entry file '%s': %v", tc.entryFile, err)
			}
			absExpectedFile, err := filepath.Abs(tc.expectedOutputFile)
			if err != nil {
				t.Fatalf("Could not get absolute path for expected file '%s': %v", tc.expectedOutputFile, err)
			}
			absProjectRoot, err := filepath.Abs(tc.projectRoot)
			if err != nil {
				t.Fatalf("Could not get absolute path for project root '%s': %v", tc.projectRoot, err)
			}

			// Run the function to be tested.
			actualBundle := GenerateBundle(absEntryFile, tc.entryType, "", absProjectRoot)

			// Read the expected output file.
			expectedBytes, err := ioutil.ReadFile(absExpectedFile)
			if err != nil {
				t.Fatalf("Failed to read expected output file '%s': %v", absExpectedFile, err)
			}
			expectedBundle := string(expectedBytes)

			// Normalize both strings to account for OS-specific line endings and whitespace.
			normalizedActual := normalizeString(actualBundle)
			normalizedExpected := normalizeString(expectedBundle)

			// Compare actual vs. expected.
			if normalizedActual != normalizedExpected {
				t.Errorf("Bundle mismatch for test '%s'.\n\n--- EXPECTED ---\n%s\n\n--- ACTUAL ---\n%s", tc.name, expectedBundle, actualBundle)
			}
		})
	}
}
