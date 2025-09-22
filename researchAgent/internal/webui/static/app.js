// State
let agents = {};
let cfgFeatures = { websocket: false, streaming: false };

// DOM refs
let chatMessages, messageInput, chatForm, agentSelect, agentName, agentDescription;
let layoutRoot, leftPane, rightPane, toggleDebugBtn, toggleConfigBtn;
let themeSelect;
let traceDiagramEl, traceRawEl, traceLinearCheckbox, refreshTraceBtn, traceCodeToggle;
let configEditor, configStatus, reloadConfigBtn, saveConfigBtn;

// WS
let ws = null; let wsReady = false;
let currentAgentMessageDiv = null; let currentAgentContentDiv = null;

document.addEventListener('DOMContentLoaded', () => {
    // Core chat refs
    chatMessages = document.getElementById('chat-messages');
    messageInput = document.getElementById('message-input');
    chatForm = document.getElementById('chat-form');
    agentSelect = document.getElementById('agent-select');
    agentName = document.getElementById('agent-name');
    agentDescription = document.getElementById('agent-description');

    // Layout & panes
    layoutRoot = document.getElementById('layoutRoot');
    leftPane = document.getElementById('leftPane');
    rightPane = document.getElementById('rightPane');
    toggleDebugBtn = document.getElementById('toggle-debug');
    toggleConfigBtn = document.getElementById('toggle-config');
    themeSelect = document.getElementById('theme-select');

    // Trace
    traceDiagramEl = document.getElementById('trace-diagram');
    traceRawEl = document.getElementById('trace-raw');
        traceLinearCheckbox = document.getElementById('trace-linear');
        traceCodeToggle = document.getElementById('trace-code-toggle');
    refreshTraceBtn = document.getElementById('refresh-trace');

    // Config
    configEditor = document.getElementById('config-editor');
    configStatus = document.getElementById('config-status');
    reloadConfigBtn = document.getElementById('reload-config');
    saveConfigBtn = document.getElementById('save-config');

    // Events
    chatForm.addEventListener('submit', handleSubmit);
    agentSelect.addEventListener('change', handleAgentChange);
    messageInput.addEventListener('keydown', handleKeyDown);

    toggleDebugBtn.addEventListener('click', () => {
        const isHidden = leftPane.getAttribute('aria-hidden') === 'true';
        leftPane.setAttribute('aria-hidden', isHidden ? 'false' : 'true');
        toggleDebugBtn.setAttribute('aria-expanded', isHidden ? 'true' : 'false');
        layoutRoot.classList.toggle('left-open', isHidden);
        if (isHidden) loadTraceDiagram();
    });
    toggleConfigBtn.addEventListener('click', () => {
        const isHidden = rightPane.getAttribute('aria-hidden') === 'true';
        rightPane.setAttribute('aria-hidden', isHidden ? 'false' : 'true');
        toggleConfigBtn.setAttribute('aria-expanded', isHidden ? 'true' : 'false');
        layoutRoot.classList.toggle('right-open', isHidden);
        if (isHidden && configEditor.value.trim().length === 0) loadRawConfig();
    });
        traceLinearCheckbox.addEventListener('change', () => loadTraceDiagram());
    refreshTraceBtn.addEventListener('click', () => loadTraceDiagram());
        if (traceCodeToggle) traceCodeToggle.addEventListener('change', () => applyTraceCodeVisibility());
    reloadConfigBtn.addEventListener('click', () => loadRawConfig());
    saveConfigBtn.addEventListener('click', () => saveRawConfig());

    if (themeSelect) {
        // Initialize from localStorage
        const savedThemePref = localStorage.getItem('traceTheme') || 'system';
        themeSelect.value = savedThemePref;
        applyPageTheme(themeSelect.value);
        themeSelect.addEventListener('change', () => {
            localStorage.setItem('traceTheme', themeSelect.value);
            applyPageTheme(themeSelect.value);
            loadTraceDiagram();
        });
    }

    // Init
    loadConfig().then(() => {
        loadAgents();
        if (cfgFeatures.websocket) connectWebSocket();
    });
    messageInput.focus();
});

