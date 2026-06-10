---
title: "Mail"
description: "MailModule provides email sending functionality. Supports multiple providers: SMTP, SES, SendGrid, Mailgun, Resend"
type: "module"
author: "Kthulu Core"
stars: 0
icon: "Box"
---

# Mail

MailModule provides email sending functionality. Supports multiple providers: SMTP, SES, SendGrid, Mailgun, Resend

## Features


- Auto-configured Fx Module
- Clean Architecture structure



## Configuration

The module is configured via environment variables:

| Variable | Description |
|----------|-------------|
| - | No environment variables detected |


## Installation

Add this module to your project:

```bash
kthulu add module mail
```

## Components

This module provides the following components to the application:

- **Backend**: Go module wired with Uber Fx.


## Usage

```go
func NewSignupService(mailer mail.Mailer) *SignupService { ... }

mailer.Send(ctx, &mail.Message{
    To:      []string{"user@example.com"},
    Subject: "Welcome!",
    HTML:    "<h1>Hello</h1>",
})
```

HTML templates:

```go
tm, _ := mail.NewTemplateMailer(mailer, "templates/mail")
tm.SendTemplate(ctx, []string{"user@example.com"}, "Welcome", "welcome.html", data)
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `MAIL_DRIVER` | `log` | `smtp`, `log` (SES/SendGrid/Mailgun/Resend are scaffolded stubs) |
| `MAIL_HOST` / `MAIL_PORT` | — | SMTP server |
| `MAIL_USERNAME` / `MAIL_PASSWORD` | — | SMTP credentials |
| `MAIL_FROM_ADDRESS` / `MAIL_FROM_NAME` | — | Sender identity |

The provider-SDK drivers (SES, SendGrid, Mailgun, Resend) are generated as typed stubs with clear errors pointing at the SDK dependency to add — drop in the SDK and fill in one method.







