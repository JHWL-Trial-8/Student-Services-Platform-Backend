package http

import (
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"

    "student-services-platform-backend/internal/config"
)

func CORS(cfg config.CORSConfig) gin.HandlerFunc {
    allowed := normalize(cfg.AllowedOrigins) // 正则化，方便匹配

    return func(c *gin.Context) {
        origin := c.GetHeader("Origin")
        if origin != "" {
            // 如果是 "*" 且 credentials=false，直接返回 "*"。
            // 如果 credentials=true，必须回显特定的 Origin 以符合浏览器的要求。
            allowOrigin := ""
            if contains(allowed, "*") {
                if cfg.AllowCredentials {
                    allowOrigin = origin
                } else {
                    allowOrigin = "*"
                }
            } else if originAllowed(allowed, origin) {
                allowOrigin = origin
            }

            if allowOrigin != "" {
                c.Header("Access-Control-Allow-Origin", allowOrigin)
                if cfg.AllowCredentials {
                    c.Header("Access-Control-Allow-Credentials", "true")
                }
                if len(cfg.AllowedHeaders) > 0 {
                    c.Header("Access-Control-Allow-Headers", strings.Join(cfg.AllowedHeaders, ", "))
                }
                if len(cfg.AllowedMethods) > 0 {
                    c.Header("Access-Control-Allow-Methods", strings.Join(cfg.AllowedMethods, ", "))
                }
            }
        }

        if c.Request.Method == http.MethodOptions {
            c.Status(http.StatusNoContent)
            c.Abort()
            return
        }
        c.Next()
    }
}

func normalize(xs []string) []string {
    out := make([]string, 0, len(xs))
    for _, x := range xs {
        out = append(out, strings.TrimSpace(strings.ToLower(x)))
    }
    return out
}

func contains(xs []string, target string) bool {
    for _, x := range xs {
        if x == target {
            return true
        }
    }
    return false
}

func originAllowed(allowed []string, origin string) bool {
    o := strings.ToLower(origin)
    // 精确匹配
    for _, a := range allowed {
        if a == o {
            return true
        }
        // 轻量级的子域名通配符支持
        if strings.HasPrefix(a, ".") && strings.HasSuffix(o, a) {
            return true
        }
    }
    return false
}