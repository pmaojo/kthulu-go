Feature: User Management
  As an API consumer
  I want to manage users
  So that I can create and retrieve user data

  Scenario: Create a new user
    Given the API is running
    When I POST to "/api/users" with:
      | name  | email           |
      | Alice | alice@test.com  |
    Then the response status should be 201
    And the response should contain "Alice"

  Scenario: List all users
    Given the API is running
    When I GET "/api/users"
    Then the response status should be 200
    And the response should be a JSON array

  Scenario: Health check
    Given the API is running
    When I GET "/health"
    Then the response status should be 200
    And the response should contain "ok"
