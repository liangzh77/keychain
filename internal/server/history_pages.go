package server

import (
	"fmt"
	"html/template"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/liangzh77/keychain/internal/admin"
	"github.com/liangzh77/keychain/internal/auth"
)

var historyPageTemplate = template.Must(template.New("history").Parse(`<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Keychain 调用历史</title>
  <link rel="icon" href="/assets/keychain-icon.png">
  <style>
    :root { --bg: #f4f1ea; --surface: #fffcf5; --surface-muted: #f7f2ea; --line: #d7cdbf; --line-soft: #ece4d8; --text: #242824; --muted: #6f7169; --accent: #31594a; --accent-soft: #e8efe8; --accent-line: #9fb9aa; --secondary: #5a5448; --danger: #9b3d35; --warn: #b26b2f; }
    * { box-sizing: border-box; }
    body { margin: 0; font-family: ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; background: var(--bg); color: var(--text); }
    header { height: 56px; display: flex; align-items: center; justify-content: space-between; padding: 0 20px; background: var(--surface); border-bottom: 1px solid var(--line); }
    header form { margin: 0; }
    .brand { display: flex; align-items: baseline; gap: 10px; }
    .brand-logo { width: 26px; height: 26px; object-fit: contain; border-radius: 5px; align-self: center; }
    .brand strong { font-size: 16px; }
    .tab { padding: 8px 10px; color: var(--muted); text-decoration: none; border-radius: 6px; font-weight: 700; font-size: 14px; }
    .tab.active, .tab:hover { background: var(--accent-soft); color: var(--accent); }
    .app { height: calc(100vh - 56px); display: grid; grid-template-columns: 320px minmax(0, 1fr); overflow: hidden; }
    .app[aria-busy="true"] { cursor: progress; }
    aside { border-right: 1px solid var(--line); background: var(--surface-muted); overflow: auto; padding: 16px; }
    main { overflow: auto; padding: 16px 20px 24px; }
    h1, h2, h3 { margin: 0; line-height: 1.2; }
    h1 { font-size: 22px; }
    h2 { font-size: 18px; }
    h3 { font-size: 15px; }
    p { margin: 0; }
    form { display: grid; gap: 10px; }
    label { display: grid; gap: 5px; font-size: 12px; font-weight: 700; color: #45483f; }
    input, select { width: 100%; min-width: 0; padding: 9px 10px; border: 1px solid #d2c7b7; border-radius: 6px; font: inherit; background: #fffdf8; color: var(--text); }
    button { height: 38px; padding: 0 12px; border: 0; border-radius: 6px; background: var(--accent); color: white; cursor: pointer; font-weight: 700; white-space: nowrap; }
    button.ghost { background: #efe9df; color: #3e433d; }
    .muted { color: var(--muted); }
    .small { font-size: 12px; }
    .panel { background: var(--surface); border: 1px solid var(--line); border-radius: 8px; }
    .panel-pad, .content { padding: 14px; }
    .stack { display: grid; gap: 12px; }
    .topline { display: flex; justify-content: space-between; align-items: start; gap: 16px; margin-bottom: 14px; }
    .filter-grid { display: grid; gap: 10px; }
    .filter-actions { display: grid; grid-template-columns: 1fr 1fr; gap: 8px; }
    .stats-grid { display: grid; grid-template-columns: repeat(3, minmax(120px, 1fr)); gap: 10px; margin-bottom: 12px; }
    .stat { padding: 13px 14px; }
    .stat strong { display: block; font-size: 24px; line-height: 1.1; margin-bottom: 4px; }
    .stat span { color: var(--muted); font-size: 12px; font-weight: 700; }
	.chart-card { margin-bottom: 12px; }
	.chart-wrap { width: 100%; overflow: hidden; border: 1px solid var(--line-soft); border-radius: 7px; background: #fffdf8; }
	.chart-empty { padding: 40px 16px; text-align: center; color: var(--muted); }
	.chart-labels { display: flex; justify-content: space-between; gap: 12px; padding-top: 8px; color: var(--muted); font-size: 12px; }
	.chart-header-tools { display: flex; align-items: center; justify-content: flex-end; gap: 12px; flex-wrap: wrap; }
	.legend { display: flex; align-items: center; gap: 12px; color: var(--muted); font-size: 12px; font-weight: 700; }
	.legend-item { display: inline-flex; align-items: center; gap: 6px; white-space: nowrap; }
	.legend-swatch { width: 22px; height: 3px; border-radius: 999px; background: var(--accent); }
	.legend-swatch.warn { background: var(--warn); }
	.axis-label { fill: #6f7169; font-size: 12px; font-weight: 700; }
	.axis-tick { fill: #7a7d73; font-size: 11px; }
	.point-value { fill: #31594a; font-size: 11px; font-weight: 800; }
	.chart-hover-line { stroke: #6f7169; stroke-width: 1.5; stroke-dasharray: 4 5; pointer-events: none; }
	.chart-hover-tip rect { fill: #242824; opacity: .92; rx: 6; }
	.chart-hover-tip text { fill: #fffcf5; font-size: 12px; font-weight: 700; }
	.chart-hit-area { fill: transparent; cursor: crosshair; }
    .table-card { overflow: hidden; }
    table { width: 100%; border-collapse: collapse; font-size: 13px; }
    th, td { padding: 10px 12px; border-bottom: 1px solid var(--line-soft); text-align: left; vertical-align: top; }
    th { position: sticky; top: 0; background: var(--surface); color: #4b4f46; font-size: 12px; z-index: 1; }
    tbody tr:hover { background: #f8f3ea; }
    .table-scroll { max-height: calc(100vh - 390px); overflow: auto; min-height: 260px; }
    .tag { display: inline-flex; align-items: center; padding: 2px 8px; border-radius: 999px; background: var(--accent-soft); color: var(--accent); font-size: 12px; font-weight: 700; white-space: nowrap; }
    .tag.warn { background: #faead8; color: var(--warn); }
	.failure { color: var(--danger); min-width: 220px; max-width: 360px; }
	.failure-inline { display: grid; grid-template-columns: minmax(0, 1fr) 28px; align-items: center; gap: 6px; }
	.failure-text { min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
	.copy-icon { width: 28px; height: 28px; padding: 0; display: inline-flex; align-items: center; justify-content: center; border: 1px solid var(--line-soft); border-radius: 6px; background: #fffdf8; color: var(--muted); }
	.copy-icon:hover { color: var(--accent); border-color: var(--accent-line); background: var(--accent-soft); }
	.copy-icon svg { width: 15px; height: 15px; }
    .notice { margin-bottom: 14px; padding: 10px 12px; border-radius: 6px; background: #fff6df; color: #7a5a22; }
    .empty { padding: 28px; text-align: center; color: var(--muted); }
    .pagination { display: flex; justify-content: space-between; align-items: center; gap: 12px; padding: 12px 14px; }
    .pager-actions { display: flex; gap: 8px; }
    .pager-actions a { display: inline-flex; align-items: center; height: 34px; padding: 0 12px; border-radius: 6px; background: #efe9df; color: #3e433d; font-weight: 700; text-decoration: none; }
    @media (max-width: 1320px) { .stats-grid { grid-template-columns: repeat(3, minmax(120px, 1fr)); } }
    @media (max-width: 980px) { .app { grid-template-columns: 1fr; height: auto; overflow: visible; } aside, main { overflow: visible; } .table-scroll { max-height: none; } }
    @media (max-width: 760px) { .stats-grid { grid-template-columns: repeat(2, minmax(120px, 1fr)); } .topline { flex-direction: column; } }
  </style>
</head>
<body>
  <header>
    <div class="brand"><img class="brand-logo" src="/assets/keychain-icon.png" alt="" aria-hidden="true"><strong>Keychain</strong><span class="muted small">admin console</span><a class="tab" href="/admin">Providers</a><a class="tab" href="/admin/access">渠道与授权</a><a class="tab active" href="/admin/history">调用历史</a></div>
    <form method="post" action="/logout"><button class="ghost" type="submit">退出</button></form>
  </header>
  <div class="app">
    <aside>
      <section class="panel panel-pad stack">
        <div>
          <h2>筛选条件</h2>
        </div>
        <form method="get" action="/admin/history" class="filter-grid">
          <label>时间范围
            <select name="range">
              {{range .RangeOptions}}<option value="{{.Value}}" {{if .Active}}selected{{end}}>{{.Label}}</option>{{end}}
            </select>
          </label>
          <label>开始时间<input type="datetime-local" name="startTime" value="{{.StartInput}}"></label>
          <label>结束时间<input type="datetime-local" name="endTime" value="{{.EndInput}}"></label>
          <label>渠道
            <select name="channelId">
              <option value="">全部渠道</option>
              {{range .Channels}}<option value="{{.ID}}" {{if .Active}}selected{{end}}>{{.Name}}</option>{{end}}
            </select>
          </label>
          <label>用户
            <select name="userId">
              <option value="">全部用户</option>
              {{range .Users}}<option value="{{.ID}}" {{if .Active}}selected{{end}}>{{.Name}}</option>{{end}}
            </select>
          </label>
          <label>Provider
            <select name="providerId">
              <option value="">全部 Providers</option>
              {{range .Providers}}<option value="{{.ID}}" {{if .Active}}selected{{end}}>{{.Name}}</option>{{end}}
            </select>
          </label>
          <label>模型
            <select name="modelId" {{if not .Models}}disabled{{end}}>
              <option value="">全部模型</option>
              {{range .Models}}<option value="{{.ID}}" {{if .Active}}selected{{end}}>{{.Name}}</option>{{end}}
            </select>
          </label>
          <label>Key
            <select name="keyId" {{if not .Keys}}disabled{{end}}>
              <option value="">全部 Key</option>
              {{range .Keys}}<option value="{{.ID}}" {{if .Active}}selected{{end}}>{{.Name}}</option>{{end}}
            </select>
          </label>
          <label>状态
            <select name="status">
              {{range .StatusOptions}}<option value="{{.Value}}" {{if .Active}}selected{{end}}>{{.Label}}</option>{{end}}
            </select>
          </label>
          <label>时间排序
            <select name="sort">
              {{range .SortOptions}}<option value="{{.Value}}" {{if .Active}}selected{{end}}>{{.Label}}</option>{{end}}
            </select>
          </label>
          <label>每页条数
            <select name="pageSize">
              {{range .PageSizeOptions}}<option value="{{.Value}}" {{if .Active}}selected{{end}}>{{.Label}}</option>{{end}}
            </select>
          </label>
          <div class="filter-actions">
            <button type="submit">筛选</button>
            <a class="tab" href="/admin/history">清空</a>
          </div>
        </form>
      </section>
    </aside>
    <main>
      {{if .Error}}<div class="notice">{{.Error}}</div>{{end}}
      <div class="topline">
        <div>
          <h1>调用历史</h1>
          <p class="muted small">查看 Key 调用记录、失败情况和调用量趋势。</p>
        </div>
      </div>
      <section class="stats-grid">
        <div class="panel stat"><strong>{{.Stats.TotalCount}}</strong><span>总调用</span></div>
        <div class="panel stat"><strong>{{.Stats.SuccessCount}}</strong><span>成功调用</span></div>
        <div class="panel stat"><strong>{{.Stats.FailedCount}}</strong><span>失败调用</span></div>
      </section>
      <section class="panel content chart-card">
        <div class="topline">
          <div>
            <h2>调用量曲线</h2>
            <p class="muted small">按{{.BucketLabel}}聚合，深色为总调用，暖色为失败。</p>
          </div>
          <div class="chart-header-tools">
            <div class="legend" aria-label="图例">
              <span class="legend-item"><span class="legend-swatch"></span>总调用</span>
              <span class="legend-item"><span class="legend-swatch warn"></span>失败调用</span>
            </div>
            <span class="tag">{{.Stats.FailureRatePercent}}% 失败率</span>
          </div>
        </div>
        {{if .ChartPolyline}}
          <div class="chart-wrap">
            <svg data-history-chart viewBox="0 0 960 300" width="100%" height="300" role="img" aria-label="调用量时间曲线">
              <text class="axis-label" x="20" y="24">调用次数</text>
              <text class="axis-label" x="500" y="286" text-anchor="middle">时间</text>
              <text class="axis-tick" x="58" y="40" text-anchor="end">{{.ChartMaxCount}}</text>
              <text class="axis-tick" x="58" y="188" text-anchor="end">0</text>
              <line x1="70" y1="184" x2="930" y2="184" stroke="#d7cdbf" />
              <line x1="70" y1="40" x2="70" y2="184" stroke="#d7cdbf" />
              <line x1="70" y1="40" x2="930" y2="40" stroke="#ece4d8" stroke-dasharray="4 6" />
              {{range .ChartPoints}}
                <line x1="{{.X}}" y1="184" x2="{{.X}}" y2="190" stroke="#d7cdbf" />
                <text class="axis-tick" x="{{.X}}" y="202" text-anchor="end" transform="rotate(-45 {{.X}} 202)">{{.TickLabel}}</text>
              {{end}}
              <polyline points="{{.ChartPolyline}}" fill="none" stroke="#31594a" stroke-width="3" stroke-linejoin="round" stroke-linecap="round" />
              {{if .FailurePolyline}}<polyline points="{{.FailurePolyline}}" fill="none" stroke="#b26b2f" stroke-width="2.5" stroke-linejoin="round" stroke-linecap="round" />{{end}}
              {{range .ChartPoints}}
                <text class="point-value" x="{{.X}}" y="{{.ValueLabelY}}" text-anchor="middle">{{.TotalCount}}</text>
                <circle data-chart-point data-x="{{.X}}" data-y="{{.Y}}" data-date="{{.Label}}" data-total="{{.TotalCount}}" data-failed="{{.FailedCount}}" cx="{{.X}}" cy="{{.Y}}" r="3.5" fill="#31594a"><title>{{.Label}}：{{.TotalCount}}次，失败{{.FailedCount}}次</title></circle>
              {{end}}
              <line data-hover-line class="chart-hover-line" x1="70" y1="36" x2="70" y2="184" visibility="hidden" />
              <g data-hover-tip class="chart-hover-tip" visibility="hidden">
                <rect width="174" height="58"></rect>
                <text data-tip-date x="10" y="19"></text>
                <text data-tip-total x="10" y="36"></text>
                <text data-tip-failed x="10" y="52"></text>
              </g>
              <rect class="chart-hit-area" x="70" y="28" width="860" height="190" />
            </svg>
          </div>
          <div class="chart-labels"><span>{{.ChartStartLabel}}</span><span>{{.ChartEndLabel}}</span></div>
        {{else}}
          <div class="chart-wrap"><div class="chart-empty">当前筛选范围内还没有调用记录。</div></div>
        {{end}}
      </section>
      <section class="panel table-card">
        <div class="content topline">
          <div>
            <h2>明细</h2>
            <p class="muted small">共 {{.Pagination.TotalItems}} 条记录</p>
          </div>
        </div>
        <div class="table-scroll">
          <table>
            <thead>
              <tr>
                <th>时间</th>
                <th>渠道</th>
                <th>用户</th>
                <th>Provider</th>
                <th>模型</th>
                <th>Key</th>
                <th>状态</th>
                <th>失败信息</th>
              </tr>
            </thead>
            <tbody>
              {{range .Rows}}
                <tr>
                  <td>{{.CreatedAtText}}</td>
                  <td>{{.ChannelName}}</td>
                  <td>{{.UserDisplayName}}</td>
                  <td>{{.ProviderName}}</td>
                  <td>{{.ModelName}}</td>
                  <td>{{.KeyAlias}}</td>
                  <td><span class="tag {{if .IsFailed}}warn{{end}}">{{.StatusText}}</span></td>
                  <td class="failure">
                    <div class="failure-inline" title="{{.FailureText}}">
                      <span class="failure-text">{{.FailureText}}</span>
                      {{if .CanCopyFailure}}<button class="copy-icon" type="button" data-copy-text="{{.FailureText}}" aria-label="复制失败信息" title="复制失败信息">
                        <svg viewBox="0 0 24 24" aria-hidden="true" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                          <rect x="9" y="9" width="13" height="13" rx="2"></rect>
                          <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>
                        </svg>
                      </button>{{end}}
                    </div>
                  </td>
                </tr>
              {{else}}
                <tr><td colspan="8" class="empty">没有匹配的调用记录。</td></tr>
              {{end}}
            </tbody>
          </table>
        </div>
        <div class="pagination">
          <span class="muted small">第 {{.Pagination.Page}} / {{.Pagination.TotalPagesDisplay}} 页</span>
          <div class="pager-actions">
            {{if .PrevURL}}<a href="{{.PrevURL}}">上一页</a>{{end}}
            {{if .NextURL}}<a href="{{.NextURL}}">下一页</a>{{end}}
          </div>
        </div>
      </section>
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
      if (url && url !== window.location.href) window.history.pushState({}, '', url);
      initAdminPage();
    }
    async function navigateAdmin(url) {
      const app = document.querySelector('.app');
      if (app) app.setAttribute('aria-busy', 'true');
      try {
        const response = await fetch(url, { headers: { 'X-Requested-With': 'fetch' } });
        if (!response.ok) throw new Error('Navigation failed');
        replaceAdminApp(await response.text(), response.url);
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
          headers: { 'Content-Type': 'application/x-www-form-urlencoded;charset=UTF-8', 'X-Requested-With': 'fetch' }
        });
        replaceAdminApp(await response.text(), response.url);
      } catch (error) {
        form.submit();
      } finally {
        if (app) app.removeAttribute('aria-busy');
        if (button) button.disabled = false;
      }
    }
    function initAdminPage() {
      document.querySelectorAll('[data-dirty-form]').forEach((form) => {
        const save = form.querySelector('[data-save]');
        if (!save) return;
        const initial = JSON.stringify(Array.from(new FormData(form).entries()));
        const sync = () => { save.disabled = JSON.stringify(Array.from(new FormData(form).entries())) === initial; };
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
          try {
            await navigator.clipboard.writeText(button.dataset.copyText || '');
            button.setAttribute('aria-label', '已复制');
            window.setTimeout(() => button.setAttribute('aria-label', '复制失败信息'), 1200);
          } catch (error) {
            const fallback = document.createElement('textarea');
            fallback.value = button.dataset.copyText || '';
            fallback.style.position = 'fixed';
            fallback.style.left = '-9999px';
            document.body.appendChild(fallback);
            fallback.select();
            document.execCommand('copy');
            fallback.remove();
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
        if (!form.matches('.app form')) return;
        if ((form.method || 'get').toLowerCase() === 'get') {
          event.preventDefault();
          const url = new URL(form.action, window.location.href);
          new URLSearchParams(new FormData(form)).forEach((value, key) => {
            if (value) url.searchParams.set(key, value);
          });
          navigateAdmin(url.href);
          return;
        }
        if (!form.matches('.app form[method="post"]')) return;
        event.preventDefault();
        submitAdminForm(form, event.submitter);
      });
      window.addEventListener('popstate', () => navigateAdmin(window.location.href));
    }
    initAdminPage();
  </script>
</body>
</html>`))

