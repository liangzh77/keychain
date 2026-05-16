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
  <link rel="icon" href="/assets/keychain-icon.png">
  <style>
    body { margin: 0; min-height: 100vh; display: grid; place-items: center; font-family: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; background: #f4f1ea; color: #242824; }
    main { width: min(420px, calc(100vw - 32px)); background: #fffcf5; border: 1px solid #d7cdbf; border-radius: 8px; padding: 28px; box-shadow: 0 18px 52px rgba(52,45,33,.1); }
    .login-brand { display: flex; align-items: center; gap: 10px; margin-bottom: 8px; }
    .brand-logo { width: 34px; height: 34px; object-fit: contain; border-radius: 6px; }
    h1 { margin: 0 0 8px; font-size: 24px; }
    p { margin: 0 0 24px; color: #6f7169; }
    form { display: grid; gap: 14px; }
    label { display: grid; gap: 6px; font-weight: 600; }
    input { width: 100%; box-sizing: border-box; padding: 10px 12px; border: 1px solid #d2c7b7; border-radius: 6px; font: inherit; background: #fffdf8; color: #242824; }
    button { padding: 11px 14px; border: 0; border-radius: 6px; background: #31594a; color: white; font: inherit; font-weight: 700; cursor: pointer; }
    .error { margin-bottom: 16px; padding: 10px 12px; border-radius: 6px; background: #f6e8e4; color: #8f332c; }
  </style>
</head>
<body>
  <main>
    <div class="login-brand"><img class="brand-logo" src="/assets/keychain-icon.png" alt="" aria-hidden="true"><h1>Keychain</h1></div>
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
  <link rel="icon" href="/assets/keychain-icon.png">
  <style>
    :root { --bg: #f4f1ea; --surface: #fffcf5; --surface-muted: #f7f2ea; --line: #d7cdbf; --line-soft: #ece4d8; --text: #242824; --muted: #6f7169; --accent: #31594a; --accent-soft: #e8efe8; --accent-line: #9fb9aa; --secondary: #5a5448; --danger: #9b3d35; --ok: #31594a; }
    * { box-sizing: border-box; }
    body { margin: 0; font-family: ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; background: var(--bg); color: var(--text); }
    header { height: 56px; display: flex; align-items: center; justify-content: space-between; padding: 0 20px; background: var(--surface); border-bottom: 1px solid var(--line); }
    header form { margin: 0; }
    .brand { display: flex; align-items: baseline; gap: 10px; }
    .brand-logo { width: 26px; height: 26px; object-fit: contain; border-radius: 5px; align-self: center; }
    .brand strong { font-size: 16px; }
    .app { height: calc(100vh - 56px); display: grid; grid-template-columns: 300px minmax(0, 1fr); overflow: hidden; }
    .app[aria-busy="true"] { cursor: progress; }
    aside { border-right: 1px solid var(--line); background: var(--surface-muted); overflow: auto; padding: 16px; }
    main { overflow: auto; padding: 16px 20px 24px; }
    h1, h2, h3 { margin: 0; line-height: 1.2; }
    h1 { font-size: 22px; letter-spacing: 0; }
    h2 { font-size: 18px; }
    h3 { font-size: 15px; }
    p { margin: 0; }
    .muted { color: var(--muted); }
    .small { font-size: 12px; }
    .panel { background: var(--surface); border: 1px solid var(--line); border-radius: 8px; }
    .panel-pad { padding: 16px; }
    .stack { display: grid; gap: 12px; }
    .topline { display: flex; justify-content: space-between; align-items: start; gap: 16px; margin-bottom: 14px; }
    .provider-list, .compact-list { display: grid; gap: 6px; margin-top: 10px; }
    .provider-link { display: block; padding: 10px 12px; border: 1px solid transparent; border-radius: 7px; color: inherit; text-decoration: none; }
    .provider-link:hover { background: #f0ebe2; }
    .provider-link.active { background: var(--accent-soft); border-color: var(--accent-line); box-shadow: inset 3px 0 0 var(--accent); }
    .provider-row { display: flex; justify-content: space-between; gap: 12px; align-items: center; }
    .provider-code { font-family: ui-monospace, SFMono-Regular, Consolas, monospace; color: var(--muted); font-size: 12px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    .count { color: var(--muted); font-size: 12px; white-space: nowrap; }
    .tab { padding: 8px 10px; color: var(--muted); text-decoration: none; border-radius: 6px; font-weight: 700; font-size: 14px; }
    .tab.active, .tab:hover { background: var(--accent-soft); color: var(--accent); }
    .content { padding: 14px; }
    form { display: grid; gap: 10px; }
    form[id^="delete-"] { display: none; }
    .form-grid { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 10px; align-items: end; }
    .form-grid > *, .settings-grid > *, .detail-form > * { min-width: 0; }
    .settings-grid { display: grid; grid-template-columns: minmax(180px, 1fr) minmax(180px, 220px) auto auto; gap: 10px; align-items: end; }
    .resource-grid { display: grid; grid-template-columns: repeat(2, minmax(360px, 1fr)); gap: 12px; align-items: stretch; }
    .resource-grid > .panel { min-height: 430px; height: 100%; }
    .detail-grid { display: grid; grid-template-columns: minmax(132px, 200px) minmax(0, 1fr); gap: 12px; align-items: start; }
    .model-detail-grid { grid-template-columns: minmax(180px, 4.2fr) minmax(168px, 5.8fr); }
    .detail-form { display: grid; gap: 10px; align-items: end; }
    .key-form { grid-template-columns: minmax(140px, 1fr) minmax(220px, 1.4fr) auto auto; }
    .model-form { grid-template-columns: minmax(0, 1fr) 120px 120px; }
    label { display: grid; gap: 5px; font-size: 12px; font-weight: 700; color: #45483f; }
    input, select { width: 100%; min-width: 0; padding: 9px 10px; border: 1px solid #d2c7b7; border-radius: 6px; font: inherit; background: #fffdf8; color: var(--text); }
    input[type="checkbox"] { width: 17px; height: 17px; }
    .check { display: inline-flex; align-items: center; justify-content: center; gap: 8px; min-height: 38px; padding: 0 12px; border: 1px solid #d2c7b7; border-radius: 6px; background: #faf7f0; color: #3e433d; font-size: 13px; font-weight: 800; white-space: nowrap; }
    button { height: 38px; padding: 0 12px; border: 0; border-radius: 6px; background: var(--accent); color: white; cursor: pointer; font-weight: 700; white-space: nowrap; }
    button:disabled { cursor: not-allowed; opacity: .48; }
    button.secondary { background: var(--secondary); }
    button.danger { background: var(--danger); }
    button.ghost { background: #efe9df; color: #3e433d; }
    details.add-panel { position: relative; }
    details.add-panel > summary { list-style: none; display: flex; justify-content: center; align-items: center; height: 38px; border-radius: 6px; background: var(--accent); color: white; font-weight: 700; cursor: pointer; }
    details.add-panel > summary::-webkit-details-marker { display: none; }
    details.add-panel[open] > summary { margin-bottom: 12px; background: var(--secondary); }
    details.add-panel.wide-add > summary { min-width: 138px; padding: 0 18px; }
    details.add-panel .add-cancel { display: none; position: absolute; top: 7px; right: 7px; width: 24px; height: 24px; padding: 0; border-radius: 999px; background: #efe9df; color: #5a5448; font-size: 18px; line-height: 1; }
    details.add-panel[open] .add-cancel { display: inline-flex; align-items: center; justify-content: center; }
    .section-title { display: grid; grid-template-columns: minmax(0, 1fr) auto; align-items: start; gap: 12px; }
    .section-title h2 { font-size: 18px; line-height: 1.2; font-weight: 700; }
    .section-title p { font-size: 12px; line-height: 1.2; }
    .section-title > details.add-panel[open] { grid-column: 1 / -1; }
    .section-title > details.add-panel[open] > summary { width: max-content; min-width: 138px; margin-left: auto; margin-right: 34px; }
    .scroll-list { height: 276px; overflow-y: auto; padding-right: 2px; align-content: start; }
    .mini-link { display: block; min-height: 40px; padding: 9px 10px; border: 1px solid var(--line-soft); border-radius: 7px; color: inherit; text-decoration: none; background: #fffdf8; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    .mini-link:hover { background: #f6f1e8; }
    .mini-link.active { border-color: var(--accent-line); background: var(--accent-soft); box-shadow: inset 3px 0 0 var(--accent); }
    .sortable-list { gap: 6px; }
    .sortable-item { display: flex; align-items: center; cursor: grab; }
    .sortable-item.dragging { opacity: .55; }
    .sortable-item .mini-link { flex: 1; min-width: 0; }
    .sortable-item .copy-key-button { flex: 0 0 auto; margin-left: 6px; }
    .mini-title { display: flex; justify-content: space-between; gap: 10px; align-items: center; }
    .pane { display: grid; gap: 8px; min-width: 0; }
    .pane-title { display: flex; align-items: center; justify-content: space-between; min-height: 28px; padding: 0 2px; color: #5f6259; font-size: 12px; font-weight: 800; letter-spacing: 0; }
    .pane-title::before { content: ""; width: 4px; height: 16px; border-radius: 999px; background: var(--accent); }
    .pane-title span { margin-right: auto; margin-left: 8px; }
    .actions { display: flex; justify-content: flex-end; gap: 6px; flex-wrap: wrap; }
    .inline-checks { display: flex; align-items: center; justify-content: flex-start; gap: 8px; flex-wrap: wrap; }
    .detail-form.key-form { grid-template-columns: minmax(0, .7fr) minmax(0, 1.3fr); }
    .detail-form.key-form .actions { grid-column: 1 / -1; }
    .detail-form.model-form { grid-template-columns: minmax(0, 1fr) 90px 104px; }
    .model-detail-grid .detail-form.model-form { grid-template-columns: minmax(0, 1fr); }
    .model-detail-grid .detail-form.model-form .actions { justify-content: flex-start; }
    .tag { display: inline-flex; align-items: center; padding: 2px 8px; border-radius: 999px; background: var(--accent-soft); color: var(--accent); font-size: 12px; font-weight: 700; }
    .tag.off { background: #e9e4db; color: #746f66; }
    .notice { margin-bottom: 14px; padding: 10px 12px; border-radius: 6px; background: #fff6df; color: #7a5a22; }
    .empty { padding: 28px; text-align: center; color: var(--muted); }
    .half-card { width: min(720px, 100%); }
    @media (max-width: 1320px) { .resource-grid { grid-template-columns: 1fr; } .settings-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); } .settings-grid .actions { justify-content: flex-start; } }
    @media (max-width: 1120px) { .key-form { grid-template-columns: repeat(2, minmax(0, 1fr)); } }
    @media (max-width: 980px) { .app { grid-template-columns: 1fr; height: auto; overflow: visible; } aside, main { overflow: visible; } .form-grid, .settings-grid, .detail-form, .key-form, .model-form { grid-template-columns: 1fr; } .actions { justify-content: flex-start; } }
    @media (max-width: 760px) { .detail-grid, .model-detail-grid, .section-title { grid-template-columns: 1fr; } .topline { flex-direction: column; align-items: stretch; } .section-title > details.add-panel > summary { width: 100%; margin-left: 0; } }
  </style>
</head>
<body>
  <header>
    <div class="brand"><img class="brand-logo" src="/assets/keychain-icon.png" alt="" aria-hidden="true"><strong>Keychain</strong><span class="muted small">admin console</span><a class="tab active" href="/admin">Providers</a><a class="tab" href="/admin/access">渠道与授权</a><a class="tab" href="/admin/history">调用历史</a></div>
    <form method="post" action="/logout"><button class="ghost" type="submit">退出</button></form>
  </header>
  <div class="app">
    <aside>
      <div class="stack">
        <details class="panel panel-pad add-panel">
          <summary>添加 Provider</summary>
          <button class="add-cancel" type="button" data-close-add aria-label="取消添加">×</button>
          <form method="post" action="/admin/providers">
            <label>名称<input name="name" placeholder="OpenAI" required></label>
            <label>密钥分发策略
              <select name="rotationStrategy">
                <option value="ROUND_ROBIN">轮询分发</option>
                <option value="STICKY_FIRST_AVAILABLE">优先第一个可用密钥</option>
              </select>
            </label>
            <label class="check"><input type="checkbox" name="isEnabled" value="1" checked> 启用 provider</label>
            <button type="submit">添加 provider</button>
          </form>
        </details>
        <section>
          <h2>Providers</h2>
          <div class="provider-list">
            {{range .Providers}}
              <a class="provider-link {{if .IsActive}}active{{end}}" href="/admin?providerId={{.Provider.ID}}">
                <div class="provider-row">
                  <strong>{{.Provider.Name}}</strong>
                  {{if .Provider.IsEnabled}}<span class="tag">启用</span>{{else}}<span class="tag off">停用</span>{{end}}
                </div>
                <div class="provider-row"><span class="count">{{.ModelCount}}个模型 · {{.KeyCount}}个密钥</span></div>
              </a>
            {{else}}
              <div class="panel empty">还没有 provider。</div>
            {{end}}
          </div>
        </section>
      </div>
    </aside>
    <main>
      {{if .Error}}<div class="notice">{{.Error}}</div>{{end}}
      {{if .Selected}}
        <div class="stack">
        <section class="panel content half-card">
          <div class="topline">
            <div>
              <h2>Provider 详情</h2>
            </div>
            {{if .Selected.Provider.IsEnabled}}<span class="tag">启用</span>{{else}}<span class="tag off">停用</span>{{end}}
          </div>
          <form class="settings-grid" method="post" action="/admin/providers/update" data-dirty-form>
            <input type="hidden" name="providerId" value="{{.Selected.Provider.ID}}">
            <input type="hidden" name="code" value="{{.Selected.Provider.Code}}">
            <label>名称<input name="name" value="{{.Selected.Provider.Name}}" required></label>
            <label>密钥分发策略
              <select name="rotationStrategy">
                <option value="ROUND_ROBIN" {{if eq .Selected.Provider.RotationStrategy "ROUND_ROBIN"}}selected{{end}}>轮询分发</option>
                <option value="STICKY_FIRST_AVAILABLE" {{if eq .Selected.Provider.RotationStrategy "STICKY_FIRST_AVAILABLE"}}selected{{end}}>优先第一个可用密钥</option>
              </select>
            </label>
            <label class="check"><input type="checkbox" name="isEnabled" value="1" {{if .Selected.Provider.IsEnabled}}checked{{end}}> 启用</label>
            <span class="actions">
              <button class="secondary" type="submit" data-save disabled>保存</button>
              <button class="danger" type="submit" form="delete-provider-{{.Selected.Provider.ID}}">删除</button>
            </span>
          </form>
          <form id="delete-provider-{{.Selected.Provider.ID}}" method="post" action="/admin/providers/delete">
            <input type="hidden" name="providerId" value="{{.Selected.Provider.ID}}">
          </form>
        </section>
        <div class="resource-grid">
        <section class="panel content">
          <div class="content">
            <div class="section-title">
              <div>
                <h2>模型</h2>
                <p class="muted small">列表只显示名称，最多显示 6 行。</p>
              </div>
              <details class="add-panel wide-add">
                <summary>添加模型</summary>
                <button class="add-cancel" type="button" data-close-add aria-label="取消添加">×</button>
                <form class="form-grid model-form" method="post" action="/admin/models">
                  <input type="hidden" name="providerId" value="{{.Selected.Provider.ID}}">
                  <label>名称<input name="name" placeholder="GPT 4.1" required></label>
                  <label class="check"><input type="checkbox" name="isEnabled" value="1" checked> 启用</label>
                  <button type="submit">添加模型</button>
                </form>
              </details>
            </div>
            {{if .Selected.Models}}
              <div class="detail-grid model-detail-grid" style="margin-top:12px">
                <div class="pane">
                  <div class="pane-title"><span>Model 列表</span></div>
                  <div class="compact-list scroll-list">
                    {{range .Selected.Models}}
                      <a class="mini-link {{if eq $.SelectedModelID .ID}}active{{end}}" href="/admin?providerId={{$.Selected.Provider.ID}}&keyId={{$.SelectedKeyID}}&modelId={{.ID}}">
                        {{.Name}}
                      </a>
                    {{end}}
                  </div>
                </div>
                {{if .SelectedModel}}
                  <div class="pane">
                    <div class="pane-title"><span>Model 详情</span></div>
                    <form class="detail-form model-form" method="post" action="/admin/models/update" data-dirty-form>
                      <input type="hidden" name="providerId" value="{{.Selected.Provider.ID}}">
                      <input type="hidden" name="modelId" value="{{.SelectedModel.ID}}">
                      <input type="hidden" name="code" value="{{.SelectedModel.Code}}">
                      <label>名称<input name="name" value="{{.SelectedModel.Name}}" required></label>
                      <label class="check"><input type="checkbox" name="isEnabled" value="1" {{if .SelectedModel.IsEnabled}}checked{{end}}> 启用</label>
                      <span class="actions">
                        <button class="secondary" type="submit" data-save disabled>保存</button>
                        <button class="danger" type="submit" form="delete-model-{{.SelectedModel.ID}}">删除</button>
                      </span>
                    </form>
                  </div>
                  <form id="delete-model-{{.SelectedModel.ID}}" method="post" action="/admin/models/delete">
                    <input type="hidden" name="providerId" value="{{.Selected.Provider.ID}}">
                    <input type="hidden" name="modelId" value="{{.SelectedModel.ID}}">
                  </form>
                {{end}}
              </div>
            {{else}}<div class="empty">这个 provider 还没有 model。</div>{{end}}
          </div>
        </section>
        <section class="panel content">
          <div class="section-title">
            <div>
              <h2>密钥</h2>
              <p class="muted small">列表只显示别名，最多显示 6 行。</p>
            </div>
            <details class="add-panel wide-add">
              <summary>添加密钥</summary>
              <button class="add-cancel" type="button" data-close-add aria-label="取消添加">×</button>
              <form class="form-grid key-form" method="post" action="/admin/keys">
                <input type="hidden" name="providerId" value="{{.Selected.Provider.ID}}">
                <input type="hidden" name="sortOrder" value="{{.Selected.NextKeySortOrder}}">
                <label>别名<input name="alias" placeholder="openai-main-01" required></label>
                <label>Key 明文<input name="secretValue" placeholder="sk-..." required></label>
                <span class="inline-checks">
                  <label class="check"><input type="checkbox" name="isEnabled" value="1" checked> 启用</label>
                  <label class="check"><input type="checkbox" name="isAvailable" value="1" checked> 可用</label>
                </span>
                <button type="submit">添加密钥</button>
              </form>
            </details>
          </div>
          {{if .Selected.Keys}}
            <div class="detail-grid" style="margin-top:12px">
              <div class="pane">
                <div class="pane-title"><span>Key 列表</span></div>
                <form class="compact-list scroll-list sortable-list" method="post" action="/admin/keys/reorder" data-sortable-keys>
                  <input type="hidden" name="providerId" value="{{.Selected.Provider.ID}}">
                  <input type="hidden" name="keyId" value="{{.SelectedKeyID}}">
                  <input type="hidden" name="modelId" value="{{.SelectedModelID}}">
                  {{range .Selected.Keys}}
                    <div class="sortable-item" draggable="true" data-sortable-item>
                      <a class="mini-link {{if eq $.SelectedKeyID .ID}}active{{end}}" draggable="false" href="/admin?providerId={{$.Selected.Provider.ID}}&keyId={{.ID}}&modelId={{$.SelectedModelID}}">
                        {{.Alias}}
                      </a>
                      <button class="ghost copy-key-button" type="button" data-copy-secret-url="/api/keys/{{.ID}}/secret">复制</button>
                      <input type="hidden" name="keyIds" value="{{.ID}}">
                    </div>
                  {{end}}
                </form>
              </div>
              {{if .SelectedKey}}
                <div class="pane">
                  <div class="pane-title"><span>密钥详情</span></div>
                  <form class="detail-form key-form" method="post" action="/admin/keys/update" data-dirty-form>
                    <input type="hidden" name="providerId" value="{{.Selected.Provider.ID}}">
                    <input type="hidden" name="keyId" value="{{.SelectedKey.ID}}">
                    <input type="hidden" name="sortOrder" value="{{.SelectedKey.SortOrder}}">
                    <label>别名<input name="alias" value="{{.SelectedKey.Alias}}" required></label>
                    <label>替换明文<input name="secretValue" placeholder="{{.SelectedKey.MaskedValue}}，留空不替换"></label>
                    <span class="inline-checks">
                      <label class="check"><input type="checkbox" name="isEnabled" value="1" {{if .SelectedKey.IsEnabled}}checked{{end}}> 启用</label>
                      <label class="check"><input type="checkbox" name="isAvailable" value="1" {{if .SelectedKey.IsAvailable}}checked{{end}}> 可用</label>
                    </span>
                    <span class="actions">
                      <button class="ghost" type="button" data-copy-secret-url="/api/keys/{{.SelectedKey.ID}}/secret">复制</button>
                      <button class="secondary" type="submit" data-save disabled>保存</button>
                      <button class="danger" type="submit" form="delete-key-{{.SelectedKey.ID}}">删除</button>
                    </span>
                  </form>
                </div>
                <form id="delete-key-{{.SelectedKey.ID}}" method="post" action="/admin/keys/delete">
                  <input type="hidden" name="providerId" value="{{.Selected.Provider.ID}}">
                  <input type="hidden" name="keyId" value="{{.SelectedKey.ID}}">
                </form>
              {{end}}
            </div>
          {{else}}<div class="empty">这个 provider 还没有 key。</div>{{end}}
        </section>
        </div>
        </div>
      {{else}}
        <section class="panel empty">先在左侧添加一个 provider。</section>
      {{end}}
    </main>
  </div>
  <script>
    function replaceAdminApp(html, url) {
      const parsed = new DOMParser().parseFromString(html, 'text/html');
      const nextApp = parsed.querySelector('.app');
      const currentApp = document.querySelector('.app');
      const nextBrand = parsed.querySelector('.brand');
      const currentBrand = document.querySelector('.brand');
      const nextStyle = parsed.querySelector('style');
      const currentStyle = document.querySelector('style');
      if (!nextApp || !currentApp) {
        window.location.href = url || window.location.href;
        return;
      }
      const currentAside = currentApp.querySelector('aside');
      const asideScrollTop = currentAside ? currentAside.scrollTop : 0;
      if (nextStyle && currentStyle) currentStyle.replaceWith(nextStyle);
      currentApp.replaceWith(nextApp);
      if (nextBrand && currentBrand) currentBrand.replaceWith(nextBrand);
      if (parsed.title) document.title = parsed.title;
      const nextAside = document.querySelector('.app aside');
      if (nextAside) nextAside.scrollTop = asideScrollTop;
      if (url && url !== window.location.href) {
        window.history.pushState({}, '', url);
      }
      initAdminPage();
    }

    async function navigateAdmin(url) {
      const app = document.querySelector('.app');
      if (app) app.setAttribute('aria-busy', 'true');
      try {
        const response = await fetch(url, {
          headers: { 'X-Requested-With': 'fetch' }
        });
        if (!response.ok) throw new Error('Navigation failed');
        const html = await response.text();
        replaceAdminApp(html, response.url);
      } catch (error) {
        window.location.href = url;
      } finally {
        if (app) app.removeAttribute('aria-busy');
      }
    }

    async function submitAdminForm(form, submitter) {
      const app = document.querySelector('.app');
      const button = submitter && submitter.tagName === 'BUTTON' ? submitter : null;
      if (app) app.setAttribute('aria-busy', 'true');
      if (button) button.disabled = true;
      try {
        const response = await fetch(form.action, {
          method: form.method || 'POST',
          body: new URLSearchParams(new FormData(form)),
          headers: {
            'Content-Type': 'application/x-www-form-urlencoded;charset=UTF-8',
            'X-Requested-With': 'fetch'
          }
        });
        const html = await response.text();
        replaceAdminApp(html, response.url);
      } catch (error) {
        form.submit();
      } finally {
        if (app) app.removeAttribute('aria-busy');
        if (button) button.disabled = false;
      }
    }

    async function copyToClipboard(value) {
      try {
        await navigator.clipboard.writeText(value || '');
      } catch (error) {
        const fallback = document.createElement('textarea');
        fallback.value = value || '';
        fallback.style.position = 'fixed';
        fallback.style.left = '-9999px';
        document.body.appendChild(fallback);
        fallback.select();
        document.execCommand('copy');
        fallback.remove();
      }
    }

    function initAdminPage() {
    document.querySelectorAll('[data-dirty-form]').forEach((form) => {
      const save = form.querySelector('[data-save]');
      if (!save) return;
      const snapshot = new FormData(form);
      const initial = JSON.stringify(Array.from(snapshot.entries()));
      const sync = () => {
        save.disabled = JSON.stringify(Array.from(new FormData(form).entries())) === initial;
      };
      form.addEventListener('input', sync);
      form.addEventListener('change', sync);
      sync();
    });
    document.querySelectorAll('[data-close-add]').forEach((button) => {
      button.addEventListener('click', () => {
        const details = button.closest('details');
        if (!details) return;
        const form = details.querySelector('form');
        if (form) form.reset();
        details.open = false;
      });
    });
    document.querySelectorAll('[data-copy-text]').forEach((button) => {
      button.addEventListener('click', async () => {
        await copyToClipboard(button.dataset.copyText || '');
        button.setAttribute('aria-label', '已复制');
        window.setTimeout(() => button.setAttribute('aria-label', '复制失败信息'), 1200);
      });
    });
    document.querySelectorAll('[data-copy-secret-url]').forEach((button) => {
      button.addEventListener('click', async () => {
        const originalText = button.textContent;
        button.disabled = true;
        try {
          const response = await fetch(button.dataset.copySecretUrl, {
            headers: { 'Accept': 'application/json', 'X-Requested-With': 'fetch' }
          });
          if (!response.ok) throw new Error('Copy failed');
          const payload = await response.json();
          await copyToClipboard(payload.secretValue || '');
          button.textContent = '已复制';
        } catch (error) {
          button.textContent = '复制失败';
        } finally {
          window.setTimeout(() => {
            button.textContent = originalText;
            button.disabled = false;
          }, 1200);
        }
      });
    });
    document.querySelectorAll('[data-history-chart]').forEach((chart) => {
      const points = Array.from(chart.querySelectorAll('[data-chart-point]')).map((point) => ({
        node: point,
        x: Number(point.dataset.x),
        y: Number(point.dataset.y),
        date: point.dataset.date || '',
        total: point.dataset.total || '0',
        failed: point.dataset.failed || '0'
      }));
      const hoverLine = chart.querySelector('[data-hover-line]');
      const hoverTip = chart.querySelector('[data-hover-tip]');
      const tipDate = chart.querySelector('[data-tip-date]');
      const tipTotal = chart.querySelector('[data-tip-total]');
      const tipFailed = chart.querySelector('[data-tip-failed]');
      if (!points.length || !hoverLine || !hoverTip) return;
      const hideHover = () => {
        hoverLine.setAttribute('visibility', 'hidden');
        hoverTip.setAttribute('visibility', 'hidden');
        points.forEach((point) => point.node.setAttribute('r', '3.5'));
      };
      chart.addEventListener('pointermove', (event) => {
        const ctm = chart.getScreenCTM();
        if (!ctm) return;
        const svgPoint = chart.createSVGPoint();
        svgPoint.x = event.clientX;
        svgPoint.y = event.clientY;
        const cursor = svgPoint.matrixTransform(ctm.inverse());
        let nearest = points[0];
        for (const point of points) {
          if (Math.abs(point.x - cursor.x) < Math.abs(nearest.x - cursor.x)) nearest = point;
        }
        points.forEach((point) => point.node.setAttribute('r', point === nearest ? '5.5' : '3.5'));
        hoverLine.setAttribute('x1', nearest.x);
        hoverLine.setAttribute('x2', nearest.x);
        hoverLine.setAttribute('visibility', 'visible');
        tipDate.textContent = nearest.date;
        tipTotal.textContent = '总调用：' + nearest.total + ' 次';
        tipFailed.textContent = '失败调用：' + nearest.failed + ' 次';
        const tipX = Math.min(Math.max(nearest.x + 12, 78), 748);
        const tipY = Math.max(34, nearest.y - 70);
        hoverTip.setAttribute('transform', 'translate(' + tipX + ' ' + tipY + ')');
        hoverTip.setAttribute('visibility', 'visible');
      });
      chart.addEventListener('pointerleave', hideHover);
    });
    document.querySelectorAll('[data-check-all], [data-check-none]').forEach((button) => {
      button.addEventListener('click', () => {
        const form = button.closest('form');
        if (!form) return;
        const checked = button.hasAttribute('data-check-all');
        form.querySelectorAll('input[name="allowedModelIds"], input[name="allowedKeyIds"]').forEach((checkbox) => {
          checkbox.checked = checked;
        });
        form.dispatchEvent(new Event('change', { bubbles: true }));
      });
    });
    document.querySelectorAll('[data-sortable-keys]').forEach((form) => {
      let dragged = null;
      let changed = false;

      form.querySelectorAll('[data-sortable-item]').forEach((item) => {
        item.addEventListener('dragstart', (event) => {
          dragged = item;
          changed = false;
          item.classList.add('dragging');
          event.dataTransfer.effectAllowed = 'move';
          event.dataTransfer.setData('text/plain', item.querySelector('input[name="keyIds"]').value);
        });
        item.addEventListener('dragend', () => {
          item.classList.remove('dragging');
          if (changed) form.requestSubmit();
          dragged = null;
        });
      });

      form.addEventListener('dragover', (event) => {
        if (!dragged) return;
        event.preventDefault();
        const target = event.target.closest('[data-sortable-item]');
        if (!target || target === dragged || !form.contains(target)) return;
        const rect = target.getBoundingClientRect();
        const after = event.clientY > rect.top + rect.height / 2;
        form.insertBefore(dragged, after ? target.nextSibling : target);
        changed = true;
      });
      form.addEventListener('drop', (event) => {
        if (dragged) event.preventDefault();
      });
    });
    }
    if (!window.__keychainAdminNavigationReady) {
      window.__keychainAdminNavigationReady = true;
      document.addEventListener('click', (event) => {
        const link = event.target.closest('a[href]');
        if (!link) return;
        if (event.defaultPrevented || event.metaKey || event.ctrlKey || event.shiftKey || event.altKey || event.button !== 0) return;
        const url = new URL(link.href, window.location.href);
        if (url.origin !== window.location.origin || !url.pathname.startsWith('/admin')) return;
        event.preventDefault();
        navigateAdmin(url.href);
      });
      document.addEventListener('submit', (event) => {
        const form = event.target;
        if (!form.matches('.app form[method="post"]')) return;
        event.preventDefault();
        submitAdminForm(form, event.submitter);
      });
      window.addEventListener('popstate', () => {
        navigateAdmin(window.location.href);
      });
    }
    initAdminPage();
  </script>
</body>
</html>`))

type adminPageData struct {
	Username        string
	Error           string
	Providers       []providerNavItem
	Selected        *providerPanel
	SelectedKey     *admin.APIKey
	SelectedKeyID   string
	SelectedModel   *admin.Model
	SelectedModelID string
}

type providerNavItem struct {
	Provider   admin.Provider
	IsActive   bool
	ModelCount int
	KeyCount   int
}

type providerPanel struct {
	Provider         admin.Provider
	Models           []admin.Model
	Keys             []admin.APIKey
	NextKeySortOrder int
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
		registerHistoryRoutes(mux, authService, store)
		mux.HandleFunc("POST /admin/providers", formCreateProviderHandler(store))
		mux.HandleFunc("POST /admin/providers/update", formUpdateProviderHandler(store))
		mux.HandleFunc("POST /admin/providers/delete", formDeleteProviderHandler(store))
		mux.HandleFunc("POST /admin/models", formCreateModelHandler(store))
		mux.HandleFunc("POST /admin/models/update", formUpdateModelHandler(store))
		mux.HandleFunc("POST /admin/models/delete", formDeleteModelHandler(store))
		mux.HandleFunc("POST /admin/keys", formCreateKeyHandler(store))
		mux.HandleFunc("POST /admin/keys/update", formUpdateKeyHandler(store))
		mux.HandleFunc("POST /admin/keys/reorder", formReorderKeysHandler(store))
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
		data := adminPageData{Username: session.Username, Error: r.URL.Query().Get("error"), SelectedKeyID: r.URL.Query().Get("keyId"), SelectedModelID: r.URL.Query().Get("modelId")}
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
		code := r.FormValue("code")
		if code == "" {
			code = r.FormValue("name")
		}
		provider, err := store.CreateProvider(r.Context(), admin.CreateProviderInput{
			Name:             r.FormValue("name"),
			Code:             code,
			IsEnabled:        r.FormValue("isEnabled") == "1",
			RotationStrategy: r.FormValue("rotationStrategy"),
		})
		if err != nil {
			redirectAdminError(w, r, "添加 provider 失败："+err.Error())
			return
		}
		redirectToProvider(w, r, provider.ID)
	}
}

func formUpdateProviderHandler(store *admin.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			redirectAdminError(w, r, "provider 表单格式不正确")
			return
		}
		providerID := r.FormValue("providerId")
		code := r.FormValue("code")
		if code == "" {
			code = r.FormValue("name")
		}
		_, err := store.UpdateProvider(r.Context(), providerID, admin.UpdateProviderInput{
			Name:             r.FormValue("name"),
			Code:             code,
			IsEnabled:        r.FormValue("isEnabled") == "1",
			RotationStrategy: r.FormValue("rotationStrategy"),
		})
		if err != nil {
			redirectAdminError(w, r, "保存 provider 失败："+err.Error())
			return
		}
		redirectToProvider(w, r, providerID)
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
		code := r.FormValue("code")
		if code == "" {
			code = r.FormValue("name")
		}
		model, err := store.CreateModel(r.Context(), admin.CreateModelInput{
			ProviderID: providerID,
			Name:       r.FormValue("name"),
			Code:       code,
			IsEnabled:  r.FormValue("isEnabled") == "1",
		})
		if err != nil {
			redirectAdminError(w, r, "添加 model 失败："+err.Error())
			return
		}
		http.Redirect(w, r, "/admin?providerId="+template.URLQueryEscaper(providerID)+"&modelId="+template.URLQueryEscaper(model.ID), http.StatusFound)
	}
}

func formUpdateModelHandler(store *admin.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			redirectAdminError(w, r, "model 表单格式不正确")
			return
		}
		providerID := r.FormValue("providerId")
		code := r.FormValue("code")
		if code == "" {
			code = r.FormValue("name")
		}
		_, err := store.UpdateModel(r.Context(), r.FormValue("modelId"), admin.UpdateModelInput{
			Name:      r.FormValue("name"),
			Code:      code,
			IsEnabled: r.FormValue("isEnabled") == "1",
		})
		if err != nil {
			redirectAdminError(w, r, "保存 model 失败："+err.Error())
			return
		}
		http.Redirect(w, r, "/admin?providerId="+template.URLQueryEscaper(providerID)+"&modelId="+template.URLQueryEscaper(r.FormValue("modelId")), http.StatusFound)
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
		redirectToProvider(w, r, providerID)
	}
}

func formCreateKeyHandler(store *admin.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			redirectAdminError(w, r, "key 表单格式不正确")
			return
		}
		providerID := r.FormValue("providerId")
		apiKey, err := store.CreateAPIKey(r.Context(), admin.CreateAPIKeyInput{
			ProviderID:  providerID,
			Alias:       r.FormValue("alias"),
			SecretValue: r.FormValue("secretValue"),
			IsEnabled:   r.FormValue("isEnabled") == "1",
			IsAvailable: r.FormValue("isAvailable") == "1",
			SortOrder:   parseOptionalInt(r.FormValue("sortOrder")),
		})
		if err != nil {
			redirectAdminError(w, r, "添加密钥失败："+err.Error())
			return
		}
		http.Redirect(w, r, "/admin?providerId="+template.URLQueryEscaper(providerID)+"&keyId="+template.URLQueryEscaper(apiKey.ID), http.StatusFound)
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
			redirectAdminError(w, r, "保存密钥失败："+err.Error())
			return
		}
		http.Redirect(w, r, "/admin?providerId="+template.URLQueryEscaper(providerID)+"&keyId="+template.URLQueryEscaper(r.FormValue("keyId")), http.StatusFound)
	}
}

func formReorderKeysHandler(store *admin.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			redirectAdminError(w, r, "key 排序表单格式不正确")
			return
		}
		providerID := r.FormValue("providerId")
		if err := store.ReorderAPIKeys(r.Context(), providerID, r.Form["keyIds"]); err != nil {
			redirectAdminError(w, r, "保存密钥排序失败："+err.Error())
			return
		}

		target := "/admin?providerId=" + template.URLQueryEscaper(providerID)
		if keyID := r.FormValue("keyId"); keyID != "" {
			target += "&keyId=" + template.URLQueryEscaper(keyID)
		}
		if modelID := r.FormValue("modelId"); modelID != "" {
			target += "&modelId=" + template.URLQueryEscaper(modelID)
		}
		http.Redirect(w, r, target, http.StatusFound)
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
			redirectAdminError(w, r, "删除密钥失败："+err.Error())
			return
		}
		redirectToProvider(w, r, providerID)
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
			if data.SelectedKeyID == "" && len(keys) > 0 {
				data.SelectedKeyID = keys[0].ID
			}
			if data.SelectedModelID == "" && len(models) > 0 {
				data.SelectedModelID = models[0].ID
			}
			nextKeySortOrder := 1
			for _, key := range keys {
				if key.SortOrder >= nextKeySortOrder {
					nextKeySortOrder = key.SortOrder + 1
				}
			}
			selectedProvider := providerPanel{Provider: provider, Models: models, Keys: keys, NextKeySortOrder: nextKeySortOrder}
			data.Selected = &selectedProvider
			for _, key := range keys {
				if key.ID == data.SelectedKeyID {
					selectedKey := key
					data.SelectedKey = &selectedKey
					break
				}
			}
			for _, model := range models {
				if model.ID == data.SelectedModelID {
					selectedModel := model
					data.SelectedModel = &selectedModel
					break
				}
			}
		}
	}
	return data, nil
}

func redirectToProvider(w http.ResponseWriter, r *http.Request, providerID string) {
	http.Redirect(w, r, "/admin?providerId="+template.URLQueryEscaper(providerID), http.StatusFound)
}

func redirectAdminError(w http.ResponseWriter, r *http.Request, message string) {
	http.Redirect(w, r, "/admin?error="+template.URLQueryEscaper(message), http.StatusFound)
}
