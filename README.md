# analyzer-ts

[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/Flying-Bird1999/analyzer-ts)

go mod tidy

go run main.go

问题记录：

1. ✅ 解析“name"报错  type A ={ "name": string }
2. ✅ import **as** 的语法 在依赖分析会有问题，找成了as前的。
3. ✅ {size: allTypes.Size} 这种语法解析也有类型的问题
4. ✅ 解析这类语法有问题：export interface Class8 extends Omit<Class2, 'age'> {name:string}
5. ✅ [keyinSupportedLanguages]?:string[] |string;
6. ✅ typePersonName=Translations["name"]
7. 声明在 .d.ts
8. export { Package_Name as Package_Name_export };
9. export default Package_Name;
