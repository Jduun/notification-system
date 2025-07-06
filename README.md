# Notification System

## Features
- REST API for interaction with notifications.
- Asynchronous processing notifications using Kafka.
- Database stores notifications for status tracking.
- Graceful Shutdown.

## Tech Stack
Go, Gin, PostgreSQL, Apache Kafka, Docker

## Setup
1. **Clone the repository**:
   ```bash
   git clone https://github.com/Jduun/notification-system.git
   cd notification_system
   ```

2. **Configure environment**:
   Copy `.env.example` to `.env` and set variables.

3. **Build application**:
   ```bash
   docker-compose up -d
   ```
