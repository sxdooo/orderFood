## ADDED Requirements

### Requirement: Collect delivery addresses

The system SHALL collect all order addresses for a given delivery date to prepare route planning.

#### Scenario: Seller requests address list

- **WHEN** a seller requests the delivery list for a delivery date after cutoff
- **THEN** the system returns all confirmed orders with buyer name, address, phone, and coordinates

#### Scenario: Orders without coordinates flagged

- **WHEN** an order address cannot be geocoded
- **THEN** the system includes the order in the list with a flag indicating missing coordinates

### Requirement: Calculate optimal delivery route

The system SHALL call a map API (Amap/Gaode) to calculate an optimized delivery route based on order addresses.

#### Scenario: Route generation after cutoff

- **WHEN** a seller triggers route generation for a delivery date with multiple orders
- **THEN** the system calls the map API with all geocoded addresses and returns an optimized stop sequence with total distance and estimated duration

#### Scenario: Single order route

- **WHEN** a seller triggers route generation with only one order
- **THEN** the system returns a route with a single destination from the seller's location

#### Scenario: Map API failure

- **WHEN** the map API returns an error
- **THEN** the system returns an error message and allows the seller to view the unordered address list

### Requirement: Display route on map

The system SHALL display the calculated route on a map in the mini-program.

#### Scenario: Seller views route on map

- **WHEN** a seller opens the delivery route page
- **THEN** the system displays a map with markers for each stop and a polyline showing the route

### Requirement: Manual route adjustment

The system SHALL allow sellers to manually reorder delivery stops and save the adjusted sequence.

#### Scenario: Seller reorders stops

- **WHEN** a seller drags to reorder delivery stops and saves
- **THEN** the system persists the new stop order and updates the route display

### Requirement: Route persistence

The system SHALL persist calculated routes so sellers can reference them during delivery.

#### Scenario: Route saved for delivery date

- **WHEN** a route is generated or manually adjusted
- **THEN** the system saves the route with stop order, distance, and duration linked to the delivery date
