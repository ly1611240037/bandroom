## Purpose

为排练房固定设备和门店公共设备提供清单、状态、库存、预约占用和故障反馈管理。

## ADDED Requirements

### Requirement: Public equipment inventory
The system SHALL allow the owner to manage public equipment names, descriptions, quantities, statuses, and hourly rental prices, while the customer experience SHALL treat the rental price as zero for v1.

#### Scenario: Equipment is available
- **WHEN** a public equipment item is available with positive inventory
- **THEN** the customer can select a quantity up to the remaining inventory

#### Scenario: Equipment is unavailable
- **WHEN** a public equipment item is marked under maintenance or disabled
- **THEN** the customer cannot select it

### Requirement: Equipment reservation inventory
The system SHALL reserve selected public equipment quantities for the booking's occupied interval, including its cleaning buffer, and SHALL release them when the booking is cancelled.

#### Scenario: Inventory is insufficient
- **WHEN** selected quantity plus overlapping reservations exceeds total inventory
- **THEN** the system rejects the booking and shows that the equipment is unavailable

### Requirement: Equipment issue feedback
The system SHALL allow a customer to submit an issue description for a room or equipment item, and SHALL allow the owner to review it and change the affected equipment status.

#### Scenario: Customer reports a fault
- **WHEN** a customer submits a valid issue report
- **THEN** the report appears in the owner's backend for processing
