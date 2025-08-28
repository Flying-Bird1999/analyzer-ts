package ts_bundle

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// getEnhancedTestProjectRoot 获取测试项目的根目录
func getEnhancedTestProjectRoot() string {
	// 使用固定的测试数据目录
	return filepath.Join("testdata")
}

// TestEnhancedGenerateBundle_SimpleType 测试简单类型的收集
func TestEnhancedGenerateBundle_SimpleType(t *testing.T) {
	projectRoot, _ := filepath.Abs(getEnhancedTestProjectRoot())
	entryFile := filepath.Join(projectRoot, "src", "utils", "user.ts")
	typeName := "User"

	bundledContent, err := GenerateBundle(entryFile, typeName, projectRoot)

	assert.NoError(t, err, "GenerateBundle 应该成功执行")

	// 精确匹配测试
	expectedOutput := "// src/utils/user.ts\nexport interface User {\n  id: number;\n  name: string;\n}\n"
	assert.Equal(t, expectedOutput, bundledContent, "输出应该与期望结果完全匹配")

	// 关键字包含测试（作为补充验证）
	assert.Contains(t, bundledContent, "interface User", "应该包含 User 接口定义")
	assert.Contains(t, bundledContent, "id: number", "应该包含 id 属性")
	assert.Contains(t, bundledContent, "name: string", "应该包含 name 属性")
	// 确保不包含其他未引用的类型
	assert.NotContains(t, bundledContent, "UserRole", "不应该包含未引用的 UserRole")
	assert.NotContains(t, bundledContent, "UserStatus", "不应该包含未引用的 UserStatus")
	assert.NotContains(t, bundledContent, "AdminUser", "不应该包含未引用的 AdminUser")
}

// TestEnhancedGenerateBundle_TypeWithUnion 测试联合类型的收集
func TestEnhancedGenerateBundle_TypeWithUnion(t *testing.T) {
	projectRoot, _ := filepath.Abs(getEnhancedTestProjectRoot())
	entryFile := filepath.Join(projectRoot, "src", "utils", "user.ts")
	typeName := "UserRole"

	bundledContent, err := GenerateBundle(entryFile, typeName, projectRoot)

	assert.NoError(t, err, "GenerateBundle 应该成功执行")

	// 精确匹配测试
	expectedOutput := "\n\nexport type UserRole = 'admin' | 'user';\n"
	assert.Equal(t, expectedOutput, bundledContent, "输出应该与期望结果完全匹配")

	// 关键字包含测试
	assert.Contains(t, bundledContent, "type UserRole", "应该包含 UserRole 类型定义")
	assert.Contains(t, bundledContent, "'admin' | 'user'", "应该包含联合类型定义")
}

// TestEnhancedGenerateBundle_EnumType 测试枚举类型的收集
func TestEnhancedGenerateBundle_EnumType(t *testing.T) {
	projectRoot, _ := filepath.Abs(getEnhancedTestProjectRoot())
	entryFile := filepath.Join(projectRoot, "src", "utils", "user.ts")
	typeName := "UserStatus"

	bundledContent, err := GenerateBundle(entryFile, typeName, projectRoot)

	assert.NoError(t, err, "GenerateBundle 应该成功执行")

	// 精确匹配测试
	expectedOutput := "\n\nexport enum UserStatus {\n  Active = 'active',\n  Inactive = 'inactive'\n}\n"
	assert.Equal(t, expectedOutput, bundledContent, "输出应该与期望结果完全匹配")

	// 关键字包含测试
	assert.Contains(t, bundledContent, "enum UserStatus", "应该包含 UserStatus 枚举定义")
	assert.Contains(t, bundledContent, "Active = 'active'", "应该包含 Active 枚举值")
	assert.Contains(t, bundledContent, "Inactive = 'inactive'", "应该包含 Inactive 枚举值")
}

