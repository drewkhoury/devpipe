
# ai-docs

This directory contains documentation for Windsurf AI.

## Global Rules

`~/.codeium/windsurf/memories/global_rules.md`

SCOPE:
- All projects
- All repos
- Always on
- Highest authority

## Project Guide

`./.windsurf/rules/project-guide.md`

SCOPE:
- Only this workspace
- Overrides .windsurfrules
- Overrides repo ai-docs
- Does NOT override global_rules.md

## Windsurf Rules

`./.windsurfrules`

SCOPE:
- This repo only
- Team shared
- Always on when repo is open

# Project Docs (this folder)

`./ai-docs/*.md`

- Scoped to the current repo only
- Visible to Windsurf only when that repo is open
- NOT automatically enforced
- NOT globally persistent
- NOT stored in memory
- NOT higher priority than .windsurfrules
- NOT higher priority than UI Project Rules
- NOT higher priority than Global Rules

They only take effect when ONE of these is true:

- You explicitly reference them in a prompt
- .windsurfrules explicitly tells Cascade to consult them
- A human pastes their contents into chat