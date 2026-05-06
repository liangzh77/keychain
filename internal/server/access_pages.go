package server

import (
	"html/template"
	"net/http"

	"github.com/liangzh77/keychain/internal/admin"
	"github.com/liangzh77/keychain/internal/auth"
)

var accessPageTemplate = template.Must(template.New("access").Parse(`<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Keychain 渠道与授权</title>
  <style>
    :root { --bg: #f7f8fa; --surface: #fff; --line: #d9dee7; --line-soft: #edf0f4; --text: #17202a; --muted: #687385; --accent: #2463eb; --danger: #b42318; }
    * { box-sizing: border-box; }
    body { margin: 0; font-family: ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; background: var(--bg); color: var(--text); }
    header { height: 56px; display: flex; align-items: center; justify-content: space-between; padding: 0 20px; background: var(--surface); border-bottom: 1px solid var(--line); }
    header form { margin: 0; }
    .brand { display: flex; align-items: baseline; gap: 10px; }
    .brand strong { font-size: 16px; }
    .app { height: calc(100vh - 56px); display: grid; grid-template-columns: 340px minmax(0, 1fr); overflow: hidden; }
    aside { border-right: 1px solid var(--line); background: #fbfcfd; overflow: auto; padding: 16px; }
    main { overflow: auto; padding: 20px 24px 28px; }
    h1, h2, h3 { margin: 0; line-height: 1.2; }
    h1 { font-size: 24px; }
    h2 { font-size: 18px; }
    h3 { font-size: 15px; }
    p { margin: 0; }
    .muted { color: var(--muted); }
    .small { font-size: 12px; }
    .panel { background: var(--surface); border: 1px solid var(--line); border-radius: 8px; }
    .panel-pad, .content { padding: 16px; }
    .stack { display: grid; gap: 14px; }
    .topline { display: flex; justify-content: space-between; align-items: start; gap: 16px; margin-bottom: 18px; }
    .tab { padding: 8px 10px; color: var(--muted); text-decoration: none; border-radius: 6px; font-weight: 700; font-size: 14px; }
    .tab.active, .tab:hover { background: #eef4ff; color: var(--accent); }
    .channel-list, .user-list, .row-list { display: grid; gap: 8px; margin-top: 14px; }
    .channel-link, .user-link { display: block; padding: 10px 12px; border: 1px solid transparent; border-radius: 7px; color: inherit; text-decoration: none; }
    .channel-link:hover, .user-link:hover { background: #f1f4f8; }
    .channel-link.active, .user-link.active { background: #eef4ff; border-color: #bed3ff; }
    .meta-row { display: flex; justify-content: space-between; gap: 12px; align-items: center; }
    .mono { font-family: ui-monospace, SFMono-Regular, Consolas, monospace; color: var(--muted); font-size: 12px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    form { display: grid; gap: 10px; }
    form[id^="delete-"] { display: none; }
    label { display: grid; gap: 5px; font-size: 12px; font-weight: 700; color: #384252; }
    input, select { width: 100%; min-width: 0; padding: 9px 10px; border: 1px solid #cbd3df; border-radius: 6px; font: inherit; background: #fff; color: var(--text); }
    input[type="checkbox"] { width: auto; }
    .check { display: flex; align-items: center; gap: 8px; min-height: 38px; }
    button { height: 38px; padding: 0 12px; border: 0; border-radius: 6px; background: var(--accent); color: white; cursor: pointer; font-weight: 700; }
    button.secondary { background: #46515f; }
    button.danger { background: var(--danger); }
    button.ghost { background: #eef1f5; color: #303846; }
    .grid-two { display: grid; grid-template-columns: minmax(280px, 360px) 1fr; gap: 14px; align-items: start; }
    .form-grid { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 10px; align-items: end; }
    .row-head, .perm-row { display: grid; grid-template-columns: 1.1fr 1.1fr 160px 120px; gap: 8px; align-items: center; }
    .row-head { color: var(--muted); font-size: 12px; font-weight: 700; padding: 0 8px; }
    .perm-row { padding: 8px; border: 1px solid var(--line-soft); border-radius: 7px; background: #fff; }
    .notice { margin-bottom: 14px; padding: 10px 12px; border-radius: 6px; background: #fff7e6; color: #8a5a00; }
    .empty { padding: 28px; text-align: center; color: var(--muted); }
    .tag { display: inline-flex; align-items: center; padding: 2px 8px; border-radius: 999px; background: #eef4ff; color: #1f4f9a; font-size: 12px; font-weight: 700; }
    .tag.off { background: #f2f3f5; color: #697386; }
    @media (max-width: 980px) { .app, .grid-two, .form-grid, .row-head, .perm-row { grid-template-columns: 1fr; height: auto; overflow: visible; } aside, main { overflow: visible; } }
  </style>
</head>
<body>
  <header>
    <div class="brand"><strong>Keychain</strong><span class="muted small">admin console</span><a class="tab" href="/admin">Providers</a><a class="tab active" href="/admin/access">渠道与授权</a></div>
    <form method="post" action="/logout"><button class="ghost" type="submit">退出</button></form>
  </header>
  <div class="app">
    <aside>
      <div class="stack">
        <section class="panel panel-pad">
          <form method="post" action="/admin/access/demo">
            <button type="submit">写入演示渠道和用户</button>
          </form>
        </section>
        <section class="panel panel-pad">
          <h2>添加渠道</h2>
          <form method="post" action="/admin/channels">
            <label>名称<input name="name" placeholder="本校默认渠道" required></label>
            <label>代码<input name="code" placeholder="school-default" required></label>
            <label>默认权限
              <select name="defaultPermissionMode">
                <option value="DENY">默认关闭</option>
                <option value="ALLOW">默认打开</option>
              </select>
            </label>
            <label class="check"><input type="checkbox" name="isEnabled" value="1" checked> 启用渠道</label>
            <button type="submit">添加渠道</button>
          </form>
        </section>
        <section>
          <h2>Channels</h2>
          <div class="channel-list">
            {{range .Channels}}
              <a class="channel-link {{if .IsActive}}active{{end}}" href="/admin/access?channelId={{.Channel.ID}}">
                <div class="meta-row"><strong>{{.Channel.Name}}</strong>{{if .Channel.IsEnabled}}<span class="tag">启用</span>{{else}}<span class="tag off">停用</span>{{end}}</div>
                <div class="meta-row"><span class="mono">{{.Channel.Code}}</span><span class="muted small">{{.UserCount}} users · {{.Channel.DefaultPermissionMode}}</span></div>
              </a>
            {{else}}
              <div class="panel empty">还没有渠道。</div>
            {{end}}
          </div>
        </section>
      </div>
    </aside>
    <main>
      <div class="topline">
        <div>
          <h1>渠道与授权</h1>
          <p class="muted">先选渠道，再维护用户、渠道默认权限和用户显式权限。</p>
        </div>
      </div>
      {{if .Error}}<div class="notice">{{.Error}}</div>{{end}}
      {{if .Selected}}
        <section class="grid-two">
          <div class="stack">
            <section class="panel panel-pad">
              <h2>{{.Selected.Channel.Name}}</h2>
              <form method="post" action="/admin/channels/update">
                <input type="hidden" name="channelId" value="{{.Selected.Channel.ID}}">
                <label>名称<input name="name" value="{{.Selected.Channel.Name}}" required></label>
                <label>代码<input name="code" value="{{.Selected.Channel.Code}}" required></label>
                <label>默认权限
                  <select name="defaultPermissionMode">
                    <option value="DENY" {{if eq .Selected.Channel.DefaultPermissionMode "DENY"}}selected{{end}}>默认关闭</option>
                    <option value="ALLOW" {{if eq .Selected.Channel.DefaultPermissionMode "ALLOW"}}selected{{end}}>默认打开</option>
                  </select>
                </label>
                <label class="check"><input type="checkbox" name="isEnabled" value="1" {{if .Selected.Channel.IsEnabled}}checked{{end}}> 启用渠道</label>
                <button type="submit">保存渠道</button>
              </form>
            </section>
            <section class="panel panel-pad">
              <h2>添加用户</h2>
              <form method="post" action="/admin/users">
                <input type="hidden" name="channelId" value="{{.Selected.Channel.ID}}">
                <label>外部用户 ID<input name="externalUserId" placeholder="stu-2026-001" required></label>
                <label>显示名<input name="displayName" placeholder="教学演示用户 001" required></label>
                <label class="check"><input type="checkbox" name="isEnabled" value="1" checked> 启用用户</label>
                <button type="submit">添加用户</button>
              </form>
            </section>
            <section class="panel panel-pad">
              <h2>Users</h2>
              <div class="user-list">
                {{range .Selected.Users}}
                  <a class="user-link {{if eq $.SelectedUserID .ID}}active{{end}}" href="/admin/access?channelId={{$.Selected.Channel.ID}}&userId={{.ID}}">
                    <div class="meta-row"><strong>{{.DisplayName}}</strong>{{if .IsEnabled}}<span class="tag">启用</span>{{else}}<span class="tag off">停用</span>{{end}}</div>
                    <span class="mono">{{.ExternalUserID}}</span>
                  </a>
                {{else}}
                  <div class="empty">这个渠道还没有用户。</div>
                {{end}}
              </div>
            </section>
          </div>
          <div class="stack">
            <section class="panel content">
              <h2>渠道默认授权</h2>
              <p class="muted small">用于该渠道下没有显式用户权限时的默认判断。</p>
              {{if .Selected.ChannelPermissions}}
                <div class="row-list">
                  <div class="row-head"><span>Provider</span><span>Model</span><span>默认允许</span><span>操作</span></div>
                  {{range .Selected.ChannelPermissions}}
                    <form class="perm-row" method="post" action="/admin/channel-permissions">
                      <input type="hidden" name="channelId" value="{{$.Selected.Channel.ID}}">
                      <input type="hidden" name="providerId" value="{{.ProviderID}}">
                      <input type="hidden" name="modelId" value="{{.ModelID}}">
                      <span>{{.ProviderName}}<br><span class="mono">{{.ProviderCode}}</span></span>
                      <span>{{.ModelName}}<br><span class="mono">{{.ModelCode}}</span></span>
                      <label class="check"><input type="checkbox" name="allowed" value="1" {{if .DefaultAllowed}}checked{{end}}> 允许</label>
                      <button class="secondary" type="submit">保存</button>
                    </form>
                  {{end}}
                </div>
              {{else}}<div class="empty">还没有 provider/model 可授权。</div>{{end}}
            </section>
            <section class="panel content">
              <h2>用户显式授权</h2>
              <p class="muted small">选择左侧用户后，可覆盖渠道默认授权。</p>
              {{if .SelectedUser}}
                <h3>{{.SelectedUser.DisplayName}}</h3>
                {{if .UserPermissions}}
                  <div class="row-list">
                    <div class="row-head"><span>Provider</span><span>Model</span><span>显式允许</span><span>操作</span></div>
                    {{range .UserPermissions}}
                      <form class="perm-row" method="post" action="/admin/user-permissions">
                        <input type="hidden" name="channelId" value="{{$.Selected.Channel.ID}}">
                        <input type="hidden" name="userId" value="{{$.SelectedUser.ID}}">
                        <input type="hidden" name="providerId" value="{{.ProviderID}}">
                        <input type="hidden" name="modelId" value="{{.ModelID}}">
                        <span>{{.ProviderName}}<br><span class="mono">{{.ProviderCode}}</span></span>
                        <span>{{.ModelName}}<br><span class="mono">{{.ModelCode}}</span></span>
                        <label class="check"><input type="checkbox" name="allowed" value="1" {{if .Allowed}}checked{{end}}> 允许</label>
                        <button class="secondary" type="submit">保存</button>
                      </form>
                    {{end}}
                  </div>
                {{else}}<div class="empty">还没有 provider/model 可授权。</div>{{end}}
              {{else}}<div class="empty">先选择一个用户。</div>{{end}}
            </section>
          </div>
        </section>
      {{else}}
        <section class="panel empty">先添加或选择一个渠道。</section>
      {{end}}
    </main>
  </div>
</body>
</html>`))

