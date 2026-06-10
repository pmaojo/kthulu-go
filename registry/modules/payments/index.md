---
title: "Payments"
description: "PaymentsModule provides Stripe-style payment processing: plans, subscriptions, payments, and invoicing."
type: "module"
author: "Kthulu Core"
stars: 0
icon: "CreditCard"
---

# Payments

PaymentsModule provides a complete billing infrastructure inspired by Stripe: subscription plans, active subscriptions, payment records, and invoices.

## Features

- Subscription plan management (monthly, yearly, one-time)
- Stripe integration for subscriptions and payments
- Invoice generation and lifecycle tracking
- Clean Architecture structure
- Auto-configured Fx Module

## Configuration

The module is configured via environment variables:

| Variable | Description |
|----------|-------------|
| `STRIPE_SECRET_KEY` | Stripe secret API key |
| `STRIPE_WEBHOOK_SECRET` | Stripe webhook signing secret |

## Installation

Add this module to your project:

```bash
kthulu add module payments
```

## Components

This module provides the following components to the application:

- **Backend**: Go module wired with Uber Fx.

## Usage

```go
// Create a new subscription
sub, err := subscriptionService.Create(ctx, userID, planID)

// Record a payment
payment, err := paymentService.Record(ctx, userID, amountCents, currency, providerID)

// List invoices for a user
invoices, err := invoiceService.ListByUser(ctx, userID)
```

## Recipe

To scaffold the full payments stack (plans, subscriptions, payments, and invoices) in one command:

```bash
kthulu add recipe payments
```

Use with: `kthulu add recipe payments`
