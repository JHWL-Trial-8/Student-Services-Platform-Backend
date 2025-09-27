package contextkeys

// CtxKey 使用自定义类型避免上下文键冲突
type CtxKey string

const (
	// UserIDKey 用于在上下文中存储用户 ID (uint)
	UserIDKey CtxKey = "uid"
	// UserRoleKey 用于在上下文中存储用户角色 (db.Role)
	UserRoleKey CtxKey = "role"
)