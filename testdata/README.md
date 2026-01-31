# Test Data

This directory contains test projects and fixtures for testing analyzer-ts plugins.

## Structure

```
testdata/
└── test_project/          # Sample TypeScript component library
    ├── .analyzer/
    │   └── component-manifest.json
    └── src/
        └── components/
            ├── Button/    # Base button component
            ├── Input/     # Input component (depends on Button)
            └── Select/    # Select component (depends on Button and Input)
```

## Component Dependency Graph

```
Button (base)
  ↑
  ├── Input (depends on Button)
  │     ↑
  │     └── Select (depends on Input and Button)
  │
  └── Select (directly depends on Button)
```

## Usage in Tests

From `analyzer_plugin/project_analyzer/impact_analysis/`:
```go
testProjectPath := "../../../testdata/test_project"
```

From `analyzer_plugin/project_analyzer/component_deps_v2/`:
```go
testProjectPath := "../../../testdata/test_project"
```

## Running Tests with Real Command

```bash
# Component dependency analysis
./analyzer-ts analyze component-deps-v2 \
  -i testdata/test_project \
  -p "component-deps-v2.manifest=testdata/test_project/.analyzer/component-manifest.json"

# Impact analysis (requires deps file from previous command)
./analyzer-ts analyze impact-analysis \
  -i testdata/test_project \
  -p "impact-analysis.changeFile=examples/changes.json" \
  -p "impact-analysis.depsFile=/path/to/deps-output.json"
```
