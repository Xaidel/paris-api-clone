# Transaction Upload Preview API Spec Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a new request-spec YAML file under `paris-api/transactions/` for `GET /api/v1/transaction-uploads/:id/preview` that matches the existing collection format.

**Architecture:** Reuse the existing `paris-api/transactions/*.yml` request-spec pattern rather than introducing a new API documentation format. Keep the change additive: create one new endpoint file that mirrors the shape, headers, auth, and settings used by the existing transaction-upload request specs.

**Tech Stack:** YAML request-spec files, existing `paris-api/transactions` collection format

---

## File structure

### Create
- `paris-api/transactions/get-transaction-upload-preview.yml` — request-spec file for the preview endpoint.

### Modify
- None by default.
- `paris-api/transactions/folder.yml` only if verification shows the request collection requires explicit folder metadata updates for new files.

---

### Task 1: Add preview endpoint request-spec file

**Files:**
- Create: `paris-api/transactions/get-transaction-upload-preview.yml`
- Reference: `paris-api/transactions/get-transaction-upload.yml`
- Reference: `paris-api/transactions/upload-transaction.yml`

- [ ] **Step 1: Inspect the existing request-spec shape before editing**

Read the existing files and compare the fields they use:

```yml
info:
  name: get-transaction-upload
  type: http
  seq: 4

http:
  method: GET
  url: http://localhost:9000/api/v1/transaction-uploads
  headers:
    - name: X-Actor-User-ID
      value: 01962b8f-aeb2-7e03-a8ff-1edce1300002
    - name: X-Actor-Group-ID
      value: 01962b8f-aeb2-7e03-a8ff-1edce1300001
  body:
    type: multipart-form
  auth: inherit

settings:
  encodeUrl: true
  timeout: 0
  followRedirects: true
  maxRedirects: 5
```

Use the same field ordering and indentation style for the new file.

- [ ] **Step 2: Create the new request-spec file with the preview endpoint**

Create `paris-api/transactions/get-transaction-upload-preview.yml` with this content:

```yml
info:
  name: get-transaction-upload-preview
  type: http
  seq: 9

http:
  method: GET
  url: http://localhost:9000/api/v1/transaction-uploads/01962b8f-aeb2-7e03-a8ff-1edce1300201/preview
  headers:
    - name: X-Actor-User-ID
      value: 01962b8f-aeb2-7e03-a8ff-1edce1300002
    - name: X-Actor-Group-ID
      value: 01962b8f-aeb2-7e03-a8ff-1edce1300001
  body:
    type: none
  auth: inherit

settings:
  encodeUrl: true
  timeout: 0
  followRedirects: true
  maxRedirects: 5
```

Notes:
- `seq: 9` is the next unused sequence value after the highest observed existing value (`8`).
- The sample upload ID should remain concrete, matching the style used elsewhere in the request collection.
- `body.type: none` makes the GET-without-body shape explicit.

- [ ] **Step 3: Verify the new file is structurally consistent with the folder pattern**

Check that the new file includes all required top-level blocks:

```text
info
http
settings
```

Check that the request-specific fields are present:

```text
http.method = GET
http.url ends with /api/v1/transaction-uploads/<id>/preview
http.headers contains both X-Actor-User-ID and X-Actor-Group-ID
http.body.type = none
http.auth = inherit
```

- [ ] **Step 4: Verify folder metadata does not require changes**

Read `paris-api/transactions/folder.yml` and confirm it is only:

```yml
info:
  name: transactions
  type: folder
  seq: 3

request:
  auth: inherit
```

Expected: no change needed. If the request collection tool auto-discovers files in the folder, leave `folder.yml` untouched.

- [ ] **Step 5: Verify the working tree only contains the intended request-spec addition**

Run:

```powershell
git status --short
```

Expected: the new file appears as an added file under `paris-api/transactions/`, with no unrelated request-spec changes required for this task.

---

## Self-review checklist

- Spec coverage: the plan adds the new request-spec file in `paris-api/transactions/` and keeps the change additive and format-consistent.
- Placeholder scan: no `TODO`, `TBD`, or vague implementation steps remain.
- Type consistency: file naming, URL, headers, and YAML block structure match the observed request-spec pattern.
