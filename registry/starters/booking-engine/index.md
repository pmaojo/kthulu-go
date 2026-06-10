---
title: "Booking Engine"
description: "Reservations for venues, rooms or courts: availability, bookings and reminders."
type: "starter"
author: "Kthulu Team"
stars: 27
icon: "Calendar"
---

# Booking Engine

Reservations for venues, rooms or courts: availability, bookings and reminders.

## Highlights

- Venues, guests and time-slot bookings with relational integrity
- Booking lifecycle enforced with status enums
- Queue + scheduler runtime for reminder emails
- Mail driver wired for confirmations

## Domain Model

| Entity | Fields |
|--------|--------|
| `venue` | name, address, capacity |
| `guest` | name, email, phone |
| `booking` | starts_at, ends_at, status, party_size, place, holder |

Every entity ships with a typed GORM model, repository, service with
validation, REST API, and a GTH (Templ + HTMX) admin page.

## Get Started

```bash
# Copy the plan from this page into my-app-plan.yaml, then:
kthulu create my-app --from-plan=my-app-plan.yaml
cd my-app && go run ./cmd/server
```

The full plan file ships with the framework at
`registry/starters/booking-engine/plan.yaml`.
