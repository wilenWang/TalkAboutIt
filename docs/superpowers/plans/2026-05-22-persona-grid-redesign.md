# Persona Grid Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Refactor PersonaManagePage into a square-card grid with archetype-based filtering, add archetype field to persona schema, and support image avatars.

**Architecture:** Add `archetype` string field to backend Persona schema and API summary response. Frontend creates reusable `PersonaCard` and `ArchetypeFilter` components. Grid uses CSS `aspect-ratio: 1/1` with hover overlay for description/actions. Avatar supports both emoji and image URLs.

**Tech Stack:** Go 1.26 (backend), React 19 + TypeScript + Tailwind CSS (frontend), SQLite (persistence), Vite (build)

---

## File Structure

| File | Action | Responsibility |
|------|--------|---------------|
| `backend/internal/persona/persona.go` | Modify | Add `Archetype` field to Persona struct |
| `backend/internal/api/router.go` | Modify | Add `Archetype` to PersonaSummary and ListPersonas mapping |
| `backend/personas/*.json` (5 files) | Modify | Add `archetype` field to each persona |
| `frontend/src/types/persona.ts` | Modify | Add `archetype: string` to Persona interface |
| `frontend/src/types/index.ts` | Modify | Add `archetype: string` to PersonaSummary |
| `frontend/src/i18n/translations.ts` | Modify | Add archetype and filter-related translations |
| `frontend/src/components/PersonaCard.tsx` | Create | Square card with avatar, name, archetype, hover overlay |
| `frontend/src/components/ArchetypeFilter.tsx` | Create | Pill filter bar for archetypes |
| `frontend/src/pages/PersonaManagePage.tsx` | Modify | Replace list with grid + filter integration |
| `frontend/src/components/PersonaEditor.tsx` | Modify | Add archetype text input |
| `frontend/public/personas/steve-jobs.png` | Create | Steve Jobs pixel avatar image |
| `backend/personas/steve-jobs.json` | Modify | Update avatar to image path |

---

## Task 1: Backend — Add Archetype to Persona Schema

**Files:**
- Modify: `backend/internal/persona/persona.go`

- [ ] **Step 1: Add `Archetype` field to Persona struct**

```go
// In backend/internal/persona/persona.go, add to the Persona struct:
type Persona struct {
	SchemaVersion    string          `json:"schema_version"`
	ID               string          `json:"id"`
	Name             string          `json:"name"`
	DisplayName      string          `json:"display_name"`
	Avatar           string          `json:"avatar"`
	RoleTitle        string          `json:"role_title"`
	Description      string          `json:"description"`
	Tags             []string        `json:"tags"`
	Archetype        string          `json:"archetype"` // ADD THIS LINE
	Language         Language        `json:"language"`
	// ... rest of fields
}
```

- [ ] **Step 2: Verify build compiles**

Run:
```bash
cd /Users/wilenwang/TalkAboutIt/backend
go build ./...
```

Expected: No errors.

- [ ] **Step 3: Commit**

```bash
git add backend/internal/persona/persona.go
git commit -m "feat(persona): add Archetype field to schema"
```

---

## Task 2: Backend — Expose Archetype in API Summary

**Files:**
- Modify: `backend/internal/api/router.go`

- [ ] **Step 1: Add Archetype to PersonaSummary struct**

```go
// In backend/internal/api/router.go, modify PersonaSummary:
type PersonaSummary struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name"`
	Avatar      string   `json:"avatar"`
	RoleTitle   string   `json:"role_title"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Archetype   string   `json:"archetype"` // ADD THIS LINE
}
```

- [ ] **Step 2: Add Archetype to ListPersonas mapping**

```go
// In backend/internal/api/router.go, in the ListPersonas function,
// inside the for loop, add Archetype to the summary:
for _, p := range personas {
	summaries = append(summaries, PersonaSummary{
		ID:          p.ID,
		Name:        p.Name,
		DisplayName: p.DisplayName,
		Avatar:      p.Avatar,
		RoleTitle:   p.RoleTitle,
		Description: p.Description,
		Tags:        p.Tags,
		Archetype:   p.Archetype, // ADD THIS LINE
	})
}
```

