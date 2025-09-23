package templates

const WebUICSSTemplate = `/* CSS Variables for theming */
:root {
  --primary-color: #2563eb;
  --primary-hover: #1d4ed8;
  --secondary-color: #64748b;
  --background-color: #f8fafc;
  --surface-color: #ffffff;
  --border-color: #e2e8f0;
  --text-primary: #1e293b;
  --text-secondary: #64748b;
  --shadow-sm: 0 1px 2px 0 rgba(0,0,0,0.05);
  --shadow-md: 0 4px 6px -1px rgba(0,0,0,0.1), 0 2px 4px -1px rgba(0,0,0,0.06);
  --radius-sm: 0.375rem;
  --radius-md: 0.5rem;
  --radius-lg: 0.75rem;
}

@media (prefers-color-scheme: dark) {
  :root {
    --primary-color: #3b82f6;
    --primary-hover: #2563eb;
    --secondary-color: #94a3b8;
    --background-color: #0f172a;
    --surface-color: #1e293b;
    --border-color: #334155;
    --text-primary: #f1f5f9;
    --text-secondary: #94a3b8;
  }
}

/* Explicit theme overrides applied via <html class="theme-light"> or "theme-dark" */
:root.theme-light {
  --primary-color: #2563eb;
  --primary-hover: #1d4ed8;
  --secondary-color: #64748b;
  --background-color: #f8fafc;
  --surface-color: #ffffff;
  --border-color: #e2e8f0;
  --text-primary: #0f172a;
  --text-secondary: #475569;
}

:root.theme-dark {
  --primary-color: #3b82f6;
  --primary-hover: #2563eb;
  --secondary-color: #94a3b8;
  --background-color: #0b1222;
  --surface-color: #111827;
  --border-color: #243244;
  --text-primary: #e5e7eb;
  --text-secondary: #9ca3af;
}

* { margin:0; padding:0; box-sizing:border-box; }
body { font-family: -apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,'Helvetica Neue',Arial,sans-serif; background:var(--background-color); color:var(--text-primary); height:100vh; overflow:hidden; line-height:1.6; }

.app-container { display:flex; flex-direction:column; height:100vh; max-width:100vw; }
.app-header { background:var(--surface-color); border-bottom:1px solid var(--border-color); height:44px; padding:0 0.75rem; box-shadow:var(--shadow-sm); display:flex; align-items:center; }
.header-content { display:flex; align-items:center; justify-content:space-between; width:100%; max-width:1400px; margin:0 auto; }
.app-title { font-size:1rem; font-weight:600; }
.header-actions { display:flex; gap:0.75rem; align-items:center; }
.header-btn { border:1px solid var(--border-color); background:var(--surface-color); color:var(--text-primary); padding:0.35rem 0.6rem; border-radius:var(--radius-sm); cursor:pointer; }
.header-btn:hover { background:var(--background-color); }

.agent-selector-container { display:flex; align-items:center; gap:0.5rem; }
.agent-label { font-size:0.85rem; color:var(--text-secondary); }
.agent-select { padding:0.25rem 0.5rem; border:1px solid var(--border-color); border-radius:var(--radius-sm); background:var(--surface-color); color:var(--text-primary); font-size:0.875rem; }
.agent-select:focus { outline:none; border-color:var(--primary-color); box-shadow:0 0 0 2px rgba(37,99,235,0.15); }
.orchestration-toggle { display:flex; align-items:center; gap:0.4rem; color:var(--text-secondary); }
.orchestration-toggle input[type="checkbox"] { width:1rem; height:1rem; border:1px solid var(--border-color); border-radius:0.25rem; background:var(--surface-color); }
.orchestration-toggle input[type="checkbox"]:checked { background:var(--primary-color); border-color:var(--primary-color); }

/* Three pane grid */
.grid-layout { flex:1; display:grid; grid-template-rows:auto 1fr; }
.main-3pane { display:grid; grid-template-columns: 0 1fr 0; gap:0; height:calc(100vh - 44px); }
.grid-layout.left-open .main-3pane { grid-template-columns: 1fr 2fr 0; }
.grid-layout.right-open .main-3pane { grid-template-columns: 0 2fr 1fr; }
.grid-layout.left-open.right-open .main-3pane { grid-template-columns: 1fr 2fr 1fr; }

.pane { background:var(--surface-color); border-right:1px solid var(--border-color); display:flex; flex-direction:column; overflow:hidden; }
.pane.right { border-left:1px solid var(--border-color); border-right:none; }
/* Keep hidden panes in DOM so grid column positions remain stable */
.pane[aria-hidden="true"] { visibility: hidden; }
.pane-header { display:flex; align-items:center; justify-content:space-between; padding:0.5rem 0.75rem; border-bottom:1px solid var(--border-color); }
.pane-body { flex:1; overflow:auto; padding:0.5rem; }
.mini-btn { border:1px solid var(--border-color); background:var(--surface-color); border-radius:var(--radius-sm); padding:0.25rem 0.5rem; cursor:pointer; }
.mini-btn.primary { background:var(--primary-color); color:#fff; border-color:var(--primary-color); }
.mini-btn:hover { background:var(--background-color); }

.trace-diagram { min-height:200px; }
.trace-diagram text, .trace-diagram g, .trace-diagram path { pointer-events: all; }
.trace-raw { background:var(--background-color); border:1px dashed var(--border-color); padding:0.5rem; border-radius:var(--radius-sm); white-space:pre-wrap; }

/* Center column (chat) */
.center { display:flex; flex-direction:column; overflow:hidden; }
.agent-info-panel { background:var(--surface-color); border-bottom:1px solid var(--border-color); padding:0.75rem; }
.agent-details h3 { font-size:1.05rem; font-weight:600; margin-bottom:0.25rem; }
.agent-details p { font-size:0.9rem; color:var(--text-secondary); }
.chat-container { flex:1; display:flex; flex-direction:column; overflow:hidden; background:var(--surface-color); }
.chat-messages { flex:1; overflow:auto; padding:1rem; scroll-behavior:smooth; }

.welcome-message { text-align:center; padding:1.5rem; color:var(--text-secondary); }
.welcome-message h3 { font-size:1.1rem; margin-bottom:0.4rem; }

.message { margin-bottom:1rem; display:flex; flex-direction:column; }
.message.user { align-items:flex-end; }
.message.agent { align-items:flex-start; }
.message-content { max-width:70%; padding:0.75rem 1rem; border-radius:var(--radius-lg); word-wrap:break-word; white-space:pre-wrap; }
.message.user .message-content { background:var(--primary-color); color:#fff; border-bottom-right-radius:var(--radius-sm); }
.message.agent .message-content { background:var(--background-color); color:var(--text-primary); border:1px solid var(--border-color); border-bottom-left-radius:var(--radius-sm); }
.message-info { font-size:0.75rem; color:var(--text-secondary); margin-top:0.25rem; padding:0 0.25rem; }

.chat-input-container { background:var(--surface-color); border-top:1px solid var(--border-color); padding:0.75rem; }
.chat-form { display:flex; gap:0.5rem; max-width:1400px; margin:0 auto; }
.message-input { flex:1; padding:0.7rem 1rem; border:1px solid var(--border-color); border-radius:var(--radius-md); background:var(--surface-color); color:var(--text-primary); font-size:1rem; }
.message-input:focus { outline:none; border-color:var(--primary-color); box-shadow:0 0 0 3px rgba(37,99,235,0.1); }
.message-input::placeholder { color:var(--text-secondary); }
.send-button { padding:0.7rem; background:var(--primary-color); color:#fff; border:none; border-radius:var(--radius-md); cursor:pointer; display:flex; align-items:center; justify-content:center; }
.send-button:hover { background:var(--primary-hover); }
.send-button:disabled { opacity:0.5; cursor:not-allowed; }

/* Config editor (CodeJar + Prism) */
.config-editor { width:100%; min-height:280px; background:var(--background-color); color:var(--text-primary); border:1px solid var(--border-color); border-radius:var(--radius-sm); padding:0.75rem; font-family: ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace; font-size:0.9rem; line-height:1.5; overflow:auto; white-space:pre; tab-size:2; caret-color: var(--text-primary); }
.config-editor:focus { outline:none; box-shadow:0 0 0 3px rgba(37,99,235,0.12); }
.config-editor[contenteditable="true"] { -webkit-user-modify: read-write-plaintext-only; }
.config-editor code { font-family: inherit; }
.config-status { margin-top:0.5rem; color:var(--text-secondary); }

/* Typing indicator */
.typing-indicator { display:flex; align-items:center; gap:0.5rem; color:var(--text-secondary); font-style:italic; padding:0.5rem 1rem; }
.typing-dots { display:flex; gap:0.25rem; }
.typing-dot { width:0.5rem; height:0.5rem; background:var(--text-secondary); border-radius:50%; animation: typingPulse 1.4s infinite ease-in-out; }
.typing-dot:nth-child(1){ animation-delay:-0.32s; } .typing-dot:nth-child(2){ animation-delay:-0.16s; }
@keyframes typingPulse { 0%,80%,100% { opacity:0.3; transform:scale(0.8);} 40% { opacity:1; transform:scale(1);} }

/* Scrollbars */
.chat-messages::-webkit-scrollbar { width:0.5rem; }
.chat-messages::-webkit-scrollbar-track { background:var(--background-color); }
.chat-messages::-webkit-scrollbar-thumb { background:var(--border-color); border-radius:var(--radius-sm); }
.chat-messages::-webkit-scrollbar-thumb:hover { background:var(--secondary-color); }

/* A11y */
.sr-only { position:absolute; width:1px; height:1px; padding:0; margin:-1px; overflow:hidden; clip:rect(0,0,0,0); border:0; }

@media (max-width: 1024px) {
  .grid-layout.left-open .main-3pane { grid-template-columns: 1fr 2fr 0; }
  .grid-layout.right-open .main-3pane { grid-template-columns: 0 2fr 1fr; }
}

@media (max-width: 768px) {
  .header-actions { gap:0.5rem; }
  .app-header { height:42px; }
  .main-3pane { height:calc(100vh - 42px); }
  .agent-details h3 { font-size:1rem; }
  .agent-details p { font-size:0.85rem; }
  .message-content { max-width:85%; }
}
`
