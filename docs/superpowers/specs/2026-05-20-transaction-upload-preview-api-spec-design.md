# Transaction Upload Preview API Spec File Design

## Summary

Add a new request-spec YAML file inside `paris-api/transactions/` for the new preview endpoint:

- `GET /api/v1/transaction-uploads/:id/preview`

The new file should mirror the existing request collection pattern already used under `paris-api/transactions/`.

## Why

The repository already keeps manual API request specs under `paris-api/transactions/` as one YAML file per endpoint. Now that the preview endpoint exists in the application, the request collection should include a matching entry so developers can discover and exercise it the same way they do the other transaction and transaction-upload endpoints.

## Existing pattern observed

The existing files under `paris-api/transactions/` use a lightweight request-spec format with this structure:

- `info.name`
- `info.type: http`
- `info.seq`
- `http.method`
- `http.url`
- `http.headers`
- `http.body`
- `http.auth: inherit`
- `settings`

Examples include:

- `paris-api/transactions/upload-transaction.yml`
- `paris-api/transactions/get-transaction-upload.yml`
- `paris-api/transactions/retry-transaction-upload-classification.yml`

This indicates the correct approach is to add a new YAML request file rather than introducing OpenAPI, Markdown reference docs, or a different spec format.

## Recommended approach

Create a new file:

- `paris-api/transactions/get-transaction-upload-preview.yml`

This matches the naming style already used in the folder and keeps the preview request colocated with the other transaction-upload request specs.

## File contents

The new YAML file should follow the same structure as the existing request specs.

Recommended shape:

```yml
info:
  name: get-transaction-upload-preview
  type: http
  seq: <next available sequence>

http:
  method: GET
  url: http://localhost:9000/api/v1/transaction-uploads/<sample-upload-id>/preview
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

## Specific decisions

### Endpoint location

Use the existing `paris-api/transactions/` folder rather than creating a new top-level folder.

Reason:

- that folder already contains transaction-upload-related requests,
- preview is part of the transaction-upload API surface,
- keeping it there matches current developer expectations.

### Request naming

Use:

- `get-transaction-upload-preview`

Reason:

- it matches the current naming pattern,
- it is consistent with `get-transaction-upload.yml`,
- it makes the relationship between the two endpoints obvious.

### Request body

Use:

- `body.type: none`

Reason:

- the endpoint is a `GET`,
- there is no request payload,
- this is the clearest explicit representation for this collection format.

### Sequence number

Choose the next unused sequence number in the folder rather than trying to renumber existing files.

Current files already contain duplicate and non-sequential `seq` values, so this change should stay additive and minimal.

## Scope

In scope:

- add the new request-spec YAML file for the preview endpoint.

Out of scope:

- changing existing request specs,
- renumbering existing `seq` values,
- converting the folder to OpenAPI,
- adding response examples unless the existing request-spec pattern already supports them elsewhere,
- changing `folder.yml` unless the request tool requires explicit folder metadata updates.

## Testing / verification

Verification should be lightweight:

1. confirm the new file matches the existing YAML structure,
2. confirm the URL points to `/api/v1/transaction-uploads/:id/preview`,
3. confirm actor headers match the established examples in the same folder,
4. confirm the request-spec file appears alongside the existing `transactions` request specs.

## Implementation notes

- Prefer copying the style of `get-transaction-upload.yml` and adjusting only the fields needed for the preview endpoint.
- Keep this as a single-file additive change unless `folder.yml` proves to require updates.
- Do not introduce a new documentation system for this request.
