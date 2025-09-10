package templates

const WebUIJSTemplate = `// Agent configurations
const agents = {
    assistant: {
        name: 'Assistant',
        description: 'General purpose assistant for various tasks'
    },
    coder: {
        name: 'Coder', 
        description: 'Programming and code analysis specialist'
    },
    writer: {
        name: 'Writer',
        description: 'Content creation and writing expert'
    },
    analyst: {
        name: 'Analyst',
        description: 'Data analysis and insights specialist'
    }
};

// DOM elements
let chatMessages, messageInput, chatForm, agentSelect, agentName, agentDescription;

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

    // Focus input
    messageInput.focus();

    // Update initial agent info
    updateAgentInfo();
});

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
        // Send message to API
        const response = await fetch('/api/chat', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                message: message,
                agent: selectedAgent
            })
        });

        if (!response.ok) {
            throw new Error(` + "`HTTP error! status: ${response.status}`" + `);
        }

        const data = await response.json();
        
        // Remove typing indicator
        removeTypingIndicator(typingId);
        
        // Add agent response
        addMessage('agent', data.response, agents[selectedAgent].name);

    } catch (error) {
        console.error('Error:', error);
        
        // Remove typing indicator
        removeTypingIndicator(typingId);
        
        // Add error message
        addMessage('agent', 'Sorry, I encountered an error. Please try again.', 'System');
    } finally {
        // Re-enable form
        setFormDisabled(false);
        messageInput.focus();
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
    }
}

// Add message to chat
function addMessage(sender, content, senderName) {
    const messageDiv = document.createElement('div');
    messageDiv.className = ` + "`message ${sender}`" + `;
    
    const contentDiv = document.createElement('div');
    contentDiv.className = 'message-content';
    contentDiv.textContent = content;
    
    const infoDiv = document.createElement('div');
    infoDiv.className = 'message-info';
    infoDiv.textContent = ` + "`${senderName} • ${new Date().toLocaleTimeString()}`" + `;
    
    messageDiv.appendChild(contentDiv);
    messageDiv.appendChild(infoDiv);
    
    // Remove welcome message if it exists
    const welcomeMessage = chatMessages.querySelector('.welcome-message');
    if (welcomeMessage) {
        welcomeMessage.remove();
    }
    
    chatMessages.appendChild(messageDiv);
    scrollToBottom();
}

// Add typing indicator
function addTypingIndicator(agentName) {
    const typingDiv = document.createElement('div');
    const typingId = 'typing-' + Date.now();
    typingDiv.id = typingId;
    typingDiv.className = 'message agent';
    
    typingDiv.innerHTML = ` + "`" + `
        <div class="message-content typing-indicator">
            <span>${agents[agentName].name} is typing</span>
            <div class="typing-dots">
                <div class="typing-dot"></div>
                <div class="typing-dot"></div>
                <div class="typing-dot"></div>
            </div>
        </div>
    ` + "`" + `;
    
    chatMessages.appendChild(typingDiv);
    scrollToBottom();
    
    return typingId;
}

// Remove typing indicator
function removeTypingIndicator(typingId) {
    const typingDiv = document.getElementById(typingId);
    if (typingDiv) {
        typingDiv.remove();
    }
}

// Enable/disable form
function setFormDisabled(disabled) {
    messageInput.disabled = disabled;
    const submitButton = chatForm.querySelector('button[type="submit"]');
    submitButton.disabled = disabled;
}

// Scroll chat to bottom
function scrollToBottom() {
    chatMessages.scrollTop = chatMessages.scrollHeight;
}

// Auto-resize functionality for potential future use
function autoResize(element) {
    element.style.height = 'auto';
    element.style.height = element.scrollHeight + 'px';
}`
