// AgenticGoKit Chat Interface JavaScript

class AgenticChatApp {
    constructor() {
        this.websocket = null;
        this.currentSessionId = null;
        this.currentAgent = null;
        this.agents = [];
        this.isConnected = false;
        this.features = { websocket: false, streaming: false };
        this.messageQueue = [];
        this.sessions = new Map();
        this.settings = {
            serverUrl: null,
            autoReconnect: true,
            reconnectDelay: 3000,
            maxReconnectAttempts: 5,
            typingIndicatorDelay: 1000,
        };
        this.reconnectAttempts = 0;
        this.isTyping = false;
    this.typingTimeout = null;
    this.sendingMessage = false; // Add debouncing flag
    this.activeStreams = new Map(); // key: agentName -> { el, content, lastIndex }
        
        this.init();
    }

    init() {
        this.setupEventListeners();
    this.loadSettings();
    this.fetchConfig().then(() => this.connect());
        this.loadAgents();
        this.loadSessions();
    }

    setupEventListeners() {
        // Send button and enter key
        const sendButton = document.getElementById('send-button');
        const chatInput = document.getElementById('chat-input');
        
        sendButton.addEventListener('click', () => this.sendMessage());
        chatInput.addEventListener('keypress', (e) => {
            if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                this.sendMessage();
            }
        });

        // Auto-resize textarea
        chatInput.addEventListener('input', () => this.autoResizeTextarea(chatInput));

        // New chat button
        document.getElementById('new-chat-btn').addEventListener('click', () => this.createNewSession());

        // Settings button
        document.getElementById('settings-btn').addEventListener('click', () => this.toggleSettings());
        document.getElementById('close-settings').addEventListener('click', () => this.closeSettings());

        // Settings form
        document.getElementById('settings-form').addEventListener('submit', (e) => {
            e.preventDefault();
            this.saveSettings();
        });

    // Config editor actions
    const loadBtn = document.getElementById('load-config-btn');
    const saveBtn = document.getElementById('save-config-btn');
    if (loadBtn) loadBtn.addEventListener('click', () => this.loadAgentflowToml());
    if (saveBtn) saveBtn.addEventListener('click', () => this.saveAgentflowToml());

    // Diagram viewer
    const showDiagramBtn = document.getElementById('show-diagram-btn');
    if (showDiagramBtn) showDiagramBtn.addEventListener('click', () => this.refreshDiagram());

        // Mobile menu toggle (for responsive design)
        const menuToggle = document.getElementById('menu-toggle');
        if (menuToggle) {
            menuToggle.addEventListener('click', () => this.toggleMobileMenu());
        }

        // Handle window resize for responsive behavior
        window.addEventListener('resize', () => this.handleResize());

