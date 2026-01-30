package vfstest

import (
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/microsoft/typescript-go/internal/vfs"
)

// memFS is a simple in-memory file system implementing vfs.FS
type memFS struct {
	mu      sync.RWMutex
	files   map[string]*memFile
	useCase bool
	modTime time.Time
}

type memFile struct {
	data    string
	modTime time.Time
	mode    fs.FileMode
}

func newMemFS(m map[string]string, useCaseSensitiveFileNames bool) *memFS {
	fs := &memFS{
		files:   make(map[string]*memFile),
		useCase: useCaseSensitiveFileNames,
		modTime: time.Now(),
	}
	for p, content := range m {
		normPath := strings.TrimPrefix(p, "/")
		fs.files[normPath] = &memFile{
			data:    content,
			modTime: fs.modTime,
			mode:    0644,
		}
	}
	return fs
}

func (m *memFS) UseCaseSensitiveFileNames() bool {
	return m.useCase
}

func (m *memFS) FileExists(p string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	path := strings.TrimPrefix(p, "/")
	_, ok := m.files[path]
	return ok
}

func (m *memFS) ReadFile(p string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	path := strings.TrimPrefix(p, "/")
	f, ok := m.files[path]
	if !ok {
		return "", false
	}
	return f.data, true
}

func (m *memFS) WriteFile(p string, data string, writeByteOrderMark bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	path := strings.TrimPrefix(p, "/")
	m.files[path] = &memFile{
		data:    data,
		modTime: time.Now(),
		mode:    0644,
	}
	return nil
}

func (m *memFS) Remove(p string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	path := strings.TrimPrefix(p, "/")
	delete(m.files, path)
	return nil
}

func (m *memFS) Chtimes(p string, aTime, mTime time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	path := strings.TrimPrefix(p, "/")
	if f, ok := m.files[path]; ok {
		f.modTime = mTime
	}
	return nil
}

func (m *memFS) DirectoryExists(p string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	path := strings.TrimPrefix(p, "/")
	prefix := path
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	for fp := range m.files {
		if strings.HasPrefix(fp, prefix) {
			return true
		}
	}
	return false
}

func (m *memFS) GetAccessibleEntries(p string) vfs.Entries {
	m.mu.RLock()
	defer m.mu.RUnlock()
	path := strings.TrimPrefix(p, "/")
	prefix := path
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	var result vfs.Entries
	seen := make(map[string]bool)

	for fp := range m.files {
		if !strings.HasPrefix(fp, prefix) {
			continue
		}
		rel := strings.TrimPrefix(fp, prefix)
		if idx := strings.Index(rel, "/"); idx >= 0 {
			// This is in a subdirectory, get the directory name
			dir := rel[:idx]
			if !seen[dir] {
				result.Directories = append(result.Directories, dir)
				seen[dir] = true
			}
		} else if rel != "" {
			// This is a file in this directory
			result.Files = append(result.Files, rel)
		}
	}

	sort.Strings(result.Directories)
	sort.Strings(result.Files)
	return result
}

func (m *memFS) Stat(p string) vfs.FileInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()
	path := strings.TrimPrefix(p, "/")
	f, ok := m.files[path]
	if !ok {
		return nil
	}
	return &memFileInfo{
		name:    filepath.Base(p),
		size:    int64(len(f.data)),
		mode:    f.mode,
		modTime: f.modTime,
	}
}

func (m *memFS) WalkDir(root string, walkFn vfs.WalkDirFunc) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	rootPath := strings.TrimPrefix(root, "/")

	// Collect all paths under root
	var paths []string
	for p := range m.files {
		if strings.HasPrefix(p, rootPath) {
			paths = append(paths, p)
		}
	}
	sort.Strings(paths)

	for _, p := range paths {
		fullPath := "/" + p
		info := m.Stat(fullPath)
		if info == nil {
			continue
		}
		d := &memDirEntry{fi: info.(*memFileInfo)}
		if err := walkFn(fullPath, d, nil); err != nil {
			return err
		}
	}
	return nil
}

func (m *memFS) Realpath(p string) string {
	// For in-memory FS, realpath is the same as the input path
	return p
}

type memFileInfo struct {
	name    string
	size    int64
	mode    fs.FileMode
	modTime time.Time
}

func (fi *memFileInfo) Name() string       { return fi.name }
func (fi *memFileInfo) Size() int64        { return fi.size }
func (fi *memFileInfo) Mode() fs.FileMode  { return fi.mode }
func (fi *memFileInfo) ModTime() time.Time { return fi.modTime }
func (fi *memFileInfo) IsDir() bool        { return fi.mode&fs.ModeDir != 0 }
func (fi *memFileInfo) Sys() any           { return nil }

type memDirEntry struct {
	fi *memFileInfo
}

func (d *memDirEntry) Name() string               { return d.fi.name }
func (d *memDirEntry) IsDir() bool                { return d.fi.IsDir() }
func (d *memDirEntry) Type() fs.FileMode          { return d.fi.mode & fs.ModeType }
func (d *memDirEntry) Info() (fs.FileInfo, error) { return d.fi, nil }

// FromMapString creates a new vfs.FS from a map of paths to file contents (strings).
func FromMapString(m map[string]string, useCaseSensitiveFileNames bool) vfs.FS {
	return newMemFS(m, useCaseSensitiveFileNames)
}

// FromMapAny creates a new vfs.FS from a map of paths to file contents (any type).
func FromMapAny(m map[string]any, useCaseSensitiveFileNames bool) vfs.FS {
	mStr := make(map[string]string)
	for k, v := range m {
		switch val := v.(type) {
		case string:
			mStr[k] = val
		case []byte:
			mStr[k] = string(val)
		}
	}
	return newMemFS(mStr, useCaseSensitiveFileNames)
}
