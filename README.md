# Notification System

## Features
- REST API for interaction with notifications.  
- Asynchronous processing notifications using Kafka.
- Status tracking.
- Retry mechanism.
- Graceful Shutdown.

## Tech Stack
Go, Gin, PostgreSQL, Apache Kafka, Docker

## Setup
1. Clone the repository:
   ```bash
   git clone https://github.com/Jduun/notification-system.git
   cd notification_system
   ```

2. Set environment variables:
   ```bash
   cp .env.example .env
   nano .env
   ```

3. Build application:
   ```bash
   docker-compose up -d
   ```
4. Run tests:
   ```bash
   go test -count=1 -v ./tests
   ```
