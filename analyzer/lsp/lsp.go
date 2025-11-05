package lsp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/astnav"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/bundled"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ls"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/lsp/lsproto"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/project"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/project/logging"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/vfs/vfstest"
	"github.com/Flying-Bird1999/analyzer-ts/analyzer/utils"
)

// dummyClient 是 project.Client 接口的一个空实现。
type dummyClient struct{}

func (c *dummyClient) WatchFiles(ctx context.Context, id project.WatcherID, watchers []*lsproto.FileSystemWatcher) error {
	return nil
}
func (c *dummyClient) UnwatchFiles(ctx context.Context, id project.WatcherID) error {
	return nil
}
func (c *dummyClient) RefreshDiagnostics(ctx context.Context) error {
	return nil
}

// dummyNpmExecutor 是 project.ata.NpmExecutor 接口的一个空实现。
type dummyNpmExecutor struct{}

func (n *dummyNpmExecutor) NpmInstall(cwd string, args []string) ([]byte, error) {
	return nil, nil
}

// Service 管理 TypeScript 项目的 LSP 语言服务会话。
type Service struct {
	session      *project.Session
	rootPath     string
	sourcesCache map[string]any
}

// NewService 从物理磁盘创建服务。
func NewService(rootPath string) (*Service, error) {
	files := make(map[string]any)
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			content, readErr := os.ReadFile(path)
			if readErr != nil {
				return nil
			}
			virtualPath, err := filepath.Rel(rootPath, path)
			if err != nil {
				return err
			}
			files["/"+filepath.ToSlash(virtualPath)] = string(content)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("遍历项目文件失败: %w", err)
	}

	return NewServiceForTest(files)
}

// NewServiceForTest 是一个专为测试设计的构造函数，从内存 map 创建服务。
func NewServiceForTest(files map[string]any) (*Service, error) {
	fs := bundled.WrapFS(vfstest.FromMap(files, false))

	session := project.NewSession(&project.SessionInit{
		Options: &project.SessionOptions{
			CurrentDirectory:   "/",
			DefaultLibraryPath: bundled.LibPath(),
			WatchEnabled:       false,
			LoggingEnabled:     false,
		},
		FS:          fs,
		Client:      &dummyClient{},
		NpmExecutor: &dummyNpmExecutor{},
		Logger:      logging.NewLogger(os.Stderr),
	})

	service := &Service{
		session:      session,
		rootPath:     "/",
		sourcesCache: files,
	}

	// 显式地“打开”所有文件，特别是 tsconfig.json，以提示语言服务创建配置项目
	var firstURI lsproto.DocumentUri
	for path, content := range files {
		uri := ls.FileNameToDocumentURI(path)
		if firstURI == "" && strings.HasSuffix(path, ".ts") {
			firstURI = uri
		}
		session.DidOpenFile(context.Background(), uri, 0, content.(string), "typescript") // ScriptKind 可能需要更精确
	}

	// 尝试获取一次语言服务，这可能会触发项目的完整构建
	if firstURI != "" {
		_, _ = session.GetLanguageService(context.Background(), firstURI)
	}

	return service, nil
}

