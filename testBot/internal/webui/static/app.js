// Agent configurations - loaded dynamically from API
let agents = {};

// DOM elements
let chatMessages, messageInput, chatForm, agentSelect, agentName, agentDescription;

// WebSocket state
let ws = null;
let wsReady = false;
let currentAgentMessageDiv = null; // container for streaming updates
let currentAgentContentDiv = null; // content element for streaming text
let cfgFeatures = { websocket: false, streaming: false };

// Initialize the application
document.addEventListener('DOMContentLoaded', function() {
    // Get DOM elements
    chatMessages = document.getElementById('chat-messages');
    messageInput = document.getElementById('message-input');
    chatForm = document.getElementById('chat-form');
    agentSelect = document.getElementById('agent-select');
    agentName = document.getElementById('agent-name');
    agentDescription = document.getElementById('agent-description');

    // Set up event listeners
    chatForm.addEventListener('submit', handleSubmit);
    agentSelect.addEventListener('change', handleAgentChange);
    messageInput.addEventListener('keydown', handleKeyDown);

    // Settings panel & actions
    const settingsPanel = document.getElementById('settings-panel');
    const openSettingsBtn = document.getElementById('open-settings');
    const closeSettingsBtn = document.getElementById('close-settings');
    const loadBtn = document.getElementById('load-config');
    const saveBtn = document.getElementById('save-config');

    // Left Trace Pane elements
    const tracePane = document.getElementById('trace-pane');
    const openTraceBtn = document.getElementById('open-trace');
    const closeTraceBtn = document.getElementById('close-trace');
    const refreshTraceBtn = document.getElementById('refresh-trace');
    const toggleTraceCode = document.getElementById('toggle-trace-code');
    const toggleTraceRender = document.getElementById('toggle-trace-render');

    if (openSettingsBtn && settingsPanel) openSettingsBtn.addEventListener('click', () => settingsPanel.classList.add('open'));
    if (closeSettingsBtn && settingsPanel) closeSettingsBtn.addEventListener('click', () => settingsPanel.classList.remove('open'));
    if (loadBtn) loadBtn.addEventListener('click', loadAgentflowToml);
    if (saveBtn) saveBtn.addEventListener('click', saveAgentflowToml);

    if (openTraceBtn && tracePane) openTraceBtn.addEventListener('click', async () => {
        tracePane.classList.add('open');
        try { await refreshTrace(); } catch {}
    });
    if (closeTraceBtn && tracePane) closeTraceBtn.addEventListener('click', () => tracePane.classList.remove('open'));
    if (refreshTraceBtn) refreshTraceBtn.addEventListener('click', refreshTrace);
    if (toggleTraceCode) toggleTraceCode.addEventListener('click', () => setViewMode('trace', 'code'));
    if (toggleTraceRender) toggleTraceRender.addEventListener('click', () => setViewMode('trace', 'render'));

    // Load config to set defaults, then agents and WS
    loadConfig().then(() => {
        loadAgents();
        if (cfgFeatures.websocket) connectWebSocket();
    });

    // Focus input
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
    } catch (e) {
        console.debug('config fetch failed', e);
    }
}

// Config editor handlers
async function loadAgentflowToml() {
    try {
        const res = await fetch('/api/config/raw');
        const data = await res.json();
        if (res.ok && data.status === 'success' && data.data) {
            const editor = document.getElementById('config-editor');
            const pathEl = document.getElementById('config-path');
            if (editor) editor.value = data.data.content || '';
            if (pathEl) pathEl.textContent = data.data.path || '';
        } else {
            alert('Failed to load config');
        }
    } catch (e) {
        alert('Error loading config: ' + e.message);
    }
}

async function saveAgentflowToml() {
    try {
        const editor = document.getElementById('config-editor');
        const toml = editor ? editor.value : '';
        const res = await fetch('/api/config/raw', { method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ toml }) });
        const data = await res.json();
        if (res.ok && data.status === 'success') {
            alert('Config saved');
        } else {
            alert('Failed to save config: ' + (data.message || res.status));
        }
    } catch (e) {
        alert('Error saving config: ' + e.message);
    }
}

