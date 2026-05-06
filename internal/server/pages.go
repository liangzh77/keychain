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
    :root { --bg: #f7f8fa; --surface: #fff; --line: #d9dee7; --line-soft: #edf0f4; --text: #17202a; --muted: #687385; --accent: #2463eb; --danger: #b42318; --ok: #18794e; }
    * { box-sizing: border-box; }
    body { margin: 0; font-family: ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; background: var(--bg); color: var(--text); }
    header { height: 56px; display: flex; align-items: center; justify-content: space-between; padding: 0 20px; background: var(--surface); border-bottom: 1px solid var(--line); }
    header form { margin: 0; }
    .brand { display: flex; align-items: baseline; gap: 10px; }
    .brand strong { font-size: 16px; }
    .app { height: calc(100vh - 56px); display: grid; grid-template-columns: 320px minmax(0, 1fr); overflow: hidden; }
    aside { border-right: 1px solid var(--line); background: #fbfcfd; overflow: auto; padding: 16px; }
    main { overflow: auto; padding: 20px 24px 28px; }
    h1, h2, h3 { margin: 0; line-height: 1.2; }
    h1 { font-size: 24px; letter-spacing: 0; }
    h2 { font-size: 18px; }
    h3 { font-size: 15px; }
    p { margin: 0; }
    .muted { color: var(--muted); }
    .small { font-size: 12px; }
    .panel { background: var(--surface); border: 1px solid var(--line); border-radius: 8px; }
    .panel-pad { padding: 16px; }
    .stack { display: grid; gap: 14px; }
    .topline { display: flex; justify-content: space-between; align-items: start; gap: 16px; margin-bottom: 18px; }
    .provider-list { display: grid; gap: 8px; margin-top: 14px; }
    .provider-link { display: block; padding: 10px 12px; border: 1px solid transparent; border-radius: 7px; color: inherit; text-decoration: none; }
    .provider-link:hover { background: #f1f4f8; }
    .provider-link.active { background: #eef4ff; border-color: #bed3ff; }
    .provider-row { display: flex; justify-content: space-between; gap: 12px; align-items: center; }
    .provider-code { font-family: ui-monospace, SFMono-Regular, Consolas, monospace; color: var(--muted); font-size: 12px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    .count { color: var(--muted); font-size: 12px; white-space: nowrap; }
    .tabs { display: flex; gap: 4px; border-bottom: 1px solid var(--line); padding: 0 16px; }
    .tab { padding: 12px 12px; color: var(--muted); text-decoration: none; border-bottom: 2px solid transparent; font-weight: 700; font-size: 14px; }
    .tab.active { color: var(--accent); border-color: var(--accent); }
    .content { padding: 16px; }
    form { display: grid; gap: 10px; }
    form[id^="delete-"] { display: none; }
    .form-grid { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 10px; align-items: end; }
    .form-grid.models { grid-template-columns: 1.2fr 1.2fr 120px 100px; }
    .form-grid.keys { grid-template-columns: 1fr 1.3fr 90px 110px 110px 100px; }
    .settings-grid { display: grid; grid-template-columns: 1fr 1fr 220px 120px 120px; gap: 10px; align-items: end; }
    label { display: grid; gap: 5px; font-size: 12px; font-weight: 700; color: #384252; }
    input, select { width: 100%; min-width: 0; padding: 9px 10px; border: 1px solid #cbd3df; border-radius: 6px; font: inherit; background: #fff; color: var(--text); }
    input[type="checkbox"] { width: auto; }
    .check { display: flex; align-items: center; gap: 8px; height: 38px; }
    button { height: 38px; padding: 0 12px; border: 0; border-radius: 6px; background: var(--accent); color: white; cursor: pointer; font-weight: 700; }
    button.secondary { background: #46515f; }
    button.danger { background: var(--danger); }
    button.ghost { background: #eef1f5; color: #303846; }
    .row-list { display: grid; gap: 8px; margin-top: 14px; }
    .row-head, .row-edit { display: grid; gap: 8px; align-items: center; }
    .row-head { color: var(--muted); font-size: 12px; font-weight: 700; padding: 0 8px; }
    .row-edit { padding: 8px; border: 1px solid var(--line-soft); border-radius: 7px; background: #fff; }
    .row-edit.models, .row-head.models { grid-template-columns: 1fr 1fr 130px 150px; }
    .row-edit.keys, .row-head.keys { grid-template-columns: 1fr 1.4fr 90px 170px 150px; }
    .actions { display: flex; gap: 6px; }
    .tag { display: inline-flex; align-items: center; padding: 2px 8px; border-radius: 999px; background: #eef4ff; color: #1f4f9a; font-size: 12px; font-weight: 700; }
    .tag.off { background: #f2f3f5; color: #697386; }
    .notice { margin-bottom: 14px; padding: 10px 12px; border-radius: 6px; background: #fff7e6; color: #8a5a00; }
    .empty { padding: 28px; text-align: center; color: var(--muted); }
    @media (max-width: 980px) { .app { grid-template-columns: 1fr; height: auto; overflow: visible; } aside, main { overflow: visible; } .form-grid, .form-grid.models, .form-grid.keys, .settings-grid, .row-edit.models, .row-head.models, .row-edit.keys, .row-head.keys { grid-template-columns: 1fr; } }
  </style>
</head>
<body>
  <header>
    <div class="brand"><strong>Keychain</strong><span class="muted small">admin console</span><a class="tab" href="/admin">Providers</a><a class="tab" href="/admin/access">渠道与授权</a></div>
    <form method="post" action="/logout"><button class="ghost" type="submit">退出</button></form>
  </header>
  <div class="app">
    <aside>
      <div class="stack">
        <section class="panel panel-pad">
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
            <label class="check"><input type="checkbox" name="isEnabled" value="1" checked> 启用 provider</label>
            <button type="submit">添加 provider</button>
          </form>
        </section>
        <section>
          <h2>Providers</h2>
          <div class="provider-list">
            {{range .Providers}}
              <a class="provider-link {{if .IsActive}}active{{end}}" href="/admin?providerId={{.Provider.ID}}&tab={{$.ActiveTab}}">
                <div class="provider-row">
                  <strong>{{.Provider.Name}}</strong>
                  {{if .Provider.IsEnabled}}<span class="tag">启用</span>{{else}}<span class="tag off">停用</span>{{end}}
                </div>
                <div class="provider-row">
                  <span class="provider-code">{{.Provider.Code}}</span>
                  <span class="count">{{.ModelCount}} models · {{.KeyCount}} keys</span>
                </div>
              </a>
            {{else}}
              <div class="panel empty">还没有 provider。</div>
            {{end}}
          </div>
        </section>
      </div>
    </aside>
    <main>
      <div class="topline">
        <div>
          <h1>后台控制台</h1>
          <p class="muted">已登录为 {{.Username}}。在左侧选择 provider，右侧用 tab 管理模型、keys 和设置。</p>
        </div>
      </div>
      {{if .Error}}<div class="notice">{{.Error}}</div>{{end}}
      {{if .Selected}}
        <section class="panel">
          <div class="content">
            <div class="topline">
              <div>
                <h2>{{.Selected.Provider.Name}}</h2>
                <p class="muted">{{.Selected.Provider.Code}} · {{.Selected.Provider.RotationStrategy}}</p>
              </div>
              {{if .Selected.Provider.IsEnabled}}<span class="tag">启用</span>{{else}}<span class="tag off">停用</span>{{end}}
            </div>
          </div>
          <nav class="tabs">
            <a class="tab {{if eq .ActiveTab "models"}}active{{end}}" href="/admin?providerId={{.Selected.Provider.ID}}&tab=models">Models</a>
            <a class="tab {{if eq .ActiveTab "keys"}}active{{end}}" href="/admin?providerId={{.Selected.Provider.ID}}&tab=keys">Keys</a>
            <a class="tab {{if eq .ActiveTab "settings"}}active{{end}}" href="/admin?providerId={{.Selected.Provider.ID}}&tab=settings">Provider 设置</a>
          </nav>
          <div class="content">
            {{if eq .ActiveTab "models"}}
              <form class="form-grid models" method="post" action="/admin/models">
                <input type="hidden" name="providerId" value="{{.Selected.Provider.ID}}">
                <label>名称<input name="name" placeholder="GPT 4.1" required></label>
                <label>代码<input name="code" placeholder="gpt-4.1" required></label>
                <label class="check"><input type="checkbox" name="isEnabled" value="1" checked> 启用</label>
                <button type="submit">添加 model</button>
              </form>
              {{if .Selected.Models}}
                <div class="row-list">
                  <div class="row-head models"><span>名称</span><span>代码</span><span>状态</span><span>操作</span></div>
                  {{range .Selected.Models}}
                    <form class="row-edit models" method="post" action="/admin/models/update">
                      <input type="hidden" name="providerId" value="{{$.Selected.Provider.ID}}">
                      <input type="hidden" name="modelId" value="{{.ID}}">
                      <input name="name" value="{{.Name}}" required>
                      <input name="code" value="{{.Code}}" required>
                      <label class="check"><input type="checkbox" name="isEnabled" value="1" {{if .IsEnabled}}checked{{end}}> 启用</label>
                      <span class="actions">
                        <button class="secondary" type="submit">保存</button>
                        <button class="danger" type="submit" form="delete-model-{{.ID}}">删除</button>
                      </span>
                    </form>
                    <form id="delete-model-{{.ID}}" method="post" action="/admin/models/delete">
                      <input type="hidden" name="providerId" value="{{$.Selected.Provider.ID}}">
                      <input type="hidden" name="modelId" value="{{.ID}}">
                    </form>
                  {{end}}
                </div>
              {{else}}<div class="empty">这个 provider 还没有 model。</div>{{end}}
            {{end}}
            {{if eq .ActiveTab "keys"}}
              <form class="form-grid keys" method="post" action="/admin/keys">
                <input type="hidden" name="providerId" value="{{.Selected.Provider.ID}}">
                <label>别名<input name="alias" placeholder="openai-main-01" required></label>
                <label>Key 明文<input name="secretValue" placeholder="sk-..." required></label>
                <label>排序<input name="sortOrder" type="number" value="0"></label>
                <label class="check"><input type="checkbox" name="isEnabled" value="1" checked> 启用</label>
                <label class="check"><input type="checkbox" name="isAvailable" value="1" checked> 可用</label>
                <button type="submit">添加 key</button>
              </form>
              {{if .Selected.Keys}}
                <div class="row-list">
                  <div class="row-head keys"><span>别名</span><span>掩码 / 替换明文</span><span>排序</span><span>状态</span><span>操作</span></div>
                  {{range .Selected.Keys}}
                    <form class="row-edit keys" method="post" action="/admin/keys/update">
                      <input type="hidden" name="providerId" value="{{$.Selected.Provider.ID}}">
                      <input type="hidden" name="keyId" value="{{.ID}}">
                      <input name="alias" value="{{.Alias}}" required>
                      <input name="secretValue" placeholder="{{.MaskedValue}}，留空不替换">
                      <input name="sortOrder" type="number" value="{{.SortOrder}}">
                      <span>
                        <label class="check"><input type="checkbox" name="isEnabled" value="1" {{if .IsEnabled}}checked{{end}}> 启用</label>
                        <label class="check"><input type="checkbox" name="isAvailable" value="1" {{if .IsAvailable}}checked{{end}}> 可用</label>
                      </span>
                      <span class="actions">
                        <button class="secondary" type="submit">保存</button>
                        <button class="danger" type="submit" form="delete-key-{{.ID}}">删除</button>
                      </span>
                    </form>
                    <form id="delete-key-{{.ID}}" method="post" action="/admin/keys/delete">
                      <input type="hidden" name="providerId" value="{{$.Selected.Provider.ID}}">
                      <input type="hidden" name="keyId" value="{{.ID}}">
                    </form>
                  {{end}}
                </div>
              {{else}}<div class="empty">这个 provider 还没有 key。</div>{{end}}
            {{end}}
            {{if eq .ActiveTab "settings"}}
              <form class="settings-grid" method="post" action="/admin/providers/update">
                <input type="hidden" name="providerId" value="{{.Selected.Provider.ID}}">
                <label>名称<input name="name" value="{{.Selected.Provider.Name}}" required></label>
                <label>代码<input name="code" value="{{.Selected.Provider.Code}}" required></label>
                <label>Key 分发策略
                  <select name="rotationStrategy">
                    <option value="ROUND_ROBIN" {{if eq .Selected.Provider.RotationStrategy "ROUND_ROBIN"}}selected{{end}}>轮询分发</option>
                    <option value="STICKY_FIRST_AVAILABLE" {{if eq .Selected.Provider.RotationStrategy "STICKY_FIRST_AVAILABLE"}}selected{{end}}>优先第一个可用 key</option>
                  </select>
                </label>
                <label class="check"><input type="checkbox" name="isEnabled" value="1" {{if .Selected.Provider.IsEnabled}}checked{{end}}> 启用</label>
                <button type="submit">保存</button>
              </form>
              <form method="post" action="/admin/providers/delete" style="margin-top:14px">
                <input type="hidden" name="providerId" value="{{.Selected.Provider.ID}}">
                <button class="danger" type="submit">删除这个 provider</button>
              </form>
            {{end}}
          </div>
        </section>
      {{else}}
        <section class="panel empty">先在左侧添加一个 provider。</section>
      {{end}}
    </main>
  </div>
</body>
</html>`))

type adminPageData struct {
	Username  string
	Error     string
	Providers []providerNavItem
	Selected  *providerPanel
	ActiveTab string
}

type providerNavItem struct {
	Provider   admin.Provider
	IsActive   bool
	ModelCount int
	KeyCount   int
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
		registerAccessRoutes(mux, authService, store)
		mux.HandleFunc("POST /admin/providers", formCreateProviderHandler(store))
		mux.HandleFunc("POST /admin/providers/update", formUpdateProviderHandler(store))
		mux.HandleFunc("POST /admin/providers/delete", formDeleteProviderHandler(store))
		mux.HandleFunc("POST /admin/models", formCreateModelHandler(store))
		mux.HandleFunc("POST /admin/models/update", formUpdateModelHandler(store))
		mux.HandleFunc("POST /admin/models/delete", formDeleteModelHandler(store))
		mux.HandleFunc("POST /admin/keys", formCreateKeyHandler(store))
		mux.HandleFunc("POST /admin/keys/update", formUpdateKeyHandler(store))
		mux.HandleFunc("POST /admin/keys/delete", formDeleteKeyHandler(store))
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
		data := adminPageData{Username: session.Username, Error: r.URL.Query().Get("error"), ActiveTab: activeTab(r)}
		if store != nil {
			loaded, err := loadAdminPageData(r, store, data)
			if err != nil {
				loaded.Error = "加载 provider 数据失败"
			}
			data = loaded
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
		provider, err := store.CreateProvider(r.Context(), admin.CreateProviderInput{
			Name:             r.FormValue("name"),
			Code:             r.FormValue("code"),
			IsEnabled:        r.FormValue("isEnabled") == "1",
			RotationStrategy: r.FormValue("rotationStrategy"),
		})
		if err != nil {
			redirectAdminError(w, r, "添加 provider 失败："+err.Error())
			return
		}
		redirectToProvider(w, r, provider.ID, "settings")
	}
}

func formUpdateProviderHandler(store *admin.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			redirectAdminError(w, r, "provider 表单格式不正确")
			return
		}
		providerID := r.FormValue("providerId")
		_, err := store.UpdateProvider(r.Context(), providerID, admin.UpdateProviderInput{
			Name:             r.FormValue("name"),
			Code:             r.FormValue("code"),
			IsEnabled:        r.FormValue("isEnabled") == "1",
			RotationStrategy: r.FormValue("rotationStrategy"),
		})
		if err != nil {
			redirectAdminError(w, r, "保存 provider 失败："+err.Error())
			return
		}
		redirectToProvider(w, r, providerID, "settings")
	}
}

func formDeleteProviderHandler(store *admin.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			redirectAdminError(w, r, "provider 表单格式不正确")
			return
		}
		if err := store.DeleteProvider(r.Context(), r.FormValue("providerId")); err != nil {
			redirectAdminError(w, r, "删除 provider 失败："+err.Error())
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
		providerID := r.FormValue("providerId")
		_, err := store.CreateModel(r.Context(), admin.CreateModelInput{
			ProviderID: providerID,
			Name:       r.FormValue("name"),
			Code:       r.FormValue("code"),
			IsEnabled:  r.FormValue("isEnabled") == "1",
		})
		if err != nil {
			redirectAdminError(w, r, "添加 model 失败："+err.Error())
			return
		}
		redirectToProvider(w, r, providerID, "models")
	}
}

func formUpdateModelHandler(store *admin.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			redirectAdminError(w, r, "model 表单格式不正确")
			return
		}
		providerID := r.FormValue("providerId")
		_, err := store.UpdateModel(r.Context(), r.FormValue("modelId"), admin.UpdateModelInput{
			Name:      r.FormValue("name"),
			Code:      r.FormValue("code"),
			IsEnabled: r.FormValue("isEnabled") == "1",
		})
		if err != nil {
			redirectAdminError(w, r, "保存 model 失败："+err.Error())
			return
		}
		redirectToProvider(w, r, providerID, "models")
	}
}

func formDeleteModelHandler(store *admin.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			redirectAdminError(w, r, "model 表单格式不正确")
			return
		}
		providerID := r.FormValue("providerId")
		if err := store.DeleteModel(r.Context(), r.FormValue("modelId")); err != nil {
			redirectAdminError(w, r, "删除 model 失败："+err.Error())
			return
		}
		redirectToProvider(w, r, providerID, "models")
	}
}

func formCreateKeyHandler(store *admin.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			redirectAdminError(w, r, "key 表单格式不正确")
			return
		}
		providerID := r.FormValue("providerId")
		_, err := store.CreateAPIKey(r.Context(), admin.CreateAPIKeyInput{
			ProviderID:  providerID,
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
		redirectToProvider(w, r, providerID, "keys")
	}
}

func formUpdateKeyHandler(store *admin.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			redirectAdminError(w, r, "key 表单格式不正确")
			return
		}
		providerID := r.FormValue("providerId")
		_, err := store.UpdateAPIKey(r.Context(), r.FormValue("keyId"), admin.UpdateAPIKeyInput{
			Alias:       r.FormValue("alias"),
			SecretValue: r.FormValue("secretValue"),
			IsEnabled:   r.FormValue("isEnabled") == "1",
			IsAvailable: r.FormValue("isAvailable") == "1",
			SortOrder:   parseOptionalInt(r.FormValue("sortOrder")),
		})
		if err != nil {
			redirectAdminError(w, r, "保存 key 失败："+err.Error())
			return
		}
		redirectToProvider(w, r, providerID, "keys")
	}
}

func formDeleteKeyHandler(store *admin.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			redirectAdminError(w, r, "key 表单格式不正确")
			return
		}
		providerID := r.FormValue("providerId")
		if err := store.DeleteAPIKey(r.Context(), r.FormValue("keyId")); err != nil {
			redirectAdminError(w, r, "删除 key 失败："+err.Error())
			return
		}
		redirectToProvider(w, r, providerID, "keys")
	}
}

func loadAdminPageData(r *http.Request, store *admin.Store, data adminPageData) (adminPageData, error) {
	providers, err := store.ListProviders(r.Context())
	if err != nil {
		return data, err
	}
	selectedID := r.URL.Query().Get("providerId")
	if selectedID == "" && len(providers) > 0 {
		selectedID = providers[0].ID
	}

	data.Providers = make([]providerNavItem, 0, len(providers))
	for _, provider := range providers {
		models, err := store.ListModels(r.Context(), provider.ID, "")
		if err != nil {
			return data, err
		}
		keys, err := store.ListAPIKeys(r.Context(), provider.ID)
		if err != nil {
			return data, err
		}
		item := providerNavItem{
			Provider:   provider,
			IsActive:   provider.ID == selectedID,
			ModelCount: len(models),
			KeyCount:   len(keys),
		}
		data.Providers = append(data.Providers, item)
		if provider.ID == selectedID {
			selectedProvider := providerPanel{Provider: provider, Models: models, Keys: keys}
			data.Selected = &selectedProvider
		}
	}
	return data, nil
}

func activeTab(r *http.Request) string {
	switch r.URL.Query().Get("tab") {
	case "keys":
		return "keys"
	case "settings":
		return "settings"
	default:
		return "models"
	}
}

func redirectToProvider(w http.ResponseWriter, r *http.Request, providerID string, tab string) {
	http.Redirect(w, r, "/admin?providerId="+template.URLQueryEscaper(providerID)+"&tab="+template.URLQueryEscaper(tab), http.StatusFound)
}

func redirectAdminError(w http.ResponseWriter, r *http.Request, message string) {
	http.Redirect(w, r, "/admin?error="+template.URLQueryEscaper(message), http.StatusFound)
}
