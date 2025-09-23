# AgenticGoKit Architecture Documentation Index

This index provides a comprehensive overview of all architecture documents created for AgenticGoKit's website (www.agenticgokit.com).

## Core Architecture Documents

1. **[SYSTEM_ARCHITECTURE.md](SYSTEM_ARCHITECTURE.md)**
   - Comprehensive system architecture overview
   - Core components and interfaces
   - Design patterns and data flows
   - Deployment architectures and comparisons

2. **[WORKFLOW_ENGINE.md](WORKFLOW_ENGINE.md)**
   - Workflow engine architecture
   - Orchestration patterns (Sequential, Collaborative, Loop, etc.)
   - Event processing and error handling
   - Advanced workflow features

3. **[MEMORY_SYSTEM.md](MEMORY_SYSTEM.md)**
   - Memory system architecture
   - Short-term and long-term memory components
   - Vector database integration
   - RAG implementation patterns

4. **[MCP_INTEGRATION.md](MCP_INTEGRATION.md)**
   - Model Context Protocol integration
   - Tool discovery and execution
   - Plugin registration system
   - Advanced MCP features

5. **[WEBSITE_DIAGRAMS.md](WEBSITE_DIAGRAMS.md)**
   - Mermaid diagrams for website integration
   - Visual representations of all key components
   - Interactive diagrams for system visualization
   - Deployment model diagrams

## Diagram Types Available

### System Architecture Diagrams

- Overall system architecture (layers and components)
- Plugin system and registry pattern
- Factory pattern for component creation
- Component integration diagrams
- Framework comparison

### Workflow Engine Diagrams

- Runner and orchestrator relationships
- Agent interaction patterns
- Event processing flow
- Orchestration patterns:
  - Route pattern
  - Collaborative pattern
  - Sequential pattern
  - Loop pattern
  - Mixed pattern

### Memory System Diagrams

- Memory interface and providers
- Vector database integration
- Short-term vs. long-term memory
- Memory data flow
- Embedding model integration

### MCP Integration Diagrams

- Tool discovery and routing
- Plugin registration system
- Tool execution flow
- Tool categories and capabilities

### Deployment Model Diagrams

- Single-process deployment
- Microservices architecture
- Serverless/event-driven architecture

## Using These Diagrams

### Website Integration

These diagrams can be directly integrated into the AgenticGoKit website using the Mermaid.js library. Add the Mermaid script to your website:

```html
<script src="https://cdn.jsdelivr.net/npm/mermaid/dist/mermaid.min.js"></script>
<script>mermaid.initialize({startOnLoad:true});</script>
```

Then include the diagrams as code blocks with the `mermaid` class:

```html
<div class="mermaid">
  flowchart TD
    A[Component A] --> B[Component B]
    B --> C[Component C]
</div>
```

### Static Image Export

To export diagrams as static images (PNG/SVG), you can use:

1. The Mermaid CLI tool:
   ```
   mmdc -i input.mmd -o output.svg
   ```

2. The Mermaid Live Editor (https://mermaid.live/) to export images manually

3. For batch processing, consider using a script to extract and convert all diagrams:
   ```
   # Example Python script to extract and render diagrams
   import re
   import subprocess
   
   def extract_diagrams(file_path):
       with open(file_path, 'r') as f:
           content = f.read()
       
       # Extract diagrams between ```mermaid and ```
       diagrams = re.findall(r'```mermaid\n(.*?)\n```', content, re.DOTALL)
       return diagrams
   
   def render_diagram(diagram_content, output_path):
       with open('temp.mmd', 'w') as f:
           f.write(diagram_content)
       
       subprocess.run(['mmdc', '-i', 'temp.mmd', '-o', output_path])
   
   # Example usage
   diagrams = extract_diagrams('WEBSITE_DIAGRAMS.md')
   for i, diagram in enumerate(diagrams):
       render_diagram(diagram, f'diagram_{i}.svg')
   ```

## Customizing Diagrams

The diagrams provided use a consistent color scheme and styling:

- API Layer: Light gray (#f9f9f9)
- Core Components: Light blue (#e6f3ff)
- Plugins: Light green (#e6ffe6)
- External Systems: Light orange (#fff5e6)

To customize these styles for your website:

1. Adjust the classDef declarations in the Mermaid code
2. Use your website's color scheme for consistency
3. Consider light/dark mode toggling if your website supports it

## Future Diagram Additions

As AgenticGoKit evolves, consider adding:

1. **User Journey Diagrams** - showing how users interact with agents
2. **Sequence Diagrams** - detailed interaction flows between components
3. **State Diagrams** - agent state transitions during processing
4. **Benchmarking Visualizations** - performance comparisons
5. **Integration Architecture** - patterns for integrating with external systems

---

This index provides a comprehensive guide to all architecture documentation created for the AgenticGoKit website.
