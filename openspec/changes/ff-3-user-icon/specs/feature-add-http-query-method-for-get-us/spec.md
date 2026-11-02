## ADDED Requirements

### Requirement: User icon in the UI
The UI SHALL show a decorative user icon on the User Details page and in each users list row, without changing existing text or column headers.

#### Scenario: Details page shows the icon
- **WHEN** a user's details have loaded
- **THEN** one user icon is shown beside the heading and the full name

#### Scenario: List rows show the icon
- **WHEN** the users list is rendered
- **THEN** each row contains a user icon and the column headers are unchanged
