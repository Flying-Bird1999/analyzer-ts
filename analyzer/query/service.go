package query

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

// Service 管理一个 TypeScript 项目的分析会话。
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

// FindReferences 在给定位置查找一个符号的所有引用。
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

// GetSymbolAt 获取给定文件位置的符号信息。
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

// Close 关闭会话以释放资源。
func (s *Service) Close() {
	s.session.Close()
}