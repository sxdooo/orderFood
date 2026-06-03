## ADDED Requirements

### Requirement: Browse tomorrow menu only

The system SHALL only display published menus where the delivery date equals tomorrow (Asia/Shanghai timezone).

#### Scenario: Buyer views tomorrow menu

- **WHEN** a buyer opens the menu page and a published menu exists for tomorrow
- **THEN** the system displays the menu with all available dishes

#### Scenario: No tomorrow menu available

- **WHEN** a buyer opens the menu page and no published menu exists for tomorrow
- **THEN** the system displays an empty state indicating no menu is available

#### Scenario: Future menus hidden from buyer

- **WHEN** a buyer browses menus
- **THEN** the system does NOT display menus for delivery dates beyond tomorrow

### Requirement: Place order with profile pre-fill

The system SHALL allow buyers to select dishes and submit orders, pre-filling delivery info from default profile with per-order override support.

#### Scenario: Checkout pre-fills default profile

- **WHEN** a buyer with a completed profile opens order confirmation
- **THEN** the system pre-fills contact name, phone, and address from the buyer's default profile

#### Scenario: Successful order placement

- **WHEN** a buyer selects dishes, confirms delivery info, and submits before cutoff
- **THEN** the system creates an order with `delivery_date = tomorrow`, status `pending_payment`, snapshots contact/address fields, and returns WeChat payment parameters

#### Scenario: Order confirmed after payment

- **WHEN** a buyer completes WeChat payment successfully
- **THEN** the system updates the order status to `pending` (awaiting seller confirmation)

#### Scenario: Order blocked after cutoff

- **WHEN** a buyer attempts to place an order after the cutoff time for today
- **THEN** the system rejects the order with a cutoff error message

#### Scenario: Order requires valid address

- **WHEN** a buyer submits an order without a complete address
- **THEN** the system rejects the order with a validation error

### Requirement: Order status tracking

The system SHALL track order status through: `pending_payment`, `pending`, `confirmed`, `delivering`, `completed`, `refunded`, and `cancelled`.

#### Scenario: Buyer views order status

- **WHEN** a buyer views their order detail
- **THEN** the system displays the current status and order items with prices

#### Scenario: Buyer sees refunded order

- **WHEN** a seller refunds an order
- **THEN** the buyer sees the order status as `refunded` (已退单) in order list and detail

#### Scenario: Buyer cancels before cutoff

- **WHEN** a buyer cancels their paid order before cutoff time
- **THEN** the system changes order status to `cancelled` and initiates a WeChat refund if payment was completed

### Requirement: Buyer order history

The system SHALL allow buyers to view their historical orders grouped by delivery date.

#### Scenario: Buyer views order history

- **WHEN** a buyer requests their order history
- **THEN** the system returns orders sorted by delivery date descending with status and total amount
