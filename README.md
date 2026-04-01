# oauth_tools

WPS365 开放平台 OAuth Access Token 命令行工具。

## 功能

- **用户 token**：自动打开浏览器完成用户授权（`authorization_code` 流程），获取后自动查询并打印当前用户信息
- **租户 token**：直接用 AppID/Secret 获取（`client_credentials` 流程），无需浏览器
- 输出 HTTP `Authorization` header，可直接复制使用
- 支持 KSO-1 请求签名（`sign` 包）

## 安装

```bash
git clone https://github.com/hunter-xue/wps365-oauth-cli.git
cd wps365-oauth-cli
```

### 构建

构建当前平台：

```bash
go build -o oauth_tools .
```

交叉编译所有平台（输出到 `build/` 目录）：

```bash
make all        # 全平台
make macos      # macOS   → build/macos/   (amd64 + arm64)
make linux      # Linux   → build/linux/   (amd64 + arm64)
make windows    # Windows → build/windows/ (amd64)
make clean      # 清理 build/ 目录
```

## 配置

复制示例配置，**只需替换 `APP_ID` 和 `SECRET` 两个字段**，其余保持默认即可：

```bash
cp .env.example .env
```

```bash
# 只需修改这两行：
APP_ID=your_app_id_here    # 替换为你的 AppID
SECRET=your_secret_here    # 替换为你的 Secret
```

`AUTH_URL` 中的 `client_id` 会自动引用 `${APP_ID}` 的值，无需手动修改。

<details>
<summary>完整字段说明</summary>

| 字段 | 必填 | 说明 |
|------|------|------|
| `APP_ID` | 是 | 应用的 AppID（accessKey） |
| `SECRET` | 是 | 应用的 Secret（secretKey） |
| `ENDPOINT` | 是 | Token 端点，默认已填写 WPS365 地址 |
| `AUTH_URL` | 用户 token 必填 | 授权 URL，`${APP_ID}` 会自动替换，无需修改 |
| `SCOPES` | 否 | 权限列表，逗号分隔，会自动注入到 `AUTH_URL` |
| `API_BASE_URL` | 否 | API 根地址，用于获取用户信息及输出提示 |

</details>

## 使用

### 获取用户 token（默认）

```bash
./oauth_tools token
```

1. 自动在本地启动回调服务器
2. 自动打开浏览器，引导用户授权
3. 授权完成后自动获取 token 并打印当前用户信息

### 获取租户 token

```bash
./oauth_tools token -type tenant
```

直接使用 AppID/Secret 获取租户级别的 access token，无需浏览器，不查询用户信息。

### 输出示例

**用户 token**（`./oauth_tools token`）：

```
Listening for OAuth callback on http://127.0.0.1:18000/koa-callback
Opening browser: https://openapi.wps.cn/oauth2/auth?...
────────────────────────────────────────
access_token:  eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9...
token_type:    bearer
expires_in:    7200s
refresh_token: kso_rt_xxxxxxxxxxxxxxxxxxxxxx
expires_at:    2026-04-01 16:15:50 UTC
api_base_url:  https://openapi.wps.cn/v7
────────────────────────────────────────
Authorization: Bearer eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9...
────────────────────────────────────────
user_id:       KANabcxyz
user_name:     金山云SA
company_id:    lLxyzabc
avatar:        https://img.example.com/avatar.jpg
```

**租户 token**（`./oauth_tools token -type tenant`）：

```
────────────────────────────────────────
access_token:  eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9...
token_type:    bearer
expires_in:    7200s
expires_at:    2026-04-01 16:46:34 UTC
api_base_url:  https://openapi.wps.cn/v7
────────────────────────────────────────
Authorization: Bearer eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9...
```

### 输出格式

```bash
# 仅输出 access_token（适合脚本）
./oauth_tools token -token-only
./oauth_tools token -type tenant -token-only

# 导出为环境变量
export TOKEN=$(./oauth_tools token -token-only)

# JSON 格式
./oauth_tools token -json
./oauth_tools token -type tenant -json
```

### 指定配置文件

```bash
./oauth_tools -env /path/to/.env token
./oauth_tools -env /path/to/.env token -type tenant
```

## 帮助

```bash
./oauth_tools -help
./oauth_tools token -help
```

## KSO-1 签名

获取 token 后，调用 WPS365 API 需要对请求进行 KSO-1 签名。`sign` 包提供了签名能力：

```go
signer, _ := sign.New(appID, secret)
signer.Apply(req, body) // 自动设置 X-Kso-Date 和 X-Kso-Authorization
```
