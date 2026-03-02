# Go Project Template
This template implements a **Package-Oriented Architecture**. It prioritizes Go idioms over rigid enterprise patterns, following the lead of major projects like **Kubernetes** and **Terraform**.


## Core Philosophy
- **Accept Interfaces, Return Structs:** We define interfaces where they are consumed (in ports.go), not where they are implemented.
- **Encapsulation over Layering:** Logic is grouped by business domain (e.g., user, order) rather than technical role (service, controller).
- **Dependency Orchestration:** Cross-package logic is handled at the api/ layer to prevent circular dependencies.


## Project Structure 
/myapp
  ├── cmd/
  │    └── server/main.go        // App entry point & dependency injection (wiring)
  ├── internal/
  │    ├── domain/               // Shared DTOs and Entities (prevents circular deps)
  │    │    ├── user.go
  │    │    └── order.go
  │    ├── user/                 // Domain-specific package
  │    │    ├── service.go       // Core business logic (Concrete Structs)
  │    │    ├── service_test.go 
  │    │    ├── ports.go         // Requirements defined as Interfaces
  │    │    └── repository/      // Infrastructure implementations (Postgres, Mock, etc.)
  │    │         └── pg.go
  │    └── order/              
  ├── pkg/                       // Reusable library code (Logger, Auth, Crypto)
  └── api/                       // Transport layer (REST Handlers, gRPC, Middleware)
       └── rest/
            └── user_handler.go  // Orchestrates calls between domain services


## Key Design Rules
**1. Avoiding Circular Dependencies**
To prevent the common ```import cycle not allowed``` error:
**A. Shared Types:** All structs used by more than one package live in ```internal/domain```.
**B. Orchestration:** If a feature requires calling both ```user``` and ```order``` services, that logic lives in the ```api/rest``` handler or a dedicated orchestrator.

**2. Mocking and Testability**
Testability is achieved through ```Constructor Injection```. Each service in ```internal/``` defines its dependencies as interfaces in ```ports.go```. 
During testing, you simply pass a mock implementation into the service constructor.

**3. The ```internal/``` Boundary**
Code inside ```internal/``` cannot be imported by any code outside this project. 
This ensures your core business logic remains private and cannot be "leaked" into external tools or libraries.

### Code Example for ```Orchestration```
This example demonstrates a **"Checkout"** flow. The ```api``` layer orchestrates the logic, keeping the ```user``` and ```order``` packages decoupled.

#### 1. The Service (internal/order/service.go)
Each service is "pure" and only manages its own domain logic.
```Go
package order

import (
    "context"
    "myproject/internal/domain"
)

type Service struct {
    repo OrderRepository
}

func (s *Service) CreateOrder(ctx context.Context, userID int, items []domain.Item) (*domain.Order, error) {
    // Domain-specific logic only
    return s.repo.Save(ctx, &domain.Order{UserID: userID, Items: items})
}
```

#### 2. The Orchestrator (api/rest/order_handler.go)
The handler imports multiple services to coordinate a multi-step workflow.
```Go
package rest

import (
    "myproject/internal/user"
    "myproject/internal/order"
)

type OrderHandler struct {
    userSvc  *user.Service
    orderSvc *order.Service
}

func (h *OrderHandler) HandleCheckout(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    userID := getIDFromToken(r)

    // 1. check user status (call User Domain)
    u, _ := h.userSvc.GetProfile(ctx, userID)
    if u.IsBanned {
        http.Error(w, "User is banned", http.StatusForbidden)
        return
    }

    // 2. create the order (call Order Domain)
    newOrder, _ := h.orderSvc.CreateOrder(ctx, userID, items)

    // 3. respond
    renderJSON(w, newOrder)
}
```