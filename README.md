# Tutorial: Securing a REST API with JWT (JSON Web Tokens)

**Topic:** Securing a REST API
**Repository:** [FleetCommander](https://github.com/ATchibo/FleetCommander)
**Difficulty:** Beginner / Intermediate

---

## 1. Introduction

In modern microservices architecture, we cannot rely on traditional "Session Cookies" because backend services often live on different domains, servers, or containers. Instead, we use **JWT (JSON Web Tokens)**.

Think of a JWT as a **Digital Hotel Key Card**:
1.  **Login:** You show your ID (Username/Password) to the Receptionist (Auth Service).
2.  **Token Issue:** They give you a Key Card (JWT) signed with a secret stamp.
3.  **Access:** To enter your room (Protected API), you just show the Key Card. The door doesn't need to call the receptionist; it just checks if the signature is valid.

In this tutorial, we will build a standalone **"Bank API"** using **Go (Golang)**. It has one public endpoint (`/login`) and one private endpoint (`/balance`).

---

## 2. Prerequisites

* **Go** installed (`go version` >= 1.20)
* **Terminal** (Bash, Zsh, or PowerShell)
* **Postman** or `curl` for testing.

---

## 3. The Implementation

We will create a single file `main.go` that acts as both the Auth Server (Issuer) and the Resource Server (Verifier).

### Step 1: Project Setup

Create a folder and initialize the Go module:

```bash
mkdir jwt-tutorial
cd jwt-tutorial
go mod init jwt-tutorial

# Install Fiber (Web Framework) and JWT Middleware
go get [github.com/gofiber/fiber/v2](https://github.com/gofiber/fiber/v2)
go get [github.com/gofiber/jwt/v3](https://github.com/gofiber/jwt/v3)
go get [github.com/golang-jwt/jwt/v4](https://github.com/golang-jwt/jwt/v4)
```

### Step 2: The Code (`main.go`)

Create a file named `main.go` and paste the following code.

```go
package main

import (
    "time"
    "[github.com/gofiber/fiber/v2](https://github.com/gofiber/fiber/v2)"
    jwtware "[github.com/gofiber/jwt/v3](https://github.com/gofiber/jwt/v3)"
    "[github.com/golang-jwt/jwt/v4](https://github.com/golang-jwt/jwt/v4)"
)

// ⚠️ WARNING: In production, store this in an Environment Variable!
const SECRET_KEY = "super-secret-key-123"

func main() {
    app := fiber.New()

    // --- 1. PUBLIC ROUTES (Open to everyone) ---
    app.Post("/login", login)

    // --- 2. MIDDLEWARE (The Security Guard) ---
    // Any route registered BELOW this line requires a valid Token.
    // The middleware automatically checks the "Authorization: Bearer <token>" header.
    app.Use(jwtware.New(jwtware.Config{
        SigningKey: []byte(SECRET_KEY),
        ErrorHandler: func(c *fiber.Ctx, err error) error {
            return c.Status(401).JSON(fiber.Map{
                "error": "Unauthorized: Invalid or Missing Token",
            })
        },
    }))

    // --- 3. PROTECTED ROUTES ( VIP Only ) ---
    app.Get("/balance", getBalance)

    app.Listen(":3000")
}

// Handler: Validates credentials and issues the JWT
func login(c *fiber.Ctx) error {
    type LoginRequest struct {
        User string `json:"user"`
        Pass string `json:"pass"`
    }
    var req LoginRequest
    c.BodyParser(&req)

    // A. Verify Credentials (Mock Database Check)
    // In a real app, you would check SQL/Mongo here.
    if req.User == "admin" && req.Pass == "password123" {
        
        // B. Create the Claims ( The Data inside the token )
        claims := jwt.MapClaims{
            "name":  "John Doe",
            "admin": true,
            "exp":   time.Now().Add(time.Hour * 72).Unix(), // Expires in 72 hours
        }

        // C. Create token
        token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

        // D. Sign token with our secret key
        t, _ := token.SignedString([]byte(SECRET_KEY))

        return c.JSON(fiber.Map{"token": t})
    }

    return c.Status(401).JSON(fiber.Map{"error": "Bad Credentials"})
}

// Handler: Only runs if the JWT is valid
func getBalance(c *fiber.Ctx) error {
    // The middleware has already parsed the token and put it in Locals("user")
    user := c.Locals("user").(*jwt.Token)
    claims := user.Claims.(jwt.MapClaims)
    
    // We can trust this data because the signature matched
    name := claims["name"].(string)

    return c.JSON(fiber.Map{
        "user":    name,
        "balance": "$1,000,000",
        "status":  "Access Granted",
    })
}
```

---

## 4. Testing the Security

We will use `curl` to prove the security works.

### Scenario A: The Intruder (No Token)
Try to access the balance without logging in.

```bash
curl http://localhost:3000/balance
```
**Result:**
```json
{"error": "Unauthorized: Invalid or Missing Token"}
```
> **Success:** The middleware blocked the request because no key was provided.

### Scenario B: Getting the Key (Login)
Let's log in as the admin user.

```bash
curl -X POST -H "Content-Type: application/json" \
     -d '{"user":"admin", "pass":"password123"}' \
     http://localhost:3000/login
```

**Result:**
```json
{"token": "eyJhGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."}
```
> **Action:** Copy that long string starting with `ey...`. This is your signed JWT.

### Scenario C: The Authorized User (With Token)
Now we access the balance again, but this time we attach the token in the `Authorization` header.

```bash
# Replace <YOUR_TOKEN> with the actual string you copied
curl -H "Authorization: Bearer <YOUR_TOKEN>" http://localhost:3000/balance
```

**Result:**
```json
{
  "user": "John Doe",
  "balance": "$1,000,000",
  "status": "Access Granted"
}
```

---

## 5. How it Works (The Theory)



A JWT consists of three parts separated by dots (`.`):

1.  **Header:** Tells the server "I am a JWT signed with the HS256 algorithm".
2.  **Payload:** Contains the user data (`name: John Doe`). This part is readable by anyone (Base64 encoded), so **never put passwords here**.
3.  **Signature:** This is the cryptographic seal.
    * `Signature = Hash(Header + Payload + SECRET_KEY)`

When the server receives the token, it takes the Header and Payload, adds its own private `SECRET_KEY`, and re-calculates the hash. If the calculated hash matches the Signature on the token, the server knows:
* The token was created by us (Authenticity).
* The data hasn't been changed by a hacker (Integrity).

## 6. Conclusion

This tutorial demonstrated the standard security pattern used in distributed systems (and in our **FleetCommander** project). By implementing JWTs, we made our API **stateless**—the server doesn't need to use memory to remember who is logged in, allowing us to scale to millions of users easily.
