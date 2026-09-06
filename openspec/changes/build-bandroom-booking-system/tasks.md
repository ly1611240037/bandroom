## 1. Project foundation

- [x]  1.1 Create the React frontend and Go backend directory structure, and verify both development entry points start successfully.
- [x]  1.2 Configure environment variables for SQLite, sessions, mail delivery, and application URLs, and verify the app starts with a documented example configuration.
- [x]  1.3 Add shared error response, request validation, time-zone, and half-hour time utilities, and verify unit tests cover valid and invalid inputs.

## 2. Database and seed data

- [x]  2.1 Create versioned SQLite migrations for users, verification tokens, password resets, membership plans, membership cards, rooms, fixed equipment, public equipment, schedules, closures, bookings, booking equipment, notifications, issue reports, venue content, and audit logs; verify migrations apply to a fresh database.
- [x]  2.2 Implement data-access repositories and transaction boundaries, and verify basic create/read/update operations with integration tests.
- [x]  2.3 Add repeatable demo-data initialization with owner, customer, rooms, equipment, membership cards, and sample bookings; verify a fresh demo database contains all documented accounts and examples.

## 3. Authentication and authorization

- [x]  3.1 Implement customer registration, password hashing, email verification, login, logout, and session handling; verify unverified customers cannot create bookings.
- [x]  3.2 Implement password-reset request and one-time expiring reset links; verify a used or expired link cannot reset a password again.
- [x]  3.3 Implement owner authentication and role checks for all backend management operations; verify a customer cannot access owner endpoints or another customer's data.
- [x]  3.4 Build customer account pages for profile, verification state, and password recovery; verify the complete customer auth flow in the browser.

## 4. Membership cards

- [x]  4.1 Implement owner management of monthly and count-based plans, including enable/disable and offline payment records; verify plan changes appear in the owner backend.
- [x]  4.2 Implement manual card activation, 30-day monthly validity, sequential future monthly cards, and unlimited count-card validity; verify date and remaining-count calculations.
- [x]  4.3 Implement monthly-card priority, count deduction, cancellation restoration, and eligibility checks; verify unit and integration tests cover both card types and exhausted counts.
- [x]  4.4 Build owner membership pages and customer membership status pages; verify card status, expiry, remaining counts, and purchase details are visible to the correct user.

## 5. Rooms, schedules, and equipment

- [x]  5.1 Implement room CRUD, photos, fixed equipment, capacity display, and room status; verify maintenance and suspended rooms cannot be selected for new bookings.
- [x]  5.2 Implement weekly schedules, special closed dates, partial-day closures, and configurable cleaning buffer with defaults of 08:00–23:00 and 30 minutes; verify affected slots are unavailable.
- [x]  5.3 Implement public equipment CRUD, quantities, statuses, and availability calculation; verify unavailable equipment cannot be selected.
- [x]  5.4 Implement customer room and equipment browsing, including room photos, fixed equipment, and remaining public-equipment quantities; verify unauthenticated browsing works.
- [x]  5.5 Implement customer issue reports and owner review/status updates; verify a submitted report appears in the owner backend and can mark equipment under maintenance.

## 6. Booking engine

- [x]  6.1 Implement availability calculation from opening hours, closures, room status, existing bookings, and 30-minute cleaning buffers; verify 0.5–3 hour selections and boundary times.
- [x]  6.2 Implement transactional booking creation with membership, one-future-booking, monthly-expiry, room-conflict, and equipment-inventory checks; verify conflicting concurrent requests cannot overbook resources.
- [x]  6.3 Implement booking records with customer, band name, contact details, room, interval, equipment quantities, notes, membership usage, and status; verify successful booking data is complete.
- [x]  6.4 Implement customer booking list/detail pages and date-first booking flow; verify a customer can browse a date, choose an available room and slot, select equipment, and book immediately.
- [x]  6.5 Implement cancellation, immediate resource release, count restoration, cancellation notifications, and the 24-hour booking rule; verify cancellation and rebooking behavior.
- [x]  6.6 Implement automatic completion after end time plus cleaning buffer, owner no-show marking, and owner edit/cancel/exception handling without adding owner-created bookings in v1; verify lifecycle transitions and audit entries.

## 7. Notifications and scheduled processing

- [x]  7.1 Implement in-app notification creation, unread/read state, and notification list pages; verify customer and owner receive the appropriate booking events.
- [x]  7.2 Implement an email adapter with test-mail configuration and templates for verification, password reset, booking success, cancellation, owner alerts, and expiry reminders; verify emails render with booking details.
- [x]  7.3 Implement reliable scheduled processing for completion, monthly-card reminders three days before expiry, and notification retries/failures; verify a mail failure does not change booking state.

## 8. Owner backend

- [x]  8.1 Build the owner dashboard and date-based all-room booking calendar; verify booking details show customer, room, time, equipment, membership usage, notes, and status.
- [x]  8.2 Build owner pages for rooms, fixed equipment, public equipment, schedules, closures, issue reports, memberships, notifications, and venue content; verify each page supports the documented operations.
- [x]  8.3 Add booking/member search and filters plus CSV exports; verify filtered exports contain the expected records and UTF-8 Chinese text.
- [x]  8.4 Add dashboard statistics and audit-log viewing; verify statistics reflect booking, membership, equipment, cancellation, and no-show data.

## 9. Venue website and content

- [ ]  9.1 Build the public homepage with BandRoom 音乐排练空间 branding, venue information, membership instructions, room highlights, opening rules, and booking entry point; verify it works without login.
- [ ]  9.2 Build responsive customer layouts for desktop and mobile browsers with the agreed dark music-themed visual style; verify core booking actions are usable at both viewport sizes.
- [ ]  9.3 Build owner editing for venue information, contact details, announcements, membership instructions, opening rules, and room photos; verify saved content appears on the public site.

## 10. Integration, packaging, and project materials

- [ ]  10.1 Add end-to-end tests for registration, card activation, valid booking, room conflict, equipment conflict, cancellation, count restoration, automatic completion, and owner management; verify the full test suite passes on a fresh SQLite database.
- [ ]  10.2 Build the React production bundle and serve it from Go alongside the API; verify the complete demo works from one local URL without a separate frontend server.
- [ ]  10.3 Add startup instructions, demo credentials, test-mail instructions, backup/restore guidance, and sample demonstration scenarios; verify another person can run the project from the documentation.
- [ ]  10.4 Prepare project overview, architecture diagram, database description, user manual, screenshots, test report, and feature demonstration script; verify the materials match the implemented v1 behavior.
