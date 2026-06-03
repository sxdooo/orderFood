## ADDED Requirements

### Requirement: WeChat login

The system SHALL authenticate users via WeChat mini-program login using `wx.login` code exchange for OpenID.

#### Scenario: First-time login creates buyer account

- **WHEN** a user logs in with a valid WeChat code for the first time
- **THEN** the system creates a new user record with role `buyer` and returns a JWT token

#### Scenario: Returning user login

- **WHEN** a user with an existing OpenID logs in with a valid WeChat code
- **THEN** the system returns a JWT token without creating a duplicate user

#### Scenario: Invalid WeChat code

- **WHEN** a user submits an invalid or expired WeChat code
- **THEN** the system returns an authentication error and does not issue a token

### Requirement: Role-based access

The system SHALL enforce role-based access control with two roles: `buyer` (default) and `seller`.

#### Scenario: Buyer accesses buyer endpoints

- **WHEN** a user with role `buyer` accesses buyer-only endpoints
- **THEN** the system allows the request

#### Scenario: Buyer denied seller endpoints

- **WHEN** a user with role `buyer` accesses seller-only endpoints
- **THEN** the system returns HTTP 403 Forbidden

#### Scenario: Seller accesses seller endpoints

- **WHEN** a user with role `seller` accesses seller-only endpoints
- **THEN** the system allows the request

### Requirement: Seller authorization

The system SHALL allow seller role assignment only through admin approval or backend configuration.

#### Scenario: Admin promotes user to seller

- **WHEN** an administrator sets a user's role to `seller`
- **THEN** the user gains access to all seller endpoints on next login

#### Scenario: Unapproved seller request

- **WHEN** a buyer requests seller status without admin approval
- **THEN** the system keeps the user as `buyer` and returns a pending status if applicable

### Requirement: Session management

The system SHALL manage user sessions using JWT tokens with Redis-backed invalidation support.

#### Scenario: Valid token access

- **WHEN** a user sends a request with a valid non-expired JWT
- **THEN** the system processes the request with the authenticated user context

#### Scenario: Logout invalidates token

- **WHEN** a user logs out
- **THEN** the system adds the token to a Redis blacklist and subsequent requests with that token are rejected
