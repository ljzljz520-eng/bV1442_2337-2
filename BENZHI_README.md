# 售后工单描述清理工具

这是一个纯 Go 的本地 HTTP 工具：提交售后工单描述后，服务会清理 HTML 标记，返回清理后的文本、字符数、词语数和问题分类。浏览器页面包含提交中的加载状态、明确错误和结果展示。

## 运行

```bash
go run .
```

服务默认监听 `http://localhost:8080`。也可以先构建再运行：

```bash
go build -o ticket-cleaner .
./ticket-cleaner
```

接口为 `POST /api/tickets/clean`，请求体格式为 `{"description":"物流一直没有更新"}`。

## 业务测试

从模块根目录运行：

```bash
go test -count=1 ./...
```

测试覆盖正常清理统计、空白描述校验、清理器不可用和清理器失败等业务路径。
