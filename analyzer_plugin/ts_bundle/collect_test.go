package ts_bundle

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// getTestProjectRoot 获取测试项目的根目录
func getTestProjectRoot() string {
	// 使用固定的测试数据目录
	return filepath.Join("testdata")
}

// TestGenerateBundle_SimpleType 测试简单类型的收集
func TestGenerateBundle_SimpleType(t *testing.T) {
	projectRoot, _ := filepath.Abs(getTestProjectRoot())
	entryFile := filepath.Join(projectRoot, "src", "utils", "user.ts")
	typeName := "User"
	
	bundledContent, err := GenerateBundle(entryFile, typeName, projectRoot)
	
	assert.NoError(t, err, "GenerateBundle 应该成功执行")
	assert.Contains(t, bundledContent, "interface User", "应该包含 User 接口定义")
	assert.Contains(t, bundledContent, "id: number", "应该包含 id 属性")
	assert.Contains(t, bundledContent, "name: string", "应该包含 name 属性")
	// 确保不包含其他未引用的类型
	assert.NotContains(t, bundledContent, "UserRole", "不应该包含未引用的 UserRole")
}

// TestGenerateBundle_TypeWithUnion 测试联合类型的收集
func TestGenerateBundle_TypeWithUnion(t *testing.T) {
	projectRoot, _ := filepath.Abs(getTestProjectRoot())
	entryFile := filepath.Join(projectRoot, "src", "utils", "user.ts")
	typeName := "UserRole"
	
	bundledContent, err := GenerateBundle(entryFile, typeName, projectRoot)
	
	assert.NoError(t, err, "GenerateBundle 应该成功执行")
	assert.Contains(t, bundledContent, "type UserRole", "应该包含 UserRole 类型定义")
	assert.Contains(t, bundledContent, "'admin' | 'user'", "应该包含联合类型定义")
}

// TestGenerateBundle_EnumType 测试枚举类型的收集
func TestGenerateBundle_EnumType(t *testing.T) {
	projectRoot, _ := filepath.Abs(getTestProjectRoot())
	entryFile := filepath.Join(projectRoot, "src", "utils", "user.ts")
	typeName := "UserStatus"
	
	bundledContent, err := GenerateBundle(entryFile, typeName, projectRoot)
	
	assert.NoError(t, err, "GenerateBundle 应该成功执行")
	assert.Contains(t, bundledContent, "enum UserStatus", "应该包含 UserStatus 枚举定义")
	assert.Contains(t, bundledContent, "Active = 'active'", "应该包含 Active 枚举值")
	assert.Contains(t, bundledContent, "Inactive = 'inactive'", "应该包含 Inactive 枚举值")
}

// TestGenerateBundle_InterfaceExtends 测试接口继承的类型收集
func TestGenerateBundle_InterfaceExtends(t *testing.T) {
	projectRoot, _ := filepath.Abs(getTestProjectRoot())
	entryFile := filepath.Join(projectRoot, "src", "utils", "user.ts")
	typeName := "AdminUser"
	
	bundledContent, err := GenerateBundle(entryFile, typeName, projectRoot)
	
	assert.NoError(t, err, "GenerateBundle 应该成功执行")
	assert.Contains(t, bundledContent, "interface AdminUser", "应该包含 AdminUser 接口定义")
	assert.Contains(t, bundledContent, "extends User", "应该包含 extends User")
	assert.Contains(t, bundledContent, "role: UserRole", "应该包含 role 属性")
	assert.Contains(t, bundledContent, "status: UserStatus", "应该包含 status 属性")
	// 确保也包含了被继承的 User 接口和依赖的 UserRole、UserStatus
	assert.Contains(t, bundledContent, "interface User", "应该包含被继承的 User 接口定义")
	assert.Contains(t, bundledContent, "type UserRole", "应该包含依赖的 UserRole 类型定义")
	assert.Contains(t, bundledContent, "enum UserStatus", "应该包含依赖的 UserStatus 枚举定义")
}

