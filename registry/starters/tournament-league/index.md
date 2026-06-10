---
title: "Tournament League"
description: "Complete tournament management: teams, players, fixtures and standings with background jobs."
type: "starter"
author: "Kthulu Team"
stars: 31
icon: "Trophy"
---

# Tournament League

Complete tournament management: teams, players, fixtures and standings with background jobs.

## Highlights

- Full bracket domain: tournaments, teams, players and matches with typed relations
- Validation everywhere: jersey numbers 1-99, status enums, email checks
- Background jobs runtime for fixture scheduling and notifications
- Admin UI with live HTMX search for every entity

## Domain Model

| Entity | Fields |
|--------|--------|
| `tournament` | title, starts_at, max_teams, status |
| `team` | name, city, wins, losses |
| `player` | name, email, number, squad |
| `match` | played_at, home_score, away_score, home, away |

Every entity ships with a typed GORM model, repository, service with
validation, REST API, and a GTH (Templ + HTMX) admin page.

## Get Started

One command — the blueprint ships inside the kthulu binary:

```bash
kthulu marketplace install tournament-league my-app
cd my-app && go run ./cmd/server
```

Or scaffold manually from the plan file at
`registry/starters/tournament-league/plan.yaml`:

```bash
kthulu create my-app --from-plan=registry/starters/tournament-league/plan.yaml
```