// FindReferences 在给定位置查找一个符号的所有引用 (LSP 实现)。
func (s *Service) FindReferences(ctx context.Context, filePath string, line, char int) (response lsproto.ReferencesResponse, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered from panic in FindReferences: %v", r)
		}
	}()

	virtualPath, err := filepath.Rel(s.rootPath, filePath)
	if err != nil {
		return lsproto.ReferencesResponse{}, fmt.Errorf("计算相对路径失败: %w", err)
	}
	virtualPath = "/" + filepath.ToSlash(virtualPath)

	content, ok := s.sourcesCache[virtualPath].(string)
	if !ok {
		return lsproto.ReferencesResponse{}, fmt.Errorf("无法从缓存中找到文件内容: %s", virtualPath)
	}
	uri := ls.FileNameToDocumentURI(virtualPath)
	s.session.DidOpenFile(ctx, uri, 0, content, "typescript")

	langService, err := s.session.GetLanguageService(ctx, uri)
	if err != nil {
		return lsproto.ReferencesResponse{}, fmt.Errorf("无法获取语言服务: %w", err)
	}

	params := &lsproto.ReferenceParams{
		TextDocument: lsproto.TextDocumentIdentifier{Uri: uri},
		Position:     lsproto.Position{Line: uint32(line - 1), Character: uint32(char - 1)},
		Context:      &lsproto.ReferenceContext{IncludeDeclaration: true},
	}

	response, err = langService.ProvideReferences(ctx, params)
	if err != nil {
		return lsproto.ReferencesResponse{}, fmt.Errorf("查找引用失败: %w", err)
	}

	return response, nil
}

// GotoDefinition 在给定位置查找符号的定义位置 (LSP 实现)。
func (s *Service) GotoDefinition(ctx context.Context, filePath string, line, char int) (response lsproto.DefinitionResponse, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered from panic in GotoDefinition: %v", r)
		}
	}()

	virtualPath, err := filepath.Rel(s.rootPath, filePath)
	if err != nil {
		return lsproto.DefinitionResponse{}, fmt.Errorf("计算相对路径失败: %w", err)
	}
	virtualPath = "/" + filepath.ToSlash(virtualPath)

	content, ok := s.sourcesCache[virtualPath].(string)
	if !ok {
		return lsproto.DefinitionResponse{}, fmt.Errorf("无法从缓存中找到文件内容: %s", virtualPath)
	}
	uri := ls.FileNameToDocumentURI(virtualPath)
	s.session.DidOpenFile(ctx, uri, 0, content, "typescript")

	langService, err := s.session.GetLanguageService(ctx, uri)
	if err != nil {
		return lsproto.DefinitionResponse{}, fmt.Errorf("无法获取语言服务: %w", err)
	}

	response, err = langService.ProvideDefinition(ctx, uri, lsproto.Position{
		Line:      uint32(line - 1),
		Character: uint32(char - 1),
	})
	if err != nil {
		return lsproto.DefinitionResponse{}, fmt.Errorf("查找定义失败: %w", err)
	}

	return response, nil
}

// GetSymbolAt 获取给定文件位置的符号信息 (LSP 实现)。
func (s *Service) GetSymbolAt(ctx context.Context, filePath string, line, char int) (*ast.Symbol, error) {
	virtualPath, err := filepath.Rel(s.rootPath, filePath)
	if err != nil {
		return nil, fmt.Errorf("计算相对路径失败: %w", err)
	}
	virtualPath = "/" + filepath.ToSlash(virtualPath)
	uri := ls.FileNameToDocumentURI(virtualPath)

	langService, err := s.session.GetLanguageService(ctx, uri)
	if err != nil {
		return nil, fmt.Errorf("无法获取语言服务: %w", err)
	}
	program := langService.GetProgram()
	if program == nil {
		return nil, fmt.Errorf("无法从语言服务获取 program")
	}
	file := program.GetSourceFile(virtualPath)
	if file == nil {
		return nil, fmt.Errorf("无法在 program 中找到文件: %s", virtualPath)
	}

	checker, done := program.GetTypeCheckerForFile(ctx, file)
	defer done()

	content, ok := s.sourcesCache[virtualPath].(string)
	if !ok {
		return nil, fmt.Errorf("无法从缓存中找到文件内容: %s", virtualPath)
	}
	lines := strings.Split(content, "\n")
	if line-1 >= len(lines) {
		return nil, fmt.Errorf("行号 %d 超出文件范围", line)
	}
	pos := 0
	for i := 0; i < line-1; i++ {
		pos += len(lines[i]) + 1
	}
	pos += char - 1

	node := astnav.GetTouchingPropertyName(file, pos)

	return checker.GetSymbolAtLocation(node), nil
}