// TestGenerateBundle_TypeWithImport 测试带导入的类型收集
func TestGenerateBundle_TypeWithImport(t *testing.T) {
	projectRoot, _ := filepath.Abs(getTestProjectRoot())
	entryFile := filepath.Join(projectRoot, "src", "index.ts")
	typeName := "UserProfile"
	
	bundledContent, err := GenerateBundle(entryFile, typeName, projectRoot)
	
	assert.NoError(t, err, "GenerateBundle 应该成功执行")
	assert.Contains(t, bundledContent, "interface UserProfile", "应该包含 UserProfile 接口定义")
	assert.Contains(t, bundledContent, "extends User", "应该包含 extends User")
	assert.Contains(t, bundledContent, "address: Address", "应该包含 address 属性")
	// 检查重命名后的类型
	assert.Contains(t, bundledContent, "tags: Common_CommonType[]", "应该包含重命名后的 tags 属性")
	// 确保包含了所有依赖的类型
	assert.Contains(t, bundledContent, "interface User", "应该包含依赖的 User 接口定义")
	assert.Contains(t, bundledContent, "interface Address", "应该包含依赖的 Address 接口定义")
	assert.Contains(t, bundledContent, "type Common_CommonType", "应该包含重命名后的 CommonType 类型定义")
}

// TestGenerateBundle_NamespaceImport 测试命名空间导入的类型收集
func TestGenerateBundle_NamespaceImport(t *testing.T) {
	projectRoot, _ := filepath.Abs(getTestProjectRoot())
	entryFile := filepath.Join(projectRoot, "src", "index.ts")
	typeName := "UserId"
	
	bundledContent, err := GenerateBundle(entryFile, typeName, projectRoot)
	
	assert.NoError(t, err, "GenerateBundle 应该成功执行")
	assert.Contains(t, bundledContent, "type UserId", "应该包含 UserId 类型定义")
	// 检查重命名后的类型
	assert.Contains(t, bundledContent, "Common_CommonInterface['id']", "应该包含重命名后的索引访问类型")
	// 确保包含了 CommonInterface
	assert.Contains(t, bundledContent, "interface Common_CommonInterface", "应该包含重命名后的 CommonInterface 接口定义")
	assert.Contains(t, bundledContent, "id: CommonType", "应该包含 CommonInterface 的 id 属性")
	assert.Contains(t, bundledContent, "type CommonType", "应该包含 CommonType 类型定义")
}

// TestGenerateBundle_TypeComposition 测试类型组合的收集
func TestGenerateBundle_TypeComposition(t *testing.T) {
	projectRoot, _ := filepath.Abs(getTestProjectRoot())
	entryFile := filepath.Join(projectRoot, "src", "index.ts")
	typeName := "FullUser"
	
	bundledContent, err := GenerateBundle(entryFile, typeName, projectRoot)
	
	assert.NoError(t, err, "GenerateBundle 应该成功执行")
	assert.Contains(t, bundledContent, "type FullUser", "应该包含 FullUser 类型定义")
	// 确保包含了所有组合的类型
	assert.Contains(t, bundledContent, "interface UserProfile", "应该包含依赖的 UserProfile 接口定义")
	assert.Contains(t, bundledContent, "interface AdminUser", "应该包含依赖的 AdminUser 接口定义")
	assert.Contains(t, bundledContent, "interface User", "应该包含依赖的 User 接口定义")
	assert.Contains(t, bundledContent, "type UserRole", "应该包含依赖的 UserRole 类型定义")
	assert.Contains(t, bundledContent, "enum UserStatus", "应该包含依赖的 UserStatus 枚举定义")
	assert.Contains(t, bundledContent, "interface Address", "应该包含依赖的 Address 接口定义")
	// 检查重命名后的类型
	assert.Contains(t, bundledContent, "type Common_CommonType", "应该包含重命名后的 CommonType 类型定义")
}

