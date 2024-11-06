# http2tcp

将 HTTP 链接转换为 TCP 通道。参考了 [http2tcp](https://github.com/movsb/http2tcp) 的实现。

## 安装

在 GitHub release 页面下载 GitHub action 自动构建发布的二进制文件，或者自行构建。

## 使用

如下命令产生的结果：服务端监听 `8080` 端口，客户端将 `8081` 端口的 TCP 链接转发到服务端的 `6379` 端口。

```bash
./http2tcp server -l :8080 -a longlongauthtoken
```

```bash
./http2tcp client -s serverhost:8080 -a longlongauthtoken -t 127.0.0.1:6379 -l 127.0.0.1:8081
```

### 作为 `ssh` 的 `ProxyCommand` 使用

```bash
./http2tcp client -s serverhost:8080 -a longlongauthtoken -t 127.0.0.1:22 -l -
```

## 原理

HTTP 规范里，携带 `Upgrade` 头的请求可以将 HTTP 协议的链接转换为其他协议的链接，在服务端返回 `101` 状态码之后，链接经过的七层代理服务（例如 `nginx`）将转变为四层代理。`http2tcp` 利用这一点，将 HTTP 链接转换为加密的 TCP 通道。


### windows使用
```shell
# 本地开启监听15433端口，将 15433 的tcp请求通过 http://127.0.0.1:8080 server端 转发到 127.0.0.1:5433
 .\http2tcp-windows-amd64.exe client -s http://127.0.0.1:8080 -a longlongauthtoken -t 127.0.0.1:5433 -l 127.0.0.1:15433 -m tcp

 .\http2tcp-windows-amd64.exe client -s 127.0.0.1:8080 -a longlongauthtoken -t 127.0.0.1:3306 -l 127.0.0.1:13306 -m tcp

 .\http2tcp-windows-amd64.exe server -l :8080 -a longlongauthtoken
```