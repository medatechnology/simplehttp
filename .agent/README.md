# .agent Directory

This directory contains AI agent context, rules, and workflows for the SimpleHttp project.

## Purpose

The `.agent` directory serves as the **persistent memory and rule system** for AI coding assistants working on this project. It ensures:

1. **Consistency** - All agents follow the same conventions and patterns
2. **Context Preservation** - Knowledge persists across sessions
3. **Onboarding** - New agents can quickly understand the project
4. **Quality** - Mandatory rules prevent common mistakes

## Contents

### 📄 `context.md`
**Comprehensive project context and documentation**

Contains:
- Project overview and architecture
- Code conventions and patterns
- Development guidelines
- Dependency management
- Quick reference guides

**When to read:** 
- First time working on the project
- Before making significant changes
- When unsure about patterns

**When to update:**
- Before session ends (mandatory)
- When new patterns are introduced
- When architecture changes
- When dependencies are added/removed

### 📋 `rules.md`
**Mandatory coding rules and conventions**

Contains:
- Critical rules that must never be violated
- Naming conventions
- Security guidelines
- Code review checklist
- Common mistakes to avoid

**When to read:**
- Before writing any code
- When reviewing code
- When adding new features

**When to update:**
- When new rules are established
- When patterns evolve
- When antipatterns are discovered

### 📁 `workflows/`
**Automated workflows and procedures**

Contains workflow files (`.md`) with step-by-step procedures for common tasks.

Current workflows:
- `update-context.md` - Update project context before session end

**When to read:**
- When performing the specific task
- Before session end (for update-context)

**When to update:**
- When workflows change
- When new workflows are needed

## Usage for AI Agents

### Starting a Session
```markdown
1. Read `.agent/context.md` to understand the project
2. Review `.agent/rules.md` for mandatory conventions
3. Check for relevant workflows in `.agent/workflows/`
```

### During Development
```markdown
1. Follow patterns documented in context.md
2. Adhere to rules in rules.md
3. Reference quick-reference sections as needed
```

### Ending a Session
```markdown
1. Run `/update-context` workflow (MANDATORY)
2. Update context.md with new patterns
3. Update rules.md if conventions changed
4. Commit context changes separately
```

## File Naming Convention

- **Markdown files** (`.md`) for documentation
- **Lowercase with hyphens** for workflow files (e.g., `update-context.md`)
- **Descriptive names** that clearly indicate purpose

## Maintenance

### Who Maintains These Files?
Both AI agents and human developers should maintain these files.

### Update Frequency
- **context.md**: Every session, before ending
- **rules.md**: When conventions change
- **workflows/**: As needed

### Version Control
These files should be:
- ✅ Committed to version control
- ✅ Reviewed like code
- ✅ Updated atomically with related changes
- ✅ Given meaningful commit messages

## Benefits

### For AI Agents
- **Persistent memory** across sessions
- **Clear guidelines** for decision making
- **Reduced errors** from following rules
- **Faster onboarding** to the project

### For Developers
- **Consistent codebase** regardless of who writes it
- **Better documentation** of patterns and conventions
- **Easier onboarding** for new team members
- **Reduced bike-shedding** with established patterns

## Integration with Development Flow

```
┌─────────────────────────────────────────┐
│        Start Work Session               │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│   Read .agent/context.md & rules.md     │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│      Write Code Following Patterns      │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│    Run Relevant Workflows               │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│   Update context.md (if needed)         │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│   Run /update-context Workflow          │
│        (MANDATORY BEFORE END)           │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│           Commit Changes                │
│  (Context changes committed separately) │
└─────────────────────────────────────────┘
```

## Best Practices

### ✅ DO
- Read context before making changes
- Follow documented patterns
- Update context when patterns change
- Run `/update-context` before session end
- Keep documentation current and accurate
- Document the "why" not just the "what"
- Use clear examples in documentation

### ❌ DON'T
- Skip reading context for "quick" changes
- Invent new patterns without documenting
- Ignore the mandatory rules
- Forget to update context at session end
- Copy-paste without understanding
- Leave TODOs in context files
- Document only what code does (code should be self-documenting)

## Example Workflow

### Adding a New Feature

```markdown
1. Read context.md to understand architecture
2. Check rules.md for relevant conventions
3. Follow the pattern for similar features
4. Implement the feature
5. Add example to example/ directory
6. Update context.md with new patterns (if any)
7. Run tests/examples to verify
8. Run /update-context workflow
9. Commit feature and context separately
```

### Bugfix

```markdown
1. Read context.md for relevant section
2. Check rules.md for conventions
3. Fix the bug following patterns
4. Verify fix with examples
5. Update context.md if bug revealed missing documentation
6. Run /update-context workflow
7. Commit fix
```

## Quick Links

| File | Purpose | Update Frequency |
|------|---------|-----------------|
| [context.md](./context.md) | Project knowledge & patterns | Every session |
| [rules.md](./rules.md) | Mandatory rules & conventions | When rules change |
| [workflows/update-context.md](./workflows/update-context.md) | Context update procedure | As needed |

## Questions?

If you're an AI agent and unclear about something:
1. Check `context.md` for general information
2. Check `rules.md` for specific conventions
3. Check `workflows/` for procedures
4. Ask the user if still unclear

If you're a human developer:
1. Read through the agent documentation
2. Follow the same patterns when working on the project
3. Update these files when you discover better patterns

---

**Remember:** This directory is the project's persistent memory. Keep it current, accurate, and comprehensive!
