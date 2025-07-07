package parser

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func runTestCase(t *testing.T) {
	t.Run("随便写一个", func(t *testing.T) {
		root, _ := filepath.Abs("./ts/example.ts")

		br := NewParserResult(root)
		br.Traverse()
		assert.Equal(t, "./type2", br.ImportDeclarations[0].Source)
	})
}
