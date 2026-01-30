module github.com/Flying-Bird1999/analyzer-ts

go 1.25.0

replace (
	github.com/microsoft/typescript-go/shim/api => ./shim/api
	github.com/microsoft/typescript-go/shim/api/encoder => ./shim/api/encoder
	github.com/microsoft/typescript-go/shim/ast => ./shim/ast
	github.com/microsoft/typescript-go/shim/astnav => ./shim/astnav
	github.com/microsoft/typescript-go/shim/bundled => ./shim/bundled
	github.com/microsoft/typescript-go/shim/checker => ./shim/checker
	github.com/microsoft/typescript-go/shim/collections => ./shim/collections
	github.com/microsoft/typescript-go/shim/compiler => ./shim/compiler
	github.com/microsoft/typescript-go/shim/core => ./shim/core
	github.com/microsoft/typescript-go/shim/ls => ./shim/ls
	github.com/microsoft/typescript-go/shim/ls/lsconv => ./shim/ls/lsconv
	github.com/microsoft/typescript-go/shim/lsp/lsproto => ./shim/lsp/lsproto
	github.com/microsoft/typescript-go/shim/parser => ./shim/parser
	github.com/microsoft/typescript-go/shim/project => ./shim/project
	github.com/microsoft/typescript-go/shim/scanner => ./shim/scanner
	github.com/microsoft/typescript-go/shim/tsoptions => ./shim/tsoptions
	github.com/microsoft/typescript-go/shim/tspath => ./shim/tspath
	github.com/microsoft/typescript-go/shim/vfs => ./shim/vfs
	github.com/microsoft/typescript-go/shim/vfs/cachedvfs => ./shim/vfs/cachedvfs
	github.com/microsoft/typescript-go/shim/vfs/osvfs => ./shim/vfs/osvfs
	github.com/microsoft/typescript-go/shim/vfs/vfstest => ./shim/vfs/vfstest
)

require (
	github.com/gobwas/glob v0.2.3
	github.com/jmespath/go-jmespath v0.4.0
	github.com/mattn/go-sqlite3 v1.14.32
	github.com/microsoft/typescript-go/shim/ast v0.0.0
	github.com/microsoft/typescript-go/shim/astnav v0.0.0
	github.com/microsoft/typescript-go/shim/bundled v0.0.0
	github.com/microsoft/typescript-go/shim/core v0.0.0
	github.com/microsoft/typescript-go/shim/ls/lsconv v0.0.0
	github.com/microsoft/typescript-go/shim/lsp/lsproto v0.0.0
	github.com/microsoft/typescript-go/shim/parser v0.0.0
	github.com/microsoft/typescript-go/shim/project v0.0.0
	github.com/microsoft/typescript-go/shim/vfs/vfstest v0.0.0-00010101000000-000000000000
	github.com/samber/lo v1.50.0
	github.com/spf13/cobra v1.9.1
	github.com/stretchr/testify v1.10.0
	github.com/tidwall/jsonc v0.3.2
	golang.org/x/text v0.32.0
	golang.org/x/tools v0.40.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dlclark/regexp2 v1.11.5 // indirect
	github.com/go-json-experiment/json v0.0.0-20251027170946-4849db3c2f7e // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/klauspost/cpuid/v2 v2.0.9 // indirect
	github.com/microsoft/typescript-go v0.0.0-20260130001633-62596ef8577a // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	github.com/zeebo/xxh3 v1.0.2 // indirect
	golang.org/x/mod v0.31.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
