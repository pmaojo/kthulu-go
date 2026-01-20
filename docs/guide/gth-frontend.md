# GTH Frontend Guide

GTH (Go + Templ + HTMX) is Kthulu's hypermedia-driven frontend stack.

## Overview

| Component | Purpose                                          |
| --------- | ------------------------------------------------ |
| **Go**    | Server-side rendering and business logic         |
| **Templ** | Type-safe HTML templates that compile to Go      |
| **HTMX**  | Dynamic UI updates without JavaScript frameworks |

## Installation Requirements

```bash
# Install templ CLI
go install github.com/a-h/templ/cmd/templ@latest
```

## Project Structure

When you create a Kthulu project, the GTH frontend is generated in `internal/views/`:

```
internal/views/
├── layouts/
│   ├── base.templ        # Base HTML with HTMX, CSS design system
│   └── admin.templ       # Admin layout with sidebar navigation
├── components/
│   ├── modal.templ       # Reusable modal dialog
│   ├── <module>_table.templ   # Data table with sorting
│   └── <module>_form.templ    # CRUD form with validation
├── pages/
│   ├── dashboard.templ   # Dashboard with stats
│   └── <module>_page.templ    # Full CRUD page
└── partials/
    └── <module>_table_rows.templ  # HTMX partial for table updates
```

## How HTMX Works

Instead of fetching JSON and transforming it with JavaScript, HTMX returns HTML:

```html
<!-- Click triggers server request -->
<button hx-get="/products/new" hx-target="#modal-container">Add Product</button>

<!-- Server returns HTML that swaps into #modal-container -->
```

### Key HTMX Attributes

| Attribute      | Purpose                                 |
| -------------- | --------------------------------------- |
| `hx-get`       | Make GET request on click               |
| `hx-post`      | Submit form via POST                    |
| `hx-target`    | Where to swap response HTML             |
| `hx-swap`      | How to swap (innerHTML, outerHTML, etc) |
| `hx-trigger`   | When to fire (click, input, etc)        |
| `hx-indicator` | Loading indicator element               |

## Compiling Templates

Templ files (`.templ`) compile to Go code (`.templ.go`):

```bash
# Generate Go code from .templ files
templ generate

# Or use go generate with directive
go generate ./...
```

## Creating New Views

Use `kthulu add module` to generate views for a new module:

```bash
kthulu add module invoice name:string amount:float status:string
```

This generates:

- `internal/views/components/invoice_table.templ`
- `internal/views/components/invoice_form.templ`
- `internal/views/pages/invoice_page.templ`
- `internal/views/partials/invoice_table_rows.templ`

## Handler Integration

GTH handlers detect HTMX requests and return partials vs full pages:

```go
func (h *ProductHandler) ListPage(w http.ResponseWriter, r *http.Request) {
    products, _ := h.service.ListProducts()

    // HTMX request = return partial
    if r.Header.Get("HX-Request") == "true" {
        partials.ProductTableRows(products).Render(r.Context(), w)
        return
    }

    // Full page render
    pages.ProductsPage(products).Render(r.Context(), w)
}
```

## Disabling Frontend

For API-only projects:

```bash
kthulu new my-api --frontend none
```

## Resources

- [HTMX Documentation](https://htmx.org/docs/)
- [Templ Documentation](https://templ.guide/)
- [Go HTML/Template](https://pkg.go.dev/html/template)
