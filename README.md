# Adivel AI Backend

This is the backend API service for the Adivel AI project, built using Go.

## Technologies Used

*   **Language:** Go (1.25+)
*   **HTTP Router:** go-chi/chi (v5)
*   **Database:** PostgreSQL
*   **Database Driver:** pgx (v5)
*   **Environment Management:** godotenv

## Architecture Layout

The application follows a standard Go project layout:
*   `cmd/api/`: Application entry point.
*   `internal/`: Private application code.
    *   `config/`: Configuration setup and environment loading.
    *   `handler/`: HTTP handlers and controllers.
    *   `model/`: Data structures and domain logic.
    *   `repository/`: Database interactions and queries.
    *   `service/`: Core business logic.
    *   `client/`: External API clients.
    *   `utils/`: Helper functions.
*   `sql/`: Database schema, migrations, or setup scripts.

## Prerequisites

*   Go 1.25.0 or higher
*   PostgreSQL running locally or remotely

## Getting Started

1.  **Install Go dependencies:**
    ```bash
    go mod tidy
    ```

2.  **Database Setup:**
    Ensure your PostgreSQL server is up and running. Create a database for the application.

3.  **Environment Variables:**
    Copy the `.env.example` file to `.env` and fill in your connection details (such as `DB_URL` or equivalent configurations).
    ```bash
    cp .env.example .env
    ```

4.  **Run the application locally:**
    ```bash
    go run cmd/api/main.go
    ```
    *(Note: Modify the `main.go` path if your entry point has a different name under `cmd/api/`)*