type accessPageData struct {
	Username        string
	Error           string
	Channels        []channelNavItem
	Selected        *channelPanel
	SelectedUser    *admin.User
	SelectedUserID  string
	UserPermissions []admin.UserPermissionRow
}

type channelNavItem struct {
	Channel   admin.Channel
	IsActive  bool
	UserCount int
}

type channelPanel struct {
	Channel            admin.Channel
	Users              []admin.User
	ChannelPermissions []admin.ChannelPermissionRow
}

func registerAccessRoutes(mux *http.ServeMux, authService *auth.Service, store *admin.Store) {
	mux.HandleFunc("GET /admin/access", accessPageHandler(authService, store))
	mux.HandleFunc("POST /admin/access/demo", formSeedDemoAccessHandler(store))
	mux.HandleFunc("POST /admin/channels", formCreateChannelHandler(store))
	mux.HandleFunc("POST /admin/channels/update", formUpdateChannelHandler(store))
	mux.HandleFunc("POST /admin/users", formCreateUserHandler(store))
	mux.HandleFunc("POST /admin/users/update", formUpdateUserHandler(store))
	mux.HandleFunc("POST /admin/channel-permissions", formSetChannelPermissionHandler(store))
	mux.HandleFunc("POST /admin/user-permissions", formSetUserPermissionHandler(store))
}

