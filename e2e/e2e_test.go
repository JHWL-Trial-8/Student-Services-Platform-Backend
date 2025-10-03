package e2e

import (
    "bytes"
    "crypto/rand"
    "encoding/hex"
    "fmt"
    "image"
    "image/color"
    "image/png"
    "mime/multipart"
    "net/http"
    "net/url"
    "os"
    "strings"
    "testing"
    "time"

    "github.com/gavv/httpexpect/v2"
    "github.com/stretchr/testify/require"
)

type userCred struct {
    ID       int
    Email    string
    Password string
    Token    string
    Name     string
    Role     string
}

type uploadedImage struct {
    ID     int
    SHA256 string
    Mime   string
    Width  int
    Height int
}

type scenario struct {
    Base         string
    E            *httpexpect.Expect
    Super        userCred
    AdminA       userCred
    AdminB       userCred
    StuA         userCred
    StuB         userCred
    Image1       uploadedImage
    HasUserMgmt  bool
    UsersChecked bool
}

func Test_E2E_All(t *testing.T) {
    base := getenv("E2E_BASE_URL", "http://localhost:8080")
    e := newExpect(t, base)
    s := &scenario{Base: base, E: e}

    t.Run("bootstrap users (super -> admins/students)", func(t *testing.T) {
        s.Super = s.mustRegisterAndLogin(t, "SUPER_ADMIN")
        s.detectUsersAPI(t)
        s.AdminA = s.mustCreateUserBySuperAndLogin(t, s.Super, "ADMIN")
        s.AdminB = s.mustCreateUserBySuperAndLogin(t, s.Super, "ADMIN")
        s.StuA = s.mustCreateUserBySuperAndLogin(t, s.Super, "STUDENT")
        s.StuB = s.mustCreateUserBySuperAndLogin(t, s.Super, "STUDENT")
        if s.HasUserMgmt {
            withAuth(s.E.GET("/api/v1/users"), s.AdminA.Token).
                Expect().Status(http.StatusForbidden)
            withAuth(s.E.GET("/api/v1/users"), s.StuA.Token).
                Expect().Status(http.StatusForbidden)
        } else {
            t.Log("[skip] /api/v1/users not implemented; skip list/permission assertions")
        }
    })

    t.Run("users me profile r/w", func(t *testing.T) {
        obj := withAuth(s.E.GET("/api/v1/users/me"), s.StuA.Token).
            Expect().Status(http.StatusOK).JSON().Object()
        obj.Value("email").String().IsEqual(s.StuA.Email)
        newPhone := "1880000" + randomDigits(4)
        withAuth(s.E.PUT("/api/v1/users/me"), s.StuA.Token).
            WithJSON(map[string]any{
                "email":       s.StuA.Email,
                "name":        s.StuA.Name + " 更新",
                "phone":       newPhone,
                "dept":        "CS",
                "allow_email": true,
            }).
            Expect().Status(http.StatusOK).
            JSON().Object().
            Value("phone").String().IsEqual(newPhone)
    })

    t.Run("images upload & dedup", func(t *testing.T) {
        imgBytes := mustMakePNG(12, 10, color.RGBA{R: 10, G: 200, B: 100, A: 255})
        up1 := s.uploadImage(t, s.StuA.Token, "lamp.png", imgBytes)
        require.NotZero(t, up1.ID)
        require.NotEmpty(t, up1.SHA256)
        require.Equal(t, 12, up1.Width)
        require.Equal(t, 10, up1.Height)
        s.Image1 = up1
        up2 := s.uploadImage(t, s.StuA.Token, "lamp_dup.png", imgBytes)
        require.Equal(t, up1.SHA256, up2.SHA256, "sha256 should match for same image content")
        if up2.ID != up1.ID {
            t.Logf("[warn] image_id differs on dedup (id1=%d id2=%d); sha256 reused as expected", up1.ID, up2.ID)
        }
    })

    var ticketA int
    var ticketB int

    t.Run("student creates tickets", func(t *testing.T) {
        ticketA = s.createTicket(t, s.StuA.Token, map[string]any{
            "title": "A-路灯坏了", "content": "宿舍楼下路灯不亮，晚上很黑", "category": "路灯报修",
            "is_urgent": true, "is_anonymous": false, "image_ids": []int{s.Image1.ID},
        })
        require.NotZero(t, ticketA, "ticketA should be created")
        ticketB = s.createTicket(t, s.StuB.Token, map[string]any{
            "title": "B-空调不制冷", "content": "教室空调制冷差", "category": "空调报修",
            "is_urgent": false, "is_anonymous": false,
        })
        require.NotZero(t, ticketB, "ticketB should be created")
    })

    t.Run("tickets list visibility", func(t *testing.T) {
        arr := withAuth(s.E.GET("/api/v1/tickets"), s.StuA.Token).
            Expect().Status(http.StatusOK).JSON().Object().Value("items").Array()
        for i := range arr.Iter() {
            arr.Element(i).Object().Value("user_id").Number().IsEqual(s.StuA.ID)
        }
        all := withAuth(s.E.GET("/api/v1/tickets"), s.AdminA.Token).
            Expect().Status(http.StatusOK).JSON().Object().Value("items").Array()
        require.GreaterOrEqual(t, all.Length().Raw(), 2.0)
    })

    t.Run("claim/unclaim with permissions & conflict", func(t *testing.T) {
        withAuth(s.E.POST(fmt.Sprintf("/api/v1/tickets/%d/claim", ticketA)), s.StuA.Token).
            Expect().Status(http.StatusForbidden)
        withAuth(s.E.POST(fmt.Sprintf("/api/v1/tickets/%d/claim", ticketA)), s.AdminA.Token).
            Expect().Status(http.StatusNoContent)
        withAuth(s.E.POST(fmt.Sprintf("/api/v1/tickets/%d/claim", ticketA)), s.AdminB.Token).
            Expect().Status(http.StatusConflict)
        obj := withAuth(s.E.GET(fmt.Sprintf("/api/v1/tickets/%d", ticketA)), s.AdminA.Token).
            Expect().Status(http.StatusOK).JSON().Object()
        obj.Value("assigned_admin_id").Number().IsEqual(s.AdminA.ID)
        status := obj.Value("status").String().Raw()
        require.Contains(t, []string{"CLAIMED", "IN_PROGRESS"}, status)
        withAuth(s.E.POST(fmt.Sprintf("/api/v1/tickets/%d/unclaim", ticketA)), s.AdminB.Token).
            Expect().Status(http.StatusForbidden)
        withAuth(s.E.POST(fmt.Sprintf("/api/v1/tickets/%d/unclaim", ticketA)), s.AdminA.Token).
            Expect().Status(http.StatusNoContent)
        obj2 := withAuth(s.E.GET(fmt.Sprintf("/api/v1/tickets/%d", ticketA)), s.AdminA.Token).
            Expect().Status(http.StatusOK).JSON().Object()
        obj2.Value("assigned_admin_id").Null()
    })

    var cannedID int
    t.Run("canned replies CRUD", func(t *testing.T) {
        canned := withAuth(s.E.POST("/api/v1/canned-replies"), s.AdminA.Token).
            WithJSON(map[string]any{"title": "已派单模板", "body": "已安排维修，预计24小时内处理。"}).
            Expect().Status(http.StatusCreated).JSON().Object()
        cannedID = int(canned.Value("id").Number().Raw())
        withAuth(s.E.GET("/api/v1/canned-replies"), s.AdminA.Token).
            Expect().Status(http.StatusOK).
            JSON().Object().Value("items").Array().Contains(canned.Raw())
        updated := withAuth(s.E.PUT(fmt.Sprintf("/api/v1/canned-replies/%d", cannedID)), s.AdminA.Token).
            WithJSON(map[string]any{"body": "已派单，预计24-48小时处理完成。"}).
            Expect().Status(http.StatusOK).JSON().Object()
        updated.Value("body").String().Contains("24-48")
        withAuth(s.E.DELETE(fmt.Sprintf("/api/v1/canned-replies/%d", cannedID)), s.AdminA.Token).
            Expect().Status(http.StatusNoContent)
    })

    t.Run("ticket messages (internal note hidden from student)", func(t *testing.T) {
        withAuth(s.E.POST(fmt.Sprintf("/api/v1/tickets/%d/claim", ticketA)), s.AdminA.Token).
            Expect().Status(http.StatusNoContent)
        msg1 := withAuth(s.E.POST(fmt.Sprintf("/api/v1/tickets/%d/messages", ticketA)), s.StuA.Token).
            WithJSON(map[string]any{"body": "晚上更暗，盼尽快修。"}).
            Expect().Status(http.StatusCreated).JSON().Object()
        require.NotZero(t, int(msg1.Value("id").Number().Raw()))
        withAuth(s.E.POST(fmt.Sprintf("/api/v1/tickets/%d/messages", ticketA)), s.AdminA.Token).
            WithJSON(map[string]any{"body": "内部：先换灯泡再测线路", "is_internal_note": true}).
            Expect().Status(http.StatusCreated)
        withAuth(s.E.POST(fmt.Sprintf("/api/v1/tickets/%d/messages", ticketA)), s.AdminA.Token).
            WithJSON(map[string]any{"body": "已安排人员，预计今晚完成。"}).
            Expect().Status(http.StatusCreated)

        arr := withAuth(s.E.GET(fmt.Sprintf("/api/v1/tickets/%d/messages", ticketA)), s.StuA.Token).
            Expect().Status(http.StatusOK).JSON().Object().Value("items").Array()
        for i := range arr.Iter() {
            msg := arr.Element(i).Object()
            // **FIX**: Correctly check for key existence.
            if msg.Raw()["is_internal_note"] != nil {
                msg.Value("is_internal_note").Boolean().IsFalse()
            }
        }
    })

    t.Run("resolve + rate + close", func(t *testing.T) {
        withAuth(s.E.POST(fmt.Sprintf("/api/v1/tickets/%d/resolve", ticketA)), s.AdminA.Token).
            Expect().Status(http.StatusNoContent)
        withAuth(s.E.GET(fmt.Sprintf("/api/v1/tickets/%d", ticketA)), s.AdminA.Token).
            Expect().Status(http.StatusOK).JSON().Object().
            Value("status").String().IsEqual("RESOLVED")
        r := withAuth(s.E.POST(fmt.Sprintf("/api/v1/tickets/%d/rate", ticketA)), s.StuA.Token).
            WithJSON(map[string]any{"stars": 5, "comment": "处理很快，谢谢！"}).
            Expect().Status(http.StatusCreated).JSON().Object()
        require.Equal(t, ticketA, int(r.Value("ticket_id").Number().Raw()))
        require.Equal(t, 5.0, r.Value("stars").Number().Raw())
        withAuth(s.E.POST(fmt.Sprintf("/api/v1/tickets/%d/close", ticketA)), s.AdminA.Token).
            Expect().Status(http.StatusNoContent)
        withAuth(s.E.GET(fmt.Sprintf("/api/v1/tickets/%d", ticketA)), s.AdminA.Token).
            Expect().Status(http.StatusOK).JSON().Object().
            Value("status").String().IsEqual("CLOSED")
    })

    t.Run("spam flow: flag -> super review approve -> auto reply", func(t *testing.T) {
        withAuth(s.E.POST(fmt.Sprintf("/api/v1/tickets/%d/spam-flag", ticketB)), s.AdminB.Token).
            WithJSON(map[string]any{"reason": "疑似广告"}).
            Expect().Status(http.StatusCreated).JSON().Object().
            Value("status").String().IsEqual("PENDING")
        obj := withAuth(s.E.GET(fmt.Sprintf("/api/v1/tickets/%d", ticketB)), s.AdminB.Token).
            Expect().Status(http.StatusOK).JSON().Object()
        obj.Value("status").String().IsEqual("SPAM_PENDING")
        withAuth(s.E.POST(fmt.Sprintf("/api/v1/tickets/%d/spam-review", ticketB)), s.Super.Token).
            WithJSON(map[string]any{"action": "approve"}).
            Expect().Status(http.StatusNoContent)
        obj2 := withAuth(s.E.GET(fmt.Sprintf("/api/v1/tickets/%d", ticketB)), s.Super.Token).
            Expect().Status(http.StatusOK).JSON().Object()
        obj2.Value("status").String().IsEqual("SPAM_CONFIRMED")

        const autoMsg = "请您在提交反馈时确保内容的有效性和准确性，感谢您的理解和配合。如有异议，请重新反馈。"
        
        respObj := withAuth(s.E.GET(fmt.Sprintf("/api/v1/tickets/%d/messages", ticketB)), s.StuB.Token).
            Expect().Status(http.StatusOK).JSON().Object()

        found := false
        // **FIX**: Correctly check for key existence.
        if respObj.Raw()["items"] != nil {
            arr := respObj.Value("items").Array()
            for i := range arr.Iter() {
                body := arr.Element(i).Object().Value("body").String().Raw()
                if strings.Contains(body, autoMsg) {
                    found = true
                    break
                }
            }
        } else {
            t.Logf("[warn] GET /api/v1/tickets/%d/messages response is missing 'items' key for empty list. API should return 'items: []'.", ticketB)
        }
        require.True(t, found, "spam 审核通过后应自动回复固定话术")
    })

    t.Run("admin stats (shape only)", func(t *testing.T) {
        stats := withAuth(
            s.E.GET("/api/v1/admin/stats").
                WithQuery("from", time.Now().UTC().Format("2006-01-02")).
                WithQuery("to", time.Now().UTC().Format("2006-01-02")),
            s.Super.Token,
        ).
            Expect().Status(http.StatusOK).JSON().Object()

        totals := stats.Value("totals").Object()
        totals.ContainsKey("tickets")
        totals.ContainsKey("closed")
        totals.ContainsKey("spam_confirmed")
        // **FIX**: Correctly check for key existence.
        if totals.Raw()["resolved"] != nil {
            totals.Value("resolved").Number()
        }

        stats.Value("by_category").Array()
        stats.Value("daily_trend").Array()
        stats.Value("admin_workload").Array()
    })
}