- [ ] **Step 3: Verify build compiles**

Run:
```bash
cd /Users/wilenwang/TalkAboutIt/backend
go build ./...
```

Expected: No errors.

- [ ] **Step 4: Commit**

```bash
git add backend/internal/api/router.go
git commit -m "feat(api): expose archetype in PersonaSummary"
```

---

## Task 3: Backend — Update 5 Persona JSON Files with Archetype

**Files:**
- Modify: `backend/personas/steve-jobs.json`
- Modify: `backend/personas/elon-musk.json`
- Modify: `backend/personas/naval-ravikant.json`
- Modify: `backend/personas/zhang-xiaolong.json`
- Modify: `backend/personas/zhang-yiming.json`

- [ ] **Step 1: Add archetype to steve-jobs.json**

Add `"archetype": "Visionary"` after the `tags` field.

```json
{
  "schema_version": "persona.v1",
  "id": "steve-jobs",
  "name": "Steve Jobs",
  "display_name": "Steve Jobs",
  "avatar": "🍎",
  "role_title": "Apple Co-founder",
  "description": "...",
  "tags": ["product", "design", "consumer-tech", "visionary", "founder"],
  "archetype": "Visionary",
  "language": { ... }
}
```

- [ ] **Step 2: Add archetype to elon-musk.json**

Add `"archetype": "Engineer"` after the `tags` field.

- [ ] **Step 3: Add archetype to naval-ravikant.json**

Add `"archetype": "Philosopher"` after the `tags` field.

- [ ] **Step 4: Add archetype to zhang-xiaolong.json**

Add `"archetype": "Craftsman"` after the `tags` field.

- [ ] **Step 5: Add archetype to zhang-yiming.json**

Add `"archetype": "Operator"` after the `tags` field.

- [ ] **Step 6: Verify persona loader can parse all files**

Run a quick Go program or test:
```bash
cd /Users/wilenwang/TalkAboutIt/backend
go test ./internal/persona/ -run TestLoadAll -v
```

If no test exists, run the server briefly:
```bash
cd /Users/wilenwang/TalkAboutIt/backend
go run ./cmd/server &
SERVER_PID=$!
sleep 2
curl -s http://localhost:8080/api/v1/personas | head -c 500
kill $SERVER_PID
```

Expected: Server starts without JSON parse errors. API returns persona list.

- [ ] **Step 7: Commit**

```bash
git add backend/personas/*.json
git commit -m "feat(personas): add archetype to all 5 personas"
```

---

## Task 4: Frontend — Add Archetype to Type Definitions

**Files:**
- Modify: `frontend/src/types/persona.ts`
- Modify: `frontend/src/types/index.ts`

- [ ] **Step 1: Add archetype to Persona interface**

```typescript
// In frontend/src/types/persona.ts, add to Persona interface:
export interface Persona {
  schema_version: string;
  id: string;
  name: string;
  display_name: string;
  avatar: string;
  role_title: string;
  description: string;
  tags: string[];
  archetype: string; // ADD THIS LINE
  language: PersonaLanguage;
  // ... rest
}
```

Also add to `createEmptyPersona()`:
```typescript
export function createEmptyPersona(): Persona {
  return {
    // ... existing fields
    tags: [],
    archetype: '', // ADD THIS LINE
    language: { ... }
    // ... rest
  };
}
```

- [ ] **Step 2: Add archetype to PersonaSummary**

```typescript
// In frontend/src/types/index.ts:
export interface PersonaSummary {
  id: string;
  name: string;
  display_name: string;
  avatar: string;
  role_title: string;
  description: string;
  tags: string[];
  archetype: string; // ADD THIS LINE
}
```

- [ ] **Step 3: Verify TypeScript compiles**