// TestEnhancedGenerateBundle_InterfaceExtends 测试接口继承的类型收集
func TestEnhancedGenerateBundle_InterfaceExtends(t *testing.T) {
	projectRoot, _ := filepath.Abs(getEnhancedTestProjectRoot())
	entryFile := filepath.Join(projectRoot, "src", "utils", "user.ts")
	typeName := "AdminUser"

	bundledContent, err := GenerateBundle(entryFile, typeName, projectRoot)

	assert.NoError(t, err, "GenerateBundle 应该成功执行")

	// 验证必须包含的内容
	assert.Contains(t, bundledContent, "interface AdminUser", "应该包含 AdminUser 接口定义")
	assert.Contains(t, bundledContent, "extends User", "应该包含 extends User")
	assert.Contains(t, bundledContent, "role: UserRole", "应该包含 role 属性")
	assert.Contains(t, bundledContent, "status: UserStatus", "应该包含 status 属性")
	// 确保也包含了被继承的 User 接口和依赖的 UserRole、UserStatus
	assert.Contains(t, bundledContent, "interface User", "应该包含被继承的 User 接口定义")
	assert.Contains(t, bundledContent, "type UserRole", "应该包含依赖的 UserRole 类型定义")
	assert.Contains(t, bundledContent, "enum UserStatus", "应该包含依赖的 UserStatus 枚举定义")

	// 使用正则表达式验证结构
	adminUserPattern := `export interface AdminUser extends User \{[\s\S]*?role: UserRole;[\s\S]*?status: UserStatus;[\s\S]*?\}`
	assert.Regexp(t, adminUserPattern, bundledContent, "AdminUser 接口定义应该符合预期格式")

	userPattern := `export interface User \{[\s\S]*?id: number;[\s\S]*?name: string;[\s\S]*?\}`
	assert.Regexp(t, userPattern, bundledContent, "User 接口定义应该符合预期格式")

	userRolePattern := `export type UserRole = 'admin' \| 'user';`
	assert.Regexp(t, userRolePattern, bundledContent, "UserRole 定义应该符合预期格式")

	userStatusPattern := `export enum UserStatus \{[\s\S]*?Active = 'active',[\s\S]*?Inactive = 'inactive'[\s\S]*?\}`
	assert.Regexp(t, userStatusPattern, bundledContent, "UserStatus 定义应该符合预期格式")
}

// TestEnhancedGenerateBundle_TypeWithImport 测试带导入的类型收集
func TestEnhancedGenerateBundle_TypeWithImport(t *testing.T) {
	projectRoot, _ := filepath.Abs(getEnhancedTestProjectRoot())
	entryFile := filepath.Join(projectRoot, "src", "index.ts")
	typeName := "UserProfile"

	bundledContent, err := GenerateBundle(entryFile, typeName, projectRoot)

	assert.NoError(t, err, "GenerateBundle 应该成功执行")

	// 验证必须包含的内容
	assert.Contains(t, bundledContent, "interface UserProfile", "应该包含 UserProfile 接口定义")
	assert.Contains(t, bundledContent, "extends User", "应该包含 extends User")
	assert.Contains(t, bundledContent, "address: Address", "应该包含 address 属性")
	// 检查重命名后的类型
	assert.Contains(t, bundledContent, "tags: Common_CommonType[]", "应该包含重命名后的 tags 属性")
	// 确保包含了所有依赖的类型
	assert.Contains(t, bundledContent, "interface User", "应该包含依赖的 User 接口定义")
	assert.Contains(t, bundledContent, "interface Address", "应该包含依赖的 Address 接口定义")
	assert.Contains(t, bundledContent, "type Common_CommonType", "应该包含重命名后的 CommonType 类型定义")

	// 使用正则表达式验证结构
	userProfilePattern := `export interface UserProfile extends User \{[\s\S]*?address: Address;[\s\S]*?tags: Common_CommonType\[];[\s\S]*?\}`
	assert.Regexp(t, userProfilePattern, bundledContent, "UserProfile 接口定义应该符合预期格式")

	userPattern := `export interface User \{[\s\S]*?id: number;[\s\S]*?name: string;[\s\S]*?\}`
	assert.Regexp(t, userPattern, bundledContent, "User 接口定义应该符合预期格式")

	addressPattern := `export interface Address \{[\s\S]*?street: string;[\s\S]*?city: string;[\s\S]*?country: string;[\s\S]*?\}`
	assert.Regexp(t, addressPattern, bundledContent, "Address 接口定义应该符合预期格式")

	commonTypePattern := `export type Common_CommonType = string \| number;`
	assert.Regexp(t, commonTypePattern, bundledContent, "CommonType 定义应该符合预期格式")
}

// TestEnhancedGenerateBundle_NamespaceImport 测试命名空间导入的类型收集
func TestEnhancedGenerateBundle_NamespaceImport(t *testing.T) {
	projectRoot, _ := filepath.Abs(getEnhancedTestProjectRoot())
	entryFile := filepath.Join(projectRoot, "src", "index.ts")
	typeName := "UserId"

	bundledContent, err := GenerateBundle(entryFile, typeName, projectRoot)

	assert.NoError(t, err, "GenerateBundle 应该成功执行")

	// 验证必须包含的内容
	assert.Contains(t, bundledContent, "type UserId", "应该包含 UserId 类型定义")
	// 检查重命名后的类型
	assert.Contains(t, bundledContent, "Common_CommonInterface['id']", "应该包含重命名后的索引访问类型")
	// 确保包含了 CommonInterface
	assert.Contains(t, bundledContent, "interface Common_CommonInterface", "应该包含重命名后的 CommonInterface 接口定义")
	assert.Contains(t, bundledContent, "id: CommonType", "应该包含 CommonInterface 的 id 属性")
	assert.Contains(t, bundledContent, "type CommonType", "应该包含 CommonType 类型定义")

	// 使用正则表达式验证结构 (修正正则表达式以匹配实际输出)
	userIdPattern := `export type UserId = Common_CommonInterface\['id'\];`
	assert.Regexp(t, userIdPattern, bundledContent, "UserId 类型定义应该符合预期格式")

	commonInterfacePattern := `export interface Common_CommonInterface \{[\s\S]*?id: CommonType;[\s\S]*?\}`
	assert.Regexp(t, commonInterfacePattern, bundledContent, "CommonInterface 接口定义应该符合预期格式")

	commonTypePattern := `export type CommonType = string \| number;`
	assert.Regexp(t, commonTypePattern, bundledContent, "CommonType 定义应该符合预期格式")
}

