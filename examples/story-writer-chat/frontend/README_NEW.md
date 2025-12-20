# 🚀 Story Writer Chat - Modern React + TypeScript Frontend

## ✨ What's New

We've upgraded the Story Writer Chat with a modern, reusable React + TypeScript + TailwindCSS stack!

### Features

- ⚛️ **React 18** - Modern component-based UI
- 📘 **TypeScript** - Full type safety
- ⚡ **Vite 5** - Lightning-fast builds (Node 18 compatible)
- 🎨 **TailwindCSS** - Beautiful utility-first styling
- 🐻 **Zustand** - Lightweight state management
- 🔌 **WebSocket** - Real-time agent communication with auto-reconnect
- 🎭 **Lucide Icons** - Beautiful icon library
- 📱 **Responsive** - Works on desktop, tablet, and mobile
- 🌓 **Dark Mode** - Automatic system-based theme

## 🏃 Quick Start

```bash
# 1. Install dependencies
cd frontend
npm install

# 2. Start the Go backend (in another terminal)
cd ..
go run main.go

# 3. Start the frontend
npm run dev
```

Open http://localhost:3000 in your browser! 🎉

## 📦 What You Get

### Reusable Components

All components work with any AgenticGoKit backend:

| Component | Purpose |
|-----------|---------|
| `<ChatInterface />` | Complete chat UI |
| `<MessageList />` | Scrollable message container |
| `<MessageBubble />` | Individual messages with agent avatars |
| `<InputArea />` | Message input with send button |
| `<WorkflowVisualizer />` | Agent status and progress |
| `<AgentCard />` | Individual agent status display |

### Smart State Management

- **`useChatStore`** - Zustand store managing:
  - WebSocket connection
  - Chat messages & streaming
  - Workflow state & progress
  - Agent statuses

### Robust WebSocket Client

- Automatic reconnection
- Message type routing
- Status callbacks
- Error handling

## 🎨 Screenshots

### Chat Interface
- User messages on the right (blue)
- Agent messages on the left with icons
- Real-time streaming text
- Smooth animations

### Workflow Visualizer
- Overall progress bar
- Individual agent cards
- Status indicators (idle, working, complete)
- Active agent highlighting

## 🔧 Customization

### Adding New Agents

Edit `src/store/chat-store.ts`:

```typescript
const AGENTS: Agent[] = [
  {
    name: 'writer',
    displayName: 'Writer',
    icon: '✍️',
    color: 'blue',
    description: 'Creates initial story draft',
  },
  {
    name: 'your_agent',
    displayName: 'Your Agent',
    icon: '🤖',
    color: 'red', // blue, green, purple, red
    description: 'Does something amazing',
  },
]
```

### Changing Colors

Edit `tailwind.config.js` for theme colors or use the CSS variables in `src/index.css`.

### Modifying Layout

The main layout is in `src/App.tsx`. It uses CSS Grid for responsive design.

## 📁 Project Structure

```
frontend/
├── src/
│   ├── components/
│   │   ├── ChatInterface.tsx       # Main chat container
│   │   ├── MessageList.tsx         # Message scroll area
│   │   ├── MessageBubble.tsx       # Individual message
│   │   ├── InputArea.tsx           # Input + send button
│   │   ├── WorkflowVisualizer.tsx  # Workflow status panel
│   │   └── AgentCard.tsx           # Agent status card
│   ├── lib/
│   │   ├── utils.ts                # Helper functions
│   │   └── websocket-client.ts     # WebSocket manager
│   ├── store/
│   │   └── chat-store.ts           # Zustand state
│   ├── types/
│   │   └── index.ts                # TypeScript types
│   ├── App.tsx                     # Main app component
│   ├── main.tsx                    # Entry point
│   └── index.css                   # Global styles + Tailwind
├── package.json
├── tsconfig.json
├── vite.config.ts
├── tailwind.config.js
└── postcss.config.js
```

## 🔌 Backend Integration

The frontend expects these WebSocket messages:

### From Backend → Frontend

```typescript
{
  type: 'workflow_start',  // Workflow begins
  content: string,
  timestamp: number
}

{
  type: 'agent_start',     // Agent begins working
  agent: string,           // Agent name
  progress: number,        // 0-100
  timestamp: number
}

{
  type: 'agent_progress',  // Streaming content
  agent: string,
  content: string,         // Text chunk
  timestamp: number
}

{
  type: 'agent_complete',  // Agent finished
  agent: string,
  content: string,         // Final output
  timestamp: number
}

{
  type: 'workflow_done',   // All done!
  content: string,
  timestamp: number
}

{
  type: 'error',           // Something went wrong
  content: string,         // Error message
  timestamp: number
}
```

### From Frontend → Backend

```typescript
{
  type: 'user_message',
  content: string,
  timestamp: number
}
```

Your existing Go backend already implements this! ✅

## 🚀 Using in Other Projects

### Option 1: Copy Components

```bash
# Copy to your project
cp -r frontend/src/components /your/project/src/
cp -r frontend/src/lib /your/project/src/
cp -r frontend/src/store /your/project/src/
cp -r frontend/src/types /your/project/src/

# Install dependencies
npm install zustand clsx lucide-react
```

### Option 2: npm Package (Future)

We plan to publish this as `@agenticgokit/react-chat-ui` for easy reuse!

## 🛠️ Development

### Commands

```bash
npm run dev      # Start dev server
npm run build    # Build for production
npm run preview  # Preview production build
npm run lint     # Run ESLint
```

### Environment

- **Dev**: http://localhost:3000 (proxies to :8080 backend)
- **Production**: Build and serve `dist/` folder

## 🐛 Troubleshooting

### "Cannot find module 'zustand'"

Run `npm install` in the `frontend/` directory.

### WebSocket connection fails

1. Ensure Go backend is running on port 8080
2. Check `src/App.tsx` for correct WebSocket URL
3. Look for CORS errors in browser console

### Node version error

The package uses Vite 5 which supports Node 18+. If you see errors, upgrade Node to v18 or higher.

### Port 3000 already in use

Change the port in `vite.config.ts`:
```typescript
server: {
  port: 3001,
}
```

## 🎯 Next Steps

1. ✅ **Install dependencies**: `npm install`
2. ✅ **Start backend**: `go run main.go`
3. ✅ **Start frontend**: `npm run dev`
4. 🎨 **Customize**: Add your agents, change colors, modify layout
5. 📦 **Build**: `npm run build` when ready for production

## 💡 Tips

- Messages auto-scroll to bottom
- WebSocket reconnects automatically
- Progress bars show real-time updates
- Dark mode follows system preferences
- All components are TypeScript-safe
- Tailwind makes styling super easy

## 📚 Learn More

- [React Documentation](https://react.dev)
- [TypeScript Handbook](https://www.typescriptlang.org/docs/)
- [Vite Guide](https://vitejs.dev/guide/)
- [TailwindCSS](https://tailwindcss.com)
- [Zustand](https://github.com/pmndrs/zustand)

## 🤝 Contributing

Improvements welcome! This UI is designed to be reusable across all AgenticGoKit examples.

---

**Enjoy your modern multi-agent chat experience!** 🎉✨

Built with ❤️ using AgenticGoKit v.next
