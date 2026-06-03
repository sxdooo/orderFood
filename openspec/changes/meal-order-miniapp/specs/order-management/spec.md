## ADDED Requirements

### Requirement: View orders by delivery date

The system SHALL allow sellers to view all orders for a specific delivery date.

#### Scenario: Seller views delivery orders

- **WHEN** a seller selects a delivery date
- **THEN** the system returns all orders for that date with buyer name, address, items, total amount, and status

#### Scenario: Seller views empty delivery date

- **WHEN** a seller selects a delivery date with no orders
- **THEN** the system returns an empty list

### Requirement: Order detail view

The system SHALL provide full order details including dishes, quantities, buyer contact info, address snapshots, and timestamps.

#### Scenario: Seller views order detail

- **WHEN** a seller opens a specific order
- **THEN** the system displays all order items with price snapshots, buyer contact name, phone, address, delivery time preference, remark, and order creation time

### Requirement: Order detail map and distance

The system SHALL display buyer address, distance from seller to buyer, and a map with both locations on the seller order detail page.

#### Scenario: Seller sees distance to buyer

- **WHEN** a seller opens an order detail with valid buyer and seller coordinates
- **THEN** the system displays the straight-line or road distance between seller shop and buyer delivery address

#### Scenario: Seller sees map with two markers

- **WHEN** a seller opens an order detail page
- **THEN** the system displays a map with markers for the seller shop location and the buyer delivery address

#### Scenario: Missing coordinates handled gracefully

- **WHEN** an order address lacks geocoded coordinates
- **THEN** the system displays the text address and hides distance/map with a prompt to geocode

### Requirement: Seller refund order

The system SHALL allow sellers to refund orders with a required reason and optional remark.

#### Scenario: Seller refunds order with reason

- **WHEN** a seller submits a refund with a required `refund_reason` and optional `refund_remark`
- **THEN** the system updates the order status to `refunded`, records refund fields and timestamp, and initiates WeChat refund if payment was completed

#### Scenario: Refund without reason rejected

- **WHEN** a seller attempts to refund an order without providing `refund_reason`
- **THEN** the system rejects the refund with a validation error

#### Scenario: Buyer notified of refund

- **WHEN** a seller successfully refunds an order
- **THEN** the buyer can see the order status changed to `refunded` (已退单) in their order list and detail

### Requirement: Daily sales statistics

The system SHALL provide daily order count and revenue summary for sellers.

#### Scenario: Seller views daily stats

- **WHEN** a seller requests statistics for a delivery date
- **THEN** the system returns total order count, total revenue, and count by order status (excluding refunded from revenue)

### Requirement: Order status management

The system SHALL allow sellers to update order status (confirm, start delivery, complete).

#### Scenario: Seller confirms order

- **WHEN** a seller changes an order from `pending` to `confirmed`
- **THEN** the system updates the order status and records the timestamp

#### Scenario: Seller marks delivery complete

- **WHEN** a seller changes an order from `delivering` to `completed`
- **THEN** the system updates the order status to `completed`

### Requirement: Orders isolated by delivery date

The system SHALL store and query orders primarily by `delivery_date` field.

#### Scenario: Orders from different dates are separate

- **WHEN** a seller queries orders for delivery date A
- **THEN** the system returns only orders with `delivery_date = A` and excludes orders from other dates
