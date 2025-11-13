# Story Writer Chat - Frontend Setup

## Quick Start

```bash
# Navigate to frontend directory
cd frontend

# Install dependencies
npm install

# Start development server
npm run dev
```

The frontend will be available at http://localhost:3000

## What Was Created

### Modern React + TypeScript Stack
- ⚛️ React 18 with TypeScript
- ⚡ Vite 5 (compatible with Node 18+)
- 🎨 TailwindCSS for styling
- 🐻 Zustand for state management
- 🔌 WebSocket client with auto-reconnect
- 🎭 Lucide React icons

### Reusable Components

All components are designed to work with any AgenticGoKit backend:

1. **`<ChatInterface />`** - Complete chat UI with messages and input
2. **`<MessageList />`** - Scrollable message container
3. **`<MessageBubble />`** - Individual message with agent avatars
4. **`<InputArea />`** - Message input with send button
5. **`<WorkflowVisualizer />`** - Agent status cards and progress
6. **`<AgentCard />`** - Individual agent status display

### State Management (`useChatStore`)

Zustand store manages:
- WebSocket connection and status
- Chat messages and streaming
- Workflow state and progress
- Agent statuses

### WebSocket Client

Robust WebSocket client with:
- Automatic reconnection
- Message type routing
- Status callbacks
- Error handling

### Type Safety

Full TypeScript types for:
- Messages and agents
- Workflow states
- WebSocket events
- Component props

## File Structure

```
frontend/
├── src/
│   ├── components/          # React UI components
│   │   ├── ChatInterface.tsx
│   │   ├── MessageList.tsx
│   │   ├── MessageBubble.tsx
│   │   ├── InputArea.tsx
│   │   ├── WorkflowVisualizer.tsx
│   │   └── AgentCard.tsx
│   ├── lib/                # Utilities
│   │   ├── utils.ts
│   │   └── websocket-client.ts
│   ├── store/              # Zustand store
│   │   └── chat-store.ts
│   ├── types/              # TypeScript types
│   │   └── index.ts
│   ├── App.tsx             # Main app
│   ├── main.tsx            # Entry point
│   └── index.css           # Global styles
├── package.json
├── tsconfig.json
├── vite.config.ts
└── tailwind.config.js
```

## Using in Other Projects

### 1. Copy the components
```bash
cp -r src/components /path/to/your/project/src/
cp -r src/lib /path/to/your/project/src/
cp -r src/store /path/to/your/project/src/
cp -r src/types /path/to/your/project/src/
```

### 2. Install dependencies
```bash
npm install zustand clsx lucide-react
```

### 3. Use in your app
```tsx
import { useChatStore } from './store/chat-store'
import { ChatInterface } from './components/ChatInterface'
import { WorkflowVisualizer } from './components/WorkflowVisualizer'

function App() {
  const connect = useChatStore(state => state.connect)
  
  useEffect(() => {
    connect('ws://your-backend:8080/ws')
  }, [])

  return (
    <div>
      <ChatInterface />
      <WorkflowVisualizer />
    </div>
  )
}
```

### 4. Customize agents
Edit `src/store/chat-store.ts`:
```typescript
const AGENTS: Agent[] = [
  {
    name: 'your_agent',
    displayName: 'Your Agent',
    icon: '🤖',
    color: 'blue',
    description: 'Does something amazing',
  },
]
```

## Next Steps

1. **Install dependencies**: `npm install`
2. **Start Go backend**: Ensure backend runs on port 8080
3. **Start frontend**: `npm run dev`
4. **Test the app**: Open http://localhost:3000

## Backend Integration

The frontend expects these WebSocket message types:

```typescript
// From backend to frontend
{
  type: 'workflow_start' | 'agent_start' | 'agent_progress' | 
        'agent_complete' | 'workflow_done' | 'error',
  content?: string,
  agent?: string,
  progress?: number,
  timestamp: number
}

// From frontend to backend
{
  type: 'user_message',
  content: string,
  timestamp: number
}
```

Your Go backend already supports these message types! ✅

## Troubleshooting

### Node version error
If you see "Vite requires Node.js version 20.19+", you have two options:
1. **Recommended**: Upgrade Node.js to version 20+
2. **Alternative**: The package.json uses Vite 5 which supports Node 18+

### Port already in use
Change the port in `vite.config.ts`:
```typescript
server: {
  port: 3001, // Change here
}
```

### WebSocket connection fails
- Ensure Go backend is running on port 8080
- Check the WebSocket URL in `src/App.tsx`
- Look for CORS issues in browser console

## Features

✅ Real-time message streaming
✅ Multi-agent workflow visualization
✅ Progress tracking
✅ Auto-reconnecting WebSocket
✅ Dark mode support (from system)
✅ Responsive design
✅ TypeScript type safety
✅ Smooth animations
✅ Accessible UI components

## Customization

### Colors
Edit Tailwind colors in `tailwind.config.js`

### Agents
Edit agent config in `src/store/chat-store.ts`

### Layout
Modify `src/App.tsx` grid layout

### Styles
Global styles in `src/index.css`

Enjoy your modern multi-agent chat UI! 🚀
