# Simple Chatroom

Simple Chatroom 是一个基于 Go 和 Gin 框架的简单聊天室应用。它支持用户注册、登录、好友管理、群聊等功能。

## 功能特性

- 用户注册和登录
- JWT 认证
- 好友管理
- 群聊功能
- 消息发送和接收

## 安装步骤

1. **克隆项目**

   ```bash
   git clone https://github.com/yourusername/simple-chatroom.git

   cd simple-chatroom
   ```

2. **安装依赖**

   确保您已安装 Go 语言环境，然后运行以下命令安装依赖：

   ```bash
   go mod tidy
   ```

3. **运行项目**

   使用以下命令启动项目：

   ```bash
   go run main.go
   ```

   项目将运行在 `http://localhost:8082`。

## 使用说明

1. **注册和登录**

   - 访问 `http://localhost:8082/index` 进行用户注册和登录。
   - 登录成功后，您将获得一个 JWT token，用于后续的 API 请求认证。

2. **好友管理**

   - 登录后，您可以添加好友、查看好友列表。

3. **群聊功能**

   - 创建群聊，邀请好友加入群聊。
   - 在群聊中发送和接收消息。

### 使用前请确保：

1. MySQL 数据库服务已启动，并创建了 `ginchat` 数据库
2. Redis 服务已启动并可连接
3. 根据实际环境修改相应的连接信息

## 贡献指南

欢迎任何形式的贡献！您可以通过以下方式参与：

- 提交问题或功能请求
- 提交代码改进或修复
- 撰写或改进文档

## 许可证

该项目使用 MIT 许可证。有关更多信息，请参阅 [LICENSE](LICENSE) 文件。