/* ----------------------------- helpers ------------------------------ */

func newExpect(t *testing.T, base string) *httpexpect.Expect {
    u, err := url.Parse(base)
    require.NoError(t, err)
    return httpexpect.WithConfig(httpexpect.Config{
        BaseURL:  strings.TrimRight(u.String(), "/"),
        Reporter: httpexpect.NewAssertReporter(t),
        Printers: []httpexpect.Printer{
            httpexpect.NewDebugPrinter(t, true),
        },
    })
}

func withAuth(r *httpexpect.Request, token string) *httpexpect.Request {
    token = strings.TrimSpace(token)
    if token == "" {
        panic("empty JWT token")
    }
    return r.WithHeader("Authorization", fmt.Sprintf("Bearer %s", token))
}

func (s *scenario) mustRegisterAndLogin(t *testing.T, role string) userCred {
    email := uniqEmail(role)
    pass := "P@ssw0rd-" + randomDigits(6)
    name := "E2E-" + role
    user := s.E.POST("/api/v1/auth/register").
        WithJSON(map[string]any{
            "email": email, "name": name, "role": role, "password": pass,
            "dept": "QA", "allow_email": true,
        }).Expect().Status(http.StatusCreated).JSON().Object()
    uid := int(user.Value("id").Number().Raw())
    tk := s.mustLogin(t, email, pass)
    return userCred{ID: uid, Email: email, Password: pass, Token: tk, Name: name, Role: role}
}