type historyPageData struct {
	Username        string
	Error           string
	RangeOptions    []selectOption
	StatusOptions   []selectOption
	SortOptions     []selectOption
	PageSizeOptions []selectOption
	Channels        []selectOption
	Users           []selectOption
	Providers       []selectOption
	Models          []selectOption
	Keys            []selectOption
	Rows            []historyRowView
	Stats           admin.DispatchHistoryStats
	Pagination      historyPaginationView
	StartInput      string
	EndInput        string
	BucketLabel     string
	ChartPolyline   string
	FailurePolyline string
	ChartPoints     []historyChartPoint
	ChartMaxCount   int
	ChartStartLabel string
	ChartEndLabel   string
	PrevURL         string
	NextURL         string
}

type selectOption struct {
	Value  string
	Label  string
	ID     string
	Name   string
	Active bool
}

type historyRowView struct {
	admin.DispatchHistoryRow
	CreatedAtText  string
	StatusText     string
	IsFailed       bool
	FailureText    string
	CanCopyFailure bool
}

type historyChartPoint struct {
	X           int
	Y           int
	ValueLabelY int
	Label       string
	TickLabel   string
	TotalCount  int
	FailedCount int
}

type historyPaginationView struct {
	admin.DispatchHistoryPagination
	TotalPagesDisplay int
}