// Close 关闭 LSP 会话以释放资源。
func (s *Service) Close() {
	s.session.Close()
}

// GetNativeQuickInfoAtPosition 获取原生 TypeScript 的 QuickInfo 信息。
// 这个方法直接调用 TypeScript 的原生 QuickInfo 功能，可以获取更完整的显示部件和类型信息。
func (s *Service) GetNativeQuickInfoAtPosition(ctx context.Context, filePath string, line, char int) (*QuickInfo, error) {
	// 计算虚拟路径
	virtualPath, err := filepath.Rel(s.rootPath, filePath)
	if err != nil {
		return nil, fmt.Errorf("计算相对路径失败: %w", err)
	}
	virtualPath = "/" + filepath.ToSlash(virtualPath)
	uri := ls.FileNameToDocumentURI(virtualPath)

	// 获取语言服务实例
	langService, err := s.session.GetLanguageService(ctx, uri)
	if err != nil {
		return nil, fmt.Errorf("无法获取语言服务: %w", err)
	}

	// 使用原生 ProvideHover 方法获取 QuickInfo
	hoverResponse, err := langService.ProvideHover(ctx, uri, lsproto.Position{
		Line:      uint32(line - 1),
		Character: uint32(char - 1),
	})
	if err != nil {
		return nil, fmt.Errorf("获取原生 QuickInfo 失败: %w", err)
	}

	if hoverResponse.Hover == nil || hoverResponse.Hover.Contents.MarkupContent == nil {
		return nil, nil
	}

	// 解析 markdown 内容为 QuickInfo 结构
	return s.parseHoverContent(hoverResponse.Hover)
}

// QuickInfo 表示 TypeScript 中的类型提示信息，包含符号的类型、文档和其他相关信息。
// 这是实现类似 VSCode 悬停提示功能的核心数据结构。
type QuickInfo struct {
	// 类型文本，显示符号的完整类型信息
	TypeText string
	// 文档字符串，包含 JSDoc 注释等文档信息
	Documentation string
	// 显示部件，用于结构化地展示类型信息
	DisplayParts []SymbolDisplayPart
	// 范围信息，表示提示信息对应的源码位置范围
	Range *lsproto.Range
}

// SymbolDisplayPart 表示符号显示的一个组成部分，用于结构化地展示类型信息。
// 每个部分都有文本和类型，支持不同样式的显示。
type SymbolDisplayPart struct {
	// 显示的文本内容
	Text string
	// 部件的类型，如 "className", "parameterName", "text" 等
	Kind string
}