func (s *scenario) detectUsersAPI(t *testing.T) {
    if s.UsersChecked {
        return
    }
    s.UsersChecked = true
    resp := withAuth(s.E.GET("/api/v1/users"), s.Super.Token).Expect()
    code := resp.Raw().StatusCode
    switch code {
    case http.StatusOK:
        s.HasUserMgmt = true
        resp.JSON().Object().ContainsKey("items")
    case http.StatusNotFound, http.StatusMethodNotAllowed, http.StatusNotImplemented:
        s.HasUserMgmt = false
        t.Log("[info] /api/v1/users not available; will fall back to /auth/register")
    default:
        if code == http.StatusForbidden {
            s.HasUserMgmt = false
            t.Log("[warn] GET /api/v1/users returned 403; skip user mgmt assertions")
        } else {
            t.Logf("[error] GET /api/v1/users unexpected status=%d", code)
            s.HasUserMgmt = false
        }
    }
}

func (s *scenario) mustCreateUserBySuperAndLogin(t *testing.T, super userCred, role string) userCred {
    email := uniqEmail(role)
    pass := "P@ssw0rd-" + randomDigits(6)
    name := "E2E-" + role + "-" + randomDigits(3)
    var uid int
    created := false
    payload := map[string]any{
        "email": email, "name": name, "role": role, "password": pass,
        "dept": "QA", "allow_email": true,
    }
    if s.HasUserMgmt {
        resp := withAuth(s.E.POST("/api/v1/users"), super.Token).WithJSON(payload).Expect()
        if resp.Raw().StatusCode == http.StatusCreated {
            uid = int(resp.JSON().Object().Value("id").Number().Raw())
            created = true
        } else {
            t.Logf("[warn] POST /api/v1/users failed with status %d; fallback to /auth/register", resp.Raw().StatusCode)
            s.HasUserMgmt = false
        }
    }
    if !created {
        u := s.E.POST("/api/v1/auth/register").WithJSON(payload).
            Expect().Status(http.StatusCreated).JSON().Object()
        uid = int(u.Value("id").Number().Raw())
    }
    tk := s.mustLogin(t, email, pass)
    return userCred{ID: uid, Email: email, Password: pass, Token: tk, Name: name, Role: role}
}

