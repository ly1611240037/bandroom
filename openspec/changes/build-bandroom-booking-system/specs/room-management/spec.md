## Purpose

为门店提供多间独立排练房的资料、固定设备、照片、营业安排和可预约状态管理。

## ADDED Requirements

### Requirement: Room information
The system SHALL allow the owner to manage each room's name, description, capacity information, photos, fixed equipment list, and availability status.

#### Scenario: Customer views a room
- **WHEN** a customer opens a room detail page
- **THEN** the system shows the room information, photos, fixed equipment, and current availability status

### Requirement: Room availability status
The system SHALL prevent new customer bookings for rooms marked under maintenance or suspended while preserving their existing booking history.

#### Scenario: Room is suspended
- **WHEN** the owner marks a room as unavailable
- **THEN** the room cannot be selected for a new booking

#### Scenario: Fixed equipment is unavailable
- **WHEN** the owner marks a room's fixed equipment as under maintenance
- **THEN** the room remains bookable and the unavailable equipment is shown to customers

### Requirement: Business schedule
The system SHALL support weekly opening hours, special closed dates, partial-day closures, and a configurable cleaning buffer with defaults of 08:00–23:00 and 30 minutes.

#### Scenario: Special closure
- **WHEN** a date or room is configured as closed
- **THEN** affected times do not appear as bookable

