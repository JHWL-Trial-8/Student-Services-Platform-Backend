package midwares

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func JWTAuthMidware(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":  http.StatusUnauthorized,
				"error": "	未识别到Token，请登录后访问",
			})
			return
		}

		const bearerPrefix = "Bearer "
		if len(authHeader) <= len(bearerPrefix) || !strings.HasPrefix(authHeader, bearerPrefix) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":  http.StatusUnauthorized,
				"error": "Token 格式错误，需为 Bearer <token>",
			})
			return
		}
		tokenStr := authHeader[len(bearerPrefix):]

		type CustomClaims struct {
			ID string `json:"id"`
			jwt.RegisteredClaims
		}

		token, err := jwt.ParseWithClaims(tokenStr, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("无效的签名算法，仅支持 HS256")
			}
			return []byte(secretKey), nil
		})

		if err != nil {
			var ve *jwt.ValidationError
			if errors.As(err, &ve) {
				switch {
				case ve.Errors&jwt.ValidationErrorMalformed != 0:
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
						"code":  http.StatusUnauthorized,
						"error": "Token 格式错误",
					})
				case ve.Errors&jwt.ValidationErrorExpired != 0:
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
						"code":  http.StatusUnauthorized,
						"error": "Token 已过期，请重新登录",
					})
				case ve.Errors&jwt.ValidationErrorNotValidYet != 0:
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
						"code":  http.StatusUnauthorized,
						"error": "Token 未生效",
					})
				default:
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
						"code":  http.StatusUnauthorized,
						"error": "Token 无效",
					})
				}
			} else {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"code":  http.StatusInternalServerError,
					"error": "服务器内部错误",
				})
			}
			return
		}
		if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
			ID := claims.ID
			if ID == "" {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"code":  http.StatusUnauthorized,
					"error": "Token 中无用户信息",
				})
				return
			}
			c.Set("id", ID)
			c.Next()
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":  http.StatusUnauthorized,
				"error": "Token 验证失败",
			})
		}
	}
}
