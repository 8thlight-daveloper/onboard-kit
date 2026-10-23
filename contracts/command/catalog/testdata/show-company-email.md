# Company email

ID: `company-email`
Carrier: `section-template`
Category: `identity-equipment-and-communication`
Goal: Day one
Result: Send and receive mail
Content: `sections/company-email.yaml`

## Content

```yaml
schemaVersion: 1
id: company-email
title: Company email
summary: Send and receive mail
timing: week-one
variables:
  - {name: email_process, description: "Which approved mail service and setup instructions apply?", audience: onboarder, required: true}
  - {name: email_test_route, description: "Which business contact or safe test procedure should verify sending and receiving mail?", audience: onboarder, required: true}
body: |
  Set up the approved mail service using $email_process.
  Use the safe test procedure or contact: $email_test_route

  Confirm the mailbox is your intended company mailbox. Send a nonsecret
  test message and receive a reply or equivalent approved test. Check how
  to reach support if mail is unavailable, rather than depending on that
  mailbox to obtain its own access. Never forward business mail to a
  personal account as an onboarding workaround.
steps:
  - id: exchange-mail
    name: Verify sending and receiving
    checks:
      - ask: Have you sent and received a nonsecret test message using the intended company mailbox?
```
