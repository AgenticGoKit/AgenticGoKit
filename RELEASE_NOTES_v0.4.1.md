# AgenticGoKit v0.4.1 Release Notes

**Release Date**: September 6, 2025  
**Release Type**: Patch Release  
**Previous Version**: v0.4.0

##  Overview

AgenticGoKit v0.4.1 introduces a comprehensive knowledge management system, significant developer experience improvements, and important bug fixes. This patch release adds powerful new CLI capabilities while maintaining full backward compatibility.

##  What's New

###  Knowledge Management System

A complete knowledge base management solution has been added to AgenticGoKit, providing powerful document ingestion, semantic search, and knowledge validation capabilities.

#### New CLI Commands

```bash
# Upload documents to knowledge base
agentcli knowledge upload documents/ --collection my-docs

# Search knowledge base semantically  
agentcli knowledge search "machine learning concepts" --top-k 5

# List all knowledge collections
agentcli knowledge list --verbose

# Validate knowledge base configuration
agentcli knowledge validate --fix-config
```

#### Key Features

- **Document Ingestion**: Support for text, markdown, PDF, and multiple file formats
- **Semantic Search**: Vector-based similarity search with scoring
- **Intelligent Embedding**: Auto-detection of optimal embedding dimensions
- **Smart Validation**: Comprehensive configuration validation with helpful error messages
- **Auto-Configuration**: Dynamic database schema generation with proper vector dimensions

#### Embedding Intelligence

- **Automatic dimension detection**:
  - `nomic-embed-text`: 768 dimensions
  - `text-embedding-3-small`: 1536 dimensions
  - Smart fallbacks for other models
- **Configuration optimization** with validation and suggestions
- **Compatibility fixes** for vector dimension mismatches

### Developer Experience Improvements

#### Better Logging by Default

- **Console logs by default**: New projects now use human-readable console logs instead of JSON
- **Configurable logging**: Full control via `agentflow.toml` configuration
- **File logging support**: Optional file output with rotation capabilities
- **Better error messages**: Enhanced troubleshooting guidance and specific error context

#### Enhanced Debugging

- **Added missing logging utilities**: `DebugLogWithFields` and `DebugLog` functions
- **Structured logging support**: Convert field maps to chained logger calls automatically
- **Improved error reporting**: Better context and actionable error messages

### Bug Fixes & Improvements

#### Module Import Consistency
- **Fixed all imports** to use lowercase module name for better Go ecosystem compatibility
- **Resolved build issues** related to module naming inconsistencies

#### Knowledge Base Fixes
- **Fixed dimension mismatch** between database schema (3072) and embedding models (768)
- **Updated database schemas** to use proper vector dimensions for embedding models
- **Resolved search failures** due to incompatible vector dimensions

## Technical Impact

- **94 files changed**
- **33,919 lines added**, 8,956 lines removed
- **Major documentation expansion**: 25KB+ of new comprehensive documentation
- **New CLI command suite**: Complete knowledge management workflow
- **Enhanced logging infrastructure**: Configurable, developer-friendly logging system

## Installation

### Using Installation Script (Recommended)

```bash
# Linux/macOS
curl -fsSL https://raw.githubusercontent.com/kunalkushwaha/agenticgokit/master/install.sh | bash

# Windows PowerShell
iwr -useb https://raw.githubusercontent.com/kunalkushwaha/agenticgokit/master/install.ps1 | iex
```

### Manual Download

Download platform-specific binaries from the [GitHub Releases](https://github.com/kunalkushwaha/agenticgokit/releases/tag/v0.4.1) page:

- `agentcli-linux-amd64` - Linux (Intel/AMD 64-bit)
- `agentcli-linux-arm64` - Linux (ARM 64-bit)
- `agentcli-darwin-amd64` - macOS (Intel)
- `agentcli-darwin-arm64` - macOS (Apple Silicon)
- `agentcli-windows-amd64.exe` - Windows (64-bit)
- `agentcli-windows-arm64.exe` - Windows (ARM)

### Verify Installation

```bash
agentcli version
# Expected output: agentcli v0.4.1 (commit: xxxxxxx, built: 2025-09-06T...)
```

## 🔄 Migration Guide

### From v0.4.0 to v0.4.1

#### No Breaking Changes
This is a **patch release** with full backward compatibility. Existing projects and configurations will continue to work without modification.

#### New Features Available
- **Knowledge commands**: Available immediately for existing projects
- **Better logging**: New projects automatically get console logging; existing projects can update `agentflow.toml`:
  ```toml
  [logging]
  level = "info"
  format = "console"  # Change from "json" to "console"
  ```

#### Optional Improvements
- **Update existing projects** to use new logging format for better development experience
- **Explore knowledge management** for projects that could benefit from document ingestion and search

## Documentation Updates

### New Documentation
- **Knowledge Management Guide**: Complete tutorial for using the new knowledge system
- **CLI Quick Reference**: Comprehensive command reference with examples
- **Project Templates Guide**: Enhanced template creation and management

### Updated Documentation  
- **25KB+ of documentation improvements** across tutorials and guides
- **Enhanced debugging guides** with practical examples
- **Improved getting-started tutorials** with better flow and examples
- **Updated API reference** with new CLI capabilities

## Testing & Quality

### Knowledge Management Testing
- **Comprehensive test suite** for all knowledge commands
- **Embedding compatibility tests** across different providers
- **Vector dimension validation** tests
- **Multi-format document processing** tests

### Logging System Testing
- **Console vs JSON format** output validation
- **Configuration loading** tests
- **Error message clarity** validation

## Known Issues

### Minor Issues
- **Large document processing**: Very large documents (>10MB) may require chunking optimization
- **Embedding model downloads**: First-time Ollama model downloads may take time
- **MCP tool discovery have some issues** Should be fixed with #44

### Workarounds
- **Document chunking**: Use `--chunk-size` flag for large documents
- **Model pre-download**: Run `ollama pull nomic-embed-text` before first use

## 🔮 What's Next



- **Issues**: [GitHub Issues](https://github.com/kunalkushwaha/agenticgokit/issues)
- **Discussions**: [GitHub Discussions](https://github.com/kunalkushwaha/agenticgokit/discussions)
- **Documentation**: [AgenticGoKit Docs](https://github.com/kunalkushwaha/agenticgokit/tree/master/docs)

## 🔗 Links

- **GitHub Release**: [v0.4.1](https://github.com/kunalkushwaha/agenticgokit/releases/tag/v0.4.1)
- **Full Changelog**: [v0.4.0...v0.4.1](https://github.com/kunalkushwaha/agenticgokit/compare/v0.4.0...v0.4.1)
- **Installation Guide**: [INSTALL.md](https://github.com/kunalkushwaha/agenticgokit/blob/master/INSTALL.md)
- **Documentation**: [docs/](https://github.com/kunalkushwaha/agenticgokit/tree/master/docs)

---

**Full Changelog**: https://github.com/kunalkushwaha/agenticgokit/compare/v0.4.0...v0.4.1