// GetQuickInfoAtPosition 获取指定位置的 QuickInfo（类型提示）信息。
// 这个方法实现了类似 VSCode 中悬停提示的功能，可以显示变量、函数、类型等的类型信息和文档。
//
// 参数：
//   - ctx: 上下文对象
//   - filePath: 文件路径
//   - line: 行号（1-based）
//   - char: 列号（1-based）
//
// 返回值：
//   - *QuickInfo: 类型提示信息，如果位置没有有效符号则返回 nil
//   - error: 错误信息
//
// 示例：
//   quickInfo, err := service.GetQuickInfoAtPosition(ctx, "/path/to/file.ts", 10, 5)
//   if err != nil {
//       return err
//   }
//   if quickInfo != nil {
//       fmt.Printf("类型: %s\n", quickInfo.TypeText)
//       fmt.Printf("文档: %s\n", quickInfo.Documentation)
//   }
func (s *Service) GetQuickInfoAtPosition(ctx context.Context, filePath string, line, char int) (*QuickInfo, error) {
	// 计算虚拟路径（相对于项目根目录的路径）
	virtualPath, err := filepath.Rel(s.rootPath, filePath)
	if err != nil {
		return nil, fmt.Errorf("计算相对路径失败: %w", err)
	}
	virtualPath = "/" + filepath.ToSlash(virtualPath)
	uri := ls.FileNameToDocumentURI(virtualPath)

	// 获取语言服务实例
	langService, err := s.session.GetLanguageService(ctx, uri)
	if err != nil {
		return nil, fmt.Errorf("无法获取语言服务: %w", err)
	}

	// 获取编译程序和源文件
	program := langService.GetProgram()
	if program == nil {
		return nil, fmt.Errorf("无法从语言服务获取 program")
	}
	file := program.GetSourceFile(virtualPath)
	if file == nil {
		return nil, fmt.Errorf("无法在 program 中找到文件: %s", virtualPath)
	}

	// 获取类型检查器（使用完成后需要调用 done() 释放资源）
	checker, done := program.GetTypeCheckerForFile(ctx, file)
	defer done()

	// 计算目标位置的字符偏移量
	content, ok := s.sourcesCache[virtualPath].(string)
	if !ok {
		return nil, fmt.Errorf("无法从缓存中找到文件内容: %s", virtualPath)
	}
	lines := strings.Split(content, "\n")
	if line-1 >= len(lines) {
		return nil, fmt.Errorf("行号 %d 超出文件范围", line)
	}
	pos := 0
	for i := 0; i < line-1; i++ {
		pos += len(lines[i]) + 1
	}
	pos += char - 1

	// 使用 AST 导航找到目标位置的节点
	node := astnav.GetTouchingPropertyName(file, pos)
	if node == nil || node.Kind == ast.KindSourceFile {
		// 避免为整个源文件或无效节点提供 quickInfo
		return nil, nil
	}

	// 获取节点位置的符号信息
	symbol := checker.GetSymbolAtLocation(node)
	if symbol == nil {
		return nil, nil
	}

	// 获取符号在当前位置的类型信息
	symbolType := checker.GetTypeOfSymbolAtLocation(symbol, node)
	if symbolType == nil {
		return nil, nil
	}

	// 构建返回结果
	quickInfo := &QuickInfo{
		TypeText: checker.TypeToString(symbolType),
		// 从符号中提取文档信息
		Documentation: s.extractDocumentation(symbol),
		// 构建显示部件列表
		DisplayParts: s.buildDisplayParts(symbolType, checker),
		// 设置范围信息
		Range: s.createRange(node, content),
	}

	return quickInfo, nil
}

// extractDocumentation 从符号中提取文档信息，包括 JSDoc 注释。
// 这是一个辅助方法，用于从 TypeScript 符号中获取相关的文档字符串。
func (s *Service) extractDocumentation(symbol *ast.Symbol) string {
	if symbol == nil || symbol.Declarations == nil || len(symbol.Declarations) == 0 {
		return ""
	}

	// 当前简化实现：返回空字符串
	// 后续可以从符号的声明节点中提取 JSDoc 注释
	// 例如：遍历 symbol.Declarations，查找其中的 JSDoc 注释
	return ""
}

// parseHoverContent 解析原生 Hover 响应为 QuickInfo 结构。
// 这个方法解析 TypeScript 语言服务返回的 markdown 格式内容，提取类型信息和文档。
func (s *Service) parseHoverContent(hover *lsproto.Hover) (*QuickInfo, error) {
	markdownContent := hover.Contents.MarkupContent.Value
	if markdownContent == "" {
		return nil, nil
	}

	// 解析 markdown 内容，分离类型信息和文档
	lines := strings.Split(markdownContent, "\n")

	var typeText strings.Builder
	var documentation strings.Builder
	var displayParts []SymbolDisplayPart

	// 解析状态：0=未开始，1=在类型部分，2=在文档部分
	parseState := 0

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		if trimmedLine == "" {
			if parseState == 1 {
				// 空行结束类型部分，开始文档部分
				parseState = 2
				continue
			}
			continue
		}

		if parseState == 0 {
			// 开始解析类型部分
			parseState = 1
		}

		if parseState == 1 {
			// 类型信息部分
			if typeText.Len() > 0 {
				typeText.WriteString("\n")
			}
			typeText.WriteString(trimmedLine)

			// 同时解析显示部件
			displayParts = append(displayParts, s.parseLineToDisplayParts(trimmedLine)...)
		} else if parseState == 2 {
			// 文档信息部分
			if documentation.Len() > 0 {
				documentation.WriteString("\n")
			}
			documentation.WriteString(trimmedLine)
		}
	}

	return &QuickInfo{
		TypeText:      typeText.String(),
		Documentation: documentation.String(),
		DisplayParts:  displayParts,
		Range:         hover.Range,
	}, nil
}

