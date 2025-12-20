# Building an E-Commerce with Kthulu Go

This tutorial demonstrates how to use the Kthulu CLI to generate a production-ready e-commerce backend and how to leverage the "Agent Editing" workflow.

## 1. Build the CLI

First, we need to build the CLI tool from the repository source.

```bash
mkdir -p bin
cd backend/backend
go build -o ../../bin/kthulu-cli ./cmd/kthulu-cli
cd ../..
```

## 2. Generate the Project

Use the CLI to create a new project using the `microservice` template.

```bash
bin/kthulu-cli create shop --template=microservice --output=examples/e-commerce --database=sqlite
```

## 3. Add E-Commerce Modules

We use the `add module` command to dynamically add features to our project. The CLI automatically resolves dependencies and generates the necessary code (domain, repository, service, handler).

```bash
cd examples/e-commerce/shop
../../../bin/kthulu-cli add module products name:string description:string price:float stock:int
../../../bin/kthulu-cli add module orders total:float status:string customer_name:string
../../../bin/kthulu-cli add module payments amount:float provider:string transaction_id:string
```

## 4. Agent-Guided Editing

Kthulu generates code with special comments (like `// @kthulu:service:products`) and placeholder comments (like `// Add business logic here`). These are designed to guide AI agents (or human developers) in customizing the business logic.

### Example: Adding Validation to Product Service

In this example, we modified `examples/e-commerce/shop/internal/adapters/http/modules/products/service/products_service.go` to add price validation.

**Before:**
```go
func (s *ProductsService) CreateProducts(entity *domain.Products) error {
    // Add business logic here
    return s.repo.Create(entity)
}
```

**After:**
```go
func (s *ProductsService) CreateProducts(entity *domain.Products) error {
    // Added by Agent: Validate price
    if entity.Price < 0 {
        return fmt.Errorf("price cannot be negative")
    }
    return s.repo.Create(entity)
}
```

## 5. Next Steps

- **Run Migrations:** Use `kthulu migrate up` to apply the database schema changes.
- **Start the Server:** Run `go run cmd/server/main.go` to start your e-commerce backend.
- **Add More Features:** Use `kthulu add module` to expand your application.
