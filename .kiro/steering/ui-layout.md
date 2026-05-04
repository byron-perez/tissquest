---
inclusion: auto
---

# UI Layout Rules

## Two Audiences, Two Modes

TissQuest has two distinct user modes. Every template must belong to one of them.

### Student Mode (`data-layout="student"`)
Pages: `index.html`, `atlas_view.html`, `tissue_record_detail.html`

- Purpose: browsing, learning, exploring microscopy content
- Max width: `max-w-5xl mx-auto px-6`
- Typography: larger, more breathing room, image-first
- Nav: `{{template "main-menu" .}}` at the top
- Breadcrumb: `{{template "breadcrumb" .}}` below nav
- No action buttons (New/Edit/Delete) visible to students
- Handler must pass `"Layout": "student"` in gin.H data

### Admin Mode (`data-layout="admin"`) — default
Pages: all `*_list.html`, `*_form.html`, `slide_form.html`

- Purpose: content management, data entry
- Max width: `max-w-5xl mx-auto px-6`
- Typography: compact, density matters
- Nav: `{{template "main-menu" .}}` at the top
- Breadcrumb: `{{template "breadcrumb" .}}` below nav
- Action buttons (New/Edit/Delete) present
- Handler does NOT need to pass Layout (defaults to "admin")

## Required Template Shell

Every full page template must follow this exact outer structure:

```html
{{define "content"}}
<div class="min-h-screen bg-base-200">
  <div class="max-w-5xl mx-auto px-6 py-8 space-y-6">

    {{template "main-menu" .}}
    {{template "breadcrumb" .}}

    <!-- page content here -->

  </div>
</div>
{{end}}
```

Deviations require explicit justification.

## Breadcrumb Data

Every handler must pass `Crumbs` as a slice of structs with `Label` and `URL` fields.
The last crumb has no URL (current page).

```go
"Crumbs": []breadcrumbItem{
    {Label: "Home", URL: "/"},
    {Label: "Atlases", URL: "/atlases"},
    {Label: atlasName}, // no URL = current page
},
```

## Spacing Tokens

| Use | Token |
|---|---|
| Page outer gap | `space-y-6` (admin) / `space-y-8` (student) |
| Section padding | `py-8 px-6` |
| Card body | `p-4` or `p-6` |
| Inline gaps | `gap-2` (tight) / `gap-4` (normal) / `gap-5` (cards) |

## Component Conventions

- Tables: `table table-zebra w-full bg-base-100 shadow rounded-box`
- Row IDs: `id="{entity}-row-{.ID}"` — required for HTMX targeting
- Delete cell IDs: `id="{entity}-row-{.ID}-delete"` — required for confirm-delete swap
- Forms: `card bg-base-100 shadow rounded-box p-6 mb-4`
- Primary action button: `btn btn-primary btn-sm`
- Destructive button: `btn btn-error btn-sm btn-outline`
- Ghost/secondary: `btn btn-ghost btn-sm`

## Handler Template Parse Lists

Every handler that renders a full page must include:
```go
[]string{
    "web/templates/layouts/base.html",
    "web/templates/pages/{page}.html",
    "web/templates/includes/main-menu.html",
    "web/templates/includes/breadcrumb.html",
}
```

Fragment-only handlers (HTMX partials) do NOT include base.html or main-menu.html.

## What NOT to Do

- Do not hardcode breadcrumb HTML in page templates — always use the partial
- Do not use `max-w-6xl` — the standard is `max-w-5xl`
- Do not use `{{template "main-menu" .}}` to render an atlas card grid — that template is navigation only
- Do not mix student and admin layout in the same page
