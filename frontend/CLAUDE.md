# Frontend Guidance

## Smart Save Strategy (SnippetEditor.tsx)

**Two save modes:**
1. **Immediate save** - Tags and favorites (`saveTagsAndFavorite()`)
   - Triggers on tag/favorite change
   - No confirmation needed
2. **Explicit save** - Title, body, language (Save button)
   - Shows "수정됨" (Modified) indicator when dirty
   - Confirmation dialog on navigation if unsaved

**Dirty tracking:**
- Only title, body, language tracked (not tags/favorites)
- Prevents accidental data loss

## Component Size Guidelines

- Keep React components under 200-250 lines
- If a component exceeds this, consider extracting:
  - Custom hooks for reusable logic
  - Sub-components for distinct UI sections
  - Constants/types to separate files
- Signs a component needs refactoring:
  - More than 10 useState hooks
  - Multiple unrelated useEffect hooks
  - Long JSX with repeated patterns