func accessPageHandler(authService *auth.Service, store *admin.Store) http.HandlerFunc {
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
		data := accessPageData{Username: session.Username, Error: r.URL.Query().Get("error"), SelectedUserID: r.URL.Query().Get("userId")}
		loaded, err := loadAccessPageData(r, store, data)
		if err != nil {
			loaded.Error = "加载渠道和授权数据失败"
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = accessPageTemplate.Execute(w, loaded)
	}
}

func loadAccessPageData(r *http.Request, store *admin.Store, data accessPageData) (accessPageData, error) {
	channels, err := store.ListChannels(r.Context())
	if err != nil {
		return data, err
	}
	selectedChannelID := r.URL.Query().Get("channelId")
	if selectedChannelID == "" && len(channels) > 0 {
		selectedChannelID = channels[0].ID
	}
	for _, channel := range channels {
		users, err := store.ListUsers(r.Context(), channel.ID)
		if err != nil {
			return data, err
		}
		item := channelNavItem{Channel: channel, IsActive: channel.ID == selectedChannelID, UserCount: len(users)}
		data.Channels = append(data.Channels, item)
		if channel.ID == selectedChannelID {
			if data.SelectedUserID == "" && len(users) > 0 {
				data.SelectedUserID = users[0].ID
			}
			channelPermissions, err := store.ListChannelPermissionRows(r.Context(), channel.ID)
			if err != nil {
				return data, err
			}
			selected := channelPanel{Channel: channel, Users: users, ChannelPermissions: channelPermissions}
			data.Selected = &selected
			for _, user := range users {
				if user.ID == data.SelectedUserID {
					selectedUser := user
					data.SelectedUser = &selectedUser
					data.UserPermissions, err = store.ListUserPermissionRows(r.Context(), user.ID)
					if err != nil {
						return data, err
					}
					break
				}
			}
		}
	}
	return data, nil
}

