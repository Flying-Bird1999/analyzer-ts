package scanProject

import (
	"encoding/json"
	"fmt"
	"main/bundle/utils"
	"os"
)

type PackageJsonInfo struct {
	Name    string
	Version string
	NpmList map[string]NpmItem
}

func GetPackageJson(packageJsonPath string) (*PackageJsonInfo, error) {
	// 检查文件是否存在
	if _, err := os.Stat(packageJsonPath); os.IsNotExist(err) {
		fmt.Printf("package.json 文件不存在: %s\n", packageJsonPath)
		return nil, err
	}

	// 读取 package.json 文件内容
	data, err := utils.ReadFileContent(packageJsonPath)
	if err != nil {
		fmt.Printf("读取 package.json 文件失败: %s\n", err)
		return nil, err
	}

	// 定义结构体解析 package.json
	var packageJson struct {
		Name             string            `json:"name"`
		Version          string            `json:"version"`
		Dependencies     map[string]string `json:"dependencies"`
		DevDependencies  map[string]string `json:"devDependencies"`
		PeerDependencies map[string]string `json:"peerDependencies"`
	}

	// 解析 JSON 数据
	if err := json.Unmarshal([]byte(data), &packageJson); err != nil {
		fmt.Printf("解析 package.json 文件失败: %s\n", err)
		return nil, err
	}

	info := &PackageJsonInfo{
		Name:    packageJson.Name,
		Version: packageJson.Version,
		NpmList: make(map[string]NpmItem),
	}

	// 将 npm 包添加到 NpmList
	for name, version := range packageJson.Dependencies {
		info.NpmList[name] = NpmItem{Name: name, Version: version, Type: "dependencies"}
	}
	for name, version := range packageJson.DevDependencies {
		info.NpmList[name] = NpmItem{Name: name, Version: version, Type: "devDependencies"}
	}
	for name, version := range packageJson.PeerDependencies {
		info.NpmList[name] = NpmItem{Name: name, Version: version, Type: "peerDependencies"}
	}

	return info, nil
}