// TestGenerateBundle_ExternalPackageImport 测试外部包导入的类型收集
func TestGenerateBundle_ExternalPackageImport(t *testing.T) {
	projectRoot, _ := filepath.Abs(getTestProjectRoot())
	entryFile := filepath.Join(projectRoot, "src", "external.ts")
	typeName := "LocalTypeWithExternal"
	
	bundledContent, err := GenerateBundle(entryFile, typeName, projectRoot)
	
	assert.NoError(t, err, "GenerateBundle 应该成功执行")
	assert.Contains(t, bundledContent, "interface LocalTypeWithExternal", "应该包含 LocalTypeWithExternal 接口定义")
	assert.Contains(t, bundledContent, "externalProp: ExternalType", "应该包含 externalProp 属性")
	// 确保包含了外部类型
	assert.Contains(t, bundledContent, "interface ExternalType", "应该包含依赖的 ExternalType 接口定义")
	assert.Contains(t, bundledContent, "externalId: number", "应该包含 ExternalType 的 externalId 属性")
	assert.Contains(t, bundledContent, "externalName: string", "应该包含 ExternalType 的 externalName 属性")
}

// TestGenerateBundle_ComplexTypes 测试复杂类型的收集
func TestGenerateBundle_ComplexTypes(t *testing.T) {
	projectRoot, _ := filepath.Abs(getTestProjectRoot())
	entryFile := filepath.Join(projectRoot, "src", "complex.ts")
	typeName := "UserFields"
	
	bundledContent, err := GenerateBundle(entryFile, typeName, projectRoot)
	
	assert.NoError(t, err, "GenerateBundle 应该成功执行")
	assert.Contains(t, bundledContent, "type UserFields", "应该包含 UserFields 类型定义")
	// 检查是否包含了必要的依赖类型 (注意类型可能已被重命名)
	assert.Contains(t, bundledContent, "interface UserProfile", "应该包含依赖的 UserProfile 接口定义")
	assert.Contains(t, bundledContent, "interface AdminUser", "应该包含依赖的 AdminUser 接口定义")
	assert.Contains(t, bundledContent, "interface User", "应该包含依赖的 User 接口定义")
	// 检查重命名后的类型
	assert.Contains(t, bundledContent, "type Common_CommonType", "应该包含重命名后的 CommonType 类型定义")
}

// TestGenerateBundle_TypeWithOmit 测试使用 Omit 的类型收集
func TestGenerateBundle_TypeWithOmit(t *testing.T) {
	projectRoot, _ := filepath.Abs(getTestProjectRoot())
	entryFile := filepath.Join(projectRoot, "src", "complex.ts")
	typeName := "UserWithoutAddress"
	
	bundledContent, err := GenerateBundle(entryFile, typeName, projectRoot)
	
	assert.NoError(t, err, "GenerateBundle 应该成功执行")
	// 由于 Omit 是工具类型，实际输出应该是展开后的类型
	// 我们检查是否包含了必要的依赖类型 (注意类型可能已被重命名)
	assert.Contains(t, bundledContent, "interface UserProfile", "应该包含依赖的 UserProfile 接口定义")
	assert.Contains(t, bundledContent, "interface AdminUser", "应该包含依赖的 AdminUser 接口定义")
	assert.Contains(t, bundledContent, "interface User", "应该包含依赖的 User 接口定义")
	assert.Contains(t, bundledContent, "interface Address", "应该包含依赖的 Address 接口定义")
	// 检查重命名后的类型
	assert.Contains(t, bundledContent, "type Common_CommonType", "应该包含重命名后的 CommonType 类型定义")
}