func formSeedDemoAccessHandler(store *admin.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := store.SeedDemoAccessData(r.Context()); err != nil {
			redirectAccessError(w, r, "写入演示数据失败："+err.Error())
			return
		}
		http.Redirect(w, r, "/admin/access", http.StatusFound)
	}
}

func formCreateChannelHandler(store *admin.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			redirectAccessError(w, r, "渠道表单格式不正确")
			return
		}
		channel, err := store.CreateChannel(r.Context(), admin.CreateChannelInput{
			Name:                  r.FormValue("name"),
			Code:                  r.FormValue("code"),
			DefaultPermissionMode: r.FormValue("defaultPermissionMode"),
			IsEnabled:             r.FormValue("isEnabled") == "1",
		})
		if err != nil {
			redirectAccessError(w, r, "添加渠道失败："+err.Error())
			return
		}
		redirectAccessChannel(w, r, channel.ID, "")
	}
}

func formUpdateChannelHandler(store *admin.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			redirectAccessError(w, r, "渠道表单格式不正确")
			return
		}
		channelID := r.FormValue("channelId")
		if err := store.UpdateChannel(r.Context(), channelID, admin.UpdateChannelInput{
			Name:                  r.FormValue("name"),
			Code:                  r.FormValue("code"),
			DefaultPermissionMode: r.FormValue("defaultPermissionMode"),
			IsEnabled:             r.FormValue("isEnabled") == "1",
		}); err != nil {
			redirectAccessError(w, r, "保存渠道失败："+err.Error())
			return
		}
		redirectAccessChannel(w, r, channelID, "")
	}
}

