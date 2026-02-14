# Taskboard Cloud
A cloud-native task management system built with a microservice architecture and deployed on AWS using Terraform.

## This project demonstrates a full DevOps-style workflow:
- Multi-language microservices (Python + Go)
- Containerized services with Docker
- CI pipelines with GitHub Actions
- Infrastructure as Code with Terraform
- Real cloud deployment on AWS (ECS, ALB, RDS)

## Architecture
          User (Browser)
                │
                ▼
     React Frontend (local or static host)
                 │
                 ▼
     AWS Application Load Balancer (ALB)
          │                  │
          ▼                  ▼
     Auth Service         Tasks Service
     (Python FastAPI)     (Go + Chi)
          │                  │
          └───────┬──────────┘
                  ▼
           PostgreSQL (AWS RDS)

## Services
### auth-service
- Python + FastAPI
- Handles:
    - User registration
    - Login
    - JWT generation

- Runs:
    - Locally via Docker
    - In AWS ECS Fargate (container)

### tasks-service
- Go + Chi router
- Handles:
    - Task CRUD operations
    - JWT-protected endpoints
    - Per-user task isolation
- Runs:
    - Locally via Docker
    - In AWS ECS Fargate (container)

### web
- React + TypeScript + Vite
- Provides:
    - Login/register UI
    - Task management interface
- Runs:
- Locally via Docker
- Can connect to:
    - Local backend
    - AWS cloud backend

### db
- PostgreSQL 16
- Runs:
    - Locally: Docker container
    - Cloud: AWS RDS instance

## Tech Stack
### Backend
- Python (FastAPI)
- Go (Chi router)
- PostgreSQL

### Frontend
- React
- TypeScript
- Vite

### DevOps / Cloud
- Docker & Docker Compose
- GitHub Actions (CI)
- Terraform (Infrastructure as Code)
- AWS:
    - ECS Fargate
    - Application Load Balancer (ALB)
    - RDS (PostgreSQL)
    - ECR (container registry)
    - IAM roles

## Features
- User registration and login
- JWT authentication
- Per-user task isolation
- Create, update, delete tasks
- Microservice architecture
- Dockerized local environment
- Automated CI pipeline
- Cloud deployment with Terraform

## Local Development
### Prerequisites
- Docker
- Docker Compose

### 1. Create environment file
Create a .env file in the project root:
JWT_SECRET=dev-secret-change-me
POSTGRES_USER=taskboard
POSTGRES_PASSWORD=taskboard
POSTGRES_DB=taskboard

Or copy:
cp .env.example .env

### 2. Start the stack
From the project root:
docker compose up --build

### 3. Open the app
Frontend:
http://localhost:5173

## Local Service Endpoints
Web frontend	http://localhost:5173

Auth service	http://localhost:8001/health

Tasks service	http://localhost:8002/health

## Using Local Frontend with Cloud Backend
The frontend can connect directly to the AWS backend.
In docker-compose.yml:
web:
  environment:
    VITE_AUTH_URL: "http://<ALB-DNS>"
    VITE_TASKS_URL: "http://<ALB-DNS>"


Then run:
docker compose up -d --build web


Open:
http://localhost:5173

## Running Tests Locally
### Auth service (Python)
cd auth-service
pip install -r requirements.txt
pytest

### Tasks service (Go)
cd tasks-service-go
go test ./...

## CI Pipeline
GitHub Actions automatically runs:
- Python auth service checks
- Go tasks service tests
- Frontend build
- Docker image builds

### Triggered on:
- Push to main
- Pull requests to main

## Cloud Deployment (AWS)
All infrastructure is defined in:
infra/

Provisioned using **Terraform**.

## AWS Services Used
ECS Fargate | Runs auth and tasks containers
Application Load Balancer | Routes /auth/* and /tasks/*
RDS PostgreSQL | Production database
ECR | Container image registry
IAM | Service permissions

## Deploy Infrastructure
From the infra/ directory:

terraform init
terraform apply

### Terraform provisions:
- VPC resources (default VPC)
- ECS cluster
- ALB with routing rules
- RDS PostgreSQL instance
- Security groups
- ECR repositories
- IAM roles

## Accessing the Cloud API
### After deployment, Terraform outputs:

alb_dns_name


### Example:
http://taskboard-cloud-alb-xxxx.eu-west-1.elb.amazonaws.com


### Endpoints:
/auth/register | Register user
/auth/login | Login
/tasks/ | Task CRUD (JWT required)

## Project Structure
- auth-service/        Python FastAPI auth service
- tasks-service-go/    Go tasks service
- web/                 React frontend
- infra/               Terraform AWS infrastructure
- docker-compose.yml   Local dev stack

## Purpose of This Project
This project was built as a portfolio-ready cloud-native system to demonstrate:
- Microservice architecture
- Multi-language backend services
- Containerization
- CI pipelines
- Infrastructure as Code
- Real AWS production deployment
