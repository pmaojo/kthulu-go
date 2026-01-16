---
name: Database Migration Helper
description: Expert guidance on creating and running database migrations in Kthulu
---

# Database Migration Helper

To create a migration in Kthulu:

1. Use the `kthulu migrate create [name]` command.
2. Edit the generated SQL files in `db/migrations`.
3. Run `kthulu migrate up` to apply.

Example:

```bash
kthulu migrate create add_users_table
```
