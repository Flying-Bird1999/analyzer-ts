<details class="lake-collapse"><summary id="u8a6f2e8b"><span class="ne-text">【合规声明】基于 glm4.6 生成，提示词如下</span></summary><p id="u20614dfc" class="ne-p"><span class="ne-text">根据以下想法内容，生成一个完整需求文档</span></p><ol class="ne-ol"><li id="u56b802ed" data-lake-index-type="0"><span class="ne-text">请注意是需求文档而不是技术文档，所以重点需要讲清楚需求场景。</span></li><li id="u64f052ec" data-lake-index-type="0"><span class="ne-text">需要结合一些图例说明</span></li><li id="ub9946042" data-lake-index-type="0"><span class="ne-text">生成需求 markdown</span></li></ol><hr id="iG1aN" class="ne-hr"><p id="u1d6562e1" class="ne-p"><span class="ne-text">底层数据源支撑: 资产系统</span></p><h2 id="xYpEZ"><span class="ne-text">为什么要做</span></h2><ol class="ne-ol"><li id="u806f4357" data-lake-index-type="0"><span class="ne-text">在 UX 规范专项后，组件库基本完成了需要快速能力追平/提供阶段，开始进入迭代阶段，并且在 UX 专项后组件库在整个 shopline 前端项目中，已经有一定的覆盖率。所以当下可以开始考虑如何迭代更稳，组件性能更好。</span></li><li id="u95ca5b82" data-lake-index-type="0"><span class="ne-text">需要构建一套面向组件库的资产数据生成以及管理维护体系方便各种的能力提供</span></li></ol><h2 id="tDlYr"><span class="ne-text">组件库维度的资产</span></h2><ol class="ne-ol"><li id="u10396001" data-lake-index-type="0"><span class="ne-text">提供辅助使用工具（eslint/stylelint/playground等）</span></li><li id="ufeaa3aa5" data-lake-index-type="0"><span class="ne-text">UX 规范(token/字体/日期)</span></li><li id="u0b41b78d" data-lake-index-type="0"><span class="ne-text">规范组件</span></li><li id="ub873073e" data-lake-index-type="0"><span class="ne-text">废弃组件</span></li><li id="u52d4909d" data-lake-index-type="0"><span class="ne-text">wiki</span></li></ol><ol class="ne-list-wrap"><ol ne-level="1" class="ne-ol"><li id="u1d50f954" data-lake-index-type="0"><span class="ne-text">代码规范 &amp; 开发手册</span></li><li id="ue6ef3909" data-lake-index-type="0"><span class="ne-text">使用指南</span></li></ol></ol><h2 id="shcOQ"><span class="ne-text">组件维度看自身的资产数据</span></h2><ol class="ne-ol"><li id="ua71f090f" data-lake-index-type="0"><span class="ne-text">组件的所有 changelog 展示</span></li><li id="u0f4b7a10" data-lake-index-type="0"><span class="ne-text">组件的所有 bug 展示</span></li><li id="ua68e6e90" data-lake-index-type="0"><span class="ne-text">组件内部调用的所有依赖展示，比如 Select 内部依赖有 Dropdown，DropdownMenu, InternalTrigger, @rc/components-xxx</span></li><li id="uc0591e04" data-lake-index-type="0"><span class="ne-text">组件内部消费到的 token 展示</span></li><li id="ub00edbfc" data-lake-index-type="0"><span class="ne-text">组件使用的 best/bad 使用范式，demo 等</span></li><li id="ub0847ce3" data-lake-index-type="0"><span class="ne-text">组件的迭代需求池，roadmap，jira，figma 等</span></li><li id="ua939b013" data-lake-index-type="0"><span class="ne-text">最少运行一个组件的体积，例如抛开三方包，一个 button 运行需要加载 27kb 代码(有混淆无压缩)</span></li><li id="ua1f6d8c0" data-lake-index-type="0"><span class="ne-text">组件对应的测试用例，渲染性能压测用例等</span></li><li id="ub9be8887" data-lake-index-type="0"><span class="ne-text">组件关联的源码</span></li></ol><h3 id="vf0tJ"><span class="ne-text">实际可以应用的场景</span></h3><p id="u469c280d" class="ne-p"><span class="ne-text">面向维护者</span></p><h4 id="JpBtB"><span class="ne-text">MR 阶段（当下重点）</span></h4><ol class="ne-ol"><li id="uf7834e03" data-lake-index-type="0"><span class="ne-text">影响范围，比如一个 Button 的变更会影响到多少组件组件，提醒需要回归的范围</span></li><li id="ua698b743" data-lake-index-type="0"><span class="ne-text">这次迭代生成的性能影响、体积变更影响</span></li></ol><h4 id="WzScr"><span class="ne-text">定期排查问题（当下重点）</span></h4><ol class="ne-ol"><li id="u32207d21" data-lake-index-type="0"><span class="ne-text">发现组件内部不规范写法导致有不符合预期的运行/产物，例如 button bundler 代码中会带上多语言的代码配置。</span></li><li id="u7c4e4924" data-lake-index-type="0"><span class="ne-text">基于 1 可以基于问题去优化组件内部的写法，优化调用逻辑以及优化性能</span></li></ol><p id="uf700392b" class="ne-p"><span class="ne-text">面向使用者（nice to have）</span></p><h4 id="l6scd"><span class="ne-text">AI（当下重点）</span></h4><ol class="ne-ol"><li id="u7cbfe259" data-lake-index-type="0"><span class="ne-text">提供合规的数据给到 ai 消费，可以了解当前运行的 ai 客服流程</span></li><li id="u22dbb37a" data-lake-index-type="0"><span class="ne-text">可以辅助 LLM 生成更合规的代码</span></li><li id="u5ae65f4a" data-lake-index-type="0"><span class="ne-text">给到 LLM chat 客服使用（计划 Q4 会有），UED 有理解组件规范和能力的场景需要，研发有代码生成的场景需要</span></li></ol><p id="u7bbf244c" class="ne-p"><br></p><h4 id="Oxrik"><span class="ne-text">运用到组件文档</span></h4><ol class="ne-ol"><li id="ub9cb4b08" data-lake-index-type="0"><span class="ne-text">看组件内部关系，更好的去自行定位问题</span></li><li id="u4a4651ae" data-lake-index-type="0"><span class="ne-text">让开发者/UED更直观的理解使用组件</span></li><li id="uc2e71f1b" data-lake-index-type="0"><span class="ne-text">组件的(迭代/bug/roadmap)快速了解</span></li></ol><h4 id="gRniD"><span class="ne-text">其他工具</span></h4><ol class="ne-ol"><li id="u40ba1b93" data-lake-index-type="0"><span class="ne-text">理论上可以提供类似 rsdoctor 一样的工具给到开发者自检自己的项目</span></li></ol><p id="ub4599275" class="ne-p"><br></p></details>
## 1. 项目背景
### 1.1 现状分析
随着 UX 规范专项的完成，SHOPLINE Admin Components 组件库已进入快速迭代阶段。组件库在整个 Shopline 前端项目中已具备一定的覆盖率，现在需要重点关注：

