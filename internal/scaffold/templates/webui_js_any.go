//go:build !js
// +build !js

package templates

const WebUIJSTemplate = `// Agent configurations - loaded dynamically from API
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
`
