# Ver.2

---

1. 完善筛选器 (core/filters.go): 在 ShouldInclude 函数中，根据 FilterConfig 的字段（路径、大小、时间等）添加详细的 if/else 判断逻辑。

2. 实现服务与定时任务 (app.go):

    1. 引入 github.com/kardianos/service 来处理跨平台的服务注册。
    2. 引入 github.com/robfig/cron/v3 来解析 cron 表达式并调度备份任务。在 main 函数或 App 结构体中启动一个 cron 调度器。

3. 实现网络备份 (core/network.go):

    1. 为 S3, FTP 等创建具体的 Uploader 实现。例如，S3Uploader 会使用 AWS Go SDK。
    2. 在 BackupManager 中，检查目标路径是否是 URL，如果是，则创建一个内存 pipe，ArchiveWriter 写入 pipe，网络上传器从 pipe 读取并上传。

4. 完善前端 UI:

   1. 添加用于配置筛选器、加密选项、定时任务的 UI 组件。
   2. 将这些配置项通过 v-model 绑定到数据对象，并在调用 Go 函数时传递。
   3. 使用图表库展示备份历史或统计信息。
