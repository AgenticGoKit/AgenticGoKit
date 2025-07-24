# Contributor Documentation

> **Navigation:** [Documentation Home](../README.md) → **Contributors**

**For developers contributing to AgenticGoKit**

This section contains documentation specifically for contributors to the AgenticGoKit project. If you're looking to use AgenticGoKit in your projects, see the [main documentation](../README.md).

## 🚀 Getting Started

### **Essential Reading**
- **[Contributor Guide](ContributorGuide.md)** - Start here! Development setup and contribution workflow
- **[Core vs Internal](CoreVsInternal.md)** - Understanding the public API vs implementation details
- **[Code Style](CodeStyle.md)** - Go standards and project conventions

### **Development Process**
- **[Adding Features](AddingFeatures.md)** - How to extend AgenticGoKit with new features
- **[Testing Strategy](Testing.md)** - Unit tests, integration tests, and benchmarks
- **[Documentation Standards](DocsStandards.md)** - Writing user-focused documentation

### **Project Management**
- **[Release Process](ReleaseProcess.md)** - How releases are managed and versioned

## 🏗️ Architecture Overview

AgenticGoKit is designed with a clear separation between public APIs and internal implementation:

- **`core/` package**: Public API that users interact with
- **`internal/` package**: Implementation details not exposed to users
- **`cmd/` package**: CLI tools and utilities
- **`examples/` directory**: Working examples and tutorials

## 🧪 Development Workflow

1. **Fork and Clone**: Fork the repository and clone your fork
2. **Create Branch**: Create a feature branch for your changes
3. **Develop**: Make your changes following our code style
4. **Test**: Run tests and add new tests for your changes
5. **Document**: Update documentation as needed
6. **Submit PR**: Create a pull request with a clear description

## 📋 Contribution Guidelines

### **Code Quality**
- Follow Go best practices and our [Code Style](CodeStyle.md)
- Write comprehensive tests for new features
- Ensure all tests pass before submitting
- Use meaningful commit messages

### **Documentation**
- Update user documentation for new features
- Follow our [Documentation Standards](DocsStandards.md)
- Include code examples in documentation
- Keep contributor docs up to date

### **Communication**
- Use GitHub Issues for bug reports and feature requests
- Use GitHub Discussions for questions and community interaction
- Be respectful and constructive in all interactions

## 🔧 Development Tools

### **Required Tools**
```bash
# Go (1.21+)
go version

# Linting
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Testing
go test ./...

# Documentation generation (if applicable)
go run tools/docgen/main.go
```

### **Recommended IDE Setup**
- **VS Code** with Go extension
- **GoLand** by JetBrains
- **Vim/Neovim** with Go plugins

## 📚 Additional Resources

- **[GitHub Repository](https://github.com/kunalkushwaha/agenticgokit)**
- **[GitHub Issues](https://github.com/kunalkushwaha/agenticgokit/issues)**
- **[GitHub Discussions](https://github.com/kunalkushwaha/agenticgokit/discussions)**
- **[Main Documentation](../README.md)** - For users of AgenticGoKit

---

**Thank you for contributing to AgenticGoKit!** 🎉