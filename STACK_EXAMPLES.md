# Stack Examples

> Practical reference snippets for building new repositories on top of `Soku-Convention-Boilerplate`.

## How to Read This Document

These examples are intentionally small.  
They are not meant to define full production architecture. They are meant to show the expected tone of code:

- readable first
- explicit over magical
- easy to review
- easy to extend

## JavaScript

### Example: clear service function

```javascript
/**
 * Builds a public profile payload from a user record.
 * Returns only fields that are safe for external exposure.
 */
function buildPublicProfile(user) {
  if (!user) {
    throw new Error("User is required.");
  }

  return {
    id: user.id,
    email: user.email,
    displayName: user.displayName,
    createdAt: user.createdAt,
  };
}
```

## TypeScript

### Example: explicit DTO mapping

```typescript
type User = {
  id: string;
  email: string;
  displayName: string;
  isActive: boolean;
};

type UserResponse = {
  id: string;
  email: string;
  displayName: string;
};

export function toUserResponse(user: User): UserResponse {
  return {
    id: user.id,
    email: user.email,
    displayName: user.displayName,
  };
}
```

## Node.js

### Example: route with narrow responsibility

```javascript
import express from "express";

const app = express();

app.get("/health", (_request, response) => {
  response.status(200).json({
    status: "ok",
    service: "user-api",
  });
});

app.listen(3000, () => {
  console.log("Server started on port 3000");
});
```

## Python

### Example: readable domain logic

```python
from dataclasses import dataclass


@dataclass
class Order:
    subtotal: float
    shipping_fee: float


def calculate_total(order: Order) -> float:
    """Returns the final payable amount for the order."""
    if order.subtotal < 0:
        raise ValueError("subtotal must not be negative")

    return order.subtotal + order.shipping_fee
```

## Go

### Example: small, explicit function

```go
package user

import "errors"

type Profile struct {
  ID          string
  DisplayName string
}

func ValidateProfile(profile Profile) error {
  if profile.ID == "" {
    return errors.New("id is required")
  }

  if profile.DisplayName == "" {
    return errors.New("display name is required")
  }

  return nil
}
```

## Java

### Example: focused utility method

```java
public final class EmailValidator {

  private EmailValidator() {}

  public static boolean isValid(String email) {
    if (email == null || email.isBlank()) {
      return false;
    }

    return email.contains("@") && email.contains(".");
  }
}
```

## Spring

### Example: controller with explicit dependencies

```java
@RestController
@RequestMapping("/api/users")
public class UserController {

  private final UserService userService;

  public UserController(UserService userService) {
    this.userService = userService;
  }

  @GetMapping("/{id}")
  public ResponseEntity<UserResponse> getUser(@PathVariable String id) {
    UserResponse response = userService.getUserById(id);
    return ResponseEntity.ok(response);
  }
}
```

## MySQL

### Example: schema with clear constraints

```sql
CREATE TABLE users (
  id BIGINT NOT NULL AUTO_INCREMENT,
  email VARCHAR(255) NOT NULL,
  display_name VARCHAR(100) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_users_email (email)
);
```

## PostgreSQL

### Example: explicit table design

```sql
CREATE TABLE users (
  id BIGSERIAL PRIMARY KEY,
  email TEXT NOT NULL UNIQUE,
  display_name TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

## gcloud

### Example: readable deployment command

```bash
gcloud run deploy user-api \
  --source . \
  --region asia-northeast1 \
  --platform managed \
  --allow-unauthenticated
```

### Example: explicit configuration lookup

```bash
gcloud config list
gcloud projects list
```

## What These Examples Are Optimizing For

These examples are intentionally biased toward:

- low surprise
- clear naming
- stable patterns
- easy onboarding
- AI-friendly inference

If a shorter or more advanced pattern makes the code harder to understand at first glance, prefer the clearer version.

## Recommended Next Step

As this boilerplate evolves, use [STACK_CONFIGS.md](./STACK_CONFIGS.md) for copyable starter files and keep this document focused on code-shape examples, such as:

- `eslint` and `prettier` for JavaScript and TypeScript
- `ruff` for Python
- `golangci-lint` for Go
- `checkstyle` or `spotless` for Java
- migration and query conventions for MySQL and PostgreSQL
- deployment checklists for `gcloud`
