package server

import (
	"html/template"
	"net/http"

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
    header { display: flex; align-items: center; justify-content: space-between; padding: 16px 24px; background: #fff; border-bottom: 1px solid #dfe4ea; }
    main { max-width: 1040px; margin: 0 auto; padding: 32px 24px; }
    h1 { margin: 0 0 8px; font-size: 28px; }
    .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); gap: 16px; margin-top: 24px; }
    .card { background: #fff; border: 1px solid #dfe4ea; border-radius: 8px; padding: 18px; }
    .muted { color: #657080; }
    button { padding: 8px 12px; border: 0; border-radius: 6px; background: #46515f; color: white; cursor: pointer; }
  </style>
</head>
<body>
  <header>
    <strong>Keychain</strong>
    <form method="post" action="/logout"><button type="submit">退出</button></form>
  </header>
  <main>
    <h1>后台控制台</h1>
    <p class="muted">已登录为 {{.Username}}。管理功能会在后续阶段逐步接入。</p>
    <section class="grid">
      <div class="card"><strong>Providers</strong><p class="muted">待接入 provider 和 model 管理。</p></div>
      <div class="card"><strong>Keys</strong><p class="muted">待接入 key 管理和掩码展示。</p></div>
      <div class="card"><strong>权限</strong><p class="muted">待接入渠道、用户和批量权限设置。</p></div>
      <div class="card"><strong>审计</strong><p class="muted">待接入分发历史和失败记录。</p></div>
    </section>
  </main>
</body>
</html>`))

func registerPageRoutes(mux *http.ServeMux, authService *auth.Service) {
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/admin", http.StatusFound)
	})
	mux.HandleFunc("GET /login", showLoginPage)
	mux.HandleFunc("POST /login", formLoginHandler(authService))
	mux.HandleFunc("POST /logout", formLogoutHandler(authService))
	mux.HandleFunc("GET /admin", adminPageHandler(authService))
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

func adminPageHandler(authService *auth.Service) http.HandlerFunc {
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
		_ = adminPageTemplate.Execute(w, map[string]string{"Username": session.Username})
	}
}

func renderLoginError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusUnauthorized)
	_ = loginPageTemplate.Execute(w, map[string]string{"Error": message})
}