// parseLineToDisplayParts 将单行 QuickInfo 文本解析为显示部件。
// 这个方法识别不同的语法元素并为它们分配适当的语义类型。
func (s *Service) parseLineToDisplayParts(line string) []SymbolDisplayPart {
	var parts []SymbolDisplayPart

	// 使用正则表达式匹配常见的 TypeScript QuickInfo 模式
	patterns := []struct {
		regex   string
		kind    string
		prefix  string
	}{
		{`^\(function\) `, "functionDeclaration", "(function) "},
		{`^\(method\) `, "methodDeclaration", "(method) "},
		{`^\(property\) `, "propertyDeclaration", "(property) "},
		{`^\(parameter\) `, "parameterName", "(parameter) "},
		{`^\(local var\) `, "localVariable", "(local var) "},
		{`^\(var\) `, "variable", "(var) "},
		{`^class `, "keyword", "class "},
		{`^interface `, "keyword", "interface "},
		{`^type `, "keyword", "type "},
		{`^const `, "keyword", "const "},
		{`^let `, "keyword", "let "},
		{`^function `, "keyword", "function "},
		{`^async `, "keyword", "async "},
		{`^\* `, "keyword", "* "},
		{`^readonly `, "keyword", "readonly "},
	}

	matched := false
	for _, pattern := range patterns {
		if strings.HasPrefix(line, pattern.prefix) {
			parts = append(parts, SymbolDisplayPart{
				Text: pattern.prefix,
				Kind: pattern.kind,
			})
			remaining := strings.TrimPrefix(line, pattern.prefix)
			if remaining != "" {
				parts = append(parts, SymbolDisplayPart{
					Text: remaining,
					Kind: "text",
				})
			}
			matched = true
			break
		}
	}

	if !matched {
		// 如果没有匹配的模式，整个作为纯文本
		parts = append(parts, SymbolDisplayPart{
			Text: line,
			Kind: "text",
		})
	}

	return parts
}

// buildDisplayParts 构建类型信息的结构化显示部件。
// 这个方法将复杂类型信息拆分为多个具有语义的部件，支持不同样式的显示。
// 现在使用原生 QuickInfo 解析方法。
func (s *Service) buildDisplayParts(symbolType interface{}, checker interface{}) []SymbolDisplayPart {
	// 对于自定义实现的 QuickInfo，返回空的显示部件
	// 推荐使用 GetNativeQuickInfoAtPosition 来获取完整的显示部件信息
	return []SymbolDisplayPart{}
}

// createRange 根据 AST 节点和源文件内容创建 LSP 范围信息。
// 这个方法将 AST 节点的位置信息转换为 LSP 标准的范围格式。
func (s *Service) createRange(node *ast.Node, content string) *lsproto.Range {
	if node == nil || content == "" {
		return nil
	}

	// 获取节点的起始和结束位置
	startPos := node.Pos()
	endPos := node.End()

	// 计算起始和结束的行列号
	startLine, startChar := utils.GetLineAndCharacterOfPosition(content, startPos)
	endLine, endChar := utils.GetLineAndCharacterOfPosition(content, endPos)

	return &lsproto.Range{
		Start: lsproto.Position{
			Line:      uint32(startLine),
			Character: uint32(startChar),
		},
		End: lsproto.Position{
			Line:      uint32(endLine),
			Character: uint32(endChar),
		},
	}
}