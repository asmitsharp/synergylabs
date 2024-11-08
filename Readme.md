# Synergy Labs API

## Overview

The Synergy Labs API is a job application platform that allows users to sign up, log in, upload resumes, and apply for jobs. The API is built using Go and utilizes the Echo framework for routing, GORM for database interactions, and Redis for caching.

## API Routes

### User Routes

- **POST /signup**

  - **Request Body:**
    ```json
    {
      "name": "John Doe",
      "email": "john.doe@example.com",
      "address": "123 Main St, Anytown, USA",
      "user_type": "APPLICANT",
      "password_hash": "securepassword",
      "profile_headline": "Aspiring Software Engineer"
    }
    ```
  - **Description:** Registers a new user. Requires the following fields:
    - `name`: The full name of the user.
    - `email`: The email address (must be unique).
    - `address`: The physical address of the user.
    - `user_type`: The type of user (e.g., "APPLICANT" or "ADMIN").
    - `password_hash`: The hashed password for authentication.
    - `profile_headline`: A brief headline for the user's profile.

- **POST /login**
  - **Request Body:**
    ```json
    {
      "email": "john.doe@example.com",
      "password": "securepassword"
    }
    ```
  - **Description:** Authenticates a user and returns a JWT token.

### Resume Routes

- **POST /uploadResume**
  - **Request Body:** Form-data with a file field named `resume`.
  - **Description:** Uploads a resume for the authenticated user.

### Job Routes

- **POST /admin/job**

  - **Request Body:**
    ```json
    {
      "title": "Software Engineer",
      "description": "Job description here."
    }
    ```
  - **Description:** Creates a new job posting. Admin access required.

- **GET /admin/job/:job_id**

  - **Description:** Retrieves job details along with applicants. Admin access required.

- **GET /admin/applicants**

  - **Description:** Retrieves all applicants. Admin access required.

- **GET /admin/applicant/:applicant_id**
  - **Description:** Retrieves specific applicant data. Admin access required.

### Public Job Routes

- **GET /jobs**

  - **Description:** Retrieves job openings. Requires authentication.

- **GET /jobs/apply**
  - **Request Query Parameters:**
    - `job_id`: The ID of the job to apply for.
  - **Description:** Applies to a job. Requires authentication.

## Important Business Logic

1. **User Registration and Authentication:**

   - Passwords are hashed using bcrypt before being stored in the database.
   - JWT tokens are generated upon successful login for session management.

2. **Resume Processing:**

   - Resumes are uploaded and sent to a third-party API for parsing.
   - Parsed data is saved in the user's profile in the database.

3. **Job Applications:**
   - Users can apply to jobs, and the application is tracked in the database.
   - The total number of applications for each job is updated accordingly.

## Running the Project Locally

### Prerequisites

- Go (1.23 or higher)
- Docker and Docker Compose
- PostgreSQL
- Redis

### Steps to Run

1. **Clone the Repository:**

   ```bash
   git clone <repository-url>
   cd synergylabs
   ```

2. **Set Up Environment Variables:**
   Create a `.env` file in the root directory with the following content:

   ```
   DATABASE_URL=postgres://user:password@db:5432/synergylabs?sslmode=disable
   REDIS_ADDR=redis:6379
   JWT_SECRET=your_jwt_secret_key
   ```

3. **Build and Run with Docker Compose:**

   ```bash
   docker-compose up --build
   ```

4. **Access the API:**
   The API will be available at `http://localhost:3000`.

### Testing the API

You can use tools like Postman or curl to test the API endpoints. Make sure to include the JWT token in the Authorization header for protected routes.

## Conclusion

This API provides a robust platform for job applications, allowing users to manage their profiles and apply for jobs efficiently. The use of JWT for authentication and Redis for caching enhances the performance and security of the application.
