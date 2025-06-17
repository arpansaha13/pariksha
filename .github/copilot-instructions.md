# GitHub Copilot Instructions

## General Coding Guidelines
- Write clear and concise code, following best practices for Vue, Nuxt3, and Go.
- Ensure function names, variable names, and comments improve code readability.
- Prefer explicit over implicit logic to enhance maintainability.
- Do not remove any existing comment.

## Function Documentation
- When creating a new function, add a short comment before it describing its main purpose.

Example:

```go
// Add returns the sum of two integers.
func Add(a, b int) int {
    return a + b
}
```

## Project coding standards

### TypeScript Guidelines
- Use TypeScript for all new code
- Follow functional programming principles where possible
- Use interfaces for data structures and type definitions
- Prefer immutable data (const, readonly)
- Use optional chaining (?.) and nullish coalescing (??) operators
- Every property in a TypeScript interface should be prefixed with `readonly`.

### Golang Guidelines

- Prefer `any` over `interface{}`.
- Prefer the use of explicit integer types (like `int16`, `int32`, `int64`, `uint8`, `uint32`, `uint64`) instead of the default `int` or `uint` when generating Go code.
    - Improves clarity about size and range.
    - Ensures consistent behavior across different architectures.
    - Helps avoid platform-dependent issues (e.g. `int` is 32-bit on some   systems and 64-bit on others).

## Controller-Service-Repository Pattern in Golang Backend

The Controller-Service-Repository (CSR) pattern promotes separation of concerns, maintainability, and testability in Golang backends.

### General Structure
1. **Controller Layer:** Handles HTTP requests, validates inputs, and calls the appropriate service methods.
2. **Service Layer:** Implements business logic, processes requests, and interacts with repositories.
3. **Repository Layer:** Interacts with the database using `gorm.DB`, ensuring data persistence.

### Repository Layer Guidelines
- Every repository method must take the first parameter as `tx *gorm.DB`.
- If the service layer passes a `*gorm.DB` as the first parameter, the repository method should use that.
- If the first parameter is `nil` (that is, no `*gorm.DB` was provided), then the repository uses its own `*gorm.DB` instance (passed to it during initialization).

#### Example Repository Method
```go
package repositories

import (
    "database/sql"

    "gorm.io/gorm"

    "example.com/project/models"
    "example.com/project/utils"
)

// Paper repository
type Paper struct {
    db *gorm.DB
}

// NewPaper initializes a Paper repository
func NewPaper(db *gorm.DB) *Paper {
    return &Paper{db: db}
}

// getTx returns tx if not nil, otherwise returns r.db
func (r *Paper) getTx(tx *gorm.DB) *gorm.DB {
    if tx == nil {
        return r.db
    }
    return tx
}

func (r *Paper) Transaction(fc func(tx *gorm.DB) error, opts ...*sql.TxOptions) error {
	return utils.TransactionHandler(r.db, fc, opts...)
}

// GetByID retrieves a paper by ID from the database.
func (r *Paper) GetByID(tx *gorm.DB, paperId int64) (*models.Paper, error) {
    tx = r.getTx(tx)
    var paper models.Paper
    if err := tx.Take(&paper, paperId).Error; err != nil {
        return nil, err
    }
    return &paper, nil
}
```

### Service Layer Guidelines

- If the service layer starts a transaction, then the `*gorm.DB` of the transaction must be passed to the repository methods as the first parameter.

```go
package services

import (
    "example.com/project/models"
    "example.com/project/repositories"
)

// Paper service
type Paper struct {
    paperRepo *repositories.Paper
}

// NewPaper initializes the Paper service
func NewPaper(paperRepo *repositories.Paper) *Paper {
    return &Paper{paperRepo: paperRepo}
}

// GetPaperByID retrieves a paper by ID
func (s *Paper) GetPaperByID(paperID int64) (*models.Paper, error) {
    return s.paperRepo.GetByID(nil, paperID) // Passing nil means default DB instance in the repository is used.
}

// CreatePaper demonstrates using a transaction within the service layer
func (s *Paper) CreatePaper(paper *models.Paper) error {
    return s.paperRepo.Transaction(func(tx *gorm.DB) error {
        return s.paperRepo.Create(tx, paper) // Passing `*gorm.DB` of transaction to repository method
    })
}
```

## Security Best Practices

- Avoid hardcoding sensitive information.
- Sanitize user input and prevent injection attacks.
- Use environment variables for configuration.