func registerHistoryRoutes(mux *http.ServeMux, authService *auth.Service, store *admin.Store) {
	mux.HandleFunc("GET /admin/history", historyPageHandler(authService, store))
}

func historyPageHandler(authService *auth.Service, store *admin.Store) http.HandlerFunc {
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
		data, err := loadHistoryPageData(r, store)
		data.Username = session.Username
		data.Error = r.URL.Query().Get("error")
		if err != nil {
			data.Error = "加载调用历史失败：" + err.Error()
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = historyPageTemplate.Execute(w, data)
	}
}

func loadHistoryPageData(r *http.Request, store *admin.Store) (historyPageData, error) {
	query := r.URL.Query()
	rangeKey := strings.TrimSpace(query.Get("range"))
	if rangeKey == "" {
		rangeKey = "30d"
	}
	startTime, endTime, startInput, endInput, bucket := resolveHistoryTimeRange(query, rangeKey)
	filter := admin.DispatchHistoryFilter{
		StartTime:  startTime,
		EndTime:    endTime,
		ChannelID:  query.Get("channelId"),
		UserID:     query.Get("userId"),
		ProviderID: query.Get("providerId"),
		ModelID:    query.Get("modelId"),
		KeyID:      query.Get("keyId"),
		Status:     query.Get("status"),
		Sort:       query.Get("sort"),
		Page:       parseOptionalInt(query.Get("page")),
		PageSize:   parseOptionalInt(query.Get("pageSize")),
	}
	if filter.PageSize == 0 {
		filter.PageSize = 50
	}

	data := historyPageData{
		StartInput:  startInput,
		EndInput:    endInput,
		BucketLabel: historyBucketLabel(bucket),
		RangeOptions: []selectOption{
			{Value: "today", Label: "今天", Active: rangeKey == "today"},
			{Value: "7d", Label: "最近 7 天", Active: rangeKey == "7d"},
			{Value: "30d", Label: "最近 30 天", Active: rangeKey == "30d"},
			{Value: "custom", Label: "自定义", Active: rangeKey == "custom"},
		},
		StatusOptions: []selectOption{
			{Value: "", Label: "全部状态", Active: filter.Status == ""},
			{Value: "DISPATCHED", Label: "成功调用", Active: filter.Status == "DISPATCHED"},
			{Value: "FAILED", Label: "失败调用", Active: filter.Status == "FAILED"},
		},
		SortOptions: []selectOption{
			{Value: "desc", Label: "最新在前", Active: filter.Sort != "asc"},
			{Value: "asc", Label: "最早在前", Active: filter.Sort == "asc"},
		},
		PageSizeOptions: []selectOption{
			{Value: "25", Label: "25 条", Active: filter.PageSize == 25},
			{Value: "50", Label: "50 条", Active: filter.PageSize == 50},
			{Value: "100", Label: "100 条", Active: filter.PageSize == 100},
			{Value: "200", Label: "200 条", Active: filter.PageSize == 200},
		},
	}

	channels, err := store.ListChannels(r.Context())
	if err != nil {
		return data, err
	}
	for _, channel := range channels {
		data.Channels = append(data.Channels, selectOption{ID: channel.ID, Name: channel.Name, Active: channel.ID == filter.ChannelID})
	}
	users, err := store.ListDispatchHistoryUserOptions(r.Context(), filter.ChannelID)
	if err != nil {
		return data, err
	}
	for _, user := range users {
		name := user.DisplayName
		if filter.ChannelID == "" {
			name = user.ChannelName + " / " + user.DisplayName
		}
		data.Users = append(data.Users, selectOption{ID: user.ID, Name: name, Active: user.ID == filter.UserID})
	}

	providers, err := store.ListProviders(r.Context())
	if err != nil {
		return data, err
	}
	for _, provider := range providers {
		data.Providers = append(data.Providers, selectOption{ID: provider.ID, Name: provider.Name, Active: provider.ID == filter.ProviderID})
	}
	if filter.ProviderID != "" {
		models, err := store.ListModels(r.Context(), filter.ProviderID, "")
		if err != nil {
			return data, err
		}
		for _, model := range models {
			data.Models = append(data.Models, selectOption{ID: model.ID, Name: model.Name, Active: model.ID == filter.ModelID})
		}
		keys, err := store.ListAPIKeys(r.Context(), filter.ProviderID)
		if err != nil {
			return data, err
		}
		for _, key := range keys {
			data.Keys = append(data.Keys, selectOption{ID: key.ID, Name: key.Alias, Active: key.ID == filter.KeyID})
		}
	} else {
		filter.ModelID = ""
		filter.KeyID = ""
	}

	rows, pagination, err := store.ListDispatchHistory(r.Context(), filter)
	if err != nil {
		return data, err
	}
	stats, err := store.GetDispatchHistoryStats(r.Context(), filter)
	if err != nil {
		return data, err
	}
	series, err := store.GetDispatchHistorySeries(r.Context(), filter, bucket)
	if err != nil {
		return data, err
	}

	data.Rows = buildHistoryRows(rows)
	data.Stats = stats
	data.Pagination = historyPaginationView{DispatchHistoryPagination: pagination, TotalPagesDisplay: maxInt(1, pagination.TotalPages)}
	data.ChartPolyline, data.FailurePolyline, data.ChartPoints, data.ChartMaxCount = buildHistoryChart(series)
	if len(series) > 0 {
		data.ChartStartLabel = compactHistoryTime(series[0].BucketStart)
		data.ChartEndLabel = compactHistoryTime(series[len(series)-1].BucketStart)
	}
	data.PrevURL = buildHistoryPageURL(query, pagination.Page-1, pagination.Page > 1)
	data.NextURL = buildHistoryPageURL(query, pagination.Page+1, pagination.TotalPages > 0 && pagination.Page < pagination.TotalPages)
	return data, nil
}

