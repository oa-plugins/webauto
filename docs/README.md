# webauto Documentation

This directory contains general knowledge and guides for the webauto plugin.

## Document Organization

### Principle
- **Issue-specific documentation**: Use GitHub issues and issue comments
- **General knowledge**: Store in `docs/` directory
- **Architecture & API**: Keep in root directory (ARCHITECTURE.md, README.md)

### Available Documents

#### Implementation Guides
- **[implementation-guide.md](implementation-guide.md)**: Step-by-step implementation guide for developers

#### Performance Documentation
- **[performance-guide.md](performance-guide.md)**: Performance optimization strategies and best practices
- **[performance-baseline.md](performance-baseline.md)**: Performance benchmark baseline results

## Why Not IMPLEMENTATION_ISSUE_*.md?

Issue-specific implementation details should be documented in GitHub issue comments for these reasons:

1. **Single Source of Truth**: All issue-related information in one place
2. **Better Tracking**: GitHub automatically links commits, PRs, and discussions
3. **Searchability**: GitHub's search is optimized for issues and comments
4. **Version Control**: Issue timeline provides natural versioning
5. **Collaboration**: Better for team discussions and reviews

## Document Lifecycle

### When to Create a Document

**✅ Create in `docs/`**:
- General implementation guides
- Performance optimization strategies
- Best practices and patterns
- Architecture decisions (ADRs)
- Testing strategies
- Deployment guides

**❌ Don't Create Separate Files**:
- Issue-specific implementation summaries → Use GitHub issue comments
- Performance optimization summaries → Use GitHub issue comments
- Bug fix details → Use GitHub issue comments
- Feature implementation logs → Use GitHub issue comments

### When to Update

- Update `docs/` when general knowledge changes
- Reference GitHub issues for historical context
- Keep architecture docs (ARCHITECTURE.md) up-to-date with current state

## Related Documentation

- **Root Directory**:
  - [ARCHITECTURE.md](../ARCHITECTURE.md) - System architecture and API documentation
  - [README.md](../README.md) - Project overview and quick start
  - [CLAUDE.md](../CLAUDE.md) - Claude Code integration guide

- **GitHub Issues**:
  - Implementation details for specific features
  - Bug reports and fixes
  - Performance optimizations
