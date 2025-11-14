# Contract Testing Framework

This directory contains comprehensive contract tests for the Kthulu backend system.

## Overview

Contract testing ensures that:
1. Repository implementations satisfy their interfaces
2. HTTP endpoints return consistent response formats
3. API contracts are maintained across changes
4. Integration points work correctly

## Test Structure

### Repository Contract Tests

- **Interface Compliance**: Verifies all repository implementations satisfy their interfaces at compile time
- **Method Signatures**: Ensures repository methods have expected signatures
- **Basic CRUD Operations**: Tests create, read, update, delete operations
- **Error Handling**: Validates proper error responses
- **Data Integrity**: Ensures data consistency across operations

### HTTP Contract Tests

- **Endpoint Availability**: Verifies all documented endpoints exist
- **Request/Response Formats**: Ensures consistent JSON structures
- **Status Codes**: Validates appropriate HTTP status codes
- **Authentication**: Tests protected endpoint access
- **Error Responses**: Ensures consistent error formats

## Running Tests

### All Contract Tests
```bash
make test-contracts
```

### With Coverage
```bash
make test-contracts-coverage
```

### Individual Test Suites
```bash
# Repository contracts only
go test -v -run TestRepositoryContracts ./internal/contracts/

# HTTP contracts only  
go test -v -run TestHTTPContracts ./internal/contracts/

# Interface compliance only
go test -v -run TestBasicRepositoryContracts ./internal/contracts/
```

## Test Coverage

The contract testing framework covers:

### Repository Interfaces
- ✅ UserRepository
- ✅ RoleRepository  
- ✅ RefreshTokenRepository
- ✅ OrganizationRepository
- ✅ ContactRepository
- ✅ ProductRepository
- ✅ InvoiceRepository
- 🚧 InventoryRepository (in progress)
- 🚧 CalendarRepository (in progress)

### HTTP Endpoints
- 🚧 Authentication endpoints
- 🚧 User management endpoints
- 🚧 Organization endpoints
- 🚧 Contact endpoints
- 🚧 Product endpoints
- 🚧 Invoice endpoints
- 🚧 Health check endpoints

## Implementation Status

### Completed ✅
- Repository interface compliance testing
- Basic contract test framework
- Test coverage reporting scripts
- Makefile integration
- Documentation

### In Progress 🚧
- Full repository behavioral testing
- HTTP endpoint contract testing
- Integration with CI/CD pipeline
- Performance contract testing

### Planned 📋
- API versioning contract tests
- Database migration contract tests
- External service contract tests
- Load testing contracts

## Best Practices

1. **Interface First**: Always define repository interfaces before implementations
2. **Fail Fast**: Contract tests should fail immediately on interface violations
3. **Comprehensive Coverage**: Test all public methods and endpoints
4. **Consistent Patterns**: Use consistent test patterns across all contracts
5. **Clear Assertions**: Make test failures easy to understand and fix

## Adding New Contract Tests

When adding a new repository or endpoint:

1. Add interface compliance test in `basic_contracts_test.go`
2. Create behavioral tests in `repository_contracts_test.go`
3. Add HTTP endpoint tests in `http_contracts_test.go`
4. Update coverage documentation
5. Run full test suite to ensure no regressions

## Troubleshooting

### Common Issues

**GORM Migration Errors**: The test database setup uses GORM AutoMigrate for simplicity. Complex domain models may require manual table creation.

**Interface Violations**: If a repository doesn't implement all interface methods, the compile-time tests will fail immediately.

**HTTP Setup Complexity**: Full HTTP contract tests require complex dependency injection setup. Start with basic endpoint availability tests.

### Solutions

- Use simplified domain models for testing
- Mock complex dependencies
- Focus on interface compliance first
- Add behavioral tests incrementally

## Future Enhancements

- Integration with OpenAPI specification validation
- Automated contract test generation from interfaces
- Performance benchmarking integration
- Contract versioning and compatibility testing
- Integration with external API contract testing tools