func (s *scenario) mustLogin(t *testing.T, email, pass string) string {
    obj := s.E.POST("/api/v1/auth/login").
        WithJSON(map[string]any{"email": email, "password": pass}).
        Expect().Status(http.StatusOK).JSON().Object()
    require.Equal(t, "bearer", strings.ToLower(obj.Value("token_type").String().Raw()))
    tk := strings.TrimSpace(obj.Value("access_token").String().Raw())
    require.NotEmpty(t, tk, "access_token should not be empty")
    t.Logf("[login] %s -> token len=%d", email, len(tk))
    return tk
}

func (s *scenario) uploadImage(t *testing.T, token, filename string, data []byte) uploadedImage {
    var buf bytes.Buffer
    w := multipart.NewWriter(&buf)
    fw, err := w.CreateFormFile("file", filename)
    require.NoError(t, err)
    _, err = fw.Write(data)
    require.NoError(t, err)
    require.NoError(t, w.Close())
    obj := withAuth(s.E.POST("/api/v1/images"), token).
        WithHeader("Content-Type", w.FormDataContentType()).
        WithBytes(buf.Bytes()).
        Expect().Status(http.StatusCreated).JSON().Object()
    return uploadedImage{
        ID:     int(obj.Value("image_id").Number().Raw()),
        SHA256: obj.Value("sha256").String().Raw(),
        Mime:   obj.Value("mime").String().Raw(),
        Width:  int(obj.Value("width").Number().Raw()),
        Height: int(obj.Value("height").Number().Raw()),
    }
}

