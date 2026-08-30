## Purpose

向顾客和老板及时反馈预约、取消、拒绝、会员到期和故障等重要事件，并保留可查询的通知记录。

## ADDED Requirements

### Requirement: In-app notifications
The system SHALL create website notifications for relevant booking and membership events and SHALL allow the recipient to view unread and read notifications.

#### Scenario: New booking notification
- **WHEN** a customer booking succeeds
- **THEN** the customer and owner receive corresponding in-app notifications

### Requirement: Email notifications
The system SHALL send email notifications for booking success and cancellation, owner booking alerts, and monthly-card expiry reminders three days before expiry.

#### Scenario: Booking email
- **WHEN** a booking is created or cancelled
- **THEN** the system sends the customer an email containing the booking details and new status

#### Scenario: Email delivery failure
- **WHEN** an email provider cannot deliver a notification
- **THEN** the booking state remains unchanged and the failure is recorded for owner review
