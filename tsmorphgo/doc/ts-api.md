## 背景

在[《对外开放组件库文档技术方案》](https://shopline.yuque.com/rp76um/front-end/ud00pl015no5k59s#LRovE)的文档中有提到了关于组件 API 自生成的[方案](收集组件元数据%20ComponentMeta)，但是其中的解析逻辑尚不清晰，没有一个完整的规范，导致在 MVP 版本中存在一些[体验问题](https://shopline.yuque.com/rp76um/front-end/ve3b9vpuq9dhq5qa)，例如：

+ API 的继承链过深，收集了所有的可用字段，即所有 `extends` 下来的属性都会显示；
+ API 字段的类型显示冗余，显示了不必要的类型参数，如 `Pick<T, "">` 或 `Omit<T, "">`；

## 应用场景

需要对外提供 API 使用文档的基于 TypeScript 的任意库。

## API 分类

`<font style="color:#DF2A3F;">`API 分为两大类 `</font><font style="color:#DF2A3F;">interface</font>``<font style="color:#DF2A3F;">` 和 `</font><font style="color:#DF2A3F;">typeAlias</font>`。

### Interface API

```typescript
interface Interface extends ParentInterface {
  user: User;
  age: number;
}
```

### TypeAlias API

```typescript
type Type = TypeNode;
```

## API 字段收集

上面提到类型节点，指的是等号右侧的内容，即 `type Type = TYPE_NODE;`中的 `TYPE_NODE`，由于类型节点的分类太多，有 2 种方案选择：

+ `<font style="color:#DF2A3F;">`方案一：自上而下 `</font>`，从顶部做 AST 递归类型解析，需要纯手工做类型判断，然后根据深度收集 API 字段（`<font style="color:#DF2A3F;">`复杂度高且有可能出现覆盖不全的情况 `</font>`）。
+ `<font style="color:#DF2A3F;">`方案二：自下而上 `</font>`，利用 TypeChecker 自带的解析能力先获取所有字段，然后过滤距离声明节点为对应深度以内的字段。

:::warning

+ interface 默认深度取 1；
+ typeAlias 默认深度取 2；

:::

### 怎么定义一层深度？

穿越 `ts.InterfaceDeclaration` 或 `ts.TypeAliasDeclaration` 被认为是一层深度，以 `ButtonProps` 为例子：

```typescript
// 深度 1
export type ButtonProps = { test: string; } & 
  Partial<AnchorButtonProps & NativeButtonProps>;

// 深度 2
export type AnchorButtonProps = {
  /**
   * 链接地址Ï
   *
   * @internal
   */
  href: string;
  /**
   * 链接打开方式
   *
   * @internal
   */
  target?: string;
  /**
   * 鼠标点击事件处理函数
   */
  onClick?: React.MouseEventHandler<HTMLElement>;
} & BaseButtonProps;

// 深度 2
export type NativeButtonProps = {
  /**
   * HTML 类型
   */
  htmlType?: ButtonHTMLType;
  /**
   * 鼠标点击事件处理函数
   */
  onClick?: React.MouseEventHandler<HTMLElement>;
} & BaseButtonProps;

// 深度 3
export interface BaseButtonProps {
  type?: ButtonType | "danger";
}
```

### 收集所有 API 字段

先利用 `TypeChecker.getPropertiesOfType` 收集 `Type`（不管是 `interface` 还是 `typeAlias`） 的所有属性，即符号集合 `Symbol[]`，可能会很多。

例如下方的 `ButtonProps` 有 275 多个，`<font style="color:#DF2A3F;">`唯一好处是：够全 `</font>`。

![ButtonProps](https://cdn.nlark.com/yuque/0/2025/png/1605148/1753697076520-4bb5a81a-9dea-48d8-a6d8-542fa93d8e9d.png)

### 过滤有效 API 字段

得到了所有的字段，该如何过滤我们需要显示的字段？

根据字段符号找到对应的 AST 节点，递归向上寻找目标声明节点（`interface` 或 `typeAlias`），`<font style="color:#DF2A3F;">`若递归次数在限定的深度内（目前设置为 2），则认为是有效的字段 `</font>`，若不是，则抛弃此字段。

### 为什么推荐方案二？

因为 TypeScript 的类型系统复杂，很容易出现字段遗漏的情况，与其手动收集，还不如依赖 TypeChecker 自带的字段收集能力。

### TypeAlias 的兜底处理

当一个 `typeAlias` API 找不到字段的时候（例如 `type Type = number | CountConfig;`），则直接调用 `LanguageService.getQuickInfoAtPosition` 获取该类型的提示文案做显示。

:::warning
名词解释：

+ LanguageService 是 typescript 内置的代码分析引擎，提供智能补全、**类型提示**和跳转到定义等功能，其底层也是基于 typescript AST 分析。
+ QuickInfo 对应 LanguageService 中的**类型提示**，我们可以有效利用 ts 自带的分析能力。

:::

`<font style="color:#DF2A3F;">`同时，将会根据 `</font><font style="color:#DF2A3F;">QuickInfo.displayParts[].kind</font>``<font style="color:#DF2A3F;">` 类型提示（`</font><font style="color:#DF2A3F;">interaceName</font>``<font style="color:#DF2A3F;">` 或 `</font><font style="color:#DF2A3F;">typeName</font>``<font style="color:#DF2A3F;">`）递归衍生新的 API。`</font>`

![](https://cdn.nlark.com/yuque/0/2025/png/1605148/1753698831271-b55cb5f3-ed65-42d1-8e87-0fa145ed5a37.png)

### QuickInfo 因字段过多显示省略号？

当一个类型显示的字段过多的时候，QuickInfo 返回的文案将会出现省略号，这样会影响我们的文案显示以及 API 收集。

通过设置 `[noErrorTruncation](https://www.typescriptlang.org/tsconfig/#noErrorTruncation): true` 让 TS 编译保留完整的提示文案。

## API 字段的类型如何显示？

为了尽量还原开发者在 vscode 中所见的类型提示，我们利用 `LanuageService` 提供的 `getQuickInfoAtPosition`方法获得 `QuickInfo`，然后对 `QuickInfo.displayParts` 进行拼接即可得到类型文案显示。

```typescript
interface QuickInfo {
    displayParts?: SymbolDisplayPart[];
}

interface SymbolDisplayPart {
    /**
     * Text of an item describing the symbol.
     */
    text: string;
    /**
     * The symbol's kind (such as 'className' or 'parameterName' or plain 'text').
     */
    kind: string;
}
```

例如这里的 `htmlType` 字段，根据 vscode 编辑器提示，我们只需要显示红框内容给开发者，`QuickInfo` 直接帮我们展开了联合类型，省去了 AST 解析工作。

![ButtonProps](https://cdn.nlark.com/yuque/0/2025/png/1605148/1753674126216-597bbe8c-09d9-4ad9-b2ce-e873663d0043.png)

## API 字段如何衍生新的 API？

当 `QuickInfo.displayParts` 中含有别的类型引用（`interfaceName`/`aliasName`）时，即下图中的 `CountConfig`，将会触发衍生新的 API，以下是衍生步骤：

+ 先看在当前文件中是否存在此类型声明，通过 `SourceFile.getFullText()` 获取源代码进行字符串索引。
+ 如果找到，则利用 `LanguageService.getTypeDefinitionAtPosition` 获取类型定义。
+ 如果找不到，则证明此类型是外部引入的（即 `import`进来的），这时调用 `typeNode.getType().getText()`能获取到导入的文件路径以及声明名称，例如 `import("/PATH/TO/FILE.ts").CountConfig`。

![BadgeProps](https://cdn.nlark.com/yuque/0/2025/png/1605148/1753674202893-04f23184-61e9-4968-8b63-2de94d525306.png)

## API 注释标签支持

### `@apiFieldsDepth`

自定义 API 收集字段的深度，默认 `interface` 为 1，`typeAlias` 为 2。

### `@apiNameAlias`

声明 API 名称别名。

## API 字段注释标签支持

### `@defaultValue` 或 `@default`

表示此字段的默认值。

### `@internal`

表示此字段为内部字段，`<font style="color:#DF2A3F;">`不会被收集（包括因这个字段引发的 API 也不会收集）`</font>`

### `@description`

表示此字段的描述。

### `@deprecated`

表示此字段已被废弃。

### `@version`

表示此字段在此版本中被添加。

### `@specificTypeText`

指定此字段需要显示的类型文案。

## TypeScript 类型设计

```typescript
export type API = InterfaceAPI | TypeAliasAPI;

export interface CommonAPI {
  /**
   * API name.
   */
  name: string;

  /**
   * API fields.
   */
  fields?: Field[];

  /**
   * When fields is empty, this field will be used.
   */
  typeText?: string;

   /**
    * File URL of this API.
    */
   importPath: stirng;
}

export interface InterfaceAPI extends CommonAPI {
  /**
   * Fixed interface api name.
   */
  apiType: "interface";
}

export interface TypeAliasAPI extends CommonAPI {
  /**
   * Fixed typeAlias api name.
   */
  apiType: "typeAlias";
}

export interface Field {
  /**
   * Field name.
   */
  name: string;
  
  /**
   * Field type name.
   */
  type: string;
  
  /**
   * Field description.
   */
  description?: string;
  
  /**
   * If this field is required.
   */
  required?: boolean;
  
  /**
   * Defualt value of this field.
   */
  defaultValue?: string;
  
  /**
   * If this field is for internal usage, will be exclued from the result.
   */
  internal?: boolean;
  
  /**
   * If this field is deprecated.
   */
  deprecated?: boolean;
  
  /**
   * Specify this field is added in which version.
   */
  version?: string;
}
```

## FAQ

### 如何对非入口文件的 API 做过滤以及外链？

你可以根据 `API.importPath` 判断某 API 是否来自入口文件，并做过滤或外链。

### 类型参数显示吗？

为了减少干扰信息，不显示类型参数。

## 解析示例

| ts 写法                                                                                                                                                                                                                   | 生成结果                                                                                                                                                                                                                                                      |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| ``tsx // 深度 1 interface Props extends ParentProps {   name: string;   age: number; }  interface ParentProps {   more: any; } ``                                                                                         | ``tsx [   {     name: "Props",     type: "interface",     fields: [       { name: "name", typeText: "string" },       { name: "age", typeText: "number" },     ]   } ] ``                                                                                     |
| ``tsx // 深度 1 type Type = Type1 & Type2;  // 深度 2 type Type1 = {   name: string;   age: number; };  // 深度 2 type Type2 = {   salary: Salary; }  // 深度 3 interface Salary {   money: Money;   amount: number; } `` | ``tsx [   {     name: "Type",     type: "typeAlias",     typeAliasCategory: "intersection",     fields: [       { name: "name", typeText: "string" },       { name: "age", typeText: "number" },       { name: "salary", typeText: "Salary" },     ]   } ] `` |
| ```typescript // 深度 1 type Type = Type1                                                                                                                                                                                 | Type2;  // 新 API 深度 1 type Type1 = {   name: string;   age: number; };  // 新 API 深度 1 type Type2 = {   // 解析出新的 API 深度 1   salary: Salary; }  // 新 API 深度 1 interface Salary {   money: Money;   amount: number; } ```                        |
| ``typescript // 深度 1 type Type = {   name: string;   age: number; }; ``                                                                                                                                                 | ``typescript [   {     name: "Type",     type: "typeAlias",     typeAliasCategory: "objectLiteral",     fields: [       { name: "name", typeText: "string" },       { name: "age", typeText: "number" },     ]   } ] ``                                       |
| ``typescript // 深度 1 type Type = TypeDeep;  // 深度 2 interface TypeDeep {   name: string;   age: number; } ``                                                                                                          | ``typescript [   {     name: "Type",     type: "typeAlias",     typeAliasCategory: "reference",     fields: [       { name: "name", typeText: "string" },       { name: "age", typeText: "number" },     ]   } ] ``                                           |
| ``typescript // 深度 1 type Type = Partial<TypeDeep>;  // 深度 2 interface TypeDeep {   name: string;   age: number; } ``                                                                                                 | ``typescript [   {     name: "Type",     type: "typeAlias",     typeAliasCategory: "referencePartial",     fields: [       { name: "name", typeText: "string", required: false },       { name: "age", typeText: "number", required: false },     ]   } ] ``  |
| ``typescript // 深度 1 type Type = Readonly<TypeDeep>;  // 深度 2 interface TypeDeep {   name: string;   age: number; } ``                                                                                                | ``typescript [   {     name: "Type",     type: "typeAlias",     typeAliasCategory: "referenceReadonly",     fields: [       { name: "name", typeText: "string", readonly: true },       { name: "age", typeText: "number", readonly: true },     ]   } ] ``   |
| ``typescript // 深度 1 type Type = Pick<TypeDeep, "name">;  // 深度 2 interface TypeDeep {   name: string;   age: number; } ``                                                                                            | ``typescript [   {     name: "Type",     type: "typeAlias",     typeAliasCategory: "referencePick",     fields: [       { name: "name", typeText: "string" },     ]   } ] ``                                                                                  |
| ``typescript // 深度 1 type Type = Omit<TypeDeep, "name">;  // 深度 2 interface TypeDeep {   name: string;   age: number; } ``                                                                                            | ``typescript [   {     name: "Type",     type: "typeAlias",     typeAliasCategory: "referenceOmit",     fields: [       { name: "age", typeText: "number" },     ]   } ] ``                                                                                   |
| ``typescript // 深度 1 type Type = Type1;  // 深度 2 type Type1 = Type2;  // 深度 3 type Type2 = TypeDeep;  // 深度 4 interface TypeDeep {   name: string;   age: number; } ``                                            | ``typescript [   {     name: "Type",     type: "typeAlias",     typeAliasCategory: "reference",     typeText: "Type1"   } ] ``                                                                                                                                |
