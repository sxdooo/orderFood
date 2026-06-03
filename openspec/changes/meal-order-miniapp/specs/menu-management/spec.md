## ADDED Requirements

### Requirement: Create menu by delivery date

The system SHALL allow sellers to create a menu for a specific delivery date with a list of dishes.

#### Scenario: Seller creates draft menu

- **WHEN** a seller creates a menu for a future delivery date with dish name, price, and optional image/description
- **THEN** the system saves the menu with status `draft` and associates dishes as menu items

#### Scenario: Duplicate delivery date menu

- **WHEN** a seller attempts to create a menu for a delivery date that already has a menu
- **THEN** the system returns an error indicating a menu already exists for that date

### Requirement: Edit menu and dishes

The system SHALL allow sellers to edit draft or published menus, including adding, updating, and soft-disabling dishes.

#### Scenario: Seller edits draft menu

- **WHEN** a seller updates dishes in a `draft` menu
- **THEN** the system saves all changes immediately

#### Scenario: Seller disables dish with existing orders

- **WHEN** a seller disables a dish that has been ordered
- **THEN** the system sets `is_available = false` without deleting the dish record or affecting existing orders

### Requirement: Publish menu

The system SHALL allow sellers to publish menus, making them visible to buyers when the delivery date is tomorrow.

#### Scenario: Seller publishes menu

- **WHEN** a seller publishes a menu with at least one available dish
- **THEN** the system changes menu status to `published`

#### Scenario: Publish empty menu rejected

- **WHEN** a seller attempts to publish a menu with no available dishes
- **THEN** the system rejects the publish action with a validation error

### Requirement: Menu status lifecycle

The system SHALL manage menu statuses: `draft`, `published`, and `expired`.

#### Scenario: Menu auto-expires after delivery date

- **WHEN** the delivery date has passed
- **THEN** the system automatically sets the menu status to `expired`

#### Scenario: Seller views menus by date range

- **WHEN** a seller requests menus for a date range
- **THEN** the system returns all menus within that range with their status and dish counts

### Requirement: Multi-day menu publishing

The system SHALL allow sellers to publish menus for multiple future delivery dates in advance.

#### Scenario: Seller publishes menus for next three days

- **WHEN** a seller creates and publishes menus for tomorrow, day after tomorrow, and the third day
- **THEN** all three menus are stored independently, each keyed by its delivery date