async function loadConfig() {
    try {
        const res = await fetch('/api/config');
        if (!res.ok) return;
        const cfg = await res.json();
        const orchToggle = document.getElementById('orchestration-mode');
        if (cfg && cfg.orchestration) {
            const def = !!cfg.orchestration.default_enabled;
            orchToggle.checked = def;
            orchToggle.title = 'Workflow mode: ' + ((cfg.orchestration.mode) ? cfg.orchestration.mode : 'direct');
        }
        if (cfg && cfg.features) {
            cfgFeatures.websocket = !!cfg.features.websocket;
            cfgFeatures.streaming = !!cfg.features.streaming;
        }
    } catch (e) { console.debug('config fetch failed', e); }
}

function connectWebSocket() {
    try {
        const protocol = location.protocol === 'https:' ? 'wss' : 'ws';
        ws = new WebSocket(protocol + '://' + location.host + '/ws');
        ws.onopen = () => { wsReady = true; document.title = document.title + ' [WS]'; };
        ws.onclose = () => { wsReady = false; };
        ws.onerror = (e) => console.error('WebSocket error', e);
        ws.onmessage = (ev) => {
            let msg; try { msg = JSON.parse(ev.data); } catch { return; }
            const t = msg.type;
            if (t === 'welcome') return;
            if (t === 'agent_progress') {
                if (!currentAgentMessageDiv) {
                    const agentDisplayName = (agents[msg.agent] && agents[msg.agent].name) ? agents[msg.agent].name : (msg.agent || 'Agent');
                    const parts = beginAgentStreaming(agentDisplayName);
                    currentAgentMessageDiv = parts.container;
                    currentAgentContentDiv = parts.content;
                }
                return;
            }
            if (t === 'agent_chunk') {
                if (!currentAgentMessageDiv) {
                    const agentDisplayName = (agents[msg.agent] && agents[msg.agent].name) ? agents[msg.agent].name : (msg.agent || 'Agent');
                    const parts = beginAgentStreaming(agentDisplayName);
                    currentAgentMessageDiv = parts.container;
                    currentAgentContentDiv = parts.content;
                }
                if (currentAgentContentDiv && msg.content) {
                    currentAgentContentDiv.textContent += msg.content;
                    scrollToBottom();
                }
                return;
            }
            if (t === 'agent_complete') {
                if (!currentAgentMessageDiv) {
                    const agentDisplayName = (agents[msg.agent] && agents[msg.agent].name) ? agents[msg.agent].name : (msg.agent || 'Agent');
                    const parts = beginAgentStreaming(agentDisplayName);
                    currentAgentMessageDiv = parts.container;
                    currentAgentContentDiv = parts.content;
                }
                if (currentAgentContentDiv && msg.content) {
                    currentAgentContentDiv.textContent = msg.content;
                }
                endAgentStreaming();
                setFormDisabled(false);
                messageInput.focus();
                return;
            }
            if (t === 'error') {
                endAgentStreaming();
                addMessage('agent', 'Error: ' + (msg.content || 'Unknown error'), 'System');
                setFormDisabled(false);
                messageInput.focus();
                return;
            }
        };
    } catch (e) { console.warn('WS init failed; will use HTTP fallback', e); }
}

// Agents
async function loadAgents() {
    try {
        const response = await fetch('/api/agents');
        if (!response.ok) throw new Error('Failed to load agents: ' + response.status);
        const agentList = await response.json();
        agentSelect.innerHTML = '';
        agents = {};
        agentList.forEach(agent => {
            agents[agent.id] = { name: agent.name, description: agent.description };
            const option = document.createElement('option'); option.value = agent.id; option.textContent = agent.name; agentSelect.appendChild(option);
        });
        updateAgentInfo();
    } catch (error) {
        console.error('Error loading agents:', error);
        agents = { 'agent1': { name: 'Agent1', description: 'Default agent from configuration' }, 'agent2': { name: 'Agent2', description: 'Default agent from configuration' } };
        agentSelect.innerHTML = '';
        Object.keys(agents).forEach(id => { const o = document.createElement('option'); o.value = id; o.textContent = agents[id].name; agentSelect.appendChild(o); });
        updateAgentInfo();
    }
}

