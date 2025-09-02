# Data_Backup

---

UESTC 软件开发综合实验

## 2022级实验用，后辈可抄，同届请勿1:1copy

---

# TODO

1. 完善 pkg/filter: 实现所有筛选逻辑，特别是路径和名称的通配符匹配，以及平台相关的用户筛选。
2. 完善 pkg/archive:
    * 添加对文件权限（Mode）和时间戳（ModTime）的恢复。
3. 实现 pkg/scheduler: 引入 github.com/kardianos/service 库，并根据其文档为 Windows, macOS (launchd), 和 Linux (systemd) 实现服务注册和注销。
4. 实现 app.go 中的配置持久化: 将 GetBackupProfiles 和 SaveProfile 函数连接到一个本地的 JSON 或 SQLite 文件，以保存用户的配置。
5. 网络备份 (Feature): 为 S3, FTP 等协议设计并实现上传/下载逻辑。
6. 增强前端UI:
    * 添加一个完整的表单来编辑所有筛选条件。
    * 提供文件夹内子文件筛选功能。
    * 显示备份进度条。
    * 提供恢复文件筛选的界面。
    * 管理定时任务和服务的UI。

