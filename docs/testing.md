# Testing Guide

This document provides an overview of the testing approach and guidelines for the Student Services Platform Backend project.

## Testing Philosophy

We follow a comprehensive testing strategy that includes unit tests, integration tests, and end-to-end tests to ensure the reliability and maintainability of our codebase. Our testing approach is based on the following principles:

1. **Test-Driven Development (TDD)**: Write tests before implementing features to ensure code correctness from the start.
2. **Isolation**: Each test should be independent and not rely on the state of other tests.
3. **Comprehensive Coverage**: Aim for high test coverage across all layers of the application.
4. **Fast Feedback**: Unit tests should run quickly to provide immediate feedback during development.
5. **Realistic Scenarios**: Integration and end-to-end tests should simulate real-world usage scenarios.

## Testing Architecture

Our testing architecture follows the layered structure of the application:

```
┌─────────────────────────────────────────────────────────────┐
│                    Integration Tests                        │
├─────────────────────────────────────────────────────────────┤
│  Controller Tests  │  Middleware Tests  │  Service Tests    │
├─────────────────────────────────────────────────────────────┤
│                  Repository Tests                           │
├─────────────────────────────────────────────────────────────┤
│                    Unit Tests                               │
└─────────────────────────────────────────────────────────────┘
```

### Test Layers

1. **Unit Tests**: Test individual functions and methods in isolation using mocks.
2. **Repository Tests**: Test database operations with a test database.
3. **Service Tests**: Test business logic with mocked dependencies.
4. **Controller Tests**: Test HTTP handlers with mocked services.
5. **Middleware Tests**: Test middleware components like JWT authentication.
6. **Integration Tests**: Test the full application flow with real dependencies.

## Testing Tools and Frameworks

We use the following tools and frameworks for testing:

- **Standard Library**: `testing` package for the core testing framework
- **Testify**: For assertions, mock generation, and test suites
- **Ginkgo/Gomega**: For BDD-style testing (optional)
- **SQLMock**: For mocking database operations
- **HTTPTest**: For testing HTTP handlers
- **GoMock**: For generating mock implementations

## Test Structure

### File Organization

Test files are organized alongside the code they test:

```
app/
├── services/
│   ├── auth/
│   │   ├── service.go
│   │   └── service_test.go
│   └── user/
│       ├── get.go
│       ├── get_test.go
│       ├── update.go
│       └── update_test.go
├── controllers/
│   ├── AuthController/
│   │   ├── login.go
│   │   ├── login_test.go
│   │   ├── register.go
│   │   └── register_test.go
│   └── UserController/
│       ├── me.go
│       └── me_test.go
├── midwares/
│   ├── jwt.go
│   └── jwt_test.go
internal/
├── db/
│   ├── users_repo.go
│   └── users_repo_test.go
└── testutils/
    ├── testutils.go
    ├── mock_db.go
    └── suite.go
tests/
└── integration/
    └── api_integration_test.go
```

### Test Naming Conventions

- Test files should be named `<filename>_test.go`
- Test functions should follow the pattern `Test<FunctionName>_<Scenario>`
- Test suites should be named `<Component>TestSuite`
- Benchmark functions should be named `Benchmark<FunctionName>`

### Test Structure

Each test should follow the Arrange-Act-Assert pattern:

```go
func (suite *AuthServiceTestSuite) TestRegister_Success() {
    // Arrange - Set up test data and mocks
    userReq := openapi.UserCreate{
        Email:    "newuser@example.com",
        Name:     "New User",
        Password: "password123",
        Role:     openapi.RoleStudent,
    }

    // Act - Call the function being tested
    user, err := suite.service.Register(userReq)

    // Assert - Verify the results
    require.NoError(suite.T(), err)
    assert.NotNil(suite.T(), user)
    assert.Equal(suite.T(), userReq.Email, user.Email)
    // ... more assertions
}
```

## Test Utilities

### Test Suites

We use test suites to share common setup and teardown logic:

```go
type AuthServiceTestSuite struct {
    testutils.TestSuite
    service *auth.Service
}

func (suite *AuthServiceTestSuite) SetupTest() {
    suite.TestSuite.SetupTest()
    
    // Setup auth service with test configuration
    suite.service = auth.NewService(suite.DB, &auth.JWTConfig{
        SecretKey:      "test-secret-key",
        AccessTokenExp: time.Hour,
        Issuer:         "test-issuer",
        Audience:       "test-audience",
    })
}
```

