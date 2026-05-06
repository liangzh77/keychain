package server

import (
	"html/template"
	"net/http"

	"github.com/liangzh77/keychain/internal/admin"
	"github.com/liangzh77/keychain/internal/auth"
)

var loginPageTemplate = template.Must(template.New("login").Parse(`<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Keychain 登录</title>
  <style>
    body { margin: 0; min-height: 100vh; display: grid; place-items: center; font-family: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; background: #f6f7f9; color: #17202a; }
    main { width: min(420px, calc(100vw - 32px)); background: #fff; border: 1px solid #dfe4ea; border-radius: 8px; padding: 28px; box-shadow: 0 16px 48px rgba(23,32,42,.08); }
    h1 { margin: 0 0 8px; font-size: 24px; }
    p { margin: 0 0 24px; color: #657080; }
    form { display: grid; gap: 14px; }
    label { display: grid; gap: 6px; font-weight: 600; }
    input { width: 100%; box-sizing: border-box; padding: 10px 12px; border: 1px solid #cfd6df; border-radius: 6px; font: inherit; }
    button { padding: 11px 14px; border: 0; border-radius: 6px; background: #1f6feb; color: white; font: inherit; font-weight: 700; cursor: pointer; }
    .error { margin-bottom: 16px; padding: 10px 12px; border-radius: 6px; background: #fff1f0; color: #a8071a; }
  </style>
</head>
<body>
  <main>
    <h1>Keychain</h1>
    <p>管理员登录</p>
    {{if .Error}}<div class="error">{{.Error}}</div>{{end}}
    <form method="post" action="/login">
      <label>账号<input name="username" autocomplete="username" value="admin"></label>
      <label>密码<input type="password" name="password" autocomplete="current-password" autofocus></label>
      <button type="submit">登录</button>
    </form>
  </main>
</body>
</html>`))