        // Handle visibility change for connection management
        document.addEventListener('visibilitychange', () => {
            if (document.visibilityState === 'visible' && !this.isConnected) {
                this.connect();
            }
        });
    }

    autoResizeTextarea(textarea) {
        textarea.style.height = 'auto';
        const newHeight = Math.min(textarea.scrollHeight, 120);
        textarea.style.height = newHeight + 'px';
    }

    connect() {
        // Use HTTP-only mode if server doesn't advertise websocket
        if (!this.features.websocket) {
            console.info('[WebUI] Running in HTTP-only mode (WebSocket disabled by server config)');
            this.isConnected = false;
            this.updateConnectionStatus(false);
            return;
        }

        if (this.websocket && this.websocket.readyState === WebSocket.OPEN) {
            return;
        }

        const wsProto = window.location.protocol === 'https:' ? 'wss' : 'ws';
        const wsUrl = `${wsProto}://${window.location.host}/ws`;
        this.settings.serverUrl = `${wsProto}://${window.location.host}`;
    console.info('[WebUI] Connecting to WebSocket:', wsUrl);

        try {
            this.websocket = new WebSocket(wsUrl);
            this.setupWebSocketHandlers();
        } catch (error) {
            console.error('[WebUI] Failed to create WebSocket connection:', error);
            this.handleConnectionError();
        }
    }

    setupWebSocketHandlers() {
        this.websocket.onopen = () => {
            console.log('WebSocket connected');
            this.isConnected = true;
            this.reconnectAttempts = 0;
            this.updateConnectionStatus(true);
            this.processMessageQueue();
            
            // Request initial session from server if none exists
            if (!this.currentSessionId) {
                this.sendSessionCreate();
            }
        };

        this.websocket.onmessage = (event) => {
            try {
                const message = JSON.parse(event.data);
                this.handleWebSocketMessage(message);
            } catch (error) {
                console.error('Failed to parse WebSocket message:', error);
            }
        };

        this.websocket.onclose = (event) => {
            console.log('WebSocket disconnected:', event.code, event.reason);
            this.isConnected = false;
            this.updateConnectionStatus(false);
            
            if (this.settings.autoReconnect && this.reconnectAttempts < this.settings.maxReconnectAttempts) {
                this.scheduleReconnect();
            }
        };

        this.websocket.onerror = (error) => {
            console.error('WebSocket error:', error);
            this.handleConnectionError();
        };
    }

    handleWebSocketMessage(message) {
        console.log('Received message:', message);

        switch (message.type) {
            case 'session_status':
                this.handleSessionStatus(message);
                break;
            case 'agent_response':
                this.handleAgentResponse(message);
                break;
            case 'agent_chunk':
                this.handleAgentResponseChunk(message);
                break;
            case 'agent_complete':
                this.handleAgentResponseComplete(message);
                break;
            case 'agent_progress':
                this.showTypingIndicator();
                break;
            case 'agent_error':
                this.handleError({ message: (message.data && message.data.message) || 'Agent error' });
                break;
            case 'error':
                this.handleError(message);
                break;
            case 'pong':
                // Handle pong response
                break;
            default:
                console.warn('Unknown message type:', message.type);
        }
    }

    handleSessionStatus(message) {
        const sessionId = (message.data && message.data.session_id) || message.session_id;
        if (!sessionId) return;
        this.currentSessionId = sessionId;
        if (!this.sessions.has(sessionId)) {
            this.sessions.set(sessionId, {
                id: sessionId,
                messages: [],
                title: 'New Chat',
                created_at: new Date().toISOString(),
                last_message_at: new Date().toISOString()
            });
        }
        this.updateSessionsList();
        this.switchToSession(sessionId);
        this.processMessageQueue();
    }

    handleAgentResponse(message) {
        if (message.session_id !== this.currentSessionId) {
            return;
        }

        this.hideTypingIndicator();
        this.addMessage({
            type: 'agent',
            content: message.content,
            timestamp: new Date().toISOString(),
            agent_name: message.agent_name || 'Agent'
        });

        this.updateSessionLastMessage(message.session_id, message.content);
    }

    handleAgentResponseChunk(message) {
        if (message.session_id !== this.currentSessionId) {
            return;
        }

        // Handle streaming response chunks
        this.updateStreamingMessage(message);
    }

    handleAgentResponseComplete(message) {
        if (message.session_id !== this.currentSessionId) {
            return;
        }

        this.hideTypingIndicator();
        // Finalize streaming message
        this.finalizeStreamingMessage(message);
    }

    handleError(message) {
        console.error('Received error:', message);
        this.showError(message.message || 'An error occurred');
        this.hideTypingIndicator();
    }

    handleConnectionError() {
        this.isConnected = false;
        this.updateConnectionStatus(false);
        this.showError('Connection lost. Attempting to reconnect...');
    }

    scheduleReconnect() {
        this.reconnectAttempts++;
        const delay = this.settings.reconnectDelay * Math.pow(1.5, this.reconnectAttempts - 1);
        
        console.log(`Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts})`);
        
        setTimeout(() => {
            if (!this.isConnected) {
                this.connect();
            }
        }, delay);
    }

    updateConnectionStatus(connected) {
        const statusIndicator = document.querySelector('.status-indicator');
        const statusText = document.querySelector('.status-text');
        const headerStatus = document.querySelector('.header-status');
        
        if (!statusIndicator || !statusText) {
            return; // Elements not found, skip update
        }
        
    // In HTTP-only mode, show "Ready" instead of connection status
        if (!this.features.websocket) {
            statusIndicator.classList.remove('disconnected');
            statusText.textContent = 'Ready (HTTP)';
            if (headerStatus) headerStatus.title = 'Using HTTP API (WebSocket disabled)';
            this.hideConnectionLostBanner();
            console.info('[WebUI] Transport mode: HTTP');
            return;
        }
        
        if (connected) {
            statusIndicator.classList.remove('disconnected');
            statusText.textContent = 'Connected (WebSocket)';
            if (headerStatus) headerStatus.title = 'Using WebSocket for real-time streaming';
            console.info('[WebUI] Transport mode: WebSocket');
            this.hideConnectionLostBanner();
        } else {
            statusIndicator.classList.add('disconnected');
            statusText.textContent = 'Disconnected';
            if (headerStatus) headerStatus.title = 'WebSocket disconnected';
            this.showConnectionLostBanner();
        }
    }

    showConnectionLostBanner() {
        // Don't show connection lost banner in HTTP-only mode
        if (!this.websocket) {
            return;
        }
        
        let banner = document.querySelector('.connection-lost');
        if (!banner) {
            banner = document.createElement('div');
            banner.className = 'connection-lost';
            banner.textContent = 'Connection lost. Attempting to reconnect...';
            document.querySelector('.chat-container').insertBefore(banner, document.querySelector('.chat-main'));
        }
    }

    hideConnectionLostBanner() {
        const banner = document.querySelector('.connection-lost');
        if (banner) {
            banner.remove();
        }
    }

    processMessageQueue() {
        while (this.messageQueue.length > 0 && this.isConnected) {
            const message = this.messageQueue.shift();
            this.websocket.send(JSON.stringify(message));
        }
    }

    sendMessage() {
        // Prevent double-sending
        if (this.sendingMessage) {
            console.log('Already sending a message, ignoring duplicate call');
            return;
        }
        
        const input = document.getElementById('chat-input');
        const message = input.value.trim();
        
        if (!message) {
            return;
        }

        if (!this.currentAgent) {
            this.showError('Please select an agent first');
            return;
        }

        // Set the sending flag
        this.sendingMessage = true;

        if (!this.currentSessionId) {
            this.createNewSession(this.currentAgent);
        }

        // Add user message to UI immediately
        this.addMessage({
            type: 'user',
            content: message,
            timestamp: new Date().toISOString()
        });

        // Update session
        this.updateSessionLastMessage(this.currentSessionId, message);

        // Clear input and show typing indicator
        input.value = '';
        input.style.height = 'auto';
        this.showTypingIndicator();

        // Send via WS if connected, else HTTP fallback
        if (this.websocket && this.isConnected) {
            console.debug('[WebUI] Sending message via WebSocket');
            this.sendMessageWS(message);
        } else {
            console.debug('[WebUI] Sending message via HTTP POST /api/chat');
            this.sendMessageToAgent(message);
        }
    }

    sendSessionCreate() {
        if (!this.websocket) return;
        const msg = {
            type: 'session_create',
            session_id: '',
            message_id: this.generateId(),
            timestamp: new Date().toISOString(),
            data: { user_agent: navigator.userAgent }
        };
    console.debug('[WebUI] Sending session_create');
    this.websocket.send(JSON.stringify(msg));
    }

    sendMessageWS(message) {
        if (!this.websocket) return;
        const payload = {
            type: 'chat_message',
            session_id: this.currentSessionId || '',
            message_id: this.generateId(),
            timestamp: new Date().toISOString(),
            data: { content: message, message_type: 'text' }
        };
        if (this.isConnected && this.currentSessionId) {
            console.debug('[WebUI] chat_message sent over WebSocket');
            this.websocket.send(JSON.stringify(payload));
        } else {
            this.messageQueue.push(payload);
            if (!this.currentSessionId) this.sendSessionCreate();
        }
    }

    async fetchConfig() {
        try {
            const res = await fetch('/api/config');
            const cfg = await res.json();
            if (cfg && cfg.status === 'success' && cfg.data && cfg.data.features) {
                this.features.websocket = !!cfg.data.features.websocket;
                this.features.streaming = !!cfg.data.features.streaming;
            }
        } catch (e) {
            this.features.websocket = false;
            this.features.streaming = false;
        }
    }

    generateId() {
        return 'msg_' + Math.random().toString(36).slice(2) + Date.now().toString(36);
    }

    async sendMessageToAgent(message) {
        try {
            const response = await fetch('/api/chat', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    agent_name: this.currentAgent,
                    message: message,
                    session_id: this.currentSessionId
                })
            });

            const data = await response.json();
            this.hideTypingIndicator();

            if (data.status === 'success' && data.data && data.data.response) {
                // Add agent response to UI
                this.addMessage({
                    type: 'agent',
                    content: data.data.response,
                    agent: this.currentAgent,
                    timestamp: new Date().toISOString()
                });
            } else {
                console.error('API response validation failed:', {
                    status: data.status,
                    hasData: !!data.data,
                    hasResponse: !!(data.data && data.data.response),
                    fullData: data
                });
                this.showError('Failed to get response from agent');
            }
        } catch (error) {
            this.hideTypingIndicator();
            console.error('Error sending message:', error);
            this.showError('Failed to send message. Please try again.');
        } finally {
            // Reset the sending flag regardless of success or failure
            this.sendingMessage = false;
        }
    }

    addMessage(messageData) {
        const messagesContainer = document.getElementById('chat-messages');
        const messageElement = this.createMessageElement(messageData);
        messagesContainer.appendChild(messageElement);
        this.scrollToBottom();

        // Add to session data
        if (this.currentSessionId && this.sessions.has(this.currentSessionId)) {
            this.sessions.get(this.currentSessionId).messages.push(messageData);
        }
    }

    clearChatMessages() {
        const messagesContainer = document.getElementById('chat-messages');
        messagesContainer.innerHTML = `
            <div class="welcome-message">
                <div class="welcome-content">
                    <h2>Welcome to AgenticGoKit Chat! 🤖</h2>
                    <p>Start a conversation with your AI agents. They're ready to help you with any task.</p>
                    <div class="welcome-features">
                        <div class="feature-item">
                            <span class="feature-icon">💬</span>
                            <span>Real-time chat with AI agents</span>
                        </div>
                        <div class="feature-item">
                            <span class="feature-icon">🔄</span>
                            <span>Multi-agent collaboration</span>
                        </div>
                        <div class="feature-item">
                            <span class="feature-icon">💾</span>
                            <span>Persistent chat sessions</span>
                        </div>
                        <div class="feature-item">
                            <span class="feature-icon">🔧</span>
                            <span>Tool integration & execution</span>
                        </div>
                    </div>
                </div>
            </div>
        `;
    }

    createMessageElement(messageData) {
        const messageDiv = document.createElement('div');
        messageDiv.className = `message ${messageData.type}`;

        const avatarText = messageData.type === 'user' ? 'U' : (messageData.agent ? messageData.agent[0].toUpperCase() : 'A');
        const senderName = messageData.type === 'user' ? 'You' : (messageData.agent || messageData.agent_name || 'Agent');
        const timestamp = new Date(messageData.timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });

        messageDiv.innerHTML = `
            <div class="message-header">
                <div class="message-avatar">${avatarText}</div>
                <span class="message-sender">${senderName}</span>
                <span class="message-time">${timestamp}</span>
            </div>
            <div class="message-content">${this.escapeHtml(messageData.content)}</div>
        `;

        return messageDiv;
    }

    showTypingIndicator() {
        if (document.querySelector('.typing-indicator')) {
            return;
        }

        const messagesContainer = document.getElementById('chat-messages');
        const typingDiv = document.createElement('div');
        typingDiv.className = 'typing-indicator';
        typingDiv.innerHTML = `
            <div class="typing-avatar">A</div>
            <div class="typing-content">
                <div class="typing-dot"></div>
                <div class="typing-dot"></div>
                <div class="typing-dot"></div>
            </div>
        `;
        
        messagesContainer.appendChild(typingDiv);
        this.scrollToBottom();
    }

    hideTypingIndicator() {
        const indicator = document.querySelector('.typing-indicator');
        if (indicator) {
            indicator.remove();
        }
    }

    updateStreamingMessage(chunkData) {
        const agentName = (chunkData.data && chunkData.data.agent_name) || chunkData.agent_name || 'Agent';
        const content = (chunkData.data && chunkData.data.content) || chunkData.content || '';
        const chunkIndex = (chunkData.data && chunkData.data.chunk_index) || chunkData.chunk_index || 0;

        let stream = this.activeStreams.get(agentName);
        if (!stream) {
            const messagesContainer = document.getElementById('chat-messages');
            const messageDiv = document.createElement('div');
            messageDiv.className = 'message agent streaming';
            const timestamp = new Date().toISOString();
            messageDiv.innerHTML = `
                <div class="message-header">
                    <div class="message-avatar">${agentName[0].toUpperCase()}</div>
                    <span class="message-sender">${agentName}</span>
                    <span class="message-time">${new Date(timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}</span>
                </div>
                <div class="message-content"></div>
            `;
            messagesContainer.appendChild(messageDiv);
            this.scrollToBottom();
            stream = { el: messageDiv.querySelector('.message-content'), content: '', lastIndex: -1 };
            this.activeStreams.set(agentName, stream);
        }
        if (chunkIndex <= stream.lastIndex) {
            // out-of-order or duplicate
        }
        stream.content += content;
        stream.lastIndex = chunkIndex;
        stream.el.textContent = stream.content;
    }

    finalizeStreamingMessage(data) {
        const agentName = (data.data && data.data.agent_name) || data.agent_name || 'Agent';
        const stream = this.activeStreams.get(agentName);
        if (stream) {
            const parent = stream.el.closest('.message');
            if (parent) parent.classList.remove('streaming');
            this.activeStreams.delete(agentName);
        }
    }

    scrollToBottom() {
        const messagesContainer = document.getElementById('chat-messages');
        messagesContainer.scrollTop = messagesContainer.scrollHeight;
    }

    createNewSession(agentName = null) {
        const sessionId = 'session_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9);
        const agent = agentName || this.currentAgent;
        
        const session = {
            id: sessionId,
            agent: agent,
            title: agent ? `Chat with ${agent}` : 'New Chat',
            messages: [],
            createdAt: new Date().toISOString(),
            lastMessage: null
        };

        this.sessions.set(sessionId, session);
        this.currentSessionId = sessionId;
        
        if (agent && !this.currentAgent) {
            this.currentAgent = agent;
        }
        
        // Clear chat messages and show welcome
        this.clearChatMessages();
        this.updateSessionsList();
        this.updateActiveSession(sessionId);
        this.saveSessions();
        
        // If creating session with specific agent, select that agent
        if (agent) {
            this.selectAgent(agent);
        }
    }

    switchToSession(sessionId) {
        if (!this.sessions.has(sessionId)) {
            return;
        }

        this.currentSessionId = sessionId;
        this.loadSessionMessages(sessionId);
        this.updateActiveSession(sessionId);
    }

    loadSessionMessages(sessionId) {
        const session = this.sessions.get(sessionId);
        if (!session) return;

        const messagesContainer = document.getElementById('chat-messages');
        messagesContainer.innerHTML = '';

        session.messages.forEach(message => {
            const messageElement = this.createMessageElement(message);
            messagesContainer.appendChild(messageElement);
        });

        this.scrollToBottom();
    }

    updateActiveSession(sessionId) {
        document.querySelectorAll('.session-item').forEach(item => {
            item.classList.remove('active');
        });

        const sessionElement = document.querySelector(`[data-session-id="${sessionId}"]`);
        if (sessionElement) {
            sessionElement.classList.add('active');
        }
    }

    updateSessionsList() {
        const sessionsList = document.getElementById('sessions-list');
        sessionsList.innerHTML = '';

        const sortedSessions = Array.from(this.sessions.values())
            .sort((a, b) => new Date(b.last_message_at) - new Date(a.last_message_at));

        sortedSessions.forEach(session => {
            const sessionElement = this.createSessionElement(session);
            sessionsList.appendChild(sessionElement);
        });
    }

    createSessionElement(session) {
        const sessionDiv = document.createElement('div');
        sessionDiv.className = 'session-item';
        sessionDiv.dataset.sessionId = session.id;
        
        const lastMessage = session.messages.length > 0 ? 
            session.messages[session.messages.length - 1].content.substring(0, 50) + '...' : 
            'No messages yet';
        
        const timestamp = new Date(session.last_message_at).toLocaleString();
        
        sessionDiv.innerHTML = `
            <div class="session-title">${session.title}</div>
            <div class="session-preview">${lastMessage}</div>
            <div class="session-time">${timestamp}</div>
        `;

        sessionDiv.addEventListener('click', () => {
            this.switchToSession(session.id);
        });

        return sessionDiv;
    }

    updateSessionLastMessage(sessionId, message) {
        const session = this.sessions.get(sessionId);
        if (session) {
            session.last_message_at = new Date().toISOString();
            // Update title based on first message
            if (session.messages.length <= 1) {
                session.title = message.substring(0, 30) + (message.length > 30 ? '...' : '');
            }
            this.updateSessionsList();
        }
        
        this.saveSessions();
    }

    getCurrentSession() {
        return this.currentSessionId ? this.sessions.get(this.currentSessionId) : null;
    }

    async loadAgents() {
        try {
            const response = await fetch('/api/agents');
            const data = await response.json();
            
            if (data.status === 'success' && data.data && data.data.available) {
                this.agents = data.data.available;
                this.updateAgentsList();
            } else {
                console.error('Failed to load agents:', data);
                this.showAgentsError();
            }
        } catch (error) {
            console.error('Error loading agents:', error);
            this.showAgentsError();
        }
    }

    updateAgentsList() {
        const agentDropdown = document.getElementById('agent-dropdown');
        const agentInfo = document.getElementById('agent-info');
        
        if (!this.agents || this.agents.length === 0) {
            agentDropdown.innerHTML = '<option value="">No agents available</option>';
            agentInfo.style.display = 'none';
            return;
        }

        // Populate dropdown
        agentDropdown.innerHTML = '<option value="">Choose an agent...</option>' + 
            this.agents.map(agent => 
                `<option value="${agent.name}">🤖 ${agent.name} - ${agent.role}</option>`
            ).join('');
        
        // Add event listener for dropdown change
        agentDropdown.onchange = (e) => {
            const selectedAgentName = e.target.value;
            if (selectedAgentName) {
                this.selectAgent(selectedAgentName);
                this.showAgentInfo(selectedAgentName);
            } else {
                agentInfo.style.display = 'none';
                this.currentAgent = null;
                document.getElementById('chat-input').placeholder = 'Select an agent to start chatting...';
            }
        };
    }

    showAgentInfo(agentName) {
        const agentInfo = document.getElementById('agent-info');
        const agent = this.agents.find(a => a.name === agentName);
        
        if (agent) {
            agentInfo.querySelector('.agent-description').textContent = agent.description;
            agentInfo.querySelector('.agent-capabilities').innerHTML = 
                agent.capabilities.map(cap => `<span class="capability-tag">${cap}</span>`).join('');
            agentInfo.style.display = 'block';
        }
    }

    selectAgent(agentName) {
        this.currentAgent = agentName;
        
        // Update chat input placeholder
        const chatInput = document.getElementById('chat-input');
        const agent = this.agents.find(a => a.name === agentName);
        if (agent) {
            chatInput.placeholder = `Ask ${agent.name} anything... (${agent.role})`;
        }
        
        // Create new session if none exists or if current session has different agent
        if (!this.currentSessionId || this.getCurrentSession()?.agent !== agentName) {
            this.createNewSession(agentName);
        }
    }

    showAgentsError() {
        const agentDropdown = document.getElementById('agent-dropdown');
        const agentInfo = document.getElementById('agent-info');
        agentDropdown.innerHTML = '<option value="">Failed to load agents</option>';
        agentInfo.style.display = 'none';
    }

    loadSessions() {
        // Load sessions from localStorage for persistence
        const saved = localStorage.getItem('agenticgokit_sessions');
        if (saved) {
            try {
                const sessionsData = JSON.parse(saved);
                sessionsData.forEach(session => {
                    this.sessions.set(session.id, session);
                });
                this.updateSessionsList();
            } catch (error) {
                console.error('Failed to load saved sessions:', error);
            }
        }
    }

    saveSessions() {
        const sessionsData = Array.from(this.sessions.values());
        localStorage.setItem('agenticgokit_sessions', JSON.stringify(sessionsData));
    }

    toggleSettings() {
        const panel = document.getElementById('settings-panel');
        panel.classList.toggle('open');
    }

    closeSettings() {
        document.getElementById('settings-panel').classList.remove('open');
    }

    loadSettings() {
        const saved = localStorage.getItem('agenticgokit_settings');
        if (saved) {
            try {
                const settings = JSON.parse(saved);
                this.settings = { ...this.settings, ...settings };
                this.populateSettingsForm();
            } catch (error) {
                console.error('Failed to load settings:', error);
            }
        }
    }

    saveSettings() {
        const form = document.getElementById('settings-form');
        const formData = new FormData(form);
        
        this.settings.autoReconnect = formData.get('auto-reconnect') === 'on';
        this.settings.reconnectDelay = parseInt(formData.get('reconnect-delay')) || 3000;
        this.settings.maxReconnectAttempts = parseInt(formData.get('max-reconnect-attempts')) || 5;

        localStorage.setItem('agenticgokit_settings', JSON.stringify(this.settings));
        this.closeSettings();
        this.showSuccess('Settings saved successfully');
    }

    async loadAgentflowToml() {
        try {
            const res = await fetch('/api/config/raw');
            const data = await res.json();
            if (data.status === 'success' && data.data) {
                const editor = document.getElementById('agentflow-editor');
                if (editor) editor.value = data.data.content || '';
                const pathEl = document.getElementById('config-path');
                if (pathEl) pathEl.textContent = data.data.path || '';
                this.showSuccess('Config loaded');
            } else {
                this.showError('Failed to load config');
            }
        } catch (e) {
            console.error(e);
            this.showError('Error loading config');
        }
    }

    async saveAgentflowToml() {
        const btn = document.getElementById('save-config-btn');
        try {
            const editor = document.getElementById('agentflow-editor');
            const toml = editor ? (editor.value || '') : '';
            if (!toml.trim()) { this.showError('Config is empty. Nothing to save.'); return; }
            if (btn) btn.disabled = true;
            const res = await fetch('/api/config/raw', {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ toml })
            });
            // Try to parse JSON; if not JSON, fall back to text
            let bodyText = '';
            let data = null;
            try { data = await res.json(); } catch {
                try { bodyText = await res.text(); } catch {}
            }
            if (res.ok && data && data.status === 'success') {
                this.showSuccess('Config saved');
            } else {
                const msg = (data && data.message) ? data.message : (bodyText || `Failed to save config (HTTP ${res.status})`);
                this.showError(msg);
            }
        } catch (e) {
            console.error(e);
            this.showError('Error saving config');
        } finally {
            if (btn) btn.disabled = false;
        }
    }

    async refreshDiagram() {
        try {
            const res = await fetch('/api/visualization/composition');
            const data = await res.json();
            const pre = document.getElementById('flow-diagram');
            if (data.status === 'success' && data.data && pre) {
                pre.textContent = data.data.diagram;
            } else if (pre) {
                pre.textContent = '// Failed to load diagram';
            }
        } catch (e) {
            const pre = document.getElementById('flow-diagram');
            if (pre) pre.textContent = '// Error fetching diagram';
        }
    }

    populateSettingsForm() {
        document.getElementById('auto-reconnect').checked = this.settings.autoReconnect;
        document.getElementById('reconnect-delay').value = this.settings.reconnectDelay;
        document.getElementById('max-reconnect-attempts').value = this.settings.maxReconnectAttempts;
    }

    toggleMobileMenu() {
        const sidebar = document.querySelector('.chat-sidebar');
        sidebar.classList.toggle('mobile-open');
    }

    handleResize() {
        // Handle responsive behavior on window resize
        if (window.innerWidth > 768) {
            document.querySelector('.chat-sidebar').classList.remove('mobile-open');
        }
    }

    showError(message) {
        this.showNotification(message, 'error');
    }

    showSuccess(message) {
        this.showNotification(message, 'success');
    }

    showNotification(message, type = 'error') {
        const existing = document.querySelector('.notification');
        if (existing) {
            existing.remove();
        }

        const notification = document.createElement('div');
        notification.className = `notification ${type}`;
        notification.textContent = message;
        
        document.body.appendChild(notification);
        
        setTimeout(() => {
            notification.remove();
        }, 5000);
    }

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    // Cleanup method
    destroy() {
        if (this.websocket) {
            this.websocket.close();
        }
        this.saveSessions();
    }
}