Run:
```bash
cd /Users/wilenwang/TalkAboutIt/frontend
npx tsc --noEmit
```

Expected: No type errors.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/types/persona.ts frontend/src/types/index.ts
git commit -m "feat(types): add archetype field to Persona and PersonaSummary"
```

---

## Task 5: Frontend — Add i18n Translations

**Files:**
- Modify: `frontend/src/i18n/translations.ts`

- [ ] **Step 1: Add archetype and filter translations**

Add these entries to the `translations` object in `frontend/src/i18n/translations.ts`:

```typescript
// Archetype names
archetypeVisionary:   { 'zh-CN': '远见者', 'en-US': 'Visionary' },
archetypeEngineer:    { 'zh-CN': '工程师', 'en-US': 'Engineer' },
archetypePhilosopher: { 'zh-CN': '哲思者', 'en-US': 'Philosopher' },
archetypeCraftsman:   { 'zh-CN': '匠人',   'en-US': 'Craftsman' },
archetypeOperator:    { 'zh-CN': '运营者', 'en-US': 'Operator' },

// Filter
labelFilterAll:       { 'zh-CN': '全部',   'en-US': 'All' },
msgNoFilterResults:   { 'zh-CN': '没有符合条件的人物', 'en-US': 'No personas match this filter' },
```

Place them logically within the existing sections (archetypes under Labels, filter under Messages).

- [ ] **Step 2: Verify no duplicate keys**

Run:
```bash
cd /Users/wilenwang/TalkAboutIt/frontend
npx tsc --noEmit
```

Expected: No type errors.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/i18n/translations.ts
git commit -m "feat(i18n): add archetype name and filter translations"
```

---

## Task 6: Frontend — Create PersonaCard Component

**Files:**
- Create: `frontend/src/components/PersonaCard.tsx`

- [ ] **Step 1: Create the component file**

```tsx
// frontend/src/components/PersonaCard.tsx
import type { PersonaSummary } from '../types';
import { useLanguage } from '../i18n/LanguageContext';

interface Props {
  persona: PersonaSummary;
  onEdit: (id: string) => void;
  onDelete: (id: string) => void;
}

function isImageAvatar(avatar: string): boolean {
  return avatar.startsWith('http') || avatar.includes('/');
}

export default function PersonaCard({ persona, onEdit, onDelete }: Props) {
  const { t } = useLanguage();

  return (
    <div className="group relative bg-white rounded-xl aspect-square flex flex-col items-center justify-center border border-black/[0.06] cursor-pointer overflow-hidden hover:shadow-[rgba(0,0,0,0.04)_0px_4px_18px] transition-shadow">
      {/* Avatar */}
      {isImageAvatar(persona.avatar) ? (
        <img
          src={persona.avatar}
          alt={persona.name}
          className="w-16 h-16 rounded-xl object-cover mb-2"
        />
      ) : (
        <span className="text-[40px] mb-1">{persona.avatar}</span>
      )}

      {/* Name */}
      <div className="font-semibold text-sm text-black/95">{persona.name}</div>

      {/* Archetype */}
      {persona.archetype && (
        <div className="text-[11px] text-[#a39e98] mt-1">
          {t('archetype' + persona.archetype) || persona.archetype}
        </div>
      )}

      {/* Hover overlay */}
      <div className="absolute inset-0 bg-white/[0.98] flex flex-col items-center justify-center p-3.5 opacity-0 group-hover:opacity-100 transition-opacity duration-200">
        <div className="text-xs text-[#615d59] text-center line-clamp-5 leading-relaxed">
          {persona.description}
        </div>
        <div className="mt-2.5 flex gap-1.5">
          <button
            onClick={(e) => {
              e.stopPropagation();
              onEdit(persona.id);
            }}
            className="px-2.5 py-1 bg-[#f2f9ff] text-[#0075de] rounded text-[11px] hover:bg-[#e0f0ff] transition-colors"
          >
            {t('actionEdit')}
          </button>
          <button
            onClick={(e) => {
              e.stopPropagation();
              onDelete(persona.id);
            }}
            className="px-2.5 py-1 bg-red-50 text-red-500 rounded text-[11px] hover:bg-red-100 transition-colors"
          >
            {t('actionDelete')}
          </button>
        </div>
      </div>
    </div>
  );
}
```

