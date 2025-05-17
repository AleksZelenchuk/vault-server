# Project Documentation

## Overview

This project implements a secure **Vault Server** to manage user accounts and securely store sensitive information such as passwords and notes. It provides functionality for user authentication, account management, and CRUD operations for storing and retrieving encrypted entries in the vault. The project uses **gRPC** for communication, follows clean architecture principles, and ensures security best practices like password hashing and validation.

---

## Architecture

The project is organized into the following key components:

1. **Services** (`service` package):
   - Handles business logic for both user management and vault data operations.
   - Implements gRPC service interfaces that interact with storage.

2. **Storage** (`storage` package):
   - Manages database-level operations using SQL and `sqlx`.
   - Encryption and decryption happen here when interacting with sensitive data.

3. **Auth** (`auth` package):
   - Provides utilities for token generation and extracting user IDs from context.
   - Includes secure user session management (e.g., JWT-based authentication).

4. **Generated Protobuf** (`gen` directory):
   - Defines gRPC APIs for user services (`vaultuserpb`) and vault data services (`vaultpb`).

5. ```.env``` - should contain basic environment variables. Mandatory ones: "VAULT_MASTER_KEY" for encryption purposes, "DATABASE_URL" for db connection in postgresql e.g ```postgres://admin:password@host:port/vault-db```, "GRPC_PORT" for server

---

## Features

### 1. **User Management**
The application allows users to:
- **Register**: Create an account with secure password hashing using bcrypt.
- **Login**: Authenticate users with their credentials and issue JWT tokens.
- **Retrieve User Data**: Fetch user information via username.
- **Delete User**: Completely remove a user account.

### 2. **Vault Data Management**
Users can:
- **Create Entries**: Add vault entries such as passwords, usernames, and notes securely.
- **Retrieve Entries**: View specific vault entries, with decrypted sensitive data.
- **Delete Entries**: Remove entries from the vault based on the user ID and record ID.
- **List Entries**: Retrieve a list of vault entries by folder and filtering with specific tags.

### 3. **Security**
Key security features include:
- **Password Hashing**: Bcrypt is used to hash passwords securely before storing them in the database.
- **JWT Authentication**: Generates secure tokens for authenticated users.
- **Encryption/Decryption**: Vault entries' sensitive information like passwords are encrypted before storing in the database.
- **User Permission Validation**: Checks user access permissions for each operation.

---

## Core Functionalities

### User Service (`UserVaultService`)
Implements `VaultUserServiceServer` for managing users:
- **Register**: Validates user input, encrypts the password, and saves the user to the database.
- **Login**: Verifies user credentials using bcrypt and issues a JWT token.
- **Get User by Username**: Retrieves user info for a valid account.
- **Delete User**: Deletes a user account along with their data.

### Vault Service (`VaultService`)
Implements `VaultServiceServer` for managing vault entries:
- **Create Entry**: Encrypts sensitive data and stores it in the database.
- **Get Entry**: Decrypts and retrieves an individual vault entry by ID.
- **Delete Entry**: Deletes the entry if the user has permission.
- **List Entries**: Provides flexible filtering by folder or tags for listing entries.

---

## Database Schema

### Tables:
1. **users**
   - Contains user data such as username, email, and a bcrypt-hashed password.

2. **vault_entries**
   - Stores sensitive vault data associated with users.
   - Columns include `id`, `title`, `username`, `password` (encrypted), `notes`, `tags`, `folder`, and `user_id`.

---

## Authentication and Authorization

### JWT Authentication
- On successful login, a JWT token is issued with the user's ID as the claim.
- For all secured endpoints, the server validates the token and derives the user ID from the request context.

### Role-Based Authorization
- Users can only operate on entries they own, enforced using the `validateUserPermission` function.

---

## Installation and Setup

### Prerequisites
- **Go (>= 1.20)**
- **PostgreSQL**
- **Protobuf compiler (`protoc`)**

### Steps to Set Up
1. Clone the Repository:
   ```bash
   git clone <repository-url>
   cd <repository-directory>
   ```
2. Create a PostgreSQL database and apply the schema for `users` and `vault_entries`.
3. Install dependencies:
   ```bash
   go mod tidy
   ```
4. Run the application:
   ```bash
   go run main.go
   ```

---

## API Endpoints (gRPC Interface)

### User Service (`VaultUserService`)
#### Methods:
1. **Register(CreateUserRequest)**: Registers a new user.
2. **Login(LoginRequest)**: Authenticates a user and returns an access token.
3. **GetUserByUsername(GetUserRequest)**: Fetches user details by username.
4. **DeleteUser(DeleteUserRequest)**: Removes a user and associated data.

### Vault Service (`VaultService`)
#### Methods:
1. **CreateEntry(CreateEntryRequest)**: Adds a new entry to the user's vault.
2. **GetEntry(GetEntryRequest)**: Retrieves a specific vault entry.
3. **DeleteEntry(DeleteEntryRequest)**: Deletes an entry that the user owns.
4. **ListEntries(ListEntriesRequest)**: Lists all the user's entries with filtering.

---

## Error Handling

### Common Errors:
1. **Unauthenticated**: For endpoints that require valid user authentication.
2. **Permission Denied**: When an action is attempted on a resource owned by another user.
3. **Invalid Input**: For invalid or missing fields in user requests.
4. **Database Errors**: For errors at the database layer (e.g., connection failure, SQL issues).

---


---

## Future Enhancements

1. **Additional Encryption Algorithms**: Support for customizable encryption for entries.
2. **Audit Logs**: Track user activity for security purposes.
3. **Tag-based Search**: Improve querying by supporting tag-based search with pagination.
4. **Two-Factor Authentication (2FA)**: Add another layer of user security.

---
