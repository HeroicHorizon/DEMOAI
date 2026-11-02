# feature-demo Specification

## Purpose
Feature: DEMO

## Requirements

### Requirement: Feature: DEMO acceptance criterion 1 (#1)
The system SHALL meet this acceptance criterion: docker compose up starts the frontend, backend and MongoDB. The backend connects to MongoDB, and the frontend reaches the backend.

#### Scenario: Criterion 1
- **WHEN** the change is in place
- **THEN** docker compose up starts the frontend, backend and MongoDB. The backend connects to MongoDB, and the frontend reaches the backend.

### Requirement: Feature: DEMO acceptance criterion 2 (#1)
The system SHALL meet this acceptance criterion: The six endpoints behave as specified:

#### Scenario: Criterion 2
- **WHEN** the change is in place
- **THEN** The six endpoints behave as specified:

### Requirement: Feature: DEMO acceptance criterion 3 (#1)
The system SHALL meet this acceptance criterion: updatedAt is refreshed on update and on status change.

#### Scenario: Criterion 3
- **WHEN** the change is in place
- **THEN** updatedAt is refreshed on update and on status change.

### Requirement: Feature: DEMO acceptance criterion 4 (#1)
The system SHALL meet this acceptance criterion: The list supports server-side pagination, search (regex-escaped) and status filter. An empty result shows "No users found."

#### Scenario: Criterion 4
- **WHEN** the change is in place
- **THEN** The list supports server-side pagination, search (regex-escaped) and status filter. An empty result shows "No users found."

### Requirement: Feature: DEMO acceptance criterion 5 (#1)
The system SHALL meet this acceptance criterion: The dashboard shows total, active and inactive counts, and they stay current after any change.

#### Scenario: Criterion 5
- **WHEN** the change is in place
- **THEN** The dashboard shows total, active and inactive counts, and they stay current after any change.

### Requirement: Feature: DEMO acceptance criterion 6 (#1)
The system SHALL meet this acceptance criterion: When MongoDB is unavailable, the API returns 503 with no internals in the body. The frontend shows "Unable to reach the server. Please try again later."

#### Scenario: Criterion 6
- **WHEN** the change is in place
- **THEN** When MongoDB is unavailable, the API returns 503 with no internals in the body. The frontend shows "Unable to reach the server. Please try again later."

### Requirement: Feature: DEMO acceptance criterion 7 (#1)
The system SHALL meet this acceptance criterion: The request context is propagated to database calls.

#### Scenario: Criterion 7
- **WHEN** the change is in place
- **THEN** The request context is propagated to database calls.

### Requirement: Feature: DEMO acceptance criterion 8 (#1)
The system SHALL meet this acceptance criterion: CORS allows only the configured origin. Payload size is limited. All configuration comes from environment variables.

#### Scenario: Criterion 8
- **WHEN** the change is in place
- **THEN** CORS allows only the configured origin. Payload size is limited. All configuration comes from environment variables.

### Requirement: Feature: DEMO acceptance criterion 9 (#1)
The system SHALL meet this acceptance criterion: Backend unit, repository and API integration tests exist and pass. Frontend component tests exist and pass.

#### Scenario: Criterion 9
- **WHEN** the change is in place
- **THEN** Backend unit, repository and API integration tests exist and pass. Frontend component tests exist and pass.