- [ ] **Step 2: Verify TypeScript compiles**

Run:
```bash
cd /Users/wilenwang/TalkAboutIt/frontend
npx tsc --noEmit
```

Expected: No type errors.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/PersonaCard.tsx
git commit -m "feat(ui): add PersonaCard square component with hover overlay"
```

---

## Task 7: Frontend — Create ArchetypeFilter Component

**Files:**
- Create: `frontend/src/components/ArchetypeFilter.tsx`

- [ ] **Step 1: Create the component file**

```tsx
// frontend/src/components/ArchetypeFilter.tsx
import { useLanguage } from '../i18n/LanguageContext';

const ARCHETYPES = ['Visionary', 'Engineer', 'Philosopher', 'Craftsman', 'Operator'] as const;

interface Props {
  active: string | null;
  onChange: (archetype: string | null) => void;
}

export default function ArchetypeFilter({ active, onChange }: Props) {
  const { t } = useLanguage();

  return (
    <div className="flex flex-wrap items-center gap-2">
      <span className="text-[13px] text-[#615d59]">{t('labelFilter')}:</span>
      <button
        onClick={() => onChange(null)}
        className={`px-3 py-1 rounded-full text-[13px] transition-colors ${
          active === null
            ? 'border border-[#0075de] bg-[#f2f9ff] text-[#0075de]'
            : 'border border-black/10 bg-white text-[#615d59] hover:bg-black/[0.03]'
        }`}
      >
        {t('labelFilterAll')}
      </button>
      {ARCHETYPES.map((arch) => (
        <button
          key={arch}
          onClick={() => onChange(arch)}
          className={`px-3 py-1 rounded-full text-[13px] transition-colors ${
            active === arch
              ? 'border border-[#0075de] bg-[#f2f9ff] text-[#0075de]'
              : 'border border-black/10 bg-white text-[#615d59] hover:bg-black/[0.03]'
          }`}
        >
          {t('archetype' + arch) || arch}
        </button>
      ))}
    </div>
  );
}
```

- [ ] **Step 2: Add `labelFilter` to translations**

```typescript
// In frontend/src/i18n/translations.ts, add:
labelFilter: { 'zh-CN': '筛选', 'en-US': 'Filter' },
```

- [ ] **Step 3: Verify TypeScript compiles**

Run:
```bash
cd /Users/wilenwang/TalkAboutIt/frontend
npx tsc --noEmit
```

Expected: No type errors.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/ArchetypeFilter.tsx frontend/src/i18n/translations.ts
git commit -m "feat(ui): add ArchetypeFilter pill component"
```

---

## Task 8: Frontend — Refactor PersonaManagePage with Grid and Filter

**Files:**
- Modify: `frontend/src/pages/PersonaManagePage.tsx`

- [ ] **Step 1: Replace the entire PersonaManagePage implementation**

