package query

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/bundled"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ls"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/lsp/lsproto"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/project"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/project/logging"
	"github.com/Zzzen/typescript-go/use-at-your-own-risk/vfs/vfstest"
)

// dummyClient 是 project.Client 接口的一个空实现，
// 因为我们的命令行工具不需要与客户端（如IDE）进行文件监控等交互。
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

// dummyNpmExecutor 是 project.ata.NpmExecutor 接口的一个空实现，
// 因为我们不需要在分析过程中执行npm命令。
type dummyNpmExecutor struct{}

func (n *dummyNpmExecutor) NpmInstall(cwd string, args []string) ([]byte, error) {
	return nil, nil
}

// Service 管理一个 TypeScript 项目的分析会话。
type Service struct {
	session  *project.Session // 底层的 typescript-go 会话对象
	rootPath string      // 保存项目在物理磁盘上的真实根路径
}

// NewService 创建并初始化一个新的查询服务。
func NewService(rootPath string) (*Service, error) {
	// 核心思想：模仿 typescript-go 的内部测试环境，创建一个纯内存的虚拟文件系统（in-memory VFS），
	// 以此绕过直接使用物理文件系统时，因底层库的异步I/O操作而导致的竞争条件问题。

	// 1. 遍历物理文件系统，将所有项目文件一次性加载到内存中的一个 map 里。
	files := make(map[string]any)
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			content, readErr := os.ReadFile(path)
			if readErr != nil {
				// 忽略读取错误，例如一些临时的或无法访问的文件
				return nil
			}
			// 计算文件相对于项目根目录的路径，作为其在虚拟文件系统中的路径。
			// 虚拟路径必须以 "/" 开头，并使用 "/" 作为路径分隔符。
			virtualPath, err := filepath.Rel(rootPath, path)
			if err != nil {
				return err
			}
			files["/"+filepath.ToSlash(virtualPath)] = string(content)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("遍历项目文件以创建内存文件系统失败: %w", err)
	}

	// 2. 基于内存中的文件 map 创建一个 VFS 实例。
	fs := bundled.WrapFS(vfstest.FromMap(files, false))

	// 3. 使用这个内存 VFS 来初始化会话。
	session := project.NewSession(&project.SessionInit{
		Options: &project.SessionOptions{
			// 由于我们使用的是VFS，会话的当前工作目录必须是虚拟文件系统的根目录 "/"。
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
		session:  session,
		rootPath: rootPath,
	}

	return service, nil
}

// 重要提示：此函数在处理包含 tsconfig.json 的“配置项目”时，已知会失败并返回空结果。
// 这是由于其依赖的 typescript-go 库存在一个已知的Bug，导致在跨文件时无法正确解析符号引用。
// 更多细节请参阅: https://github.com/microsoft/typescript-go/issues/1219
//
// 然而，此函数在处理不含 tsconfig.json 的“推断项目”（例如单个文件）时可以正常工作。
//
// FindReferences 在给定位置查找一个符号的所有引用。
func (s *Service) FindReferences(ctx context.Context, filePath string, line, char int) (response lsproto.ReferencesResponse, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered from panic in FindReferences: %v", r)
		}
	}()

	// 1. 将用户传入的真实物理路径，转换为会话内部VFS所能理解的虚拟路径。
	virtualPath, err := filepath.Rel(s.rootPath, filePath)
	if err != nil {
		return lsproto.ReferencesResponse{}, fmt.Errorf("计算相对路径失败: %w", err)
	}
	virtualPath = "/" + filepath.ToSlash(virtualPath)

	// 2. 尽管VFS中已有文件内容，我们仍需调用 DidOpenFile 来触发引擎对该文件的分析，并将其设置为查询上下文。
	content, err := os.ReadFile(filePath)
	if err != nil {
		return lsproto.ReferencesResponse{}, fmt.Errorf("无法读取待查询文件: %w", err)
	}

	// 注意：用于与会话通信的URI，必须基于虚拟路径创建。
	uri := ls.FileNameToDocumentURI(virtualPath)
	s.session.DidOpenFile(ctx, uri, 0, string(content), "typescript")

	// 3. 获取特定于该文件的语言服务实例。
	langService, err := s.session.GetLanguageService(ctx, uri)
	if err != nil {
		return lsproto.ReferencesResponse{}, fmt.Errorf("无法获取语言服务: %w", err)
	}

	// 4. 构造LSP参数并发起请求。
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

// Close 关闭会话以释放资源。
func (s *Service) Close() {
	s.session.Close()
}