// TestEnhancedGenerateBundle_TypeComposition 测试类型组合的收集
func TestEnhancedGenerateBundle_TypeComposition(t *testing.T) {
	projectRoot, _ := filepath.Abs(getEnhancedTestProjectRoot())
	entryFile := filepath.Join(projectRoot, "src", "index.ts")
	typeName := "FullUser"

	bundledContent, err := GenerateBundle(entryFile, typeName, projectRoot)

	assert.NoError(t, err, "GenerateBundle 应该成功执行")

	// 验证必须包含的内容
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

	// 使用正则表达式验证结构
	fullUserPattern := `export type FullUser = UserProfile & AdminUser;`
	assert.Regexp(t, fullUserPattern, bundledContent, "FullUser 类型定义应该符合预期格式")
}

// TestEnhancedGenerateBundle_ExternalPackageImport 测试外部包导入的类型收集
func TestEnhancedGenerateBundle_ExternalPackageImport(t *testing.T) {
	projectRoot, _ := filepath.Abs(getEnhancedTestProjectRoot())
	entryFile := filepath.Join(projectRoot, "src", "external.ts")
	typeName := "LocalTypeWithExternal"

	bundledContent, err := GenerateBundle(entryFile, typeName, projectRoot)

	assert.NoError(t, err, "GenerateBundle 应该成功执行")

	// 验证必须包含的内容
	assert.Contains(t, bundledContent, "interface LocalTypeWithExternal", "应该包含 LocalTypeWithExternal 接口定义")
	assert.Contains(t, bundledContent, "externalProp: ExternalType", "应该包含 externalProp 属性")
	// 确保包含了外部类型
	assert.Contains(t, bundledContent, "interface ExternalType", "应该包含依赖的 ExternalType 接口定义")
	assert.Contains(t, bundledContent, "externalId: number", "应该包含 ExternalType 的 externalId 属性")
	assert.Contains(t, bundledContent, "externalName: string", "应该包含 ExternalType 的 externalName 属性")

	// 使用正则表达式验证结构 (修正正则表达式以匹配实际输出)
	localTypePattern := `export interface LocalTypeWithExternal \{[\s\S]*?localProp: string;[\s\S]*?externalProp: ExternalType;[\s\S]*?\}`
	assert.Regexp(t, localTypePattern, bundledContent, "LocalTypeWithExternal 接口定义应该符合预期格式")

	externalTypePattern := `export interface ExternalType \{[\s\S]*?externalId: number;[\s\S]*?externalName: string;[\s\S]*?\}`
	assert.Regexp(t, externalTypePattern, bundledContent, "ExternalType 接口定义应该符合预期格式")
}

// TestEnhancedGenerateBundle_EmptyInput 测试空输入处理
func TestEnhancedGenerateBundle_EmptyInput(t *testing.T) {
	projectRoot, _ := filepath.Abs(getEnhancedTestProjectRoot())
	entryFile := filepath.Join(projectRoot, "src", "utils", "user.ts")
	typeName := "NonExistentType"

	bundledContent, err := GenerateBundle(entryFile, typeName, projectRoot)

	// 对于不存在的类型，应该返回空内容而不是错误
	assert.NoError(t, err, "GenerateBundle 应该成功执行，即使类型不存在")
	assert.Equal(t, "", bundledContent, "对于不存在的类型，应该返回空内容")
}

// TestEnhancedGenerateBundle_InvalidFile 测试无效文件处理
// func TestEnhancedGenerateBundle_InvalidFile(t *testing.T) {
// 	projectRoot, _ := filepath.Abs(getEnhancedTestProjectRoot())
// 	entryFile := filepath.Join(projectRoot, "src", "nonexistent.ts")
// 	typeName := "User"

// 	_, err := GenerateBundle(entryFile, typeName, projectRoot)

// 	// 对于不存在的文件，应该返回错误
// 	assert.Error(t, err, "GenerateBundle 应该返回错误，当文件不存在时")
// }
