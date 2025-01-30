# Kick Points

Kick Points is a project that allows users to earn and manage points through API endpoints. The project includes a backend server built with Go, a PostgreSQL database, and a WebSocket client to interact with the Kick platform.

## API Endpoints

### Get All Users with Points

- **Path:** `/points`
- **Method:** `GET`
- **Headers:**
  ```json
  {
    "Authorization": "Bearer YOUR_API_KEY"
  }
  ```
- **Response:**
  ```json
  {
    "users": [
      {
        "id": 1,
        "username": "user1",
        "points": 100,
        "updated_at": "2023-10-01T12:00:00Z",
        "created_at": "2023-10-01T12:00:00Z"
      }
    ]
  }
  ```

### Add Points to a User

- **Path:** `/points/add`
- **Method:** `POST`
- **Headers:**
  ```json
  {
    "Authorization": "Bearer YOUR_API_KEY"
  }
  ```
- **Request Body:**
  ```json
  {
    "username": "user1",
    "points": 10
  }
  ```
- **Response:**
  ```json
  {
    "newAmount": 110
  }
  ```

### Remove Points from a User

- **Path:** `/points/remove`
- **Method:** `POST`
- **Headers:**
  ```json
  {
    "Authorization": "Bearer YOUR_API_KEY"
  }
  ```
- **Request Body:**
  ```json
  {
    "username": "user1",
    "points": 10
  }
  ```
- **Response:**
  ```json
  {
    "success": true,
    "newAmount": 90
  }
  ```

### Get Points of a User

- **Path:** `/points/:username`
- **Method:** `GET`
- **Headers:**
  ```json
  {
    "Authorization": "Bearer YOUR_API_KEY"
  }
  ```
- **Response:**
  ```json
  {
    "points": 100
  }
  ```
