# Taskboard Cloud

A small cloud-native task management system built with a microservice architecture.

## This project demonstrates:
Python authentication service
Go task service
React frontend
PostgreSQL database
Docker-based local development
CI with GitHub Actions

The goal of this project is to showcase full-stack and DevOps skills in a clean, deployable architecture.

## Services:

### auth-service
Python + FastAPI
Handles user login and JWT generation.

### tasks-service
Go + Chi router
Handles CRUD operations for tasks.

### web
React + TypeScript + Vite
Frontend UI for login and task management.

### db
PostgreSQL 16

## Tech Stack

### Backend
Python (FastAPI)
Go (Chi router)
PostgreSQL

### Frontend
React
TypeScript
Vite

### DevOps
Docker & Docker Compose
GitHub Actions CI
Terraform (planned in Phase C)

## Local Development Prerequisites
Docker
Docker Compose

### 1. Configure environment
Create a .env file in the project root:

JWT_SECRET=dev-secret-change-me
POSTGRES_USER=taskboard
POSTGRES_PASSWORD=taskboard
POSTGRES_DB=taskboard

### 2. Start the stack
From the project root:

docker compose up --build

### 3. Open the app
Frontend:
http://localhost:5173


### Service endpoints:
Service	URL
Web frontend	http://localhost:5173
Auth service	http://localhost:8001/health
Tasks service	http://localhost:8002/health

## Features
User login with JWT authentication
Create, update, delete tasks
Tasks scoped per user
Multi-service architecture
Dockerized development environment
CI pipeline with automated tests and builds

## Running Tests (Local)
### Auth service (Python)
cd auth-service
pip install -r requirements.txt
pytest

### Tasks service (Go)
cd tasks-service-go
go test ./...

## CI Pipeline
### GitHub Actions automatically runs:
Python auth service checks
Go task service tests
Frontend build
Docker image builds

### Triggered on:
Push to main
Pull requests to main


## Purpose of This Project

### This project was built as a portfolio-ready cloud-native application to demonstrate:
Microservice design
Multi-language backend architecture
Containerization
CI/CD pipelines
Infrastructure as Code (Terraform)


## Local run
### 1) Create env file:
bash
cp .env.example .env
### 2) Build + run:
docker compose up -d --build

### 3) Open:
Web: http://localhost:5173
Auth health: http://localhost:8001/health
Tasks health: http://localhost:8002/health