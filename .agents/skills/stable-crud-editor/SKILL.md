---
name: stable-crud-editor
description: Keep Next.js left-list/right-detail and inline-editor CRUD screens visually stable while switching records. Use when a MongoDB-backed detail request causes the detail pane or editor to blank, unmount, remount, flicker, show projection-row data, or accept a stale response after quick record clicks.
---

# Stable CRUD Editor

Use this for the shared BC Ai Account CRUD workbench before adding a screen-specific workaround.

## Workflow

1. Read the real row click/edit handler, detail loader, form conditional, and close/navigation paths.
2. Keep the current editor mounted while the next MongoDB detail request is pending. Do not set the form visibility false before `await`.
3. Increment one request token for every record-open action, including non-hydrated paths. Apply async success, error, and loading completion only when its token is still latest.
4. Preserve the unsaved-change confirmation. If the new detail fails, keep the prior editor visible and show the explicit API error; never substitute list-row data as editable truth.
5. Invalidate the same token when closing the editor, leaving the screen, or changing active Holding/Company so a late response cannot reopen or cross tenant context.

For a read-only detail pane, retain the previous MongoDB-hydrated record until the next detail succeeds. Mirror it in a ref when async error/race handlers need the latest hydrated record; never substitute the selected projection row or the first list row. Clear it only when the selection becomes empty or the validated Holding/Company changes. Do not gate the pane on a loading flag set inside `useEffect`, because the selection render occurs before that effect and creates one blank frame. Use `aria-busy` and guard actions without changing the pane's appearance; do not insert a spinner, overlay, opacity change, or disabled-button styling between records. Atomically replace the pane after the latest request succeeds. On failure, restore the previous row selection and show the source API error so retained data is never silent or misleading.

For an editor that receives a hydrated record prop but keeps an editable local draft:

- Remove GUID-mismatch early returns that replace the editor with loading or empty cards.
- Retain the previous hydrated record while the next record is pending and make the editor `inert` without changing its appearance.
- Sync the new record into the local draft with `useLayoutEffect`, not `useEffect`, so the new header and old draft cannot paint together for one frame.
- When no record has loaded yet, render `null`; expose instructions through the screen's `วิธีใช้` button instead of a transient pane.
- On detail failure, keep the previous hydrated editor, restore its selection, and surface the API error explicitly.

## Required Pattern

```tsx
const editorRequestRef = useRef(0);

async function openEdit(record: RecordData) {
  const request = ++editorRequestRef.current;
  const hadOpenEditor = formOpen;

  if (!hadOpenEditor) setFormOpen(false);
  setDetailLoading(true);

  try {
    const detail = await loadMongoDetail(record);
    if (request !== editorRequestRef.current) return;
    setEditing(detail);
    setForm(formFromRecord(detail));
    setFormOpen(true);
  } catch (error) {
    if (request !== editorRequestRef.current) return;
    showApiError(error);
  } finally {
    if (request === editorRequestRef.current) setDetailLoading(false);
  }
}
```

Do not add a changing `key` to the editor component. A per-record key forces remounting and recreates the flicker.

## Verification

- Switch two records repeatedly, then click several records quickly; the last click must win.
- Delay the detail request 450–500 ms and sample `requestAnimationFrame` for at least 300 ms. Assert both detail card and content remain mounted with non-zero height and no loading text appears.
- Delay A, click A then B after about 25 ms, and assert both selected row and hydrated detail remain B after A finishes.
- Close while a request is pending; a late response must not reopen the editor.
- Change Holding/Company while a request is pending; retained detail must clear and the late response must be ignored.
- Verify a failed detail request surfaces its source error and never turns the list row into editable fallback data.
- Run the frontend typecheck/build required by the current runtime mode and test the real screen without creating or cleaning unrelated DEV data.
