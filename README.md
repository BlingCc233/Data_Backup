# Data_Backup

---

UESTC 软件开发综合实验

## 2022级实验用，后辈可抄，同届请勿1:1copy

---

# TODO

1. 完善 pkg/filter: 实现所有筛选逻辑，特别是路径和名称的通配符匹配，以及平台相关的用户筛选。
2. 实现 pkg/scheduler: 引入 github.com/kardianos/service 库，并根据其文档为 Windows, macOS (launchd), 和 Linux (systemd) 实现服务注册和注销。
3. 实现 app.go 中的配置持久化: 将 GetBackupProfiles 和 SaveProfile 函数连接到一个本地的 JSON 或 SQLite 文件，以保存用户的配置。
4. 网络备份 (Feature): 为 S3, FTP 等协议设计并实现上传/下载逻辑。
5. 增强前端UI:
    * 提供文件夹内子文件筛选功能。
    * 重写备份进度条，显示加密、压缩、写入的进度。
    * 新增管理定时任务和服务的UI。
    * 对于恢复时的历史选择、加入动画与进度条等待后台任务完成。
    * 恢复成功的提示
    * 完成备份/恢复回到首页
6. **重要‼️：Restore抛弃文件预览、回退到v1.1.4的并行/分块逻辑，减小内存占用**