var adminPageTemplate = template.Must(template.New("admin").Parse(`<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Keychain 后台</title>
  <style>
    body { margin: 0; font-family: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; background: #f6f7f9; color: #17202a; }
    header { display: flex; align-items: center; justify-content: space-between; padding: 16px 24px; background: #fff; border-bottom: 1px solid #dfe4ea; position: sticky; top: 0; z-index: 10; }
    main { max-width: 1180px; margin: 0 auto; padding: 32px 24px; }
    h1 { margin: 0 0 8px; font-size: 28px; }
    h2 { margin: 0 0 14px; font-size: 18px; }
    h3 { margin: 0 0 8px; font-size: 16px; }
    .grid { display: grid; grid-template-columns: minmax(280px, 360px) 1fr; gap: 16px; align-items: start; margin-top: 24px; }
    .stack { display: grid; gap: 16px; }
    .card { background: #fff; border: 1px solid #dfe4ea; border-radius: 8px; padding: 18px; }
    .muted { color: #657080; }
    form { display: grid; gap: 10px; }
    label { display: grid; gap: 5px; font-weight: 600; }
    input, select { width: 100%; box-sizing: border-box; padding: 9px 10px; border: 1px solid #cfd6df; border-radius: 6px; font: inherit; }
    button { padding: 9px 12px; border: 0; border-radius: 6px; background: #1f6feb; color: white; cursor: pointer; font-weight: 700; }
    .secondary { background: #46515f; }
    .provider { display: grid; gap: 14px; }
    .provider-head { display: flex; gap: 12px; justify-content: space-between; align-items: start; }
    .tag { display: inline-flex; padding: 2px 8px; border-radius: 999px; background: #edf3ff; color: #1f4f9a; font-size: 12px; }
    table { width: 100%; border-collapse: collapse; font-size: 14px; }
    th, td { padding: 8px; text-align: left; border-bottom: 1px solid #edf0f4; vertical-align: top; }
    th { color: #657080; font-weight: 600; }
    .inline-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 10px; }
    .notice { margin-top: 16px; padding: 10px 12px; border-radius: 6px; background: #fff7e6; color: #8a5a00; }
    @media (max-width: 900px) { .grid, .inline-grid { grid-template-columns: 1fr; } }
  </style>
</head>
<body>
  <header>
    <strong>Keychain</strong>
    <form method="post" action="/logout"><button type="submit">退出</button></form>
  </header>
  <main>
    <h1>后台控制台</h1>
    <p class="muted">已登录为 {{.Username}}。先从左侧添加 provider，再在右侧给对应 provider 添加 models 和 keys。</p>
    {{if .Error}}<div class="notice">{{.Error}}</div>{{end}}
    <section class="grid">
      <div class="card">
        <h2>添加 Provider</h2>
        <form method="post" action="/admin/providers">
          <label>名称<input name="name" placeholder="OpenAI" required></label>
          <label>代码<input name="code" placeholder="openai" required></label>
          <label>Key 分发策略
            <select name="rotationStrategy">
              <option value="ROUND_ROBIN">轮询分发</option>
              <option value="STICKY_FIRST_AVAILABLE">优先第一个可用 key</option>
            </select>
          </label>
          <label><input type="checkbox" name="isEnabled" value="1" checked> 启用 provider</label>
          <button type="submit">添加 provider</button>
        </form>
      </div>
      <div class="stack">
        {{if not .Providers}}
          <div class="card"><h2>还没有 provider</h2><p class="muted">先添加一个 provider，例如 OpenAI、DeepSeek 或 Gemini。</p></div>
        {{end}}
        {{range .Providers}}
          <div class="card provider">
            <div class="provider-head">
              <div>
                <h2>{{.Provider.Name}}</h2>
                <div class="muted">{{.Provider.Code}} · {{.Provider.RotationStrategy}}</div>
              </div>
              {{if .Provider.IsEnabled}}<span class="tag">启用</span>{{else}}<span class="tag">停用</span>{{end}}
            </div>
            <div class="inline-grid">
              <form method="post" action="/admin/models">
                <h3>添加 Model</h3>
                <input type="hidden" name="providerId" value="{{.Provider.ID}}">
                <label>名称<input name="name" placeholder="GPT 4.1" required></label>
                <label>代码<input name="code" placeholder="gpt-4.1" required></label>
                <label><input type="checkbox" name="isEnabled" value="1" checked> 启用 model</label>
                <button class="secondary" type="submit">添加 model</button>
              </form>
              <form method="post" action="/admin/keys">
                <h3>添加 Key</h3>
                <input type="hidden" name="providerId" value="{{.Provider.ID}}">
                <label>别名<input name="alias" placeholder="openai-main-01" required></label>
                <label>Key 明文<input name="secretValue" placeholder="sk-..." required></label>
                <label>排序<input name="sortOrder" type="number" value="0"></label>
                <label><input type="checkbox" name="isEnabled" value="1" checked> 启用 key</label>
                <label><input type="checkbox" name="isAvailable" value="1" checked> 当前可用</label>
                <button class="secondary" type="submit">添加 key</button>
              </form>
            </div>
            <div>
              <h3>Models</h3>
              {{if .Models}}
                <table><thead><tr><th>名称</th><th>代码</th><th>状态</th></tr></thead><tbody>
                  {{range .Models}}<tr><td>{{.Name}}</td><td>{{.Code}}</td><td>{{if .IsEnabled}}启用{{else}}停用{{end}}</td></tr>{{end}}
                </tbody></table>
              {{else}}<p class="muted">还没有 model。</p>{{end}}
            </div>
            <div>
              <h3>Keys</h3>
              {{if .Keys}}
                <table><thead><tr><th>别名</th><th>掩码</th><th>顺序</th><th>状态</th></tr></thead><tbody>
                  {{range .Keys}}<tr><td>{{.Alias}}</td><td>{{.MaskedValue}}</td><td>{{.SortOrder}}</td><td>{{if .IsEnabled}}启用{{else}}停用{{end}} / {{if .IsAvailable}}可用{{else}}不可用{{end}}</td></tr>{{end}}
                </tbody></table>
              {{else}}<p class="muted">还没有 key。</p>{{end}}
            </div>
          </div>
        {{end}}
      </div>
    </section>
  </main>
</body>
</html>`))

type adminPageData struct {
	Username  string
	Error     string
	Providers []providerPanel
}

type providerPanel struct {
	Provider admin.Provider
	Models   []admin.Model
	Keys     []admin.APIKey
}

func registerPageRoutes(mux *http.ServeMux, authService *auth.Service, store *admin.Store) {
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/admin", http.StatusFound)
	})
	mux.HandleFunc("GET /login", showLoginPage)
	mux.HandleFunc("POST /login", formLoginHandler(authService))
	mux.HandleFunc("POST /logout", formLogoutHandler(authService))
	mux.HandleFunc("GET /admin", adminPageHandler(authService, store))
	if store != nil {
		mux.HandleFunc("POST /admin/providers", formCreateProviderHandler(store))
		mux.HandleFunc("POST /admin/models", formCreateModelHandler(store))
		mux.HandleFunc("POST /admin/keys", formCreateKeyHandler(store))
	}
}

