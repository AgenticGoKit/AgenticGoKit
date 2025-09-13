package templates

const WebUIIndexTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Config.Name}} - AI Agent Chat</title>
    <link rel="stylesheet" href="style.css">
    <script src="https://cdn.jsdelivr.net/npm/mermaid/dist/mermaid.min.js"></script>
    <script>mermaid.initialize({ startOnLoad: false, theme: 'base' });</script>
    <meta http-equiv="Content-Security-Policy" content="default-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net">
    <meta name="color-scheme" content="light dark">
    <meta name="view-transition" content="same-origin">
</head>
<body>
    <div class="app-container grid-layout" id="layoutRoot">
        <!-- Header -->
        <header class="app-header">
            <div class="header-content">
                <h1 class="app-title">{{.Config.Name}}</h1>
                <div class="header-actions">
                    <button id="toggle-debug" class="header-btn" title="Toggle Debug Trace" aria-expanded="false">📈 Debug Trace</button>
                    <button id="toggle-config" class="header-btn" title="Toggle Settings" aria-expanded="false">⚙️ Config</button>
                        <div class="theme-selector-container">
                            <label for="theme-select" class="agent-label">Theme:</label>
                            <select id="theme-select" class="agent-select" title="Mermaid diagram theme">
                                <option value="system" selected>System</option>
                                <option value="light">Light</option>
                                <option value="dark">Dark</option>
                            </select>
                        </div>
                    <div class="agent-selector-container">
                        <label for="agent-select" class="agent-label">Agent:</label>
                        <select id="agent-select" class="agent-select"></select>
                        <label class="orchestration-toggle">
                            <input type="checkbox" id="orchestration-mode" />
                            <span class="toggle-text">Full Workflow</span>
                        </label>
                    </div>
                </div>
            </div>
        </header>

        <!-- Main 3-pane Content -->
        <main class="main-3pane" id="main3pane">
            <!-- Left: Debug/Trace Pane -->
            <aside id="leftPane" class="pane left" aria-hidden="true">
                <div class="pane-header">
                    <h3>Debug Trace</h3>
                    <div class="pane-actions">
                        <button id="refresh-trace" class="mini-btn" title="Refresh">⟳</button>
                        <label class="mini-toggle" title="Linear view">
                            <input type="checkbox" id="trace-linear" checked /> Linear
                        </label>
                        <label class="mini-toggle" title="Show code">
                            <input type="checkbox" id="trace-code-toggle" /> Code
                        </label>
                    </div>
                </div>
                <div class="pane-body">
                    <div id="trace-diagram" class="trace-diagram" role="img" aria-label="Sequence diagram"></div>
                    <pre id="trace-raw" class="trace-raw" hidden></pre>
                </div>
            </aside>

            <!-- Center: Chat Column -->
            <section id="centerPane" class="center">
                <div class="agent-info-panel" id="agent-info">
                    <div class="agent-details">
                        <h3 id="agent-name">Loading...</h3>
                        <p id="agent-description">Loading agents from configuration...</p>
                    </div>
                </div>
                <div class="chat-container">
                    <div class="chat-messages" id="chat-messages">
                        <div class="welcome-message">
                            <h3>Welcome to {{.Config.Name}}!</h3>
                            <p>Choose an agent and start chatting to get assistance with your tasks.</p>
                            <p><strong>Tip:</strong> Toggle "Full Workflow" for orchestrated multi-agent processing, or leave off for direct agent responses.
                            Use 📈 and ⚙️ to open the Debug Trace and Config panels.</p>
                        </div>
                    </div>
                </div>
                <footer class="chat-input-container">
                    <form id="chat-form" class="chat-form">
                        <input type="text" id="message-input" class="message-input" placeholder="Type your message here..." autocomplete="off" />
                        <button type="submit" class="send-button" aria-label="Send message">
                            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
                                <line x1="22" y1="2" x2="11" y2="13"></line>
                                <polygon points="22,2 15,22 11,13 2,9"></polygon>
                            </svg>
                        </button>
                    </form>
                </footer>
            </section>

            <!-- Right: Settings/Config Pane -->
            <aside id="rightPane" class="pane right" aria-hidden="true">
                <div class="pane-header">
                    <h3>Settings</h3>
                    <div class="pane-actions">
                        <button id="reload-config" class="mini-btn" title="Reload">⟳</button>
                        <button id="save-config" class="mini-btn primary" title="Save">💾</button>
                    </div>
                </div>
                <div class="pane-body">
                    <label for="config-editor" class="sr-only">agentflow.toml</label>
                    <textarea id="config-editor" spellcheck="false" class="config-editor" placeholder="agentflow.toml will load here..."></textarea>
                    <div id="config-status" class="config-status" role="status" aria-live="polite"></div>
                </div>
            </aside>
        </main>
    </div>

    <script src="app.js"></script>
</body>
</html>`
