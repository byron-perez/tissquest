---
description: "Instructions for domain entities in tissquest - validation patterns, business methods, repository contracts"
applyTo: ["internal/core/**/*.go"]
---

# Domain Model Instructions

## Entity Structure
- Define structs with JSON tags for serialization
- Include ID, timestamps (CreatedAt, UpdatedAt)
- Use uint for IDs, time.Time for dates

## Validation
- Implement `Validate() error` method on entities
- Check required fields, length limits, format constraints
- Return descriptive error messages

## Business Logic
- Add methods for entity manipulation (e.g., AddTissueRecord, RemoveTissueRecord)
- Keep business rules in domain layer
- Avoid database concerns

## Repository Interfaces
- Define RepositoryInterface with CRUD operations
- Include domain-specific queries (FindByName, ListWithPagination)
- Use dependency injection in services