# Loan Tracker

A Loan Management System built with Go, utilizing Clean Architecture and MongoDB. This project provides functionalities for user management, loan application, and system logging.

## Features

- User registration, login, and password reset
- Loan application and status tracking
- Admin functionalities for loan management and user management
- System logging and viewing logs

## Architecture

This project follows the Clean Architecture principles with the following structure:

- **Delivery**: Handles HTTP requests and responses
  - `controller`: Contains HTTP handlers for user, loan, and log operations
  - `router`: Configures the routing and middleware
- **Infrastructure**: Handles external services and integrations
- **Repository**: Manages database interactions
  - `userRepository`, `loanRepository`, `logRepository`: Interfaces for data access
- **Usecase**: Contains business logic and application use cases
  - `userUsecase`, `loanUsecase`, `logUsecase`: Business logic for users, loans, and logs
- **Domain**: Defines the core business models and entities

## Installation

### Install Dependencies

Ensure you have Go installed on your machine. Then, run the following command to install the required dependencies:

```bash
go mod tidy
```
## Environment Variables

Create a `.env` file in the root of the project with the following content:

```dotenv
# MongoDB Connection URL
MONGO_URL

# JWT Secret Key
JWT_SECRET_KEY

# SMTP Configuration
SMTP_HOST=smtp.email.com
SMTP_PORT=587
SMTP_USERNAME=
SMTP_PASSWORD=
SMTP_FROM=
``` 
## Running the Application

### Load Environment Variables

Ensure you have a `.env` file with the necessary environment variables.

### Start the Server

Run the application using:

```bash
go run main.go
```
The server will start on port `8080`. You can change the port by modifying the `router.Run(":8080")` line in `main.go`.

## API Endpoints

### Public Routes

- **Register User**
  - `POST /users/register`
  - Request Body: JSON with user details

- **Login User**
  - `POST /users/login`
  - Request Body: JSON with login credentials

- **Refresh Token**
  - `POST /users/token/refresh`
  - Request Body: JSON with refresh token

- **Forgot Password**
  - `POST /users/password-reset`
  - Request Body: JSON with email

- **Reset Password**
  - `GET /users/password-reset/:token`
  - Query Parameter: Token for password reset

- **Verify Email**
  - `GET /users/verify-email/:token`
  - Query Parameter: Verification token

### Authenticated User Routes

- **Get User Profile**
  - `GET /users/profile/:id`
  - Requires authentication

- **Change Password**
  - `PUT /users/password-reset`
  - Requires authentication

### Loan Routes

- **Create Loan**
  - `POST /loans`
  - Requires authentication

- **View Loan Status**
  - `GET /loans/:id`
  - Requires authentication

### Admin Routes

- **View All Loans**
  - `GET /admin/loans`
  - Requires admin authentication

- **Approve/Reject Loan**
  - `PATCH /admin/loans/:id/status`
  - Requires admin authentication

- **Delete Loan**
  - `DELETE /admin/loans/:id`
  - Requires admin authentication

- **Get All Users**
  - `GET /admin/users`
  - Requires admin authentication

- **Delete User**
  - `DELETE /admin/users/:id`
  - Requires admin authentication

- **View Logs**
  - `GET /admin/logs`
  - Requires admin authentication


## Contact

For any questions or feedback, please reach out to [nebiyu](mailto:nebiyumusbah378@gmail.com).