func showLoginPage(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = loginPageTemplate.Execute(w, map[string]string{})
}

func formLoginHandler(authService *auth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			renderLoginError(w, "表单格式不正确")
			return
		}
		if !authService.Authenticate(r.FormValue("username"), r.FormValue("password")) {
			renderLoginError(w, "账号或密码错误")
			return
		}
		token, session, err := authService.CreateSession(r.Context())
		if err != nil {
			renderLoginError(w, "创建登录会话失败")
			return
		}
		setSessionCookie(w, r, token, session.ExpiresAt)
		http.Redirect(w, r, "/admin", http.StatusFound)
	}
}

func formLogoutHandler(authService *auth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if cookie, err := r.Cookie(auth.CookieName); err == nil {
			_ = authService.DeleteSession(r.Context(), cookie.Value)
		}
		clearSessionCookie(w, r)
		http.Redirect(w, r, "/login", http.StatusFound)
	}
}

func adminPageHandler(authService *auth.Service, store *admin.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(auth.CookieName)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		session, ok, err := authService.GetSession(r.Context(), cookie.Value)
		if err != nil || !ok {
			clearSessionCookie(w, r)
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		data := adminPageData{Username: session.Username, Error: r.URL.Query().Get("error")}
		if store != nil {
			data.Providers, err = loadProviderPanels(r, store)
			if err != nil {
				data.Error = "加载 provider 数据失败"
			}
		}
		_ = adminPageTemplate.Execute(w, data)
	}
}

func renderLoginError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusUnauthorized)
	_ = loginPageTemplate.Execute(w, map[string]string{"Error": message})
}

func formCreateProviderHandler(store *admin.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			redirectAdminError(w, r, "provider 表单格式不正确")
			return
		}
		_, err := store.CreateProvider(r.Context(), admin.CreateProviderInput{
			Name:             r.FormValue("name"),
			Code:             r.FormValue("code"),
			IsEnabled:        r.FormValue("isEnabled") == "1",
			RotationStrategy: r.FormValue("rotationStrategy"),
		})
		if err != nil {
			redirectAdminError(w, r, "添加 provider 失败："+err.Error())
			return
		}
		http.Redirect(w, r, "/admin", http.StatusFound)
	}
}

func formCreateModelHandler(store *admin.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			redirectAdminError(w, r, "model 表单格式不正确")
			return
		}
		_, err := store.CreateModel(r.Context(), admin.CreateModelInput{
			ProviderID: r.FormValue("providerId"),
			Name:       r.FormValue("name"),
			Code:       r.FormValue("code"),
			IsEnabled:  r.FormValue("isEnabled") == "1",
		})
		if err != nil {
			redirectAdminError(w, r, "添加 model 失败："+err.Error())
			return
		}
		http.Redirect(w, r, "/admin", http.StatusFound)
	}
}

func formCreateKeyHandler(store *admin.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			redirectAdminError(w, r, "key 表单格式不正确")
			return
		}
		_, err := store.CreateAPIKey(r.Context(), admin.CreateAPIKeyInput{
			ProviderID:  r.FormValue("providerId"),
			Alias:       r.FormValue("alias"),
			SecretValue: r.FormValue("secretValue"),
			IsEnabled:   r.FormValue("isEnabled") == "1",
			IsAvailable: r.FormValue("isAvailable") == "1",
			SortOrder:   parseOptionalInt(r.FormValue("sortOrder")),
		})
		if err != nil {
			redirectAdminError(w, r, "添加 key 失败："+err.Error())
			return
		}
		http.Redirect(w, r, "/admin", http.StatusFound)
	}
}

func loadProviderPanels(r *http.Request, store *admin.Store) ([]providerPanel, error) {
	providers, err := store.ListProviders(r.Context())
	if err != nil {
		return nil, err
	}
	panels := make([]providerPanel, 0, len(providers))
	for _, provider := range providers {
		models, err := store.ListModels(r.Context(), provider.ID, "")
		if err != nil {
			return nil, err
		}
		keys, err := store.ListAPIKeys(r.Context(), provider.ID)
		if err != nil {
			return nil, err
		}
		panels = append(panels, providerPanel{Provider: provider, Models: models, Keys: keys})
	}
	return panels, nil
}

func redirectAdminError(w http.ResponseWriter, r *http.Request, message string) {
	http.Redirect(w, r, "/admin?error="+template.URLQueryEscaper(message), http.StatusFound)
}
