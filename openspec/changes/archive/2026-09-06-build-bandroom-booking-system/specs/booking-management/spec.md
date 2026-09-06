## Purpose

为会员提供多房间、半小时粒度的自助预约，并自动处理房间、设备、会员资格和清场时间冲突。

## ADDED Requirements

### Requirement: Booking eligibility
The system SHALL allow a verified customer to book only when they have eligible membership, no other future booking, and a selected date that is not beyond the active monthly card's expiry when a monthly card is used.

#### Scenario: Eligible customer books
- **WHEN** a verified customer with valid membership submits an otherwise valid booking
- **THEN** the system creates the booking immediately with a successful status

#### Scenario: Customer already has a future booking
- **WHEN** a customer with one future booking submits another booking
- **THEN** the system rejects the new booking

### Requirement: Booking time rules
The system SHALL accept start times and durations in 30-minute increments, SHALL limit each booking to 0.5 through 3 hours, SHALL enforce weekly and special opening hours, and SHALL reserve an additional 30-minute cleaning buffer without charging for it.

#### Scenario: Valid time selection
- **WHEN** a customer selects a half-hour-aligned interval within opening hours and no conflict exists
- **THEN** the interval can be submitted for booking

#### Scenario: Cleaning buffer conflict
- **WHEN** a new booking begins before the previous booking's end plus 30 minutes
- **THEN** the system rejects the new booking as conflicting

### Requirement: Booking and equipment conflict prevention
The system SHALL atomically check room availability and public equipment inventory when creating a booking so overlapping successful bookings cannot exceed room or equipment capacity.

#### Scenario: Room conflict
- **WHEN** another booking occupies the selected room or its cleaning buffer
- **THEN** the system rejects the booking

### Requirement: Customer cancellation
The system SHALL allow a customer to cancel their booking at any time before completion, release the room and public equipment immediately, and restore one consumed count-card use when applicable.

#### Scenario: Cancellation
- **WHEN** a customer cancels an active future booking
- **THEN** the booking becomes cancelled, resources are released, and a cancellation notification is sent

### Requirement: Booking lifecycle and owner override
The system SHALL automatically mark a booking as completed after its end time plus the cleaning buffer, allow the owner to mark a booking as a no-show, and allow the owner to edit, cancel, or override booking restrictions for exceptional cases. The v1 system SHALL create customer bookings only through the customer booking flow.

#### Scenario: Automatic completion
- **WHEN** the booking end time plus cleaning buffer has passed
- **THEN** the booking status becomes completed

#### Scenario: No-show
- **WHEN** the owner marks a booking as a no-show
- **THEN** the system records the no-show without deducting a monthly-card quota or imposing an automatic future restriction
