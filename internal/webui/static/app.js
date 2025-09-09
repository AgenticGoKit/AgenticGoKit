// AgenticGoKit Chat Interface JavaScript

class AgenticChatApp {
    constructor() {
        this.websocket = null;
        this.currentSessionId = null;
        this.isConnected = false;
        this.messageQueue = [];
        this.sessions = new Map();
        this.settings = {
            serverUrl: window.location.origin.replace('http', 'ws'),
            autoReconnect: true,
            reconnectDelay: 3000,
            maxReconnectAttempts: 5,
            typingIndicatorDelay: 1000,
        };
        this.reconnectAttempts = 0;
        this.isTyping = false;
        this.typingTimeout = null;
        
        this.init();
    }

    init() {
        this.setupEventListeners();
        this.loadSettings();
        this.connect();
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
        if (this.websocket && this.websocket.readyState === WebSocket.OPEN) {
            return;
        }

        const wsUrl = `${this.settings.serverUrl}/ws`;
        console.log('Connecting to WebSocket:', wsUrl);

        try {
            this.websocket = new WebSocket(wsUrl);
            this.setupWebSocketHandlers();
        } catch (error) {
            console.error('Failed to create WebSocket connection:', error);
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
            
            // Create initial session if none exists
            if (!this.currentSessionId) {
                this.createNewSession();
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
            case 'session_created':
                this.handleSessionCreated(message);
                break;
            case 'agent_response':
                this.handleAgentResponse(message);
                break;
            case 'agent_response_chunk':
                this.handleAgentResponseChunk(message);
                break;
            case 'agent_response_complete':
                this.handleAgentResponseComplete(message);
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

    handleSessionCreated(message) {
        this.currentSessionId = message.session_id;
        this.sessions.set(message.session_id, {
            id: message.session_id,
            messages: [],
            title: 'New Chat',
            created_at: new Date().toISOString(),
            last_message_at: new Date().toISOString()
        });
        this.updateSessionsList();
        this.switchToSession(message.session_id);
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
        
        if (connected) {
            statusIndicator.classList.remove('disconnected');
            statusText.textContent = 'Connected';
            this.hideConnectionLostBanner();
        } else {
            statusIndicator.classList.add('disconnected');
            statusText.textContent = 'Disconnected';
            this.showConnectionLostBanner();
        }
    }

    showConnectionLostBanner() {
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
        const input = document.getElementById('chat-input');
        const message = input.value.trim();
        
        if (!message || !this.currentSessionId) {
            return;
        }

        // Add user message to UI immediately
        this.addMessage({
            type: 'user',
            content: message,
            timestamp: new Date().toISOString()
        });

        // Update session
        this.updateSessionLastMessage(this.currentSessionId, message);

        // Send to server
        const chatMessage = {
            type: 'chat_message',
            session_id: this.currentSessionId,
            message: message,
            timestamp: new Date().toISOString()
        };

        if (this.isConnected) {
            this.websocket.send(JSON.stringify(chatMessage));
        } else {
            this.messageQueue.push(chatMessage);
            this.showError('Message queued. Will send when connected.');
        }

        // Clear input and show typing indicator
        input.value = '';
        input.style.height = 'auto';
        this.showTypingIndicator();
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

    createMessageElement(messageData) {
        const messageDiv = document.createElement('div');
        messageDiv.className = `message ${messageData.type}`;

        const avatarText = messageData.type === 'user' ? 'U' : (messageData.agent_name ? messageData.agent_name[0].toUpperCase() : 'A');
        const senderName = messageData.type === 'user' ? 'You' : (messageData.agent_name || 'Agent');
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
        // Implementation for handling streaming responses
        // This would update a message in real-time as chunks arrive
        console.log('Streaming chunk:', chunkData);
    }

    finalizeStreamingMessage(data) {
        // Implementation for finalizing streaming responses
        console.log('Streaming complete:', data);
    }

    scrollToBottom() {
        const messagesContainer = document.getElementById('chat-messages');
        messagesContainer.scrollTop = messagesContainer.scrollHeight;
    }

    createNewSession() {
        if (!this.isConnected) {
            this.showError('Cannot create session while disconnected');
            return;
        }

        const createMessage = {
            type: 'session_create',
            timestamp: new Date().toISOString()
        };

        this.websocket.send(JSON.stringify(createMessage));
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

// Initialize the app when the page loads
document.addEventListener('DOMContentLoaded', () => {
    window.chatApp = new AgenticChatApp();
    
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
