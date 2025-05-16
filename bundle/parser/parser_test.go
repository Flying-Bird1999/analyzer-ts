package parser

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func runTestCase(t *testing.T) {
	t.Run("随便写一个", func(t *testing.T) {
		root, _ := filepath.Abs("./ts/example.ts")
		bundle := Traverse(root)
		assert.Equal(t, "./type2", bundle.ImportDeclarations[0].Source)
	})
}