// Chat submit
async function handleSubmit(e) {
    e.preventDefault();
    const message = messageInput.value.trim(); if (!message) return;
    const selectedAgent = agentSelect.value;

    messageInput.value = '';
    setFormDisabled(true);
    addMessage('user', message, 'You');
    const typingId = addTypingIndicator(selectedAgent);

    try {
        const orch = document.getElementById('orchestration-mode').checked;
        if (wsReady && cfgFeatures.websocket) {
            removeTypingIndicator(typingId);
            const agentDisplayName = (agents[selectedAgent] && agents[selectedAgent].name) ? agents[selectedAgent].name : selectedAgent;
            const parts = beginAgentStreaming(agentDisplayName);
            currentAgentMessageDiv = parts.container; currentAgentContentDiv = parts.content;
            ws.send(JSON.stringify({ type: 'chat', agent: selectedAgent, message, useOrchestration: orch }));
            return; // re-enable on agent_complete
        }
        const response = await fetch('/api/chat', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ message, agent: selectedAgent, useOrchestration: orch }) });
        if (!response.ok) { const errorText = await response.text(); throw new Error('HTTP ' + response.status + ': ' + errorText); }
        const data = await response.json();
        removeTypingIndicator(typingId);
        addMessage('agent', data.response, agents[selectedAgent].name);
    } catch (error) {
        console.error('Chat error:', error);
        removeTypingIndicator(typingId);
        addMessage('agent', 'Sorry, I encountered an error: ' + error.message + '. See console for details.', 'System');
    } finally {
        if (!(wsReady && cfgFeatures.websocket)) { setFormDisabled(false); messageInput.focus(); }
    }
}

function handleAgentChange() { updateAgentInfo(); messageInput.focus(); }
function handleKeyDown(e) { if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); chatForm.dispatchEvent(new Event('submit')); } }

function updateAgentInfo() {
    const selectedAgent = agentSelect.value; const agent = agents[selectedAgent];
    if (agent) { agentName.textContent = agent.name; agentDescription.textContent = agent.description; }
    else { agentName.textContent = 'Loading...'; agentDescription.textContent = 'Loading agent information...'; }
}

// Messages
function addMessage(sender, content, senderName) {
    const messageDiv = document.createElement('div'); messageDiv.className = 'message ' + sender;
    const contentDiv = document.createElement('div'); contentDiv.className = 'message-content'; contentDiv.textContent = content;
    const infoDiv = document.createElement('div'); infoDiv.className = 'message-info'; infoDiv.textContent = senderName + ' • ' + new Date().toLocaleTimeString();
    messageDiv.appendChild(contentDiv); messageDiv.appendChild(infoDiv);
    const welcomeMessage = chatMessages.querySelector('.welcome-message'); if (welcomeMessage) welcomeMessage.remove();
    chatMessages.appendChild(messageDiv); scrollToBottom();
}

function beginAgentStreaming(agentDisplayName) {
    const messageDiv = document.createElement('div'); messageDiv.className = 'message agent';
    const contentDiv = document.createElement('div'); contentDiv.className = 'message-content'; contentDiv.textContent = '';
    const infoDiv = document.createElement('div'); infoDiv.className = 'message-info'; infoDiv.textContent = agentDisplayName + ' • ' + new Date().toLocaleTimeString();
    messageDiv.appendChild(contentDiv); messageDiv.appendChild(infoDiv);
    const welcomeMessage = chatMessages.querySelector('.welcome-message'); if (welcomeMessage) welcomeMessage.remove();
    chatMessages.appendChild(messageDiv); scrollToBottom();
    return { container: messageDiv, content: contentDiv };
}
function endAgentStreaming() { currentAgentMessageDiv = null; currentAgentContentDiv = null; }

function addTypingIndicator(agentNameKey) {
    const typingDiv = document.createElement('div'); const typingId = 'typing-' + Date.now(); typingDiv.id = typingId; typingDiv.className = 'message agent';
    const agentDisplayName = (agents[agentNameKey] && agents[agentNameKey].name) ? agents[agentNameKey].name : agentNameKey;
    typingDiv.innerHTML = '<div class="message-content typing-indicator">' +
        '<span>' + agentDisplayName + ' is typing</span>' +
        '<div class="typing-dots">' +
        '<div class="typing-dot"></div>' +
        '<div class="typing-dot"></div>' +
        '<div class="typing-dot"></div>' +
        '</div>' +
    '</div>';
    chatMessages.appendChild(typingDiv); scrollToBottom(); return typingId;
}
function removeTypingIndicator(typingId) { const div = document.getElementById(typingId); if (div) div.remove(); }
function setFormDisabled(disabled) { messageInput.disabled = disabled; chatForm.querySelector('button[type="submit"]').disabled = disabled; }
function scrollToBottom() { chatMessages.scrollTop = chatMessages.scrollHeight; }

