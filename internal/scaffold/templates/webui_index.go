package templates

const WebUIIndexTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Config.Name}} - AI Agent Chat</title>
    <link rel="stylesheet" href="style.css">
</head>
<body>
    <div class="app-container">
        <!-- Header -->
        <header class="app-header">
            <div class="header-content">
                <h1 class="app-title">{{.Config.Name}}</h1>
                <div class="agent-selector-container">
                    <label for="agent-select" class="agent-label">Agent:</label>
                    <select id="agent-select" class="agent-select">
                        <option value="assistant">Assistant</option>
                        <option value="coder">Coder</option>
                        <option value="writer">Writer</option>
                        <option value="analyst">Analyst</option>
                    </select>
                </div>
            </div>
        </header>

        <!-- Main Content -->
        <main class="main-content">
            <!-- Agent Info Panel -->
            <div class="agent-info-panel" id="agent-info">
                <div class="agent-details">
                    <h3 id="agent-name">Assistant</h3>
                    <p id="agent-description">General purpose assistant</p>
                </div>
            </div>

            <!-- Chat Area -->
            <div class="chat-container">
                <div class="chat-messages" id="chat-messages">
                    <div class="welcome-message">
                        <h3>Welcome to {{.Config.Name}}!</h3>
                        <p>Choose an agent and start chatting to get assistance with your tasks.</p>
                    </div>
                </div>
            </div>
        </main>

        <!-- Chat Input -->
        <footer class="chat-input-container">
            <form id="chat-form" class="chat-form">
                <input 
                    type="text" 
                    id="message-input" 
                    class="message-input" 
                    placeholder="Type your message here..." 
                    autocomplete="off"
                >
                <button type="submit" class="send-button">
                    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <line x1="22" y1="2" x2="11" y2="13"></line>
                        <polygon points="22,2 15,22 11,13 2,9"></polygon>
                    </svg>
                </button>
            </form>
        </footer>
    </div>

    <script src="app.js"></script>
</body>
</html>`
