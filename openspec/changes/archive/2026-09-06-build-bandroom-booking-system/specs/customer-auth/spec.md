## Purpose

为顾客提供安全、清晰的账号入口，使预约与个人联系人身份绑定，并支持邮箱通知和账号恢复。

## ADDED Requirements

### Requirement: Customer registration and verification
The system SHALL allow a customer to register with name, email, phone number, and password, and SHALL require email verification before booking.

#### Scenario: Successful registration
- **WHEN** a customer submits valid registration details
- **THEN** the system creates an unverified account and sends a verification email

#### Scenario: Unverified customer attempts to book
- **WHEN** an unverified customer submits a booking
- **THEN** the system rejects the booking and explains that email verification is required

### Requirement: Customer authentication
The system SHALL allow verified customers to log in and SHALL restrict each customer to their own profile, memberships, bookings, and notifications.

#### Scenario: Verified login
- **WHEN** a verified customer submits correct credentials
- **THEN** the system creates an authenticated session

#### Scenario: Wrong credentials
- **WHEN** a customer submits invalid credentials
- **THEN** the system rejects the login without revealing which credential was incorrect

### Requirement: Password recovery
The system SHALL allow a customer to request a time-limited password reset link through their registered email address.

#### Scenario: Password reset
- **WHEN** a customer uses a valid reset link and submits a new password
- **THEN** the system updates the password and invalidates the reset link