function getViewMode(kind) {
    try { return localStorage.getItem('viewMode:' + kind) || 'render'; } catch { return 'render'; }
}

function setViewMode(kind, mode) {
    try { localStorage.setItem('viewMode:' + kind, mode); } catch {}
    if (kind === 'trace') {
        const pre = document.getElementById('trace-pre');
        const container = document.getElementById('trace-render');
        if (pre && container) {
            if (mode === 'code') { pre.classList.remove('hidden'); container.classList.add('hidden'); }
            else { pre.classList.add('hidden'); container.classList.remove('hidden'); }
        }
    }
}

// Flow diagram removed as per requirement

async function refreshTrace() {
    try {
        const res = await fetch('/api/visualization/trace');
        const data = await res.json();
        const pre = document.getElementById('trace-pre');
        const container = document.getElementById('trace-render');
        if (res.ok && data.status === 'success' && data.data) {
            const code = data.data.diagram;
            const labels = Array.isArray(data.data.labels) ? data.data.labels : [];
            if (pre) pre.textContent = code;
            setViewMode('trace', getViewMode('trace'));
            if (container) {
                const ready = await ensureMermaid();
                if (!ready) {
                    container.innerHTML = '<div class="muted">Mermaid is not available (CDN blocked?). The raw code is shown above.</div>';
                } else {
                    try {
                        const traceId = 'traceDiagram-' + Date.now();
                        const { svg } = await mermaid.render(traceId, code);
                        container.innerHTML = svg;
                        // Attach tooltips using <title> elements so they work on hover
                        try { attachTraceTooltips(container, labels); } catch {}
                        try { container.scrollIntoView({ behavior: 'smooth', block: 'nearest' }); } catch {}
                    } catch (e) {
                        container.innerHTML = '<div class="muted">Mermaid render error</div>';
                    }
                }
            }
        } else {
            if (pre) pre.textContent = '// Failed to load trace diagram';
            if (container) container.innerHTML = '';
        }
    } catch (e) {
        const pre = document.getElementById('trace-pre');
        const container = document.getElementById('trace-render');
        if (pre) pre.textContent = '// Error: ' + e.message;
        if (container) container.innerHTML = '';
    }
}

// Ensure Mermaid is loaded (attempt dynamic CDN load if missing)
async function ensureMermaid() {
    if (window.mermaid) return true;
    const sources = [
        'https://cdn.jsdelivr.net/npm/mermaid@10/dist/mermaid.min.js',
        'https://unpkg.com/mermaid@10/dist/mermaid.min.js',
        'https://cdnjs.cloudflare.com/ajax/libs/mermaid/10.9.1/mermaid.min.js'
    ];
    for (const src of sources) {
        const ok = await new Promise((resolve) => {
            const s = document.createElement('script');
            s.src = src;
            s.crossOrigin = 'anonymous';
            s.referrerPolicy = 'no-referrer';
            s.onload = () => { try { mermaid.initialize({ startOnLoad: false }); } catch {} resolve(true); };
            s.onerror = () => resolve(false);
            document.head.appendChild(s);
        });
        if (ok && window.mermaid) return true;
    }
    return false;
}

// Attach tooltip to arrows by mapping label text (e.g., "M1") to full message using a simple lookup.
function attachTraceTooltips(container, labels) {
    if (!container) return;
    const map = new Map();
    for (const l of labels) { if (l && l.id && l.message) map.set(String(l.id), String(l.message)); }
    // mermaid renders arrow labels inside <text> nodes; we set a <title> as tooltip
    const texts = container.querySelectorAll('text');
    texts.forEach((t) => {
        const val = (t.textContent || '').trim();
        if (map.has(val)) {
            // Remove any existing title to avoid duplicates
            const prev = t.querySelector('title');
            if (prev) prev.remove();
            const title = document.createElementNS('http://www.w3.org/2000/svg', 'title');
            title.textContent = map.get(val);
            t.appendChild(title);
        }
    });
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
                // Auto-refresh trace view to reflect latest edges
                try { if (typeof refreshTrace === 'function') refreshTrace(); } catch {}
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
    } catch (e) {
        console.warn('WS init failed; will use HTTP fallback', e);
    }
}

