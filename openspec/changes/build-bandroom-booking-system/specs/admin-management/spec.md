## Purpose

为唯一老板管理员提供集中管理会员、房间、设备、预约、通知、统计、导出和操作追踪的后台工作台。

## ADDED Requirements

### Requirement: Owner access
The system SHALL provide a separate owner administrator account with access to all venue management data and operations.

#### Scenario: Owner opens backend
- **WHEN** the owner logs in with valid administrator credentials
- **THEN** the system shows the management dashboard

### Requirement: Booking calendar and search
The system SHALL show all room bookings in a date-based calendar and SHALL support filtering by date, room, customer, status, name, or phone number.

#### Scenario: Owner reviews a booking
- **WHEN** the owner selects a booking in the calendar
- **THEN** the system shows customer, room, time, equipment, membership usage, notes, and status

### Requirement: Data export and statistics
The system SHALL provide basic booking, membership, room-use, equipment-use, cancellation, and no-show statistics and SHALL allow the owner to export relevant records as CSV.

#### Scenario: Export records
- **WHEN** the owner applies filters and requests an export
- **THEN** the system downloads a CSV containing the filtered records

### Requirement: Operation audit log
The system SHALL record owner actions affecting membership, rooms, equipment, bookings, content, and settings with actor, time, action, target, and relevant before-and-after values.

#### Scenario: Audit an owner change
- **WHEN** the owner changes a room status
- **THEN** the system records the change in the audit log