// Trace diagram
async function loadTraceDiagram() {
        try {
        if (!traceDiagramEl) return;
            const linear = traceLinearCheckbox && traceLinearCheckbox.checked ? 'true' : 'false';
            // Determine theme preference
            let themePref = 'system';
            if (themeSelect) themePref = themeSelect.value || 'system';
            let theme = '';
            if (themePref === 'system') {
                const prefersDark = window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches;
                theme = prefersDark ? 'dark' : 'light';
            } else {
                theme = themePref;
            }
            const res = await fetch('/api/visualization/trace?linear=' + linear + '&theme=' + encodeURIComponent(theme));
        if (!res.ok) throw new Error('Failed to load trace: ' + res.status);
        const obj = await res.json();
        const data = obj && obj.data ? obj.data : {};
        const code = data.diagram || '';
        const labels = Array.isArray(data.labels) ? data.labels : [];
        traceRawEl.textContent = code;
        // Render mermaid
        try {
            const id = 'trace_' + Date.now();
            const out = await mermaid.render(id, code);
            traceDiagramEl.innerHTML = out.svg;
            if (out.bindFunctions) out.bindFunctions(traceDiagramEl);
            // Attach label tooltips: find label texts and set title attributes
                    try {
                        const labelMap = new Map(labels.map(l => [String(l.id || l.ID || l.Id), l]));
                        const svgNS = 'http://www.w3.org/2000/svg';
                        // Mermaid renders labels as <text>M1</text>; add <title> for tooltips
                        const texts = traceDiagramEl.querySelectorAll('text');
                        texts.forEach(t => {
                            const txt = (t.textContent || '').trim();
                            if (!labelMap.has(txt)) return;
                            const info = labelMap.get(txt);
                            const tip = (info.message || info.Message || '').toString();
                            if (!tip) return;
                            // Remove existing <title>
                            [...t.querySelectorAll('title')].forEach(el => el.remove());
                            const titleEl = document.createElementNS(svgNS, 'title');
                            titleEl.textContent = tip;
                            t.appendChild(titleEl);
                            // Also add to parent <g> if present to increase hover area
                            const p = t.parentElement;
                            if (p && p.namespaceURI === svgNS) {
                                [...p.querySelectorAll(':scope > title')].forEach(el => el.remove());
                                const titleEl2 = document.createElementNS(svgNS, 'title');
                                titleEl2.textContent = tip;
                                p.appendChild(titleEl2);
                            }
                        });
                    } catch {}
        } catch (mermErr) {
            console.warn('Mermaid render failed, showing raw.', mermErr);
            traceDiagramEl.textContent = code;
        }
        applyTraceCodeVisibility();
    } catch (e) {
        console.error('Trace load error', e);
        traceDiagramEl.textContent = 'Failed to load trace: ' + e.message;
    }
}

function applyPageTheme(pref) {
    // Pref can be 'system' | 'light' | 'dark'
    const root = document.documentElement; // <html>
    root.classList.remove('theme-light', 'theme-dark');
    if (pref === 'light') root.classList.add('theme-light');
    else if (pref === 'dark') root.classList.add('theme-dark');
    // 'system' means rely on media query variables
}

function applyTraceCodeVisibility() {
    if (!traceCodeToggle || !traceRawEl || !traceDiagramEl) return;
    const showCode = !!traceCodeToggle.checked;
    traceRawEl.hidden = !showCode;
    traceDiagramEl.style.display = showCode ? 'none' : '';
}

// Config raw editor
async function loadRawConfig() {
    try {
        const res = await fetch('/api/config/raw');
        if (!res.ok) throw new Error('Failed to load config: ' + res.status);
        const obj = await res.json();
        const txt = obj && obj.data && typeof obj.data.content === 'string' ? obj.data.content : '';
        configEditor.value = txt;
        configStatus.textContent = 'Config loaded at ' + new Date().toLocaleTimeString();
    } catch (e) {
        configStatus.textContent = 'Load failed: ' + e.message;
    }
}
async function saveRawConfig() {
    try {
        const body = JSON.stringify({ toml: configEditor.value });
        const res = await fetch('/api/config/raw', { method: 'PUT', headers: { 'Content-Type': 'application/json' }, body });
        if (!res.ok) { const t = await res.text(); throw new Error('Save failed: ' + res.status + ' ' + t); }
        configStatus.textContent = 'Saved at ' + new Date().toLocaleTimeString();
        // Refresh /api/config derived features optionally
        loadConfig();
    } catch (e) {
        configStatus.textContent = 'Save failed: ' + e.message;
    }
}
