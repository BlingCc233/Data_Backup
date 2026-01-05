# CcDataBak - 跨平台高速加密备份工具

> UESTC 软件开发综合实验

## 声明

> [!NOTE]\
> 22级大四上实验，同届请勿1:1copy

<br/>

**一款极致并行、安全加密、界面美观的现代化跨平台数据备份软件。**

[![构建状态](https://img.shields.io/github/actions/workflow/status/BlingCc233/Data_Backup/.github/workflows/release.yml?branch=main&style=for-the-badge)](https://github.com/BlingCc233/Data_Backup/actions)
[![最新版本](https://img.shields.io/github/v/release/BlingCc233/Data_Backup?style=for-the-badge)](https://github.com/BlingCc233/Data_Backup/releases)
[![许可证](https://img.shields.io/github/license/BlingCc233/Data_Backup?style=for-the-badge)](./LICENSE)
[![Go 版本](https://img.shields.io/badge/Go-1.2x-blue.svg?style=for-the-badge)](https://golang.org/)
[![Wails 版本](https://img.shields.io/badge/Wails-v2-red.svg?style=for-the-badge)](https://wails.io/)
[![Vue.js 版本](https://img.shields.io/badge/Vue.js-v3-green.svg?style=for-the-badge)](https://vuejs.org/)

---

**CcDataBak** 是一款为解决现代数据备份挑战而生的桌面应用程序。它利用 Go 语言的强大并发能力实现了极致的备份速度，并通过强大的加密算法确保您的数据安全无虞。无论您是开发者、摄影师还是需要保护重要文件的普通用户，CcDataBak 都能提供一个可靠、高效且美观的解决方案。

**支持的平台：** Windows, macOS, Linux

## ✨ 核心特性

*   **⚡ 极致的并行化备份**
   *   **多核性能榨取**：深度利用 Go 语言的 Goroutine 和 Channel，将备份任务智能分解为数千个微任务，并充分利用现代 CPU 的每一个核心，实现文件处理和数据传输的最大并行化。
   *   **智能 I/O 调度**：优化磁盘读写操作，通过并发 I/O 显著减少等待时间，即使面对海量小文件也能保持高速率。
   *   **即时反馈**：备份过程中的每一个文件状态都实时更新，让您对进度了如指掌。

*   **🛡️ 银行级的安全加密**
   *   **端到端加密**：在数据离开您的设备之前，使用强大的 **AES-256** 算法进行加密。这意味着即使备份目标被攻破，您的数据依然是安全的。
   *   **安全密钥管理**：您的加密密钥由您完全掌控，我们绝不存储或传输您的密钥。
   *   **TODO 数据完整性校验**：通过哈希校验确保备份数据在传输和存储过程中没有被篡改。

*   **🎨 丰富且直观的筛选器**
   *   **按规则包含/排除**：通过通配符 (`*`, `?`) 或正则表达式轻松定义需要备份或忽略的文件/文件夹。 例如，`*.log` 可以排除所有日志文件。
   *   **按文件属性筛选**：根据文件大小、修改日期等属性进行精细筛选。例如，只备份过去7天内修改过的，且小于 100MB 的文件。
   *   **模板化筛选**：内置常用筛选模板（如“代码项目”、“文档”、“照片”），一键应用，省时省力。

*   **🖼️ 美观易用的 GUI**
   *   **现代化界面**：基于 Vue.js 3 和精美 UI 库打造，提供流畅、直观的用户体验。
   *   **明暗主题切换**：支持浅色和深色两种模式，适应不同工作环境和个人偏好。
   *   **跨平台一致性**：无论您在 Windows、macOS 还是 Linux 上使用，都能获得原生应用般的体验。

## 📸 软件截图

<table>
  <tr>
    <td align="center"><strong>主界面 (浅色模式)</strong></td>
    <td align="center"><strong>新建备份任务 (深色模式)</strong></td>
  </tr>
  <tr>
    <td><img src="https://place-hold.it/400x300/f1faee/1d3557" alt="主界面截图"></td>
    <td><img src="https://place-hold.it/400x300/1d3557/f1faee" alt="新建任务截图"></td>
  </tr>
  <tr>
    <td align="center"><strong>高级筛选器设置</strong></td>
    <td align="center"><strong>备份进度详情</strong></td>
  </tr>
  <tr>
    <td><img src="https://place-hold.it/400x300/e9c46a/264653" alt="筛选器截图"></td>
    <td><img src="https://place-hold.it/400x300/2a9d8f/e9c46a" alt="进度详情截图"></td>
  </tr>
</table>


## 🛠️ 技术栈

*   **后端**: Go (Golang)
*   **GUI 框架**: Wails v2
*   **前端**: Vue.js 3
*   **UI 组件库**: Element Plus / Naive UI
*   **数据库**: SQLite (用于存储配置和任务历史)

## 🚀 安装与启动

### 预备环境

在开始之前，请确保您的系统已经安装了 Go 语言环境和 Wails 所需的依赖。具体请参考 [Wails 官方安装指南](https://wails.io/docs/gettingstarted/installation)。

### 从源码构建

1.  **克隆仓库**
    ```bash
    git clone https://github.com/BlingCc233/Data_Backup.git
    cd Data_Backup/go-backup-app
    ```

2.  **安装前端依赖**
    ```bash
    cd frontend
    npm install
    cd ..
    ```

3.  **构建并运行**
   *   **开发模式**：实时重新加载，便于开发。
       ```bash
       wails dev
       ```
   *   **生产构建**：将应用打包为单个可执行文件。
       ```bash
       wails build
       ```
       可执行文件将位于 `build/bin/` 目录下。

### 从 Releases 下载

您可以直接从 [GitHub Releases](https://github.com/BlingCc233/Data_Backup/releases) 页面下载适用于您操作系统的最新版本。

## 📖 使用说明

1.  打开 CcDataBak 应用程序。
2.  点击“新建备份任务”按钮。
3.  **选择源**：选择您想要备份的文件夹。
4.  **配置筛选器**：(可选) 添加包含或排除规则，例如 `*.tmp` 来排除临时文件。
5.  **选择目标**：选择备份数据存储的位置（本地文件夹、移动硬盘等）。
6.  **设置加密**：输入并确认您的加密密码。**请务必牢记此密码，丢失后数据将无法恢复！**
7.  点击“开始备份”，任务将立即开始。

## 🤝 如何贡献

我们热烈欢迎任何形式的贡献！无论是提交 Bug、提出新功能建议，还是直接贡献代码。

如果您想贡献代码，请遵循以下步骤：

1.  Fork 本仓库。
2.  创建一个新的分支 (`git checkout -b feature/feature-name`)。
3.  提交您的代码 (`git commit -m 'feat: Add some amazing feature'`)。
4.  将您的分支推送到 GitHub (`git push origin feature/feature-name`)。
5.  创建一个 Pull Request。

## TODO


1. 实现 pkg/scheduler: 引入 github.com/kardianos/service 库，并根据其文档为 Windows, macOS (launchd), 和 Linux (systemd) 实现服务注册和注销。
2. 网络备份 (Feature): 为 S3, FTP 等协议设计并实现上传/下载逻辑。
3. 增强前端UI:
   * 新增管理定时任务和服务的UI。
   * sqlite可展开显示每项备份的内容.

## TODO2

1. 实现服务与定时任务 (app.go):
   1. 引入 github.com/kardianos/service 来处理跨平台的服务注册。
   2. 引入 github.com/robfig/cron/v3 来解析 cron 表达式并调度备份任务。在 main 函数或 App 结构体中启动一个 cron 调度器。
2. 实现网络备份 (core/network.go):
   1. 为 S3, FTP 等创建具体的 Uploader 实现。例如，S3Uploader 会使用 AWS Go SDK。
   2. 在 BackupManager 中，检查目标路径是否是 URL，如果是，则创建一个内存 pipe，ArchiveWriter 写入 pipe，网络上传器从 pipe 读取并上传。
3. 完善前端 UI:
   1. 添加用于定时任务的 UI 组件。


## 更新日志

**v1.1.1**
- [X] 基本功能实现
- [X] 加解密测试
- [X] 压缩解压测试

**v1.1.2**
- [X] Huffman压缩并行化提升大文件（夹）备份速度
- [X] 并行化压缩测试（~~解压依旧串行、huffman算法限制~~）
- [X] 大文件(夹)并行分块读取、内存优化

**v1.1.3**
- [X] 加解密并行优化
- [X] 多文件并行io，单文件并行加解密

**v1.1.4**
- [X] 修复bug：单个有加密backup后restore：Error: failed to read next archive entry (archive may be corrupt): failed to read header json: unexpected EOF
- [X] 修复bug：有加密的情况下restore往内存读一整个qbak

**v1.1.5**
- [X] GUI重构：交互逻辑、显示逻辑
- [X] 新增多源file筛选

**v1.1.6**
- [X] 修复GUI显示BUG
- [X] 恢复逻辑重构
- [X] 恢复回退并行化、抛弃预览功能
- [X] 数据库自动追新

**v1.1.7**
- [X] 修复加解密时密码无效的bug
- [X] 进度条流程优化
- [X] 恢复时无法取消的bug
- [X] 文件冲突解决

**v1.1.8**
- [X] 图标更新
- [X] restore显示所存内容
- [X] 文件夹展开筛选

**v1.1.9**
- [X] v1.2.0之前最后做的优化
- [X] 优化垃圾回收
- [X] 优化GUI渲染

**v1.2.0**
- [X] 实现档案预设功能
- [X] 增大了缓冲池大小
- [X] 网络备份函数实现

**v1.2.1**
- [X] 修复windows直接打开恢复页面无内容bug
- [X] 新增对压缩的选项开关

**v1.2.3**
- [X] 重构UI，采用neo brutalism风格
- [X] 添加增量备份功能
- [X] 添加定时任务功能
- [X] 添加服务功能
- [X] 添加实时备份功能
- [X] 修复了一些UI交互问题
- [X] 修复了任务中无法停止的bug
- [X] 修复了异常内存占用bug

## 📄 许可证

本项目基于 `GPL-3.0 License` 开源。
