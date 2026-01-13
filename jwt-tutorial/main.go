package main

import (
    "time"
    "github.com/gofiber/fiber/v2"
    jwtware "github.com/gofiber/jwt/v3"
    "github.com/golang-jwt/jwt/v4"
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