// Load agents from the API
async function loadAgents() {
    try {
        const response = await fetch('/api/agents');
        if (!response.ok) {
            throw new Error('Failed to load agents: ' + response.status);
        }
        const agentList = await response.json();
        
        // Clear existing options
        agentSelect.innerHTML = '';
        agents = {};
        
        // Populate agents
        agentList.forEach(agent => {
            agents[agent.id] = {
                name: agent.name,
                description: agent.description
            };
            
            const option = document.createElement('option');
            option.value = agent.id;
            option.textContent = agent.name;
            agentSelect.appendChild(option);
        });
        
        // Update initial agent info
        updateAgentInfo();
        
    } catch (error) {
        console.error('Error loading agents:', error);
        
        // Fallback to default agents
        agents = {
            'agent1': { name: 'Agent1', description: 'Default agent from configuration' },
            'agent2': { name: 'Agent2', description: 'Default agent from configuration' }
        };
        
        agentSelect.innerHTML = '';
        Object.keys(agents).forEach(agentId => {
            const option = document.createElement('option');
            option.value = agentId;
            option.textContent = agents[agentId].name;
            agentSelect.appendChild(option);
        });
        
        updateAgentInfo();
    }
}

// Handle form submission
async function handleSubmit(e) {
    e.preventDefault();
    
    const message = messageInput.value.trim();
    if (!message) return;

    const selectedAgent = agentSelect.value;
    
    // Clear input and disable form
    messageInput.value = '';
    setFormDisabled(true);

    // Add user message to chat
    addMessage('user', message, 'You');

    // Add typing indicator
    const typingId = addTypingIndicator(selectedAgent);

    try {
        const orch = document.getElementById('orchestration-mode').checked;

        if (wsReady && cfgFeatures.websocket) {
            // Use WebSocket streaming path
            console.log('Using WebSocket transport');
            removeTypingIndicator(typingId);
            const agentDisplayName = (agents[selectedAgent] && agents[selectedAgent].name) ? agents[selectedAgent].name : selectedAgent;
            const parts = beginAgentStreaming(agentDisplayName);
            currentAgentMessageDiv = parts.container;
            currentAgentContentDiv = parts.content;

            ws.send(JSON.stringify({
                type: 'chat',
                agent: selectedAgent,
                message: message,
                useOrchestration: orch
            }));
            // Re-enable form in agent_complete handler
            return;
        }

        // Fallback to HTTP request/response
        console.log('Using HTTP transport');
        const response = await fetch('/api/chat', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                message: message,
                agent: selectedAgent,
                useOrchestration: orch
            })
        });

        console.log('Response status:', response.status);

        if (!response.ok) {
            const errorText = await response.text();
            console.error('HTTP error response:', errorText);
            throw new Error('HTTP error! status: ' + response.status + ', body: ' + errorText);
        }

        const data = await response.json();
        console.log('Response data:', data);
        
        // Remove typing indicator
        removeTypingIndicator(typingId);
        
        // Add agent response
        addMessage('agent', data.response, agents[selectedAgent].name);

    // Auto-refresh trace view after HTTP interaction
    try { if (typeof refreshTrace === 'function') refreshTrace(); } catch {}

    } catch (error) {
        console.error('Error details:', error);
        console.error('Error stack:', error.stack);
        
        // Remove typing indicator
        removeTypingIndicator(typingId);
        
        // Add error message with more details
        addMessage('agent', 'Sorry, I encountered an error: ' + error.message + '. Please check the console for more details.', 'System');
    } finally {
        if (!(wsReady && cfgFeatures.websocket)) {
            setFormDisabled(false);
            messageInput.focus();
        }
    }
}

