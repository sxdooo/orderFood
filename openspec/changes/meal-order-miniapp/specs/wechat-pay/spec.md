## ADDED Requirements

### Requirement: WeChat JSAPI payment on order

The system SHALL initiate WeChat mini-program JSAPI payment when a buyer submits an order.

#### Scenario: Payment initiated after order creation

- **WHEN** a buyer submits a valid order before cutoff
- **THEN** the system creates an order with status `pending_payment` and returns WeChat payment parameters (`timeStamp`, `nonceStr`, `package`, `signType`, `paySign`)

#### Scenario: Order blocked without payment

- **WHEN** a buyer creates an order but does not complete payment within the timeout period
- **THEN** the system cancels the order with status `cancelled` due to payment timeout

### Requirement: Payment callback handling

The system SHALL handle WeChat payment callback notifications to confirm successful payments.

#### Scenario: Successful payment callback

- **WHEN** the system receives a valid WeChat payment success notification for an order
- **THEN** the system updates the order status to `pending` (awaiting seller confirmation) and records the transaction ID and paid amount

#### Scenario: Invalid payment callback signature

- **WHEN** the system receives a payment notification with an invalid signature
- **THEN** the system rejects the notification and does not update the order status

### Requirement: Payment refund on cancellation

The system SHALL support refunding via WeChat Pay when a paid order is cancelled before delivery.

#### Scenario: Buyer cancels paid order before cutoff

- **WHEN** a buyer cancels a paid order before cutoff time
- **THEN** the system initiates a WeChat refund and updates the order status to `cancelled` upon refund success

#### Scenario: Seller cancels paid order via refund

- **WHEN** a seller refunds a paid order with a valid reason
- **THEN** the system initiates a full WeChat refund to the buyer, updates the order status to `refunded`, and records refund reason and remark

### Requirement: Payment record tracking

The system SHALL maintain payment records linked to orders for reconciliation.

#### Scenario: Seller views payment info on order

- **WHEN** a seller views an order detail
- **THEN** the system displays payment status, transaction ID, paid amount, and paid timestamp if payment was completed
