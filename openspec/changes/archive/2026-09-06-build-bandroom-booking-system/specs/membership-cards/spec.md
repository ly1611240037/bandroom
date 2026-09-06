## Purpose

为线下销售的月卡和次数卡提供可追踪的会员资格、使用资格和预约消耗管理。

## ADDED Requirements

### Requirement: Membership plan management
The system SHALL allow the owner to create, edit, enable, and disable monthly and count-based card plans with names, prices, and applicable rules.

#### Scenario: Owner creates a plan
- **WHEN** the owner submits a valid membership plan
- **THEN** the plan becomes available for manual activation

### Requirement: Manual card activation
The system SHALL allow the owner to activate a card for a customer after offline payment and SHALL record plan, amount, payment method, activation time, and notes.

#### Scenario: Activate monthly card
- **WHEN** the owner activates a monthly card on the purchase date
- **THEN** the card is valid for 30 days starting that date

#### Scenario: Activate a prepaid future monthly card
- **WHEN** the owner assigns a second monthly card to a customer with an active monthly card
- **THEN** the second card starts on the day after the current card expires

### Requirement: Card eligibility and consumption
The system SHALL treat an active monthly card as unlimited booking eligibility and SHALL consume one count from a count card for each successful booking when no active monthly card exists.

#### Scenario: Monthly card takes priority
- **WHEN** a customer has both an active monthly card and remaining count-card uses
- **THEN** the booking uses the monthly card and does not reduce count-card uses

#### Scenario: Count card is exhausted
- **WHEN** a customer has no active monthly card and zero remaining count-card uses
- **THEN** the system rejects the booking and explains that an active membership is required

#### Scenario: Cancel count-card booking
- **WHEN** a customer cancels a booking that consumed a count-card use
- **THEN** the system restores exactly one count-card use