// Notification styles
const notificationStyles = `
.notification {
    position: fixed;
    top: 20px;
    right: 20px;
    padding: 12px 20px;
    border-radius: 8px;
    color: white;
    font-size: 14px;
    font-weight: 500;
    z-index: 1000;
    animation: slideInRight 0.3s ease;
    max-width: 300px;
    word-wrap: break-word;
}

.notification.error {
    background-color: #d93025;
}

.notification.success {
    background-color: #34a853;
}

@keyframes slideInRight {
    from {
        transform: translateX(100%);
        opacity: 0;
    }
    to {
        transform: translateX(0);
        opacity: 1;
    }
}
`;

// Add notification styles to the page
const style = document.createElement('style');
style.textContent = notificationStyles;
document.head.appendChild(style);

// Initialize the app and make it globally available
let app;
document.addEventListener('DOMContentLoaded', function() {
    app = new AgenticChatApp();
    // Make app globally available for onclick handlers
    window.app = app;
    window.chatApp = app; // For compatibility
    
    // Cleanup on page unload
    window.addEventListener('beforeunload', () => {
        if (window.chatApp) {
            window.chatApp.destroy();
        }
    });
});

// Export for module systems
if (typeof module !== 'undefined' && module.exports) {
    module.exports = AgenticChatApp;
}