// TestGenerateBundle_TypeWithPick 测试使用 Pick 的类型收集
func TestGenerateBundle_TypeWithPick(t *testing.T) {
	projectRoot, _ := filepath.Abs(getTestProjectRoot())
	entryFile := filepath.Join(projectRoot, "src", "complex.ts")
	typeName := "UserBasicInfo"
	
	bundledContent, err := GenerateBundle(entryFile, typeName, projectRoot)
	
	assert.NoError(t, err, "GenerateBundle 应该成功执行")
	// 由于 Pick 是工具类型，实际输出应该是展开后的类型
	// 我们检查是否包含了必要的依赖类型 (注意类型可能已被重命名)
	assert.Contains(t, bundledContent, "interface UserProfile", "应该包含依赖的 UserProfile 接口定义")
	assert.Contains(t, bundledContent, "interface AdminUser", "应该包含依赖的 AdminUser 接口定义")
	assert.Contains(t, bundledContent, "interface User", "应该包含依赖的 User 接口定义")
}

// TestGenerateBundle_NamespaceTypeAccess 测试命名空间类型访问的收集
func TestGenerateBundle_NamespaceTypeAccess(t *testing.T) {
	projectRoot, _ := filepath.Abs(getTestProjectRoot())
	entryFile := filepath.Join(projectRoot, "src", "complex.ts")
	typeName := "UserTypeCheck"
	
	bundledContent, err := GenerateBundle(entryFile, typeName, projectRoot)
	
	assert.NoError(t, err, "GenerateBundle 应该成功执行")
	assert.Contains(t, bundledContent, "interface UserTypeCheck", "应该包含 UserTypeCheck 接口定义")
	// 检查重命名后的类型引用
	assert.Contains(t, bundledContent, "userId: UserUtils_User['id']", "应该包含重命名后的 userId 属性")
	assert.Contains(t, bundledContent, "userRole: UserUtils_UserRole", "应该包含重命名后的 userRole 属性")
	// 确保包含了依赖的类型 (注意类型可能已被重命名)
	assert.Contains(t, bundledContent, "interface UserUtils_User", "应该包含重命名后的 User 接口定义")
	assert.Contains(t, bundledContent, "type UserUtils_UserRole", "应该包含重命名后的 UserRole 类型定义")
}

// TestGenerateBundle_IndexedAccessType 测试索引访问类型的收集
func TestGenerateBundle_IndexedAccessType(t *testing.T) {
	projectRoot, _ := filepath.Abs(getTestProjectRoot())
	entryFile := filepath.Join(projectRoot, "src", "complex.ts")
	typeName := "UserName"
	
	bundledContent, err := GenerateBundle(entryFile, typeName, projectRoot)
	
	assert.NoError(t, err, "GenerateBundle 应该成功执行")
	// 由于是索引访问类型，实际输出应该是 string 类型
	// 我们检查是否包含了必要的依赖类型 (注意类型可能已被重命名)
	assert.Contains(t, bundledContent, "interface UserProfile", "应该包含依赖的 UserProfile 接口定义")
	assert.Contains(t, bundledContent, "interface AdminUser", "应该包含依赖的 AdminUser 接口定义")
	assert.Contains(t, bundledContent, "interface User", "应该包含依赖的 User 接口定义")
}

// TestGenerateBundle_ImportWithAlias 测试带别名的导入
func TestGenerateBundle_ImportWithAlias(t *testing.T) {
	projectRoot, _ := filepath.Abs(getTestProjectRoot())
	entryFile := filepath.Join(projectRoot, "src", "advanced.ts")
	typeName := "AdvancedUser"
	
	bundledContent, err := GenerateBundle(entryFile, typeName, projectRoot)
	
	assert.NoError(t, err, "GenerateBundle 应该成功执行")
	assert.Contains(t, bundledContent, "interface AdvancedUser", "应该包含 AdvancedUser 接口定义")
	// 检查是否正确处理了别名导入
	assert.Contains(t, bundledContent, "aliasId: number", "应该包含 aliasId 属性")
	assert.Contains(t, bundledContent, "aliasName: string", "应该包含 aliasName 属性")
	assert.Contains(t, bundledContent, "'admin' | 'user'", "应该包含 AliasRole 的联合类型定义")
}