### Mocks

We use mocks to isolate the code being tested from its dependencies:

```go
type MockAuthService struct {
    mock.Mock
}

func (m *MockAuthService) Register(req openapi.UserCreate) (*openapi.User, error) {
    args := m.Called(req)
    return args.Get(0).(*openapi.User), args.Error(1)
}
```

### Test Data Factories

We use factories to create consistent test data:

```go
func (f *UserFactory) CreateDefaultUser() *db.User {
    user := &db.User{
        Email:        "test@example.com",
        Name:         "Test User",
        Role:         db.RoleStudent,
        // ... other fields
    }
    
    err := f.db.Create(user).Error
    require.NoError(f.t, err, "Failed to create default test user")
    return user
}
```

## Running Tests

### Running All Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage
```

### Running Specific Tests

```bash
# Run tests for a specific package
go test ./app/services/auth/...

# Run a specific test function
go test -run TestRegister_Success ./app/services/auth/...

# Run tests with verbose output
go test -v ./app/services/auth/...
```

### Running Tests in CI/CD

Our CI/CD pipeline automatically runs tests on every push and pull request:

1. **Linting**: Code is checked for style and correctness
2. **Unit Tests**: Fast tests that run in isolation
3. **Integration Tests**: Slower tests that verify component interactions
4. **Coverage Report**: Generates and uploads coverage reports

## Test Coverage

We aim for a minimum of 80% test coverage across all packages. Coverage reports are generated automatically and can be viewed locally:

```bash
# Generate coverage report
make test-coverage

# View coverage report in browser
open coverage.html
```

## Best Practices

### Writing Good Tests

1. **Test Behavior, Not Implementation**: Focus on what the code should do, not how it does it.
2. **Use Descriptive Names**: Test names should clearly describe what they're testing.
3. **Keep Tests Simple**: Each test should verify one specific behavior.
4. **Avoid Test Interdependence**: Tests should not rely on the order they run in.
5. **Use Helpers for Common Setup**: Extract common setup code into helper functions.

### Mocking Guidelines

1. **Mock External Dependencies**: Mock databases, APIs, and other external services.
2. **Verify Mock Interactions**: Ensure that mocks are called with the expected parameters.
3. **Don't Mock Everything**: Only mock dependencies that are slow or unreliable.
4. **Use Real Implementations When Possible**: For fast, reliable dependencies, use real implementations.

### Integration Testing

1. **Use Test Databases**: Use in-memory SQLite or test containers for database tests.
2. **Test Real Scenarios**: Test complete user workflows, not just individual components.
3. **Clean Up After Tests**: Ensure that tests clean up any data they create.
4. **Use Test Configuration**: Use separate configuration for tests to avoid affecting production.

## Debugging Tests

### Debugging Failed Tests

When a test fails, use these strategies to debug it:

1. **Run with Verbose Output**: `go test -v` to see which tests are running.
2. **Run Only the Failing Test**: `go test -run TestName` to isolate the issue.
3. **Add Logging**: Add `t.Log()` statements to see what's happening during the test.
4. **Use a Debugger**: Use Delve or your IDE's debugger to step through the test.

### Common Issues and Solutions

1. **Race Conditions**: Use `t.Parallel()` carefully and ensure tests are independent.
2. **Flaky Tests**: Identify and eliminate non-deterministic behavior.
3. **Slow Tests**: Optimize test setup and use mocks for slow dependencies.
4. **Memory Leaks**: Ensure that tests clean up resources properly.

## Contributing

When adding new features or fixing bugs, please follow these guidelines:

1. **Write Tests First**: Implement tests before writing the actual code.
2. **Update Tests**: When changing existing code, update the corresponding tests.
3. **Maintain Coverage**: Ensure that new code is adequately tested.
4. **Review Tests**: Have your tests reviewed along with your code changes.

## Resources

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Testify Documentation](https://github.com/stretchr/testify)
- [Effective Go Testing](https://go.dev/doc/effective_go#testing)
- [Go Blog: Testing](https://go.dev/blog/testing)