Feature: Kthulu Plan
  As a developer
  I want to plan my project architecture using a CLI command
  So that I can visualize and configure my project before creating it

  Scenario: Generate a default plan
    Given I run the plan command with "my-app"
    Then a file named "kthulu-plan.yaml" should exist
    And the plan should have name "my-app"
    And the plan should have template "microservice"
    And the plan should have database "sqlite"

  Scenario: Generate a plan with custom template
    Given I run the plan command with "shop --template=ecommerce"
    Then a file named "kthulu-plan.yaml" should exist
    And the plan should have name "shop"
    And the plan should have template "ecommerce"
    And the plan should have feature "product"

  Scenario: Generate a plan with custom features
    Given I run the plan command with "blog --features=posts,comments"
    Then a file named "kthulu-plan.yaml" should exist
    And the plan should have name "blog"
    And the plan should have feature "posts"
    And the plan should have feature "comments"
