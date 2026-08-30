## Purpose

为顾客提供完整的门店展示和会员办理说明，并让老板能够维护演示网站中的基础内容。

## ADDED Requirements

### Requirement: Venue homepage content
The system SHALL display the venue name, description, address, contact information, opening rules, membership plan information, room highlights, and a booking entry point.

#### Scenario: Visitor browses homepage
- **WHEN** an unauthenticated visitor opens the homepage
- **THEN** the system displays venue information and allows browsing rooms and availability without requiring login

### Requirement: Owner manages venue content
The system SHALL allow the owner to edit venue information, contact details, membership instructions, opening rules, announcements, and room photos.

#### Scenario: Publish an announcement
- **WHEN** the owner saves a valid announcement
- **THEN** the announcement is displayed on the customer-facing site
