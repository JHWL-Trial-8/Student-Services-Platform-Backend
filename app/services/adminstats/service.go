package adminstats

import (
	"fmt"
	"sync"
	"time"

	dbpkg "student-services-platform-backend/internal/db"
	"student-services-platform-backend/internal/openapi"

	"gorm.io/gorm"
	"strings"
)

type Service struct {
	db    *gorm.DB
	mu    sync.Mutex
	cache map[string]cacheEntry
}

type cacheEntry struct {
	resp    openapi.AdminStatsGet200Response
	expires time.Time
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db, cache: make(map[string]cacheEntry)}
}

func keyFor(from, to time.Time) string {
	return from.UTC().Format("2006-01-02T15") + "|" + to.UTC().Format("2006-01-02T15")
}

// Stats 在 UTC 时间下聚合 [from, to) 区间内的统计数据。
func (s *Service) Stats(from, to time.Time) (*openapi.AdminStatsGet200Response, error) {
	if to.Before(from) {
		tmp := from
		from = to
		to = tmp
	}

	// 小型内存缓存（TTL 30秒）
	cacheKey := keyFor(from, to)
	now := time.Now().UTC()

	s.mu.Lock()
	if ce, ok := s.cache[cacheKey]; ok && now.Before(ce.expires) {
		out := ce.resp // copy
		s.mu.Unlock()
		return &out, nil
	}
	s.mu.Unlock()

	var resp openapi.AdminStatsGet200Response

	// ---- 总计 ----
	type totalsRow struct {
		Tickets       int64
		Resolved      int64
		Closed        int64
		SpamConfirmed int64
	}
	var tr totalsRow
	if err := s.db.Raw(
		`SELECT
   COUNT(*) AS tickets,
   SUM(CASE WHEN status = ? THEN 1 ELSE 0 END) AS resolved,
   SUM(CASE WHEN status = ? THEN 1 ELSE 0 END) AS closed,
   SUM(CASE WHEN status = ? THEN 1 ELSE 0 END) AS spam_confirmed
  FROM tickets
  WHERE created_at >= ? AND created_at < ?`,
		dbpkg.TicketStatusResolved, dbpkg.TicketStatusClosed, dbpkg.TicketStatusSpamConfirmed, from, to,
	).Scan(&tr).Error; err != nil {
		return nil, err
	}
	resp.Totals = openapi.AdminStatsGet200ResponseTotals{
		Tickets:       int32(tr.Tickets),
		Resolved:      int32(tr.Resolved),
		Closed:        int32(tr.Closed),
		SpamConfirmed: int32(tr.SpamConfirmed),
	}

	// ---- 按分类统计 ----
	type catRow struct {
		Category string
		Count    int64
	}
	var cats []catRow
	if err := s.db.Raw(
		`SELECT category, COUNT(*) AS count
   FROM tickets
  WHERE created_at >= ? AND created_at < ?
  GROUP BY category
  ORDER BY count DESC`,
		from, to,
	).Scan(&cats).Error; err != nil {
		return nil, err
	}
	resp.ByCategory = make([]openapi.AdminStatsGet200ResponseByCategoryInner, 0, len(cats))
	for _, r := range cats {
		resp.ByCategory = append(resp.ByCategory, openapi.AdminStatsGet200ResponseByCategoryInner{
			Category: r.Category,
			Count:    int32(r.Count),
		})
	}

	// ---- 每日趋势 ----
	var trendSQL string
	switch s.db.Dialector.Name() {
	case "postgres":
		// postgres 写法
		// to_char((created_at at time zone 'UTC')::date, 'YYYY-MM-DD')
		trendSQL = `
SELECT to_char((created_at AT TIME ZONE 'UTC')::date, 'YYYY-MM-DD') AS date, COUNT(*) AS count
   FROM tickets
  WHERE created_at >= ? AND created_at < ?
  GROUP BY 1
  ORDER BY 1`
	case "sqlite":
		// sqlite 写法
		trendSQL = `
SELECT strftime('%Y-%m-%d', created_at) AS date, COUNT(*) AS count
   FROM tickets
  WHERE created_at >= ? AND created_at < ?
  GROUP BY 1
  ORDER BY 1`
	default:
		// mysql 及其他支持 DATE() 函数的数据库
		trendSQL = `
SELECT DATE(created_at) AS date, COUNT(*) AS count
   FROM tickets
  WHERE created_at >= ? AND created_at < ?
  GROUP BY 1
  ORDER BY 1`
	}
	type trendRow struct {
		Date  string
		Count int64
	}
	var trs []trendRow
	if err := s.db.Raw(trendSQL, from, to).Scan(&trs).Error; err != nil {
		return nil, err
	}
	resp.DailyTrend = make([]openapi.AdminStatsGet200ResponseDailyTrendInner, 0, len(trs))
	for _, r := range trs {
		resp.DailyTrend = append(resp.DailyTrend, openapi.AdminStatsGet200ResponseDailyTrendInner{
			Date:  r.Date,
			Count: int32(r.Count),
		})
	}

	// ---- 管理员工作量 ----
	type wlRow struct {
		AdminID        int32
		Name           string
		TicketsHandled int64
	}
	var wls []wlRow
	wlSQL := `
  SELECT u.id AS admin_id, COALESCE(u.name, '') AS name, COUNT(*) AS tickets_handled
    FROM audit_logs al
    JOIN users u ON u.id = al.actor_user_id
   WHERE al.action IN ('ticket.resolve','ticket.close')
     AND al.created_at >= ? AND al.created_at < ?
   GROUP BY u.id, u.name
   ORDER BY tickets_handled DESC`
	if err := s.db.Raw(wlSQL, from, to).Scan(&wls).Error; err != nil {
		return nil, err
	}
	resp.AdminWorkload = make([]openapi.AdminStatsGet200ResponseAdminWorkloadInner, 0, len(wls))
	for _, r := range wls {
		name := r.Name
		if strings.TrimSpace(name) == "" {
			name = fmt.Sprintf("User#%d", r.AdminID)
		}
		resp.AdminWorkload = append(resp.AdminWorkload, openapi.AdminStatsGet200ResponseAdminWorkloadInner{
			AdminId:        r.AdminID,
			Name:           name,
			TicketsHandled: int32(r.TicketsHandled),
		})
	}

	// 存入缓存
	s.mu.Lock()
	s.cache[cacheKey] = cacheEntry{
		resp:    resp,
		expires: now.Add(30 * time.Second),
	}
	s.mu.Unlock()

	return &resp, nil
}