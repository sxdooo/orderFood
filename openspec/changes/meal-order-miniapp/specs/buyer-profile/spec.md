## ADDED Requirements

### Requirement: First-time profile onboarding

The system SHALL guide buyers to complete their default delivery profile on first login.

#### Scenario: First login redirects to profile setup

- **WHEN** a buyer logs in for the first time and `profile_completed` is false
- **THEN** the system redirects the buyer to a profile setup page requiring contact name, phone, and address

#### Scenario: Profile setup completion

- **WHEN** a buyer submits valid contact name, phone, and address on the setup page
- **THEN** the system saves the information as default delivery profile and sets `profile_completed = true`

#### Scenario: Incomplete profile blocks ordering

- **WHEN** a buyer with `profile_completed = false` attempts to place an order
- **THEN** the system redirects the buyer to complete profile setup first

### Requirement: Default delivery information

The system SHALL store each buyer's default delivery information: contact name, phone, and address with coordinates.

#### Scenario: Default info available for checkout

- **WHEN** a buyer with a completed profile opens the order confirmation page
- **THEN** the system pre-fills contact name, phone, and address from the buyer's default profile

### Requirement: Per-order override without affecting defaults

The system SHALL allow buyers to modify delivery information during checkout without changing their default profile.

#### Scenario: Buyer overrides address for one order

- **WHEN** a buyer changes the address on the order confirmation page and submits the order
- **THEN** the system saves the modified values as order snapshots and does NOT update the buyer's default profile

#### Scenario: Default profile unchanged after order

- **WHEN** a buyer places an order with a temporarily modified address
- **THEN** the buyer's default profile retains the original address for future orders

### Requirement: Manage default profile in My page

The system SHALL allow buyers to view and edit their default delivery information on the "My" page.

#### Scenario: Buyer views default profile

- **WHEN** a buyer opens the profile section on the "My" page
- **THEN** the system displays the current default contact name, phone, and address

#### Scenario: Buyer updates default profile

- **WHEN** a buyer updates their default delivery information on the "My" page
- **THEN** the system saves the new values and geocodes the address to update coordinates

#### Scenario: Updated default applies to next order

- **WHEN** a buyer updates their default profile and places a new order without modifications
- **THEN** the order confirmation page pre-fills the updated default information

### Requirement: Historical orders use snapshots

The system SHALL preserve order contact and address as snapshots independent of profile changes.

#### Scenario: Profile change does not affect past orders

- **WHEN** a buyer updates their default profile after placing an order
- **THEN** the previously placed order retains the original contact name, phone, and address from order time
