# Voice To Post High Level Tech Design

---

## 📋 Instructions for AI Assistants

**This is a living document for phased development.**

**Critical guidelines:**
- We build **ONE phase at a time** in sequence
- **Only the "Active" phase** is currently being implemented
- **Future phases have HIGH uncertainty** and will change based on learnings
- **Do NOT plan or build** beyond the current active phase
- **After each phase completes**, this document will be updated with lessons learned
- Use your planning skills to create detailed plans for the active phase only

**When you see "Status: Planned"** - treat it as a rough sketch that WILL change.

---

## Project Overview

Build a Go CLI tool that automates voice-to-blog workflow: record audio → transcribe → LLM cleanup → generate Hugo blog post.

**Core Stack:**
- CLI: Kong
- Audio: malgo
- Transcription: Deepgram (production) / Whisper API (budget)
- LLM: Claude (liushuangls/go-anthropic/v2)
- Storage: Tigris Object Store
- Site: Hugo (existing)

**Current Status:**
**Last Updated:** 2025-10-31
**Active Phase:** Phase 1

---

## Phase 1: MVP Foundation

**Status:** Active
**Goal:** Prove the end-to-end workflow with minimal complexity

**Deliverables:**
- Basic Kong CLI: `record`, `transcribe`, `process` commands
- Audio recording with malgo (local WAV files)
- Single transcription provider (Whisper API)
- Simple Hugo markdown generation (no LLM yet)
- Manual workflow (run each command separately)

**Success Criteria:**
- Can record voice memo
- Can transcribe to text
- Can generate markdown with frontmatter
- Hugo builds site with new post

**Out of Scope:**
- Config files
- Cloud storage
- LLM cleanup
- Multiple providers
- Retry logic
- Progress indicators

---

## Phase 2: AI Enhancement

**Status:** Planned (HIGH UNCERTAINTY)
**Goal:** Add intelligence to transform raw transcripts

**Anticipated:**
- Claude API integration for cleanup
- Auto-generate titles/tags/categories
- Multiple transcription providers (add Deepgram)
- Tigris storage for audio backups
- Config file support
- Basic retry logic

**Will be updated after Phase 1 completion**

---

## Phase 3: Production Robustness

**Status:** Planned (HIGH UNCERTAINTY)
**Goal:** Make reliable for daily use

**Anticipated:**
- Workflow state persistence
- Comprehensive error handling
- Progress indicators
- `publish` and `workflow` commands
- Preview mode
- Structured logging

**Will be updated after Phase 2 completion**

---

## Phase 4: Advanced Features

**Status:** Planned (VERY HIGH UNCERTAINTY)
**Goal:** Polish UX and add power features

**Possible:**
- Silence detection
- Speaker diarization
- Batch processing
- Git integration

**This phase will likely change significantly**

---

## Lessons Learned

*Updated after each phase*

### Phase 1
*To be filled*

### Phase 2
*To be filled*

### Phase 3
*To be filled*

### Phase 4
*To be filled*

---

## Key Decisions

- Kong chosen over Cobra/urfave (lightweight, no IoC)
- malgo for audio (zero system deps)
- Whisper API for Phase 1 (cost), Deepgram for Phase 2+ (quality)
- Separate `cmd/cli` from existing server

---

## Open Questions

- [ ] Should LLM processing move to Phase 1?
- [ ] Create `content/memos/` or use `content/posts/`?
- [ ] Interactive title prompt or auto-generate?