+ **迭代稳定性**：如何确保组件库迭代更加稳定可靠
+ **性能优化**：如何持续提升组件性能表现
+ **资产管理**：如何有效管理和利用组件库产生的各类资产数据

### 1.2 问题识别
当前组件库在快速迭代过程中面临以下挑战：

1. **影响范围不明确**：组件变更时难以快速识别影响范围
2. **性能监控缺失**：缺乏系统性的性能影响评估机制
3. **资产管理混乱**：组件相关数据分散，缺乏统一管理
4. **协作效率低下**：维护者、使用者、UED 之间的信息同步不够高效

## 2. 项目目标
### 2.1 总体目标
构建一套完整的组件库资产数据生成、管理和分析体系，为组件库的稳定迭代和性能优化提供数据支撑。

### 2.2 核心价值
+ **提升维护效率**：为维护者提供精准的影响范围分析和性能监控
+ **优化用户体验**：为使用者和 UED 提供更好的组件理解和使用体验
+ **赋能 AI 应用**：为 AI 辅助开发和客服提供结构化的组件数据
+ **保障质量稳定**：建立系统性的质量监控和问题发现机制

## 3. 资产体系规划
### 3.1 组件库维度资产（_示例_）
![](https://cdn.nlark.com/yuque/__mermaid_v3/e642fb81feed7dd474b98c32d1706d2b.svg)

> _注：以上为资产分类的示例说明，实际资产内容以具体实现为准_
>

### 3.2 组件维度资产（_示例_）
![](https://cdn.nlark.com/yuque/__mermaid_v3/dc1a6c3cc35f44adde411ba86a7c86f1.svg)

> _注：以上为组件资产维度的示例说明，具体资产内容会根据实际组件特性和业务需求进行调整_
>

## 4. 核心功能需求
### 4.1 面向维护者功能
#### 4.1.1 MR 阶段辅助分析 🔴 **高优先级**
**功能描述**：在代码合并阶段提供智能化的影响范围分析

**核心场景**：

![](https://cdn.nlark.com/yuque/__mermaid_v3/335c24565eaa9b13fc3d7f823de698c1.svg)

**具体需求**：

+ **依赖关系分析**：识别组件变更对其他组件的级联影响
+ **性能影响评估**：分析代码变更对运行时性能的影响
+ **体积变更监控**：评估代码变更对打包体积的影响
+ **回归范围建议**：基于影响范围推荐需要回归的测试用例

#### 4.1.2 定期质量巡检 🟡 **中优先级**
**功能描述**：系统化地发现和识别组件库中的质量问题

**核心场景**：

![](https://cdn.nlark.com/yuque/__mermaid_v3/e97c6eb4d299a236592f4424c542a191.svg)

**具体需求**：

+ **规范违规检测**：发现不符合组件库规范的代码模式
+ **性能异常识别**：识别性能表现异常的组件
+ **冗余代码发现**：发现无用的依赖和代码片段
+ **安全漏洞扫描**：检测潜在的安全风险

### 4.2 面向使用者功能
#### 4.2.1 AI 辅助开发 🟡 **中优先级**
**功能描述**：为 AI 代码生成和客服提供结构化的组件数据支撑

**核心场景**：

![](https://cdn.nlark.com/yuque/__mermaid_v3/829fecca3e01c65575e9fca18675c3b3.svg)

**具体需求**：

+ **组件数据 API**：提供标准化的组件数据接口
+ **使用模式识别**：识别和推荐正确的组件使用模式
+ **代码合规检查**：确保 AI 生成的代码符合组件库规范
+ **知识库集成**：为 AI 提供最新的组件知识

#### 4.2.2 增强型文档系统 🟡 **中优先级**
**功能描述**：基于资产数据提供更丰富的组件文档体验

**核心场景**：

![](https://cdn.nlark.com/yuque/__mermaid_v3/22f62795f5054824fd2789b2e0a68826.svg)

**具体需求**：

+ **关系图谱展示**：可视化展示组件间的依赖关系
+ **性能数据展示**：显示组件的性能表现和优化建议
+ **版本演进追踪**：展示组件的版本变更历史
+ **问题快速定位**：基于已知问题库提供故障排查指导

### 4.3 面向工具链功能
#### 4.3.1 开发者自检工具 🟢 **低优先级**
**功能描述**：提供类似 rsdoctor 的项目自检能力

**核心场景**：

![](https://cdn.nlark.com/yuque/__mermaid_v3/89f3aff07d8ab458b5e9e5cc65d3fd22.svg)

**具体需求**：

+ **使用规范检查**：检查项目中组件使用是否符合规范
+ **性能问题诊断**：识别项目中的性能瓶颈
+ **版本兼容性分析**：分析组件版本兼容性问题
+ **优化建议推荐**：基于最佳实践提供优化建议

## 5. 技术方案设计要求
### 5.1 架构原则
遵循轻量化、无服务器原则，充分利用现有基础设施能力，尽量低成本实现（服务器资源）。

### 5.2 核心要求
+ **数据维护**：可结合 Fishbone Serverless 技术栈，支持自动化采集、处理和分发
+ **多版本支持**：支持组件库多版本资产维护，CI 阶段统一处理，通过固定规则自动获取
+ **存储选型**：
    - 推荐：CDN 存储、SQLite 存储、鱼骨 NAS 存储
    - 不推荐：服务端数据库、Redis 等重型服务端组件（降低服务器资源成本）

