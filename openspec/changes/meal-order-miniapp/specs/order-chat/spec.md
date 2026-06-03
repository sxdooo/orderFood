## ADDED Requirements

### Requirement: Order-scoped chat room

The system SHALL provide a chat room for each order accessible by the buyer and seller involved in that order.

#### Scenario: Buyer opens order chat

- **WHEN** a buyer opens the chat for their order
- **THEN** the system displays the message history for that order

#### Scenario: Seller opens order chat

- **WHEN** a seller opens the chat for an order on their delivery date
- **THEN** the system displays the message history for that order

#### Scenario: Unauthorized user denied

- **WHEN** a user who is neither the buyer nor the seller attempts to access an order chat
- **THEN** the system returns HTTP 403 Forbidden

### Requirement: Send text messages

The system SHALL allow buyers and sellers to send text messages within an order chat.

#### Scenario: Buyer sends text message

- **WHEN** a buyer sends a text message in an order chat
- **THEN** the system saves the message with sender role `buyer` and returns it to the chat

#### Scenario: Seller sends text message

- **WHEN** a seller sends a text message in an order chat
- **THEN** the system saves the message with sender role `seller` and returns it to the chat

### Requirement: Send image messages

The system SHALL allow buyers and sellers to send image messages uploaded to cloud storage.

#### Scenario: User sends image message

- **WHEN** a user uploads an image in an order chat
- **THEN** the system stores the image URL and saves a message with type `image`

### Requirement: Message history

The system SHALL retain and return full message history for each order chat, ordered by creation time.

#### Scenario: Load chat history

- **WHEN** a participant opens an order chat with existing messages
- **THEN** the system returns all messages in chronological order with sender info and timestamps

### Requirement: Message polling

The system SHALL support polling for new messages with a configurable interval (default 5 seconds).

#### Scenario: Poll for new messages

- **WHEN** a user polls for messages with a `since` timestamp
- **THEN** the system returns only messages created after that timestamp

#### Scenario: Unread message indicator

- **WHEN** a user has unread messages in an order chat
- **THEN** the system displays an unread count on the order list
