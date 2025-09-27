package adminstatsapi

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// 解析 YYYY-MM-DD（仅日期）。返回 (t, ok)
func parseDate(s string) (time.Time, bool) {
	if s == "" {
		return time.Time{}, false
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Time{}, false
	}
	// 视为 UTC 时间的当天零点
	return t.UTC(), true
}

func (h *Handler) Get(c *gin.Context) {
	// 可选参数 ?from=YYYY-MM-DD&to=YYYY-MM-DD (to 日期为包含关系)
	var (
		fromStr = c.Query("from")
		toStr   = c.Query("to")
	)

	var from, to time.Time
	okFrom := false
	okTo := false
	if t, ok := parseDate(fromStr); ok {
		from = t
		okFrom = true
	}
	if t, ok := parseDate(toStr); ok {
		// 将 'to' 变为不包含当天，通过增加24小时使其变为次日零点
		to = t.Add(24 * time.Hour)
		okTo = true
	}

	now := time.Now().UTC()
	if !okTo {
		to = now
	}
	if !okFrom {
		from = to.AddDate(0, 0, -30) // 默认为最近 30 天
	}

	resp, err := h.svc.Stats(from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "统计失败", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}