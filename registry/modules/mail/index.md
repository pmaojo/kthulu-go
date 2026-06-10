---
title: "Mail"
description: "Generated mail infrastructure: SMTP and log drivers out of the box, env-configured, injected via Fx."
type: "module"
author: "Kthulu Core"
stars: 0
icon: "Box"
---

# Mail

Add `mail` to your project features and Kthulu generates `internal/infrastructure/mail/` — a `Mailer` interface with working **SMTP** and **log** drivers (the log driver is the development default), plus a template mailer for HTML emails. The mailer is provided through Fx, so any service can inject it.

```bash
kthulu create my-app --features=auth,user,mail
```

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
