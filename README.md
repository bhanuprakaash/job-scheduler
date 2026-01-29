# Distributed Job Scheduler

A high-performance job scheduling system built in **Go**. This system is designed to manage, execute, and track asynchronous background tasks with reliability and concurrency control. It uses **gRPC** for low-latency communication, **PostgreSQL** for persistent state management, and a full observability stack with **Prometheus** and **Grafana**.

## Overview

This project serves as a comprehensive implementation of a distributed system, focusing on:
* **Decoupling:** Separating task submission (Client) from execution (Worker).
* **Reliability:** Ensuring jobs are persisted and recovered in case of failures.
* **Observability:** Providing real-time insights into job processing rates, worker health, and system latency.
* **Concurrency:** Utilizing Go's primitives (Goroutines, Channels) to process jobs in parallel without resource exhaustion.

## Tech Stack

* **Language:** Go (Golang)
* **Communication:** gRPC / Protocol Buffers
* **Database:** PostgreSQL
* **Observability:** Prometheus, Grafana
* **Object Storage:** MinIO (S3 Compatible)
* **CLI Framework:** Cobra

## Job Catalog

The system includes four job implementations to demonstrate different workload patterns:

### 1. PDF Invoice Generator
* **Type:** CPU & I/O Bound
* **Description:** Generates professional PDF invoices containing line items, tax calculations, and branding.
* **Workflow:** Receives JSON payload → Generates PDF using `gofpdf` → Uploads to S3/MinIO bucket `secure/invoices/`.

### 2. Email Notifications
* **Type:** Network Bound (Latency Sensitive)
* **Description:** Sends transactional emails to users (e.g., "Welcome", "Reset Password").
* **Workflow:** Integrates with 3rd party providers (e.g., Resend) to deliver HTML content reliably with retry logic.

### 3. Image Resizer
* **Type:** CPU Intensive
* **Description:** Downloads high-resolution images and resizes them for web optimization.
* **Workflow:** Fetches image from URL → Resizes using Catmull-Rom resampling → Encodes to JPEG → Uploads to Object Storage.

### 4. Data Archival
* **Type:** Maintenance / Batch Process
* **Description:** A system maintenance task that cleans up old database records.
* **Workflow:** Selects jobs older than X days → Exports them to a JSON file → Uploads archive to S3 → Deletes records from Primary DB.

## Key Features

* **gRPC API:** Strictly typed, high-performance API for job submission and management.
* **Worker Pools:** Configurable concurrency using the Fan-Out pattern to limit active goroutines.
* **Persistent State:** All job statuses (`PENDING`, `IN_PROGRESS`, `COMPLETED`, `FAILED`) are tracked in Postgres to survive restarts.
* **Real-time Monitoring:** Native instrumentation exposing metrics like `jobs_processed_total`, `job_duration_seconds`, and `active_workers`.
* **Graceful Shutdown:** Handles `SIGINT`/`SIGTERM` signals to finish active jobs before stopping the server.
* **CLI Tooling:** A developer-friendly CLI to submit jobs and query status.
