package middleware

import (
    "net/http"

    "student-services-platform-backend/app/contextkeys"
    dbpkg "student-services-platform-backend/internal/db"

    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

// RequireRole 基于数据库中的用户角色做鉴权；需在 JWTAuth 之后使用。
func RequireRole(db *gorm.DB, allowed ...dbpkg.Role) gin.HandlerFunc {
    allowedSet := make(map[dbpkg.Role]struct{})
    for _, r := range allowed {
        allowedSet[r] = struct{}{}
    }

    return func(c *gin.Context) {
        // 读取由JWTAuth放置在contextkeys.UserIDKey下的类型为uint的用户ID（uid）
        val, ok := c.Get(string(contextkeys.UserIDKey))
        if !ok {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
            return
        }
        uid, ok := val.(uint)
        if !ok || uid == 0 {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "用户 ID 无效"})
            return
        }

        // 实时从数据库获取用户最小字段，并绑定请求上下文
        var u dbpkg.User
        if err := db.WithContext(c.Request.Context()).
            Select("id", "role").
            First(&u, uid).Error; err != nil {
            // 包括用户不存在的情况，统一返回无权限
            c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "无权限"})
            return
        }

        // 检查角色是否在允许的集合中
        if _, ok := allowedSet[u.Role]; !ok {
            c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "无权限"})
            return
        }

        // 使用类型安全的键将必要信息存入 context（保持与全局一致）
        c.Set(string(contextkeys.UserIDKey), u.ID)
        c.Set(string(contextkeys.UserRoleKey), u.Role)

        c.Next()
    }
}