---
description: "Instructions for Gin HTTP handlers in tissquest - routing patterns, response conventions, error handling"
applyTo: ["cmd/api-server-gin/**/*.go"]
---

# Handler Instructions

## Gin Routing Patterns
- Use RESTful conventions: GET for retrieval, POST for creation
- Route parameters: `/atlas/:id` for single resource access
- Query parameters for pagination: `?page=1&limit=10`

## Response Conventions
- JSON responses: Use `c.JSON(statusCode, data)` for API endpoints
- HTML responses: Use `c.HTML(statusCode, templateName, data)` for web pages
- Consistent status codes: 200 OK, 201 Created, 400 Bad Request, 404 Not Found, 500 Internal Server Error

## Error Handling
- Check for errors from service/repository calls
- Return appropriate HTTP status codes with error messages
- Avoid panicking; use proper error responses

## Template Rendering
- Templates loaded from `./web/templates` with base layout
- Use relative paths for template names (e.g., "includes/atlas_view.html")
- Pass data structs to templates for rendering