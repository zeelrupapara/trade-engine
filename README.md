# Trade Engine Service

## Local Setup of the gRPC Trade Engine Service

Follow these steps to set up the news service locally:

### 1. Install Dependencies
To install the required dependencies, run the following command:
```bash
go mod tidy
```

### 2. Create Environment Variables
Copy the example environment configuration file to create your `.env` file:
```bash
cp .env.example .env
```

### 3. Start the Database
If you have a database connection configured, set the necessary configurations in the `.env` file. Then, start the database with Docker:
```bash
docker-compose up -d
```

### 4. Apply Migrations
To apply database migrations, run:
```bash
go run main.go migrate up
```

### 5. Start the gRPC Server
Finally, start the gRPC server by running:
```bash
go run main.go grpc-server
```

Your local gRPC news service is now up and running!
Connection: in PORT 9000 default