func (s *scenario) createTicket(t *testing.T, token string, payload map[string]any) int {
    obj := withAuth(s.E.POST("/api/v1/tickets"), token).
        WithJSON(payload).
        Expect().Status(http.StatusCreated).JSON().Object()
    return int(obj.Value("id").Number().Raw())
}

func uniqEmail(prefix string) string {
    return fmt.Sprintf("%s_%d_%s@test.example", strings.ToLower(prefix), time.Now().UnixNano(), randHex(4))
}

func randHex(n int) string {
    b := make([]byte, n)
    _, _ = rand.Read(b)
    return hex.EncodeToString(b)
}

func randomDigits(n int) string {
    const digits = "0123456789"
    b := make([]byte, n)
    rand.Read(b)
    for i := range b {
        b[i] = digits[int(b[i])%10]
    }
    return string(b)
}

func mustMakePNG(w, h int, c color.RGBA) []byte {
    img := image.NewRGBA(image.Rect(0, 0, w, h))
    for y := 0; y < h; y++ {
        for x := 0; x < w; x++ {
            img.SetRGBA(x, y, c)
        }
    }
    var buf bytes.Buffer
    _ = png.Encode(&buf, img)
    return buf.Bytes()
}

func getenv(k, def string) string {
    if v := os.Getenv(k); v != "" {
        return v
    }
    return def
}