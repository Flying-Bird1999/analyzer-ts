package tsmorphgo

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Zzzen/typescript-go/use-at-your-own-risk/ast"
	"github.com/stretchr/testify/assert"
)

// complex_navigation_test.go
//
// 这个文件包含了复杂 AST 导航功能的综合测试用例，专注于验证 tsmorphgo 在处理
// 深度嵌套、控制流、类型系统和装饰器等复杂 TypeScript 代码结构时的导航能力。
//
// 主要测试场景：
// 1. 深度嵌套结构导航 - 验证在多层嵌套的对象、类、方法中的节点查找和导航
// 2. 复杂控制流导航 - 测试在 if/else、switch、循环、try/catch 等控制流中的导航
// 3. 复杂类型系统导航 - 验证在泛型、接口继承、类型别名等类型系统中的导航
// 4. 装饰器和元数据导航 - 测试在 Angular/装饰器风格代码中的节点导航
// 5. 项目级边界情况 - 验证大型项目、循环依赖、语法错误等边缘场景
//
// 测试目标：
// - 验证 GetAncestors() 方法在复杂结构中的正确性
// - 验证 GetFirstAncestorByKind() 方法在特定场景下的准确性
// - 测试在极端复杂的 AST 结构中的性能和稳定性
// - 确保在各种边缘情况下系统不会崩溃并返回合理结果

