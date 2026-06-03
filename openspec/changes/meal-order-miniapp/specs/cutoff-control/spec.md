## ADDED Requirements

### Requirement: Set daily cutoff time

The system SHALL allow sellers to set a cutoff time for each order date (today), after which buyers cannot place new orders for tomorrow.

#### Scenario: Seller sets cutoff time

- **WHEN** a seller sets today's cutoff time to 17:00
- **THEN** the system saves the cutoff setting for today's order date

#### Scenario: Seller updates cutoff time before deadline

- **WHEN** a seller changes today's cutoff time before the current cutoff has passed
- **THEN** the system updates the cutoff time and buyers can still order until the new time

### Requirement: Automatic cutoff enforcement

The system SHALL automatically stop accepting orders when the cutoff time is reached.

#### Scenario: Cutoff time reached

- **WHEN** the current time passes the configured cutoff time for today
- **THEN** the system marks today's ordering as closed in Redis and rejects new order submissions

#### Scenario: Order attempt after cutoff

- **WHEN** a buyer attempts to place an order after cutoff
- **THEN** the system returns an error indicating ordering is closed for today

### Requirement: Cutoff timezone handling

The system SHALL evaluate all cutoff times in `Asia/Shanghai` timezone.

#### Scenario: Cutoff evaluated at correct local time

- **WHEN** the server time is 17:00 Asia/Shanghai and cutoff is set to 17:00
- **THEN** the system closes ordering at exactly 17:00 local time regardless of server UTC offset

### Requirement: Cutoff status visibility

The system SHALL display the current cutoff status and remaining time to buyers.

#### Scenario: Buyer sees ordering open

- **WHEN** a buyer views the menu page before cutoff
- **THEN** the system displays the cutoff time and that ordering is open

#### Scenario: Buyer sees ordering closed

- **WHEN** a buyer views the menu page after cutoff
- **THEN** the system displays that ordering is closed and hides the order button