```tsx
// frontend/src/pages/PersonaManagePage.tsx
import { useState, useEffect, useCallback, useMemo } from 'react';
import { fetchPersonas, deletePersona } from '../api/client';
import type { PersonaSummary } from '../types';
import { useLanguage } from '../i18n/LanguageContext';
import PersonaEditor from '../components/PersonaEditor';
import PersonaCard from '../components/PersonaCard';
import ArchetypeFilter from '../components/ArchetypeFilter';

interface Props {
  onBack: () => void;
}

export default function PersonaManagePage({ onBack }: Props) {
  const { t } = useLanguage();
  const [personas, setPersonas] = useState<PersonaSummary[]>([]);
  const [loading, setLoading] = useState(true);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [isCreating, setIsCreating] = useState(false);
  const [filterArchetype, setFilterArchetype] = useState<string | null>(null);

  const load = useCallback(() => {
    setLoading(true);
    fetchPersonas()
      .then(setPersonas)
      .catch(console.error)
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  const filteredPersonas = useMemo(() => {
    if (!filterArchetype) return personas;
    return personas.filter((p) => p.archetype === filterArchetype);
  }, [personas, filterArchetype]);

  const handleDelete = async (id: string) => {
    if (!window.confirm(t('msgConfirmDelete'))) return;
    try {
      await deletePersona(id);
      load();
    } catch (e) {
      alert(t('errDeletePersona'));
    }
  };

  if (editingId || isCreating) {
    return (
      <PersonaEditor
        personaId={editingId}
        onSave={() => {
          setEditingId(null);
          setIsCreating(false);
          load();
        }}
        onCancel={() => {
          setEditingId(null);
          setIsCreating(false);
        }}
      />
    );
  }

  return (
    <div className="h-screen flex flex-col bg-white">
      {/* Header */}
      <header className="px-6 py-3 flex items-center gap-3 border-b border-black/[0.06]">
        <button
          onClick={onBack}
          className="text-[13px] text-[#615d59] hover:text-black/95 transition-colors"
        >
          ← {t('actionBack')}
        </button>
        <span className="text-lg font-bold tracking-tight">✦ TalkAboutIt</span>
        <span className="flex-1" />
        <button
          onClick={() => setIsCreating(true)}
          className="px-4 py-1.5 bg-[#0075de] text-white text-sm rounded-md hover:bg-[#0066c0] transition-colors"
        >
          {t('actionNewPersona')}
        </button>
      </header>

      {/* Content */}
      <main className="flex-1 overflow-y-auto px-6 py-8 max-w-[900px] mx-auto w-full">
        <h2 className="text-[22px] font-bold tracking-tight mb-6">{t('tabPersonas')}</h2>

        {/* Filter */}
        <div className="mb-6">
          <ArchetypeFilter active={filterArchetype} onChange={setFilterArchetype} />
        </div>

        {loading ? (
          <div className="text-sm text-[#a39e98]">{t('msgLoading')}</div>
        ) : filteredPersonas.length === 0 ? (
          <div className="text-center py-20">
            <div className="text-4xl mb-3">🔍</div>
            <div className="text-sm text-[#a39e98]">{t('msgNoFilterResults')}</div>
          </div>
        ) : (
          <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-3">
            {filteredPersonas.map((p) => (
              <PersonaCard
                key={p.id}
                persona={p}
                onEdit={setEditingId}
                onDelete={handleDelete}
              />
            ))}
          </div>
        )}
      </main>
    </div>
  );
}
```

- [ ] **Step 2: Verify TypeScript compiles**

Run:
```bash
cd /Users/wilenwang/TalkAboutIt/frontend
npx tsc --noEmit
```

Expected: No type errors.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/pages/PersonaManagePage.tsx
git commit -m "feat(personas): refactor manage page to square grid with archetype filter"
```

---

## Task 9: Frontend — Update PersonaEditor with Archetype Input

**Files:**
- Modify: `frontend/src/components/PersonaEditor.tsx`

- [ ] **Step 1: Add archetype text input in the Basic Info section**

In `frontend/src/components/PersonaEditor.tsx`, in the "Basic Info" section's grid (after the RoleTitle input), add:

```tsx
<TextInput
  label="Archetype"
  value={p.archetype}
  onChange={(v) => update('archetype', v)}
  placeholder="e.g. Visionary, Engineer, Philosopher"
/>
```

The grid currently has `grid-cols-2`, so this will flow naturally into the next row.

- [ ] **Step 2: Verify TypeScript compiles**

Run:
```bash
cd /Users/wilenwang/TalkAboutIt/frontend
npx tsc --noEmit
```

Expected: No type errors.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/PersonaEditor.tsx
git commit -m "feat(editor): add archetype input to PersonaEditor"
```

---