// TestComplexASTNavigation 测试复杂的AST导航功能
func TestComplexASTNavigation(t *testing.T) {
	// 测试用例 1: 深度嵌套的AST结构导航
	t.Run("DeepNestedNavigation", func(t *testing.T) {
		project := createTestProject(map[string]string{"/test_deep_nested.ts": `
			class OuterClass {
				private innerField: {
					nested: {
						deep: {
							value: string;
						};
						items: Array<{
							id: number;
							data: {
								content: string;
								metadata?: {
									tags: string[];
								};
							};
						}>;
					};
				};

				constructor() {
					this.innerField = {
						nested: {
							deep: {
								value: "test"
							},
							items: [{
								id: 1,
								data: {
									content: "hello",
									metadata: {
										tags: ["tag1", "tag2"]
									}
								}
							}]
						}
					};
				}

				processData(): void {
					const result = this.innerField.nested.items[0].data.content;
					console.log(result);
				}
			}
		`})
		sf := project.GetSourceFile("/test_deep_nested.ts")
		assert.NotNil(t, sf)

		// 找到最深层级的标识符 "content"
		var contentNode *Node
		sf.ForEachDescendant(func(node Node) {
			if IsIdentifier(node) && strings.TrimSpace(node.GetText()) == "content" {
				// 确保是方法中的content，而不是类型定义中的
				if parent := node.GetParent(); parent != nil {
					if grandParent := parent.GetParent(); grandParent != nil {
						if strings.Contains(grandParent.GetText(), "this.innerField.nested.items[0].data.content") {
							contentNode = &node
						}
					}
				}
			}
		})

		assert.NotNil(t, contentNode, "未能找到深层嵌套的content节点")

		// 测试复杂的祖先链导航
		ancestors := contentNode.GetAncestors()

		// 验证祖先链包含基本的节点类型
		expectedKinds := []ast.Kind{
			ast.KindPropertyAccessExpression, // .content
			ast.KindPropertyAccessExpression, // .data
			ast.KindPropertyAccessExpression, // .items
			ast.KindVariableDeclaration,      // result = ...
		}

		foundKinds := make(map[ast.Kind]bool)
		for _, ancestor := range ancestors {
			foundKinds[ancestor.Kind] = true
		}

		// 只验证必需的节点类型
		for _, expectedKind := range expectedKinds {
			assert.True(t, foundKinds[expectedKind], "应该找到祖先节点类型: %v", expectedKind)
		}
	})

	// 测试用例 2: 复杂的控制流结构导航
	t.Run("ComplexControlFlowNavigation", func(t *testing.T) {
		project := createTestProject(map[string]string{"/test_control_flow.ts": `
			function processData(items: any[]): any[] {
				const result = [];

				for (let i = 0; i < items.length; i++) {
					const item = items[i];

					if (item && item.type === 'active') {
						switch (item.category) {
							case 'important':
								result.push({
									...item,
									priority: 'high',
									processed: true
								});
								break;
							case 'normal':
								if (item.content && item.content.length > 100) {
									continue;
								}
								result.push(item);
								break;
							default:
								result.push({
									...item,
									priority: 'low'
								});
						}
					} else if (item && item.type === 'archived') {
						try {
							const archived = JSON.parse(item.data);
							if (archived && archived.restore) {
								result.push(archived.restore());
							}
						} catch (error) {
							console.error('Failed to parse archived item:', error);
						}
					}
				}

				return result.filter(Boolean);
			}
		`})
		sf := project.GetSourceFile("/test_control_flow.ts")
		assert.NotNil(t, sf)

		// 找到最深层级的 "priority" 标识符
		var priorityNode *Node
		sf.ForEachDescendant(func(node Node) {
			if IsIdentifier(node) && strings.TrimSpace(node.GetText()) == "priority" {
				if parent := node.GetParent(); parent != nil {
					// 确保是在对象字面量中的priority属性
					if grandParent := parent.GetParent(); grandParent != nil {
						if strings.Contains(grandParent.GetText(), "priority: 'high'") {
							priorityNode = &node
						}
					}
				}
			}
		})

		assert.NotNil(t, priorityNode, "未能找到priority节点")

		// 测试复杂的祖先导航，验证控制流结构
		ancestors := priorityNode.GetAncestors()

		// 验证祖先链包含基本的控制流节点类型
		expectedControlFlowKinds := []ast.Kind{
			ast.KindPropertyAssignment,      // priority: 'high'
			ast.KindObjectLiteralExpression, // { ...item, priority: 'high', ... }
			ast.KindCallExpression,          // result.push(...)
		}

		foundControlFlowKinds := make(map[ast.Kind]bool)
		for _, ancestor := range ancestors {
			foundControlFlowKinds[ancestor.Kind] = true
		}

		// 只验证必需的节点类型
		for _, expectedKind := range expectedControlFlowKinds {
			assert.True(t, foundControlFlowKinds[expectedKind], "应该找到控制流节点类型: %v", expectedKind)
		}

		// 验证能找到特定的祖先类型
		caseStatement, ok := priorityNode.GetFirstAncestorByKind(ast.KindCaseClause)
		assert.True(t, ok, "应该找到CaseClause祖先")
		assert.Contains(t, caseStatement.GetText(), "case 'important'")

		switchStatement, ok := priorityNode.GetFirstAncestorByKind(ast.KindSwitchStatement)
		assert.True(t, ok, "应该找到SwitchStatement祖先")
		assert.Contains(t, switchStatement.GetText(), "switch (item.category)")
	})

	// 测试用例 3: 复杂的泛型和类型系统导航
	t.Run("ComplexTypeSystemNavigation", func(t *testing.T) {
		project := createTestProject(map[string]string{"/test_types.ts": `
			interface BaseRepository<T, K extends keyof T> {
				findById(id: T[K]): Promise<T | null>;
			 findAll(filter: Partial<T>): Promise<T[]>;
			 create(entity: Omit<T, 'id'>): Promise<T>;
			 update(id: T[K], updates: Partial<T>): Promise<T>;
			 delete(id: T[K]): Promise<boolean>;
			}

			interface User {
			 id: number;
			 name: string;
			 email: string;
			 profile: {
				 age: number;
				 preferences: {
					 notifications: boolean;
					 theme: 'light' | 'dark';
				 };
			 };
			}

			class UserRepository implements BaseRepository<User, 'id'> {
			 async findById(id: number): Promise<User | null> {
				 // Implementation
				 return null;
			 }

			 async findAll(filter: Partial<User>): Promise<User[]> {
				 // Implementation
				 return [];
			 }

			 async create(entity: Omit<User, 'id'>): Promise<User> {
				 // Implementation
				 return entity as User;
			 }

			 async update(id: number, updates: Partial<User>): Promise<User> {
				 // Implementation
				 return {} as User;
			 }

			 async delete(id: number): Promise<boolean> {
				 // Implementation
				 return true;
			 }
			}

			type UserService = {
			 repository: BaseRepository<User, 'id'>;
			 cache: CacheService<User>;
			 logger: Logger;
			};

			interface CacheService<T> {
			 get(key: string): Promise<T | null>;
			 set(key: string, value: T, ttl?: number): Promise<void>;
			 invalidate(pattern: string): Promise<number>;
			}
		`})
		sf := project.GetSourceFile("/test_types.ts")
		assert.NotNil(t, sf)

		// 找到复杂类型中的标识符 "notifications"
		var notificationsNode *Node
		sf.ForEachDescendant(func(node Node) {
			if IsIdentifier(node) && strings.TrimSpace(node.GetText()) == "notifications" {
				if parent := node.GetParent(); parent != nil {
					// 确保是在类型定义中的notifications
					if grandParent := parent.GetParent(); grandParent != nil {
						if strings.Contains(grandParent.GetText(), "notifications: boolean") {
							notificationsNode = &node
						}
					}
				}
			}
		})

		assert.NotNil(t, notificationsNode, "未能找到notifications节点")

		// 测试复杂类型系统的祖先导航
		ancestors := notificationsNode.GetAncestors()

		// 验证祖先链包含类型系统相关的节点类型
		expectedTypeKinds := []ast.Kind{
			ast.KindPropertySignature,    // notifications: boolean
			ast.KindTypeLiteral,          // { notifications: boolean, theme: ... }
			ast.KindPropertySignature,    // preferences: { ... }
			ast.KindTypeLiteral,          // { age: number, preferences: ... }
			ast.KindPropertySignature,    // profile: { ... }
			ast.KindInterfaceDeclaration, // interface User
		}

		foundTypeKinds := make(map[ast.Kind]bool)
		for _, ancestor := range ancestors {
			foundTypeKinds[ancestor.Kind] = true
		}

		for _, expectedKind := range expectedTypeKinds {
			assert.True(t, foundTypeKinds[expectedKind], "应该找到类型系统节点类型: %v", expectedKind)
		}

		// 验证能找到特定的类型系统祖先
		userInterface, ok := notificationsNode.GetFirstAncestorByKind(ast.KindInterfaceDeclaration)
		assert.True(t, ok, "应该找到User接口")
		assert.Contains(t, userInterface.GetText(), "interface User")

		// 验证在User接口内部
		shouldFindUserInterface := false
		for _, ancestor := range ancestors {
			if ancestor.Kind == ast.KindInterfaceDeclaration &&
				strings.Contains(ancestor.GetText(), "interface User") {
				shouldFindUserInterface = true
				break
			}
		}
		assert.True(t, shouldFindUserInterface, "应该在祖先链中找到User接口")
	})

	// 测试用例 4: 复杂的装饰器和元数据导航
	t.Run("ComplexDecoratorNavigation", func(t *testing.T) {
		project := createTestProject(map[string]string{"/test_decorators.ts": `
			@Component({
				selector: 'app-user-profile',
				templateUrl: './user-profile.component.html',
				styleUrls: ['./user-profile.component.scss'],
				changeDetection: ChangeDetectionStrategy.OnPush,
				providers: [
					{ provide: UserService, useClass: UserService },
					UserRepository
				]
			})
			@AuthRequired({
				roles: ['admin', 'user-manager'],
				permissions: ['user:read', 'user:write']
			})
			@LogExecution({
				level: 'debug',
				includeParams: true,
				excludeParams: ['password']
			})
			export class UserProfileComponent implements OnInit {
				@Input() userId: number;
				@Output() userUpdated = new EventEmitter<User>();
				@HostBinding('class.active') isActive = false;
				@HostListener('click', ['$event'])
				onClick(event: MouseEvent): void {
					console.log('Component clicked:', event);
				}

				constructor(
					private userService: UserService,
					private repo: UserRepository,
					private logger: Logger
				) {}

				ngOnInit(): void {
					this.userService.findById(this.userId).subscribe(user => {
						this.userUpdated.emit(user);
					});
				}

				@Throttle(300)
				@Validate({ required: true, minLength: 3 })
				updateUserProfile(@Inject('formData') data: Partial<User>): Observable<User> {
					return this.userService.update(this.userId, data).pipe(
						tap(updatedUser => {
							this.logger.info('User updated successfully', updatedUser);
							this.userUpdated.emit(updatedUser);
						})
					);
				}
			}
		`})
		sf := project.GetSourceFile("/test_decorators.ts")
		assert.NotNil(t, sf)

		// 找到方法装饰器中的 "required" 标识符
		var requiredNode *Node
		sf.ForEachDescendant(func(node Node) {
			if IsIdentifier(node) && strings.TrimSpace(node.GetText()) == "required" {
				if parent := node.GetParent(); parent != nil {
					// 确保是在装饰器配置中的required
					if grandParent := parent.GetParent(); grandParent != nil {
						if strings.Contains(grandParent.GetText(), "required: true") {
							requiredNode = &node
						}
					}
				}
			}
		})

		assert.NotNil(t, requiredNode, "未能找到required节点")

		// 测试复杂装饰器结构的祖先导航
		ancestors := requiredNode.GetAncestors()

		// 验证祖先链包含装饰器相关的节点类型
		expectedDecoratorKinds := []ast.Kind{
			ast.KindPropertyAssignment,      // required: true
			ast.KindObjectLiteralExpression, // { required: true, minLength: 3 }
			ast.KindCallExpression,          // @Validate({ ... })
			ast.KindDecorator,               // Validate decorator
			ast.KindMethodDeclaration,       // updateUserProfile method
			ast.KindClassDeclaration,        // UserProfileComponent class
		}

		foundDecoratorKinds := make(map[ast.Kind]bool)
		for _, ancestor := range ancestors {
			foundDecoratorKinds[ancestor.Kind] = true
		}

		for _, expectedKind := range expectedDecoratorKinds {
			assert.True(t, foundDecoratorKinds[expectedKind], "应该找到装饰器节点类型: %v", expectedKind)
		}

		// 验证能找到特定的装饰器祖先
		validateDecorator, ok := requiredNode.GetFirstAncestorByKind(ast.KindDecorator)
		assert.True(t, ok, "应该找到Validate装饰器")
		assert.Contains(t, validateDecorator.GetText(), "@Validate")

		methodDeclaration, ok := requiredNode.GetFirstAncestorByKind(ast.KindMethodDeclaration)
		assert.True(t, ok, "应该找到方法声明")
		assert.Contains(t, methodDeclaration.GetText(), "updateUserProfile")

		classDeclaration, ok := requiredNode.GetFirstAncestorByKind(ast.KindClassDeclaration)
		assert.True(t, ok, "应该找到类声明")
		assert.Contains(t, classDeclaration.GetText(), "class UserProfileComponent")
	})
}