func resolveHistoryTimeRange(values url.Values, rangeKey string) (string, string, string, string, string) {
	now := time.Now()
	startInput := values.Get("startTime")
	endInput := values.Get("endTime")
	var start, end time.Time
	if startInput != "" || endInput != "" || rangeKey == "custom" {
		start = parseHistoryInputTime(startInput)
		end = parseHistoryInputTime(endInput)
	} else {
		switch rangeKey {
		case "today":
			start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
			end = now
		case "7d":
			start = now.AddDate(0, 0, -7)
			end = now
		default:
			start = now.AddDate(0, 0, -30)
			end = now
		}
		startInput = start.Format("2006-01-02T15:04")
		endInput = end.Format("2006-01-02T15:04")
	}
	bucket := "day"
	if !start.IsZero() && !end.IsZero() {
		duration := end.Sub(start)
		if duration <= 36*time.Hour {
			bucket = "hour"
		} else if duration > 90*24*time.Hour {
			bucket = "month"
		}
	}
	return formatHistoryFilterTime(start), formatHistoryFilterTime(end), startInput, endInput, bucket
}

func parseHistoryInputTime(value string) time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}
	}
	if parsed, err := time.ParseInLocation("2006-01-02T15:04", value, time.Local); err == nil {
		return parsed
	}
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed
	}
	return time.Time{}
}

func formatHistoryFilterTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}

func buildHistoryRows(rows []admin.DispatchHistoryRow) []historyRowView {
	views := make([]historyRowView, 0, len(rows))
	for _, row := range rows {
		view := historyRowView{
			DispatchHistoryRow: row,
			CreatedAtText:      compactHistoryTime(row.CreatedAt),
			StatusText:         "成功调用",
			IsFailed:           row.Status == "FAILED",
			FailureText:        "-",
		}
		if view.IsFailed {
			view.StatusText = "失败调用"
			parts := make([]string, 0, 2)
			if row.FailureErrorCode != nil && *row.FailureErrorCode != "" {
				parts = append(parts, *row.FailureErrorCode)
			}
			if row.FailureErrorMessage != nil && *row.FailureErrorMessage != "" {
				parts = append(parts, *row.FailureErrorMessage)
			}
			if len(parts) > 0 {
				view.FailureText = strings.Join(parts, "：")
				view.CanCopyFailure = true
			}
		}
		views = append(views, view)
	}
	return views
}

func buildHistoryChart(points []admin.DispatchHistoryPoint) (string, string, []historyChartPoint, int) {
	if len(points) == 0 {
		return "", "", nil, 0
	}
	maxCount := 1
	for _, point := range points {
		if point.TotalCount > maxCount {
			maxCount = point.TotalCount
		}
	}
	width := 860.0
	height := 144.0
	left := 70.0
	top := 40.0
	step := 0.0
	if len(points) > 1 {
		step = width / float64(len(points)-1)
	}
	totalPairs := make([]string, 0, len(points))
	failurePairs := make([]string, 0, len(points))
	chartPoints := make([]historyChartPoint, 0, len(points))
	for index, point := range points {
		x := left + float64(index)*step
		if len(points) == 1 {
			x = left + width/2
		} else if len(points) == 2 {
			x = left + width*(float64(index)+1)/3
		}
		totalY := top + height - float64(point.TotalCount)/float64(maxCount)*height
		failureY := top + height - float64(point.FailedCount)/float64(maxCount)*height
		totalPairs = append(totalPairs, fmt.Sprintf("%d,%d", int(math.Round(x)), int(math.Round(totalY))))
		failurePairs = append(failurePairs, fmt.Sprintf("%d,%d", int(math.Round(x)), int(math.Round(failureY))))
		chartPoints = append(chartPoints, historyChartPoint{
			X:           int(math.Round(x)),
			Y:           int(math.Round(totalY)),
			ValueLabelY: maxInt(14, int(math.Round(totalY))-8),
			Label:       compactHistoryTime(point.BucketStart),
			TickLabel:   compactHistoryAxisTick(point.BucketStart),
			TotalCount:  point.TotalCount,
			FailedCount: point.FailedCount,
		})
	}
	return strings.Join(totalPairs, " "), strings.Join(failurePairs, " "), chartPoints, maxCount
}

func compactHistoryAxisTick(value string) string {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return value
	}
	local := parsed.In(time.Local)
	return local.Format("01-02")
}

func compactHistoryTime(value string) string {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return value
	}
	return parsed.In(time.Local).Format("2006-01-02 15:04")
}

func historyBucketLabel(bucket string) string {
	switch bucket {
	case "hour":
		return "小时"
	case "month":
		return "月"
	default:
		return "天"
	}
}

func buildHistoryPageURL(values url.Values, page int, ok bool) string {
	if !ok {
		return ""
	}
	next := url.Values{}
	for key, existing := range values {
		for _, value := range existing {
			if strings.TrimSpace(value) != "" && key != "page" {
				next.Add(key, value)
			}
		}
	}
	next.Set("page", strconv.Itoa(page))
	return "/admin/history?" + next.Encode()
}

func maxInt(left int, right int) int {
	if left > right {
		return left
	}
	return right
}