## Task 10: Frontend — Add Steve Jobs Avatar Image

**Files:**
- Create: `frontend/public/personas/steve-jobs.png`
- Modify: `backend/personas/steve-jobs.json`

- [ ] **Step 1: Copy the avatar image to public directory**

Source: `/Users/wilenwang/.claude/image-cache/cf7ed989-7de3-4b64-a1a2-9e479f6c9e17/1.png`

Run:
```bash
mkdir -p /Users/wilenwang/TalkAboutIt/frontend/public/personas
cp "/Users/wilenwang/.claude/image-cache/cf7ed989-7de3-4b64-a1a2-9e479f6c9e17/1.png" \
   "/Users/wilenwang/TalkAboutIt/frontend/public/personas/steve-jobs.png"
```

- [ ] **Step 2: Update Steve Jobs persona JSON to use image path**

In `backend/personas/steve-jobs.json`, change:
```json
"avatar": "🍎",
```
to:
```json
"avatar": "/personas/steve-jobs.png",
```

- [ ] **Step 3: Verify image is accessible**

Run the frontend dev server:
```bash
cd /Users/wilenwang/TalkAboutIt/frontend
npm run dev &
```

Check: `http://localhost:5173/personas/steve-jobs.png` should display the image.

- [ ] **Step 4: Commit**

```bash
git add frontend/public/personas/ backend/personas/steve-jobs.json
git commit -m "feat(assets): add Steve Jobs pixel avatar image"
```

---

## Task 11: Integration Test — Run Full Stack and Verify

**Files:**
- None (verification only)

- [ ] **Step 1: Start backend**

```bash
cd /Users/wilenwang/TalkAboutIt/backend
go run ./cmd/server &
```

- [ ] **Step 2: Start frontend**

```bash
cd /Users/wilenwang/TalkAboutIt/frontend
npm run dev
```

- [ ] **Step 3: Manual verification checklist**

Open `http://localhost:5173/personas` and verify:

1. [ ] Grid shows 5 square cards
2. [ ] Steve Jobs card shows pixel avatar image (not emoji)
3. [ ] Other cards show emoji avatars
4. [ ] Each card shows name + archetype label below
5. [ ] Filter pills show: 全部, 远见者, 工程师, 哲思者, 匠人, 运营者 (if UI is zh-CN)
6. [ ] Clicking a filter pill highlights it and filters the grid
7. [ ] Clicking "全部" clears the filter
8. [ ] Hovering a card shows description + 编辑/删除 buttons
9. [ ] Clicking "编辑" opens PersonaEditor with archetype field populated
10. [ ] Creating a new persona allows setting archetype
11. [ ] API returns archetype: `curl http://localhost:8080/api/v1/personas`

- [ ] **Step 4: Fix any issues found**

Address any visual or functional bugs discovered during verification.

- [ ] **Step 5: Final commit**

```bash
git add -A
git commit -m "feat(personas): complete square grid with archetype filter"
```

---

## Self-Review

### Spec Coverage

| Spec Requirement | Task |
|-----------------|------|
| Add archetype to Persona schema | Task 1 |
| Add archetype to API summary | Task 2 |
| Update 5 persona JSONs | Task 3 |
| Frontend type updates | Task 4 |
| i18n translations | Task 5 |
| Square card component | Task 6 |
| Archetype filter component | Task 7 |
| Grid + filter integration | Task 8 |
| Editor archetype input | Task 9 |
| Steve Jobs avatar image | Task 10 |
| Full stack verification | Task 11 |

### Placeholder Scan

- No TBD, TODO, or incomplete sections.
- All code blocks contain complete implementations.
- All commands have exact expected outputs.

### Type Consistency

- `Archetype` / `archetype` naming is consistent across Go struct tags, JSON, TypeScript interfaces, and component props.
- Translation keys (`archetypeVisionary`, etc.) match the archetype values (`Visionary`, etc.).
- `PersonaSummary` includes `archetype` in both frontend types and backend API response.
