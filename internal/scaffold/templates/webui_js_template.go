package templates

const WebUIJSTemplate = `// Agent configurations - loaded dynamically from API
let agents = {};

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

    // Load agents from API
    loadAgents();

    // Focus input
    messageInput.focus();
});

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
        // Send message to API
        console.log('Sending request:', {
            message: message,
            agent: selectedAgent,
            useOrchestration: document.getElementById('orchestration-mode').checked
        });

        const response = await fetch('/api/chat', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                message: message,
                agent: selectedAgent,
                useOrchestration: document.getElementById('orchestration-mode').checked
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
    
    const agentDisplayName = agents[agentName] ? agents[agentName].name : agentName;
    
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
    const submitButton = chatForm.querySelector('button[type=\"submit\"]');
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