// TestGenerateBundle_DefaultImportWithAlias 测试带别名的默认导入
func TestGenerateBundle_DefaultImportWithAlias(t *testing.T) {
	projectRoot, _ := filepath.Abs(getTestProjectRoot())
	entryFile := filepath.Join(projectRoot, "src", "advanced.ts")
	typeName := "AdvancedDefaultUser2"
	
	_, err := GenerateBundle(entryFile, typeName, projectRoot)
	
	assert.NoError(t, err, "GenerateBundle 应该成功执行")
	// 注意：对于默认导入，我们可能无法直接获取到原始类型的属性，因为默认导入的是整个模块
	// 这个测试主要是验证不会出错
}

// TestGenerateBundle_ImportType 测试 import type 语法
func TestGenerateBundle_ImportType(t *testing.T) {
	projectRoot, _ := filepath.Abs(getTestProjectRoot())
	entryFile := filepath.Join(projectRoot, "src", "advanced.ts")
	typeName := "AdvancedRole"
	
	bundledContent, err := GenerateBundle(entryFile, typeName, projectRoot)
	
	assert.NoError(t, err, "GenerateBundle 应该成功执行")
	assert.Contains(t, bundledContent, "type AdvancedRole", "应该包含 AdvancedRole 类型定义")
	// 检查是否正确处理了 import type
	assert.Contains(t, bundledContent, "'admin' | 'user'", "应该包含 AliasRole 的联合类型定义")
}

// TestGenerateBundle_ReExportWithAlias 测试带别名的重新导出
func TestGenerateBundle_ReExportWithAlias(t *testing.T) {
	projectRoot, _ := filepath.Abs(getTestProjectRoot())
	entryFile := filepath.Join(projectRoot, "src", "advanced.ts")
	typeName := "RenamedAliasUser"
	
	_, err := GenerateBundle(entryFile, typeName, projectRoot)
	
	assert.NoError(t, err, "GenerateBundle 应该成功执行")
	// 注意：重新导出的类型在打包结果中可能不会直接显示，这个测试主要是验证不会出错
}

// TestGenerateBundle_PathAliasImport 测试路径别名导入
func TestGenerateBundle_PathAliasImport(t *testing.T) {
	projectRoot, _ := filepath.Abs(getTestProjectRoot())
	entryFile := filepath.Join(projectRoot, "src", "path-alias.ts")
	typeName := "PathAliasUser"
	
	bundledContent, err := GenerateBundle(entryFile, typeName, projectRoot)
	
	assert.NoError(t, err, "GenerateBundle 应该成功执行")
	assert.Contains(t, bundledContent, "interface PathAliasUser", "应该包含 PathAliasUser 接口定义")
	// 检查是否正确处理了路径别名导入
	assert.Contains(t, bundledContent, "aliasId: number", "应该包含 aliasId 属性")
	assert.Contains(t, bundledContent, "aliasName: string", "应该包含 aliasName 属性")
	assert.Contains(t, bundledContent, "'admin' | 'user'", "应该包含 AliasRole 的联合类型定义")
}

// TestGenerateBundle_CircularDependency 测试循环依赖导入
func TestGenerateBundle_CircularDependency(t *testing.T) {
	projectRoot, _ := filepath.Abs(getTestProjectRoot())
	entryFile := filepath.Join(projectRoot, "src", "circular.ts")
	typeName := "CircularType"
	
	bundledContent, err := GenerateBundle(entryFile, typeName, projectRoot)
	
	assert.NoError(t, err, "GenerateBundle 应该成功执行")
	assert.Contains(t, bundledContent, "interface CircularType", "应该包含 CircularType 接口定义")
	// 检查是否正确处理了循环依赖
	assert.Contains(t, bundledContent, "interface CircularAType", "应该包含 CircularAType 接口定义")
	assert.Contains(t, bundledContent, "aName: string", "应该包含 aName 属性")
}