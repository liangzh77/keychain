package server

import (
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

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
    :root { --bg: #f4f1ea; --surface: #fffcf5; --surface-muted: #f7f2ea; --line: #d7cdbf; --line-soft: #ece4d8; --text: #242824; --muted: #6f7169; --accent: #31594a; --accent-soft: #e8efe8; --accent-line: #9fb9aa; --secondary: #5a5448; --danger: #9b3d35; }
    * { box-sizing: border-box; }
    body { margin: 0; font-family: ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; background: var(--bg); color: var(--text); }
    header { height: 56px; display: flex; align-items: center; justify-content: space-between; padding: 0 20px; background: var(--surface); border-bottom: 1px solid var(--line); }
    header form { margin: 0; }
    .brand { display: flex; align-items: baseline; gap: 10px; }
    .brand strong { font-size: 16px; }
    .app { height: calc(100vh - 56px); display: grid; grid-template-columns: 300px minmax(0, 1fr); overflow: hidden; }
    .app[aria-busy="true"] { cursor: progress; }
    aside { border-right: 1px solid var(--line); background: var(--surface-muted); overflow: auto; padding: 16px; }
    main { overflow: auto; padding: 16px 20px 24px; }
    h1, h2, h3 { margin: 0; line-height: 1.2; }
    h1 { font-size: 22px; }
    h2 { font-size: 18px; }
    h3 { font-size: 15px; }
    p { margin: 0; }
    .muted { color: var(--muted); }
    .small { font-size: 12px; }
    .panel { background: var(--surface); border: 1px solid var(--line); border-radius: 8px; }
    .panel-pad, .content { padding: 14px; }
    .stack { display: grid; gap: 12px; }
    .topline { display: flex; justify-content: space-between; align-items: start; gap: 16px; margin-bottom: 14px; }
    .tab { padding: 8px 10px; color: var(--muted); text-decoration: none; border-radius: 6px; font-weight: 700; font-size: 14px; }
    .tab.active, .tab:hover { background: var(--accent-soft); color: var(--accent); }
    .channel-list, .user-list, .row-list { display: grid; gap: 6px; margin-top: 10px; }
    .channel-link, .user-link { display: block; padding: 10px 12px; border: 1px solid transparent; border-radius: 7px; color: inherit; text-decoration: none; }
    .user-link { min-height: 40px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    .channel-link:hover, .user-link:hover { background: #f0ebe2; }
    .channel-link.active, .user-link.active { background: var(--accent-soft); border-color: var(--accent-line); box-shadow: inset 3px 0 0 var(--accent); }
    .meta-row { display: flex; justify-content: space-between; gap: 12px; align-items: center; }
    .mono { font-family: ui-monospace, SFMono-Regular, Consolas, monospace; color: var(--muted); font-size: 12px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    form { display: grid; gap: 10px; }
    form[id^="delete-"] { display: none; }
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
    details.add-panel.wide-add > summary { min-width: 132px; padding: 0 18px; }
    details.add-panel .add-cancel { display: none; position: absolute; top: 7px; right: 7px; width: 24px; height: 24px; padding: 0; border-radius: 999px; background: #efe9df; color: #5a5448; font-size: 18px; line-height: 1; }
    details.add-panel[open] .add-cancel { display: inline-flex; align-items: center; justify-content: center; }
    .form-grid { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 10px; align-items: end; }
    .form-grid > *, .detail-form > *, .perm-form > * { min-width: 0; }
    .channel-form { grid-template-columns: repeat(2, minmax(0, 1fr)); }
    .user-form { grid-template-columns: minmax(0, 1fr) auto auto; }
    .access-grid { display: grid; grid-template-columns: repeat(2, minmax(360px, 1fr)); gap: 12px; align-items: start; }
    .split { display: grid; grid-template-columns: minmax(132px, 200px) minmax(0, 1fr); gap: 12px; align-items: start; }
    .detail-form { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)) auto; gap: 10px; align-items: end; }
    .section-title { display: grid; grid-template-columns: minmax(0, 1fr) auto; align-items: start; gap: 12px; }
    .section-title h2 { font-size: 18px; line-height: 1.2; font-weight: 700; }
    .section-title p { font-size: 12px; line-height: 1.2; }
    .section-title > details.add-panel[open] { grid-column: 1 / -1; }
    .section-title > details.add-panel[open] > summary { width: max-content; min-width: 132px; margin-left: auto; margin-right: 34px; }
    .scroll-list { height: 276px; overflow-y: auto; padding-right: 2px; align-content: start; }
    .mini-link { display: block; min-height: 40px; padding: 9px 10px; border: 1px solid var(--line-soft); border-radius: 7px; color: inherit; text-decoration: none; background: #fffdf8; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    .mini-link:hover { background: #f6f1e8; }
    .mini-link.active { border-color: var(--accent-line); background: var(--accent-soft); box-shadow: inset 3px 0 0 var(--accent); }
    .pane { display: grid; gap: 8px; min-width: 0; }
    .pane-title { display: flex; align-items: center; justify-content: space-between; min-height: 28px; padding: 0 2px; color: #5f6259; font-size: 12px; font-weight: 800; letter-spacing: 0; }
    .pane-title::before { content: ""; width: 4px; height: 16px; border-radius: 999px; background: var(--accent); }
    .pane-title span { margin-right: auto; margin-left: 8px; }
    .perm-form { display: grid; grid-template-columns: minmax(0, 1fr) auto auto; gap: 10px; align-items: end; }
    .actions { display: flex; justify-content: flex-end; gap: 6px; flex-wrap: wrap; }
    .channel-actions { grid-column: 1 / -1; display: flex; align-items: center; justify-content: flex-start; gap: 8px; flex-wrap: wrap; padding-top: 2px; }
    .detail-form.user-form { grid-template-columns: minmax(0, 1fr) auto; }
    .detail-form.user-form .actions { grid-column: 1 / -1; }
    .user-actions { display: flex; align-items: center; justify-content: flex-end; gap: 8px; flex-wrap: wrap; }
    .permission-zone { margin-top: 12px; padding-top: 12px; border-top: 1px solid var(--line-soft); }
    .model-permission-form { display: grid; gap: 10px; }
    .permission-actions { display: flex; align-items: center; justify-content: space-between; gap: 8px; flex-wrap: wrap; }
    .bulk-actions { display: flex; align-items: center; gap: 6px; flex-wrap: wrap; }
    .model-check-list { display: grid; gap: 8px; max-height: 276px; overflow-y: auto; padding-right: 2px; }
    .model-check { display: flex; align-items: center; justify-content: space-between; gap: 12px; min-height: 38px; padding: 8px 10px; border: 1px solid var(--line-soft); border-radius: 7px; background: #fffdf8; font-size: 13px; font-weight: 700; color: var(--text); }
    .model-check span { min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    .notice { margin-bottom: 14px; padding: 10px 12px; border-radius: 6px; background: #fff6df; color: #7a5a22; }
    .empty { padding: 28px; text-align: center; color: var(--muted); }
    .tag { display: inline-flex; align-items: center; padding: 2px 8px; border-radius: 999px; background: var(--accent-soft); color: var(--accent); font-size: 12px; font-weight: 700; }
    .tag.off { background: #e9e4db; color: #746f66; }
    @media (max-width: 1320px) { .access-grid { grid-template-columns: 1fr; } .channel-form, .perm-form, .detail-form.user-form { grid-template-columns: repeat(2, minmax(0, 1fr)); } .actions, .user-actions { justify-content: flex-start; } }
    @media (max-width: 980px) { .app { grid-template-columns: 1fr; height: auto; overflow: visible; } aside, main { overflow: visible; } .form-grid, .detail-form, .channel-form, .user-form, .perm-form { grid-template-columns: 1fr; } .actions, .user-actions { justify-content: flex-start; } }
    @media (max-width: 760px) { .split, .section-title { grid-template-columns: 1fr; } .topline { flex-direction: column; align-items: stretch; } .section-title > details.add-panel > summary { width: 100%; margin-left: 0; } }
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
        <details class="panel panel-pad add-panel">
          <summary>添加渠道</summary>
          <button class="add-cancel" type="button" data-close-add aria-label="取消添加">×</button>
          <form method="post" action="/admin/channels">
            <label>名称<input name="name" placeholder="本校默认渠道" required></label>
            <label>默认权限
              <select name="defaultPermissionMode">
                <option value="DENY">默认关闭</option>
                <option value="ALLOW">默认打开</option>
              </select>
            </label>
            <label>用户系统
              <select name="userManagementMode">
                <option value="EXTERNAL_MANAGED">外部自有用户系统</option>
                <option value="KEYCHAIN_HOSTED">Keychain 托管用户系统</option>
              </select>
            </label>
            <label class="check"><input type="checkbox" name="isEnabled" value="1" checked> 启用渠道</label>
            <button type="submit">添加渠道</button>
          </form>
        </details>
        <section>
          <h2>Channels</h2>
          <div class="channel-list">
            {{range .Channels}}
              <a class="channel-link {{if .IsActive}}active{{end}}" href="/admin/access?channelId={{.Channel.ID}}">
                <div class="meta-row"><strong>{{.Channel.Name}}</strong>{{if .Channel.IsEnabled}}<span class="tag">启用</span>{{else}}<span class="tag off">停用</span>{{end}}</div>
                <div class="meta-row"><span class="muted small">{{.UserCount}}个用户 · {{.DefaultPermissionText}} · {{.UserManagementText}}</span></div>
              </a>
            {{else}}
              <div class="panel empty">还没有渠道。</div>
            {{end}}
          </div>
        </section>
      </div>
    </aside>
    <main>
      {{if .Error}}<div class="notice">{{.Error}}</div>{{end}}
      {{if .Selected}}
        <div class="access-grid">
          <div class="stack">
            <section class="panel content">
              <div class="topline">
                <div>
                  <h2>渠道详情</h2>
                </div>
                {{if .Selected.Channel.IsEnabled}}<span class="tag">启用</span>{{else}}<span class="tag off">停用</span>{{end}}
              </div>
              <form class="form-grid channel-form" method="post" action="/admin/channels/update" data-dirty-form>
                <input type="hidden" name="channelId" value="{{.Selected.Channel.ID}}">
                <input type="hidden" name="code" value="{{.Selected.Channel.Code}}">
                <label>名称<input name="name" value="{{.Selected.Channel.Name}}" required></label>
                <label>默认权限
                  <select name="defaultPermissionMode">
                    <option value="DENY" {{if eq .Selected.Channel.DefaultPermissionMode "DENY"}}selected{{end}}>默认关闭</option>
                    <option value="ALLOW" {{if eq .Selected.Channel.DefaultPermissionMode "ALLOW"}}selected{{end}}>默认打开</option>
                  </select>
                </label>
                <label>用户系统
                  <select name="userManagementMode">
                    <option value="EXTERNAL_MANAGED" {{if eq .Selected.Channel.UserManagementMode "EXTERNAL_MANAGED"}}selected{{end}}>外部自有用户系统</option>
                    <option value="KEYCHAIN_HOSTED" {{if eq .Selected.Channel.UserManagementMode "KEYCHAIN_HOSTED"}}selected{{end}}>Keychain 托管用户系统</option>
                  </select>
                </label>
                <span class="channel-actions">
                  <label class="check"><input type="checkbox" name="isEnabled" value="1" {{if .Selected.Channel.IsEnabled}}checked{{end}}> 启用</label>
                  <button class="secondary" type="submit" data-save disabled>保存渠道</button>
                  <button class="danger" type="submit" form="delete-channel-{{.Selected.Channel.ID}}">删除</button>
                </span>
              </form>
              <form id="delete-channel-{{.Selected.Channel.ID}}" method="post" action="/admin/channels/delete">
                <input type="hidden" name="channelId" value="{{.Selected.Channel.ID}}">
              </form>
            </section>
          <section class="panel content">
            <div class="section-title">
              <div>
                <h2>Providers 授权</h2>
                <p class="muted small">点击一个 provider/model，右侧修改该渠道的默认允许状态。</p>
              </div>
            </div>
            {{if .Selected.ChannelPermissions}}
              <div class="split" style="margin-top:12px">
                <div class="pane">
                  <div class="pane-title"><span>授权列表</span></div>
                  <div class="row-list scroll-list">
                    {{range .Selected.ChannelPermissions}}
                      <a class="mini-link {{if and (eq $.SelectedPermProviderID .ProviderID) (eq $.SelectedPermModelID .ModelID)}}active{{end}}" href="/admin/access?channelId={{$.Selected.Channel.ID}}&userId={{$.SelectedUserID}}&permProviderId={{.ProviderID}}&permModelId={{.ModelID}}&userPermProviderId={{$.SelectedUserPermProviderID}}&userPermModelId={{$.SelectedUserPermModelID}}">
                        {{.ProviderName}} / {{.ModelName}}
                      </a>
                    {{end}}
                  </div>
                </div>
                {{if .SelectedChannelPermission}}
                  <div class="pane">
                    <div class="pane-title"><span>授权详情</span></div>
                    <form class="perm-form" method="post" action="/admin/channel-permissions" data-dirty-form>
                      <input type="hidden" name="channelId" value="{{.Selected.Channel.ID}}">
                      <input type="hidden" name="providerId" value="{{.SelectedChannelPermission.ProviderID}}">
                      <input type="hidden" name="modelId" value="{{.SelectedChannelPermission.ModelID}}">
                      <label>授权对象<input value="{{.SelectedChannelPermission.ProviderName}} / {{.SelectedChannelPermission.ModelName}}" disabled></label>
                      <label class="check"><input type="checkbox" name="allowed" value="1" {{if .SelectedChannelPermission.DefaultAllowed}}checked{{end}}> 默认允许</label>
                      <button class="secondary" type="submit" data-save disabled>保存授权</button>
                    </form>
                  </div>
                {{end}}
              </div>
            {{else}}<div class="empty">还没有 provider/model 可授权。</div>{{end}}
          </section>
          </div>
          <section class="panel content">
            <div class="section-title">
              <div>
                <h2>用户授权</h2>
                <p class="muted small">用户由外部应用通过 API 创建；这里仅维护已有用户和授权。</p>
              </div>
            </div>
            <div class="split" style="margin-top:12px">
              <div class="pane">
                <div class="pane-title"><span>用户列表</span></div>
                <div class="user-list scroll-list">
                  {{range .Selected.Users}}
                    <a class="user-link {{if eq $.SelectedUserID .ID}}active{{end}}" href="/admin/access?channelId={{$.Selected.Channel.ID}}&userId={{.ID}}&permProviderId={{$.SelectedPermProviderID}}&permModelId={{$.SelectedPermModelID}}">
                      {{.DisplayName}}
                    </a>
                  {{else}}
                    <div class="empty">这个渠道还没有用户。</div>
                  {{end}}
                </div>
              </div>
              <div class="stack">
                {{if .SelectedUser}}
                  <div class="pane-title"><span>用户详情</span></div>
                  <form class="detail-form user-form" method="post" action="/admin/users/update" data-dirty-form>
                    <input type="hidden" name="channelId" value="{{.Selected.Channel.ID}}">
                    <input type="hidden" name="userId" value="{{.SelectedUser.ID}}">
                    <input type="hidden" name="externalUserId" value="{{.SelectedUser.ExternalUserID}}">
                    <label>用户名称<input name="displayName" value="{{.SelectedUser.DisplayName}}" required></label>
                    <span class="user-actions">
                      <label class="check"><input type="checkbox" name="isEnabled" value="1" {{if .SelectedUser.IsEnabled}}checked{{end}}> 启用</label>
                      <button class="secondary" type="submit" data-save disabled>保存用户</button>
                    </span>
                  </form>
                  {{if .UserPermissionProviders}}
                    <div class="split permission-zone">
                      <div class="pane">
                        <div class="pane-title"><span>Provider 列表</span></div>
                        <div class="row-list scroll-list">
                          {{range .UserPermissionProviders}}
                            <a class="mini-link {{if eq $.SelectedUserPermProviderID .ProviderID}}active{{end}}" href="/admin/access?channelId={{$.Selected.Channel.ID}}&userId={{$.SelectedUserID}}&permProviderId={{$.SelectedPermProviderID}}&permModelId={{$.SelectedPermModelID}}&userPermProviderId={{.ProviderID}}">
                              {{.ProviderName}}
                            </a>
                          {{end}}
                        </div>
                      </div>
                      {{if .SelectedUserProviderPermissions}}
                        <div class="pane">
                          <div class="pane-title"><span>Model 授权详情</span></div>
                          <form class="model-permission-form" method="post" action="/admin/user-permissions" data-dirty-form>
                            <input type="hidden" name="channelId" value="{{.Selected.Channel.ID}}">
                            <input type="hidden" name="userId" value="{{.SelectedUser.ID}}">
                            <input type="hidden" name="providerId" value="{{.SelectedUserPermProviderID}}">
                            <div class="model-check-list">
                              {{range .SelectedUserProviderPermissions}}
                                <label class="model-check">
                                  <span>{{.ModelName}}</span>
                                  <input type="hidden" name="modelIds" value="{{.ModelID}}">
                                  <input type="checkbox" name="allowedModelIds" value="{{.ModelID}}" {{if .Allowed}}checked{{end}}>
                                </label>
                              {{end}}
                            </div>
                            <span class="permission-actions">
                              <span class="bulk-actions">
                                <button class="ghost" type="button" data-check-all>全选</button>
                                <button class="ghost" type="button" data-check-none>不选</button>
                              </span>
                              <button class="secondary" type="submit" data-save disabled>保存授权</button>
                            </span>
                          </form>
                          <div class="pane-title" style="margin-top:12px"><span>Key 授权详情</span></div>
                          {{if .SelectedUserProviderKeyPermissions}}
                            <form class="model-permission-form" method="post" action="/admin/user-key-permissions" data-dirty-form>
                              <input type="hidden" name="channelId" value="{{.Selected.Channel.ID}}">
                              <input type="hidden" name="userId" value="{{.SelectedUser.ID}}">
                              <input type="hidden" name="providerId" value="{{.SelectedUserPermProviderID}}">
                              <div class="model-check-list">
                                {{range .SelectedUserProviderKeyPermissions}}
                                  <label class="model-check">
                                    <span>{{.KeyAlias}}</span>
                                    <input type="hidden" name="keyIds" value="{{.KeyID}}">
                                    <input type="checkbox" name="allowedKeyIds" value="{{.KeyID}}" {{if .Allowed}}checked{{end}}>
                                  </label>
                                {{end}}
                              </div>
                              <span class="permission-actions">
                                <span class="bulk-actions">
                                  <button class="ghost" type="button" data-check-all>全选</button>
                                  <button class="ghost" type="button" data-check-none>不选</button>
                                </span>
                                <button class="secondary" type="submit" data-save disabled>保存 Key 授权</button>
                              </span>
                            </form>
                          {{else}}<div class="empty">这个 provider 还没有 key。</div>{{end}}
                        </div>
                      {{end}}
                    </div>
                  {{else}}<div class="empty">还没有 provider/model 可授权。</div>{{end}}
                {{else}}<div class="empty">先选择一个用户。</div>{{end}}
              </div>
            </div>
          </section>
        </div>
      {{else}}
        <section class="panel empty">先添加或选择一个渠道。</section>
      {{end}}
    </main>
  </div>
  <script>
    function replaceAdminApp(html, url) {
      const parsed = new DOMParser().parseFromString(html, 'text/html');
      const nextApp = parsed.querySelector('.app');
      const currentApp = document.querySelector('.app');
      if (!nextApp || !currentApp) {
        window.location.href = url || window.location.href;
        return;
      }
      currentApp.replaceWith(nextApp);
      if (url && url !== window.location.href) {
        window.history.pushState({}, '', url);
      }
      initAdminPage();
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

    function initAdminPage() {
    document.querySelectorAll('.app form[method="post"]').forEach((form) => {
      form.addEventListener('submit', (event) => {
        event.preventDefault();
        submitAdminForm(form, event.submitter);
      });
    });
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
    }
    initAdminPage();
  </script>
</body>
</html>`))

type accessPageData struct {
	Username                           string
	Error                              string
	Channels                           []channelNavItem
	Selected                           *channelPanel
	SelectedUser                       *admin.User
	SelectedUserID                     string
	UserPermissions                    []admin.UserPermissionRow
	UserPermissionProviders            []userPermissionProvider
	SelectedUserProviderPermissions    []admin.UserPermissionRow
	SelectedUserProviderKeyPermissions []admin.UserKeyPermissionRow
	SelectedChannelPermission          *admin.ChannelPermissionRow
	SelectedPermProviderID             string
	SelectedPermModelID                string
	SelectedUserPermission             *admin.UserPermissionRow
	SelectedUserPermProviderID         string
	SelectedUserPermModelID            string
}

type channelNavItem struct {
	Channel   admin.Channel
	IsActive  bool
	UserCount int
}

func (item channelNavItem) DefaultPermissionText() string {
	if item.Channel.DefaultPermissionMode == "ALLOW" {
		return "启用"
	}
	return "停用"
}

func (item channelNavItem) UserManagementText() string {
	if item.Channel.UserManagementMode == "KEYCHAIN_HOSTED" {
		return "托管用户系统"
	}
	return "外部用户系统"
}

type userPermissionProvider struct {
	ProviderID   string
	ProviderName string
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
	mux.HandleFunc("POST /admin/channels/delete", formDeleteChannelHandler(store))
	mux.HandleFunc("POST /admin/users/update", formUpdateUserHandler(store))
	mux.HandleFunc("POST /admin/channel-permissions", formSetChannelPermissionHandler(store))
	mux.HandleFunc("POST /admin/user-permissions", formSetUserPermissionHandler(store))
	mux.HandleFunc("POST /admin/user-key-permissions", formSetUserKeyPermissionHandler(store))
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
		data := accessPageData{
			Username:                   session.Username,
			Error:                      r.URL.Query().Get("error"),
			SelectedUserID:             r.URL.Query().Get("userId"),
			SelectedPermProviderID:     r.URL.Query().Get("permProviderId"),
			SelectedPermModelID:        r.URL.Query().Get("permModelId"),
			SelectedUserPermProviderID: r.URL.Query().Get("userPermProviderId"),
			SelectedUserPermModelID:    r.URL.Query().Get("userPermModelId"),
		}
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
			if data.SelectedPermProviderID == "" && len(channelPermissions) > 0 {
				data.SelectedPermProviderID = channelPermissions[0].ProviderID
				data.SelectedPermModelID = channelPermissions[0].ModelID
			}
			selected := channelPanel{Channel: channel, Users: users, ChannelPermissions: channelPermissions}
			data.Selected = &selected
			for _, permission := range channelPermissions {
				if permission.ProviderID == data.SelectedPermProviderID && permission.ModelID == data.SelectedPermModelID {
					selectedPermission := permission
					data.SelectedChannelPermission = &selectedPermission
					break
				}
			}
			for _, user := range users {
				if user.ID == data.SelectedUserID {
					selectedUser := user
					data.SelectedUser = &selectedUser
					data.UserPermissions, err = store.ListUserPermissionRows(r.Context(), user.ID)
					if err != nil {
						return data, err
					}
					data.UserPermissionProviders = buildUserPermissionProviders(data.UserPermissions)
					if data.SelectedUserPermProviderID == "" && len(data.UserPermissionProviders) > 0 {
						data.SelectedUserPermProviderID = data.UserPermissionProviders[0].ProviderID
					}
					data.SelectedUserProviderPermissions = filterUserPermissionsByProvider(data.UserPermissions, data.SelectedUserPermProviderID)
					data.SelectedUserProviderKeyPermissions, err = store.ListUserKeyPermissionRows(r.Context(), user.ID, data.SelectedUserPermProviderID)
					if err != nil {
						return data, err
					}
					if data.SelectedUserPermModelID == "" && len(data.SelectedUserProviderPermissions) > 0 {
						data.SelectedUserPermModelID = data.SelectedUserProviderPermissions[0].ModelID
					}
					for _, permission := range data.UserPermissions {
						if permission.ProviderID == data.SelectedUserPermProviderID && permission.ModelID == data.SelectedUserPermModelID {
							selectedPermission := permission
							data.SelectedUserPermission = &selectedPermission
							break
						}
					}
					break
				}
			}
		}
	}
	return data, nil
}

func buildUserPermissionProviders(rows []admin.UserPermissionRow) []userPermissionProvider {
	seen := make(map[string]bool, len(rows))
	providers := make([]userPermissionProvider, 0)
	for _, row := range rows {
		if seen[row.ProviderID] {
			continue
		}
		seen[row.ProviderID] = true
		providers = append(providers, userPermissionProvider{
			ProviderID:   row.ProviderID,
			ProviderName: row.ProviderName,
		})
	}
	return providers
}

func filterUserPermissionsByProvider(rows []admin.UserPermissionRow, providerID string) []admin.UserPermissionRow {
	filtered := make([]admin.UserPermissionRow, 0)
	for _, row := range rows {
		if row.ProviderID == providerID {
			filtered = append(filtered, row)
		}
	}
	return filtered
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
		code := strings.TrimSpace(r.FormValue("code"))
		if code == "" {
			code = "channel-" + strconv.FormatInt(time.Now().UnixNano(), 36)
		}
		channel, err := store.CreateChannel(r.Context(), admin.CreateChannelInput{
			Name:                  r.FormValue("name"),
			Code:                  code,
			DefaultPermissionMode: r.FormValue("defaultPermissionMode"),
			UserManagementMode:    r.FormValue("userManagementMode"),
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
		code := strings.TrimSpace(r.FormValue("code"))
		if code == "" {
			code = "channel-" + strconv.FormatInt(time.Now().UnixNano(), 36)
		}
		if err := store.UpdateChannel(r.Context(), channelID, admin.UpdateChannelInput{
			Name:                  r.FormValue("name"),
			Code:                  code,
			DefaultPermissionMode: r.FormValue("defaultPermissionMode"),
			UserManagementMode:    r.FormValue("userManagementMode"),
			IsEnabled:             r.FormValue("isEnabled") == "1",
		}); err != nil {
			redirectAccessError(w, r, "保存渠道失败："+err.Error())
			return
		}
		redirectAccessChannel(w, r, channelID, "")
	}
}

func formDeleteChannelHandler(store *admin.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			redirectAccessError(w, r, "渠道表单格式不正确")
			return
		}
		if err := store.DeleteChannel(r.Context(), r.FormValue("channelId")); err != nil {
			redirectAccessError(w, r, "删除渠道失败："+err.Error())
			return
		}
		http.Redirect(w, r, "/admin/access", http.StatusFound)
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
		displayName := r.FormValue("displayName")
		externalUserID := r.FormValue("externalUserId")
		if externalUserID == "" {
			externalUserID = displayName
		}
		if err := store.UpdateUser(r.Context(), userID, admin.UpdateUserInput{
			ExternalUserID: externalUserID,
			DisplayName:    displayName,
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
		providerID := r.FormValue("providerId")
		modelIDs := r.Form["modelIds"]
		if len(modelIDs) > 0 {
			allowedModelIDs := make(map[string]bool, len(r.Form["allowedModelIds"]))
			for _, modelID := range r.Form["allowedModelIds"] {
				allowedModelIDs[modelID] = true
			}
			for _, modelID := range modelIDs {
				if err := store.SetUserPermission(r.Context(), userID, providerID, modelID, allowedModelIDs[modelID]); err != nil {
					redirectAccessError(w, r, "保存用户授权失败："+err.Error())
					return
				}
			}
			url := "/admin/access?channelId=" + template.URLQueryEscaper(channelID) + "&userId=" + template.URLQueryEscaper(userID) + "&userPermProviderId=" + template.URLQueryEscaper(providerID)
			http.Redirect(w, r, url, http.StatusFound)
			return
		}
		if err := store.SetUserPermission(r.Context(), userID, providerID, r.FormValue("modelId"), r.FormValue("allowed") == "1"); err != nil {
			redirectAccessError(w, r, "保存用户授权失败："+err.Error())
			return
		}
		redirectAccessChannel(w, r, channelID, userID)
	}
}

func formSetUserKeyPermissionHandler(store *admin.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			redirectAccessError(w, r, "用户 Key 授权表单格式不正确")
			return
		}
		channelID := r.FormValue("channelId")
		userID := r.FormValue("userId")
		providerID := r.FormValue("providerId")
		keyIDs := r.Form["keyIds"]
		allowedKeyIDs := make(map[string]bool, len(r.Form["allowedKeyIds"]))
		for _, keyID := range r.Form["allowedKeyIds"] {
			allowedKeyIDs[keyID] = true
		}
		for _, keyID := range keyIDs {
			if err := store.SetUserKeyPermission(r.Context(), userID, providerID, keyID, allowedKeyIDs[keyID]); err != nil {
				redirectAccessError(w, r, "保存用户 Key 授权失败："+err.Error())
				return
			}
		}
		url := "/admin/access?channelId=" + template.URLQueryEscaper(channelID) + "&userId=" + template.URLQueryEscaper(userID) + "&userPermProviderId=" + template.URLQueryEscaper(providerID)
		http.Redirect(w, r, url, http.StatusFound)
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
