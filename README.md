# GoLang OTP Auth Backend Service

This is a backend service developed in Golang using the Gin framework. It implements OTP-based login and registration, along with basic user management features.

## Features

- OTP-based login & registration.
- Rate limiting for OTP requests (max 3 requests per phone number within 2 minutes).
- JWT-based authentication for protected endpoints.
- REST endpoints for user management (list users, get user by ID).
- Pagination and search for the user list.
- Containerized with Docker and Docker Compose.
- API documentation with Swagger/OpenAPI.

---

## Architecture & Design

The project follows a layered architecture to ensure a clean separation of concerns:

-   **Presentation Layer (`pkg/*/handler.go`, `internal/api/routes.go`):** Handles HTTP requests and responses using the Gin framework. It's responsible for input validation and calling the appropriate services.
-   **Application/Business Logic Layer (`pkg/*/service.go`):** Contains the core business logic. It orchestrates operations and is independent of the web framework and database.
-   **Data Access Layer (`pkg/*/repository.go`, `internal/database/*.go`):** Abstracted using the **Repository Pattern**. Interfaces define the contracts for data operations, and concrete implementations handle the interaction with the data store (currently in-memory). This makes it easy to switch to a persistent database like PostgreSQL in the future.

**Design Patterns Used:**
- **Dependency Injection:** Dependencies (like services and repositories) are created in `main.go` and injected into their consumers, promoting loose coupling and testability.
- **Repository Pattern:** Decouples the business logic from the data storage mechanism.
- **Middleware Pattern:** Used for cross-cutting concerns like JWT authentication and logging.

---

## Prerequisites

-   Go (version 1.24 or later)
-   Docker
-   Docker Compose

---

## How to Run Locally

1.  **Clone the repository:**
    ```bash
    git clone <your-repo-url>
    cd <your-repo-name>
    ```

2.  **Set up environment variables:**
    Copy the example environment file and edit it if necessary. The defaults are fine for local development.
    ```bash
    cp .env.example .env
    ```

3.  **Install dependencies:**
    ```bash
    go mod tidy
    ```

4. Before running the server locally, you might need to generate the Swagger files:
    
    Install swag if you haven't already
    ```bash
    go install github.com/swaggo/swag/cmd/swag@latest
    ```
    Generate the docs
    ```bash
    swag init -g ./cmd/app/main.go
    ```

5.  **Run the application:**
    ```bash
    go run ./cmd/app/main.go
    ```
    The server will start on `http://localhost:8080`.


6.  Generate/Update Swagger docs
    ```bash
    swag init -g ./cmd/app/main.go
    ```

---

## How to Run with Docker

The easiest way to run the application is with Docker Compose.

1.  **Build and run the container:**
    From the root of the project, run:
    ```bash
    go mod vendor
    docker-compose up --build
    ```
    The server will be available at `http://localhost:8080`.

2.  **To stop the service:**
    ```bash
    docker-compose down
    ```

---

## API Documentation

API documentation is generated using Swagger. Once the server is running, you can access the interactive Swagger UI at:

**[http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)**

![Swagger Screen Shot](/docs/swagger-screenshot.png "Swagger Screen Shot")