func formCreateUserHandler(store *admin.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			redirectAccessError(w, r, "用户表单格式不正确")
			return
		}
		channelID := r.FormValue("channelId")
		user, err := store.CreateUser(r.Context(), admin.CreateUserInput{
			ChannelID:      channelID,
			ExternalUserID: r.FormValue("externalUserId"),
			DisplayName:    r.FormValue("displayName"),
			IsEnabled:      r.FormValue("isEnabled") == "1",
		})
		if err != nil {
			redirectAccessError(w, r, "添加用户失败："+err.Error())
			return
		}
		redirectAccessChannel(w, r, channelID, user.ID)
	}
}

func formUpdateUserHandler(store *admin.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			redirectAccessError(w, r, "用户表单格式不正确")
			return
		}
		channelID := r.FormValue("channelId")
		userID := r.FormValue("userId")
		if err := store.UpdateUser(r.Context(), userID, admin.UpdateUserInput{
			ExternalUserID: r.FormValue("externalUserId"),
			DisplayName:    r.FormValue("displayName"),
			IsEnabled:      r.FormValue("isEnabled") == "1",
		}); err != nil {
			redirectAccessError(w, r, "保存用户失败："+err.Error())
			return
		}
		redirectAccessChannel(w, r, channelID, userID)
	}
}

func formSetChannelPermissionHandler(store *admin.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			redirectAccessError(w, r, "渠道授权表单格式不正确")
			return
		}
		channelID := r.FormValue("channelId")
		if err := store.SetChannelPermissionDefault(r.Context(), channelID, r.FormValue("providerId"), r.FormValue("modelId"), r.FormValue("allowed") == "1"); err != nil {
			redirectAccessError(w, r, "保存渠道授权失败："+err.Error())
			return
		}
		redirectAccessChannel(w, r, channelID, "")
	}
}

func formSetUserPermissionHandler(store *admin.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			redirectAccessError(w, r, "用户授权表单格式不正确")
			return
		}
		channelID := r.FormValue("channelId")
		userID := r.FormValue("userId")
		if err := store.SetUserPermission(r.Context(), userID, r.FormValue("providerId"), r.FormValue("modelId"), r.FormValue("allowed") == "1"); err != nil {
			redirectAccessError(w, r, "保存用户授权失败："+err.Error())
			return
		}
		redirectAccessChannel(w, r, channelID, userID)
	}
}

func redirectAccessChannel(w http.ResponseWriter, r *http.Request, channelID string, userID string) {
	url := "/admin/access?channelId=" + template.URLQueryEscaper(channelID)
	if userID != "" {
		url += "&userId=" + template.URLQueryEscaper(userID)
	}
	http.Redirect(w, r, url, http.StatusFound)
}

func redirectAccessError(w http.ResponseWriter, r *http.Request, message string) {
	http.Redirect(w, r, "/admin/access?error="+template.URLQueryEscaper(message), http.StatusFound)
}
