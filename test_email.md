# 邮件功能测试指南

## 🧪 测试步骤

### 1. 配置邮件服务
复制 `config/config.example.yaml` 为 `config/config.yaml`，并配置：
```yaml
email:
  smtp_host: "smtp.qq.com"            # 或其他邮箱服务商
  smtp_port: 587
  smtp_username: "your-email@qq.com"   # 您的邮箱
  smtp_password: "your-auth-code"      # 邮箱授权码（不是登录密码）
  from_email: "your-email@qq.com"
  from_name: "学生服务平台"
  tls_enabled: true
```

### 2. 启动服务
```bash
# 启动API服务
go run ./cmd/api
```

### 3. 测试邮件功能
```bash
# 1. 注册用户/登录获取token
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password"}'

# 2. 创建工单
curl -X POST http://localhost:8080/api/v1/tickets \
  -H "Authorization: Bearer <your-token>" \
  -H "Content-Type: application/json" \
  -d '{"title":"测试工单","content":"测试邮件通知","category":"技术支持"}'

# 3. 管理员接单
curl -X POST http://localhost:8080/api/v1/tickets/1/claim \
  -H "Authorization: Bearer <admin-token>"

# 4. 管理员标记已处理（触发邮件）
curl -X POST http://localhost:8080/api/v1/tickets/1/resolve \
  -H "Authorization: Bearer <admin-token>"
```

### 4. 检查邮件
工单创建者应该会收到"工单已处理完成"的邮件通知。

## 📧 邮件内容示例
```html
🎉 您的工单已处理完成

xxx 同学，
您提交的工单已经处理完成：

工单编号：1
标题：测试工单
处理人员：管理员姓名
处理时间：2024-01-01 12:00:00

处理结果：
您的工单已处理完成

如果您对处理结果满意，欢迎给我们评价。
```

## ⚠️ 注意事项
1. 确保邮箱配置正确
2. QQ邮箱需要使用授权码，不是登录密码
3. 检查防火墙是否允许SMTP连接
4. 第一次可能进入垃圾邮件，需要标记为信任