// TestProjectEdgeCases 测试项目层面的边界情况
func TestProjectEdgeCases(t *testing.T) {
	// 测试用例 1: 空项目和无效输入
	t.Run("EmptyProjectAndInvalidInputs", func(t *testing.T) {
		// 测试空项目
		emptyProject := createTestProject(map[string]string{})
		assert.NotNil(t, emptyProject)

		// 测试获取不存在的文件
		nonExistentFile := emptyProject.GetSourceFile("/nonexistent.ts")
		assert.Nil(t, nonExistentFile)

		// 测试创建空文件的项目
		emptyFileProject := createTestProject(map[string]string{"/empty.ts": ""})
		assert.NotNil(t, emptyFileProject)

		emptyFile := emptyFileProject.GetSourceFile("/empty.ts")
		assert.NotNil(t, emptyFile)

		// 验证空文件的基本操作
		var nodeCount int
		emptyFile.ForEachDescendant(func(node Node) {
			nodeCount++
		})
		// 空文件可能有基本的AST节点（如SourceFile），但应该很少
		assert.LessOrEqual(t, nodeCount, 2, "空文件应该只有很少的节点")
	})

	// 测试用例 2: 大型项目和性能
	t.Run("LargeProjectPerformance", func(t *testing.T) {
		// 创建一个包含多个文件的大型项目
		largeSources := make(map[string]string)

		// 创建10个文件，每个文件包含大量内容
		for i := 0; i < 10; i++ {
			content := fmt.Sprintf(`
				// File %d - Large content for testing
				import { Component, Input, Output, EventEmitter } from '@angular/core';
				import { HttpClient } from '@angular/common/http';
				import { Observable } from 'rxjs';
				import { map, tap, catchError } from 'rxjs/operators';

				interface LargeInterface%d {
					id: number;
					name: string;
					data: {
						field1: string;
						field2: number;
						field3: boolean;
						field4: Array<{
							nestedId: number;
							nestedName: string;
						}>;
					};
					metadata: {
						createdAt: Date;
						updatedAt: Date;
						version: number;
						tags: string[];
					};
				}

				class LargeClass%d {
					@Input() data: LargeInterface%d;
					@Output() dataChange = new EventEmitter<LargeInterface%d>();

					constructor(private http: HttpClient) {}

					processData(): Observable<LargeInterface%d[]> {
						return this.http.get<LargeInterface%d[]>('/api/data').pipe(
							map(items => items.map(item => ({
								...item,
								processed: true,
								timestamp: new Date()
							}))),
							tap(items => console.log('Processed', items.length, 'items')),
							catchError(error => {
								console.error('Error processing data:', error);
								throw error;
							})
						);
					}

					validateData(data: LargeInterface%d): boolean {
						return !!(data && data.id && data.name && data.data);
					}

					transformData(data: LargeInterface%d): LargeInterface%d {
						return {
							...data,
							metadata: {
								...data.metadata,
								updatedAt: new Date(),
								version: (data.metadata.version || 0) + 1
							}
						};
					}
				}

				// Utility functions
				function utilityFunction%d(input: string): number {
					return input.length * 2;
				}

				function anotherUtility%d(a: number, b: number): string {
					return (a + b).toString();
				}

				// Constants and configurations
				const CONFIG%d = {
					apiEndpoint: '/api/v%d',
					timeout: 5000,
					retries: 3,
					cache: true
				};

				// Export everything
				export { LargeInterface%d, LargeClass%d, utilityFunction%d, anotherUtility%d, CONFIG%d };
				export default LargeClass%d;
			`, i, i, i, i, i, i, i, i, i, i, i, i, i, i, i, i, i, i, i, i)
			largeSources[fmt.Sprintf("/large_file_%d.ts", i)] = content
		}

		largeProject := createTestProject(largeSources)
		assert.NotNil(t, largeProject)

		// 测试项目级别的操作 - 遍历已知文件
		knownFiles := []string{"/large_file_0.ts", "/large_file_1.ts", "/large_file_2.ts", "/large_file_3.ts", "/large_file_4.ts",
			"/large_file_5.ts", "/large_file_6.ts", "/large_file_7.ts", "/large_file_8.ts", "/large_file_9.ts"}

		// 测试每个文件的基本操作
		for _, filePath := range knownFiles {
			sf := largeProject.GetSourceFile(filePath)
			assert.NotNil(t, sf)
			assert.Equal(t, filePath, sf.GetFilePath())

			// 测试文件的基本导航
			var nodeCount int
			sf.ForEachDescendant(func(node Node) {
				nodeCount++
			})
			assert.Greater(t, nodeCount, 0, "每个文件应该有多个节点")
		}
	})

	// 测试用例 3: 语法错误和边缘语法
	t.Run("SyntaxErrorsAndEdgeSyntax", func(t *testing.T) {
		// 测试包含各种边缘语法情况的项目
		edgeCases := map[string]string{
			"/incomplete_syntax.ts": `
				const incomplete =
				function missingBrace() {
					console.log("missing closing brace")
			`,
			"/deeply_nested.ts": `
				const deep = {
					level1: {
						level2: {
							level3: {
								level4: {
									level5: {
										value: "deeply nested"
									}
								}
							}
						}
					}
				}
			`,
			"/large_array.ts": `
				const largeArray = [
					%s
				];
			`,
			"/complex_types.ts": `
				type Complex<T extends { id: number }, K extends keyof T> = {
					[P in K]: T[P] extends Array<infer U> ? U : T[P];
				} & {
					_meta: {
						originalType: T;
						selectedKeys: K[];
					};
				};

				const complexVar: Complex<{ id: number; name: string; items: string[]; }, 'id' | 'name'> = {
					id: 1,
					name: 'test',
					_meta: {
						originalType: { id: 0, name: '', items: [] },
						selectedKeys: ['id', 'name']
					}
				};
			`,
			"/unicode_and_special.ts": `
				const unicode = "Hello 世界 🌍";
				const specialChars = "Special: @#$%^&*()_+-=[]{}|;':\",./<>?";
				const templateLiteral = "Template with " + unicode + " and " + specialChars;

				interface UnicodeInterface {
					"中文属性": string;
					"property-with-dashes": number;
					"property@with@symbols": boolean;
				}
			`,
		}

		// 为large_array.ts生成内容
		var items []string
		for i := 0; i < 100; i++ {
			items = append(items, fmt.Sprintf(`{ id: %d, name: "item%d", value: %d }`, i, i, i))
		}
		edgeCases["/large_array.ts"] = fmt.Sprintf(edgeCases["/large_array.ts"], strings.Join(items, ",\n\t\t"))

		edgeProject := createTestProject(edgeCases)
		assert.NotNil(t, edgeProject)

		// 测试边缘情况文件的基本访问
		for filePath := range edgeCases {
			sf := edgeProject.GetSourceFile(filePath)
			assert.NotNil(t, sf, fmt.Sprintf("应该能获取文件: %s", filePath))

			// 验证文件内容非空（检查是否有节点）
			var hasNodes bool
			sf.ForEachDescendant(func(node Node) {
				hasNodes = true
			})
			assert.True(t, hasNodes, fmt.Sprintf("文件 %s 应该有节点", filePath))

			// 测试基本的节点遍历（不应该崩溃）
			var traversalCount int
			sf.ForEachDescendant(func(node Node) {
				traversalCount++
				// 验证节点的基本属性访问
				_ = node.Kind
				_ = node.GetText()
				_ = node.GetParent()
			})

			// 即使有语法错误，也应该能遍历到一些节点
			assert.Greater(t, traversalCount, 0, fmt.Sprintf("文件 %s 应该能遍历到节点", filePath))
		}
	})

	// 测试用例 4: 循环依赖和复杂导入
	t.Run("CircularDependenciesAndComplexImports", func(t *testing.T) {
		// 创建包含循环依赖的项目
		circularSources := map[string]string{
			"/file_a.ts": `
				import { BClass } from './file_b';
				import { CClass } from './file_c';

				export class AClass {
					constructor(public b: BClass, public c: CClass) {}
					methodA(): string {
						return "A -> " + this.b.methodB() + " -> " + this.c.methodC();
					}
				}
			`,
			"/file_b.ts": `
				import { AClass } from './file_a';
				import { CClass } from './file_c';

				export class BClass {
					constructor(public a: AClass, public c: CClass) {}
					methodB(): string {
						return "B -> " + (this.a ? this.a.methodA() : "no A") + " -> " + this.c.methodC();
					}
				}
			`,
			"/file_c.ts": `
				import { AClass } from './file_a';
				import { BClass } from './file_b';

				export class CClass {
					constructor(public a?: AClass, public b?: BClass) {}
					methodC(): string {
						return "C -> " + (this.a ? "has A" : "no A") + " -> " + (this.b ? "has B" : "no B");
					}
				}
			`,
			"/main.ts": `
				import { AClass } from './file_a';
				import { BClass } from './file_b';
				import { CClass } from './file_c';

				const a = new AClass(null as any, new CClass());
				const b = new BClass(null as any, new CClass());
				const c = new CClass();

				console.log(a.methodA());
				console.log(b.methodB());
				console.log(c.methodC());
			`,
		}

		circularProject := createTestProject(circularSources)
		assert.NotNil(t, circularProject)

		// 验证所有文件都能正确加载
		mainFile := circularProject.GetSourceFile("/main.ts")
		assert.NotNil(t, mainFile)

		fileA := circularProject.GetSourceFile("/file_a.ts")
		assert.NotNil(t, fileA)

		fileB := circularProject.GetSourceFile("/file_b.ts")
		assert.NotNil(t, fileB)

		fileC := circularProject.GetSourceFile("/file_c.ts")
		assert.NotNil(t, fileC)

		// 测试FindReferences在循环依赖中的表现
		var classANode *Node
		fileA.ForEachDescendant(func(node Node) {
			if IsIdentifier(node) && strings.TrimSpace(node.GetText()) == "AClass" {
				if parent := node.GetParent(); parent != nil && IsClassDeclaration(*parent) {
					classANode = &node
				}
			}
		})

		if classANode != nil {
			references, err := FindReferences(*classANode)
			assert.NoError(t, err)
			// 在循环依赖中应该能找到多个引用
			assert.GreaterOrEqual(t, len(references), 1, "在循环依赖中应该找到AClass的引用")
		}
	})

	// 测试用例 5: 内存和资源限制
	t.Run("MemoryAndResourceLimits", func(t *testing.T) {
		// 测试创建大量小文件
		manyFiles := make(map[string]string)
		for i := 0; i < 50; i++ {
			manyFiles[fmt.Sprintf("/small_file_%d.ts", i)] = fmt.Sprintf(`
				// Small file %d
				const constant%d = %d;
				export function smallFunction%d(): number {
					return constant%d * 2;
				}
				export default smallFunction%d;
			`, i, i, i, i, i, i)
		}

		manyFilesProject := createTestProject(manyFiles)
		assert.NotNil(t, manyFilesProject)

		// 验证所有文件都能正确加载和访问 - 遍历已知文件
		knownFiles := make([]string, 50)
		for i := 0; i < 50; i++ {
			knownFiles[i] = fmt.Sprintf("/small_file_%d.ts", i)
		}

		// 验证每个文件的功能性
		for i, filePath := range knownFiles {
			sf := manyFilesProject.GetSourceFile(filePath)
			assert.NotNil(t, sf, fmt.Sprintf("应该能获取文件: %s", filePath))
			assert.NotNil(t, sf)
			assert.Contains(t, sf.GetFilePath(), fmt.Sprintf("small_file_%d.ts", i))

			// 验证能找到预期的内容
			expectedConstant := fmt.Sprintf("constant%d", i)
			var foundConstant bool
			sf.ForEachDescendant(func(node Node) {
				if IsIdentifier(node) && strings.TrimSpace(node.GetText()) == expectedConstant {
					foundConstant = true
				}
			})
			assert.True(t, foundConstant, fmt.Sprintf("应该在文件 %d 中找到常量 %s", i, expectedConstant))
		}
	})
}