// Handle agent selection change
function handleAgentChange() {
    updateAgentInfo();
    messageInput.focus();
}

// Handle keyboard shortcuts
function handleKeyDown(e) {
    if (e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault();
        chatForm.dispatchEvent(new Event('submit'));
    }
}

// Update agent info panel
function updateAgentInfo() {
    const selectedAgent = agentSelect.value;
    const agent = agents[selectedAgent];
    
    if (agent) {
        agentName.textContent = agent.name;
        agentDescription.textContent = agent.description;
    } else {
        agentName.textContent = 'Loading...';
        agentDescription.textContent = 'Loading agent information...';
    }
}

// Add message to chat
function addMessage(sender, content, senderName) {
    const messageDiv = document.createElement('div');
    messageDiv.className = 'message ' + sender;
    
    const contentDiv = document.createElement('div');
    contentDiv.className = 'message-content';
    contentDiv.textContent = content;
    
    const infoDiv = document.createElement('div');
    infoDiv.className = 'message-info';
    infoDiv.textContent = senderName + ' • ' + new Date().toLocaleTimeString();
    
    messageDiv.appendChild(contentDiv);
    messageDiv.appendChild(infoDiv);
    
    const welcomeMessage = chatMessages.querySelector('.welcome-message');
    if (welcomeMessage) { welcomeMessage.remove(); }
    
    chatMessages.appendChild(messageDiv);
    scrollToBottom();
}

// Begin an agent streaming message; returns refs to container and content element
function beginAgentStreaming(agentDisplayName) {
    const messageDiv = document.createElement('div');
    messageDiv.className = 'message agent';

    const contentDiv = document.createElement('div');
    contentDiv.className = 'message-content';
    contentDiv.textContent = '';

    const infoDiv = document.createElement('div');
    infoDiv.className = 'message-info';
    infoDiv.textContent = agentDisplayName + ' • ' + new Date().toLocaleTimeString();

    messageDiv.appendChild(contentDiv);
    messageDiv.appendChild(infoDiv);

    const welcomeMessage = chatMessages.querySelector('.welcome-message');
    if (welcomeMessage) welcomeMessage.remove();

    chatMessages.appendChild(messageDiv);
    scrollToBottom();

    return { container: messageDiv, content: contentDiv };
}

function endAgentStreaming() {
    currentAgentMessageDiv = null;
    currentAgentContentDiv = null;
}

// Typing indicator helpers and other utilities (kept for compatibility)
function addTypingIndicator(agentNameKey) {
    const typingDiv = document.createElement('div');
    const typingId = 'typing-' + Date.now();
    typingDiv.id = typingId;
    typingDiv.className = 'message agent';
    const agentDisplayName = (agents[agentNameKey] && agents[agentNameKey].name) ? agents[agentNameKey].name : agentNameKey;
    typingDiv.innerHTML = '<div class="message-content typing-indicator">' +
        '<span>' + agentDisplayName + ' is typing</span>' +
        '<div class="typing-dots">' +
        '<div class="typing-dot"></div>' +
        '<div class="typing-dot"></div>' +
        '<div class="typing-dot"></div>' +
        '</div>' +
        '</div>';
    chatMessages.appendChild(typingDiv);
    scrollToBottom();
    return typingId;
}

function removeTypingIndicator(typingId) {
    const typingDiv = document.getElementById(typingId);
    if (typingDiv) typingDiv.remove();
}

function setFormDisabled(disabled) {
    messageInput.disabled = disabled;
    const submitButton = chatForm.querySelector('button[type="submit"]');
    submitButton.disabled = disabled;
}

function scrollToBottom() { chatMessages.scrollTop = chatMessages.scrollHeight; }
function autoResize(element) { element.style.height = 'auto'; element.style.height = element.scrollHeight + 'px'; }
