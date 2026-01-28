---
description: Update project context before session end or token limit
---

# Update Project Context Workflow

This workflow should be executed **AUTOMATICALLY** before:
- Session ends
- Token limits are approaching (when you see the token warning)
- Major architectural changes are committed
- New patterns or conventions are introduced

## Steps

### 1. Review Session Changes
Review all changes made during the current session:
- What files were modified?
- What new patterns were introduced?
- Were there any architectural changes?
- Were new dependencies added?
- Did coding conventions change?

### 2. Update Context File
Edit `.agent/context.md` to include:

**If new patterns were introduced:**
- Add to "Development Patterns" section
- Document the pattern with clear examples
- Explain when and why to use it

**If dependencies changed:**
- Update "Dependency Management" section
- Add new dependencies to the list
- Document why they were added

**If architecture changed:**
- Update architecture diagrams
- Update file structure if needed
- Document new components

**If conventions changed:**
- Update "Code Conventions" section
- Add examples of the new convention
- Update the checklist if needed

### 3. Update Version Files (if applicable)

If this was a feature addition or bug fix:

// turbo
```bash
# Read current version
cat .version
```

If version should be bumped, update `.version`:
```bash
echo "0.0.X" > .version
```

Update `.changelog`:
```bash
echo "main.0.0.X\tCommit: <description of changes>" >> .changelog
```

### 4. Verify Context Completeness

Check that `.agent/context.md` includes:
- [ ] All coding conventions are documented
- [ ] All architectural patterns are explained
- [ ] Dependencies are up to date
- [ ] Version information is current
- [ ] Examples reflect current patterns
- [ ] Agent instructions are clear
- [ ] Last reviewed date is updated

### 5. Commit Context Updates

// turbo
```bash
git add .agent/context.md
git add .version .changelog  # if updated
git commit -m "docs: update project context [session end]"
```

## Example Commit Messages

- `docs: update project context [session end]`
- `docs: update context with new middleware pattern`
- `docs: update context - added new framework adapter`
- `docs: update context and version to 0.0.4`

## What Makes a Good Context Update

✅ **DO:**
- Document new patterns with code examples
- Explain the reasoning behind changes
- Update the "Last reviewed" date at the bottom
- Be specific about what changed and why
- Include practical examples
- Update relevant checklists

❌ **DON'T:**
- Just say "updated context" without details
- Copy-paste code without explanation
- Leave sections incomplete
- Forget to update the date
- Skip updating related sections

## Quick Context Checklist

Use this checklist before ending each session:

```markdown
- [ ] All new patterns documented in context.md
- [ ] Code conventions section is current
- [ ] Architecture diagrams reflect reality
- [ ] Dependencies list is accurate
- [ ] Version and changelog updated (if needed)
- [ ] Examples are working and relevant
- [ ] "Last reviewed" date updated
- [ ] Agent instructions are clear
- [ ] No TODOs left in context file
```

## Notes

- This is a **CRITICAL** workflow - context preservation is essential
- The context file is the memory of the project across sessions
- Future AI agents (and developers) rely on this documentation
- When in doubt, over-document rather than under-document
- Context updates should be committed separately from feature changes
