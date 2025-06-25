## Http Diff

Http Diff 是一个用于对比接口响应数据的工具，使用相同参数分别调用接口A、接口B，然后对两个接口的响应数据进行 Diff，最终输出对比结果。

### 支持的功能

通过配置文件指定任务执行信息，如：

```toml
[[diff_configs]]
name = "task_1"
concurrency = 1
wait_time = "1000ms"
work_dir = "./data"
payload = "payload_task_1.txt"
url_a = "http://127.0.0.1:8080/ping"
url_b = "http://127.0.0.1:8080/ping"
method = "POST"
content_type = "application/json"
ignore_fields = "trace_id"
output_show_no_diff_line = true
log_statistics = true
success_conditions = "stat=1"
```

* 多任务同时运行，通过在配置文件中配置多个任务支持多个任务同时运行。
* 支持并发请求，通过配置文件中的 `concurrency` 参数设置每个任务的并发数。
* 限流支持，通过配置文件中的 `wait_time` 参数设置每次请求完成后等待的时间。
* 支持设置工作目录，通过配置文件中的 `work_dir` 参数设置工作目录，建议每个任务设置一个工作目录。
* 支持请求参数文件，通过配置文件中的 `payload` 参数指定请求参数文件。每行代表一个请求的参数信息，JSON 格式。
  * params：指定拼接在 URL 后面的参数，需进行 URL 编码，比如：`key1=value1&key2=value2` 编码后的数据为：`key1%3Dvalue1%26key2%3Dvalue2`。
  * headers：指定请求的 HTTP 头，格式为 JSON，需要数据进行 JSON 转义，比如：`{"Name":"aaa","traceid":"bbb"}`，转义后的数据为：`{\"Name\":\"aaa\",\"traceid\":\"bbb\"}`。
  * body：指定请求的请求体，用于 POST 请求。
  * 一个完整的请求示例：`{"params": "key1%3Dvalue1%26key2%3Dvalue2", "headers": "{\"Name\":\"aaa\",\"traceid\":\"bbb\"}", "body":"{\"ids\":\"123\",\"userId\":\"456\"}"}`。
* 支持设置请求的 URL，通过配置文件中的 `url_a` 和 `url_b` 参数分别指定两个接口的URL。
* 支持指定 HTTP 请求的 Method，目前支持 GET 和 POST。
* 支持设置请求的 Content-Type，通过配置文件中的 `content_type` 参数指定请求的内容类型，默认为空。
* 支持设置忽略的字段，设置后相应字段在对比时会被忽略，支持多级，但不支持设置忽略数组中的字段。通过配置文件中的 `ignore_fields` 参数指定在对比响应数据时忽略的字段，多个字段用逗号分隔。
* 支持设置是否在输出文件中记录没有差异的行，通过配置文件中的 `output_show_no_diff_line` 参数设置，默认为 `false`。
* 支持设置是否在日志中记录统计信息，通过配置文件中的 `log_statistics` 参数设置，默认为 `false`。
  * 可以通过命令：`cat 日志文件 | grep "Task_logStatisticsInfo_"` 查看统计信息。
* 支持设置成功条件，通过配置文件中的 `success_conditions` 参数设置，多个条件用逗号分隔。每个条件格式为 `key1=value1,key2=value2`，表示响应数据中必须包含该键值对才能认为请求成功。

### 如何使用

1. 第一步，先构建项目

```shell
go build -o http-diff main.go
```

2. 第二步，运行程序

```shell
./http-diff start
```

## todo

* 统计数据并发访问问题处理。