# Migration status

The compatibility target is the Spring Boot service in `../smart-seaman-mobile-api`.

## Foundation

- [x] Fiber application bootstrap and graceful shutdown
- [x] MySQL connection pool configuration
- [x] CORS and security headers
- [x] Legacy response envelope
- [x] JWT authentication boundary
- [x] Route inventory
- [x] Message-code lookup and localized auth errors
- [ ] Static app-link files
- [x] Object storage client (S3-compatible Spaces)
- [x] SMTP email client
- [x] FCM and Omise clients
  - [x] FCM token update
- [ ] Scheduled notification refresh and user deletion jobs

## Domains

- [ ] Authentication and sessions
  - [x] Login, BCrypt verification, JWT issuance
  - [x] Refresh token and session token update
  - [x] Online-session enforcement for protected routes
  - [x] Activate user
  - [x] Change password
  - [x] Register with default certificates and confirmation email
  - [x] Reset-password email and forgot-password completion
  - [ ] Transaction log parity
- [x] Profile
  - [x] Read and update profile
  - [x] Profile image upload/download
  - [x] Deactivate/reactivate account
- [x] Master data and Thailand addresses
  - [x] Company and position master data
  - [x] Document and course dropdowns
  - [x] Province, district, and subdistrict lookup
- [x] Delivery addresses
  - [x] Default address read, create, and update
  - [x] Renewal request address snapshot
- [ ] Certificates and documents
  - [x] Certificate list, create, update, delete, edit, and file view
  - [x] Close-to-expiration list with display date and renewal-request flags
  - [ ] Request-item file upload and profile document endpoints
- [x] Document renewal workflow
  - [x] Active renewal statuses
  - [x] Renewal documents and active prices
  - [x] Request lifecycle, mobile update, delete/resubmit, timeline, stage, and payment transaction state
  - [x] Request validation
  - [x] Profile and renewal request-item upload/preview
- [x] Omise payments and webhook
  - [x] Payment transaction persistence and webhook state update
  - [x] Live Omise charge creation and signature verification
- [ ] Notifications and FCM
  - [x] FCM token update
  - [x] Notification list and read-state updates
- [x] Schools and courses
- [x] Forms
- [x] News, banners, and vouchers
  - [x] News list, detail, and image preview

An item is complete only after request validation, SQL behavior, response JSON, status codes, authorization, side effects, and focused compatibility tests match the Java implementation.

## Event logs

Before deploying the API, apply `migrations/001_create_event_logs.sql` to the application database. The API writes `http.request.completed` events with request/trace IDs, actor ID, route, status, and duration. Request/response bodies remain in structured application logs; Base64 files, multipart content, binary responses, and secrets are excluded from both logging paths.
