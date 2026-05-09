Project Brief: LuminaLog (High-Performance Log Aggregator)
1. High-Level Goal
Build a high-performance log ingestion system using Golang and PostgreSQL. The system must be capable of receiving high-frequency JSON log data from multiple concurrent sources, buffering them in memory, and performing optimized batch writes to a partitioned database.
2. The User Persona & Motivation
The developer is a Senior Sysadmin/DBA transitioning to Backend Engineering. The project must emphasize concurrency, database optimization, and system reliability rather than just basic CRUD operations. It serves as a portfolio piece to demonstrate "operational-aware" coding.
3. Architecture & Tech Stack
•	Language: Go (Golang) using standard library concurrency primitives (Goroutines, Channels, Select).
•	Database: PostgreSQL 16+ (Focusing on JSONB, B-Tree/GIN indexing, and Table Partitioning).
•	Driver: pgx (for high-performance connection pooling and batching).
•	Infrastructure: Docker & Docker Compose for simulating a distributed microservice environment.
4. Core Features & Requirements
•	Ingestion API: A REST endpoint (POST /logs) that accepts JSON payloads: { "service": string, "level": string, "message": string, "data": jsonb, "timestamp": string }.
•	Concurrency Pattern: Implement a Worker Pool. The API should drop logs into a buffered channel. A background worker should collect logs and execute Batch Inserts (Multi-value inserts or COPY) every 500ms or when the buffer reaches 100 items.
•	Database Optimization:
•	Use Declarative Partitioning by time (e.g., daily partitions).
•	Implement GIN indexing on the JSONB field for fast metadata searching.
•	Reliability: Implement Graceful Shutdown using os/signal. The application must flush all in-memory logs to the DB before exiting upon receiving a SIGTERM.
•	Simulation: A separate "Log Spammer" utility in Go to simulate 10+ microservices sending concurrent traffic via Docker Compose.
5. Guidance for the LLM
•	Code Style: Prioritize clean, idiomatic Go. Avoid over-engineering with heavy frameworks; lean on the standard library where possible.
•	DBA Focus: When writing SQL, suggest optimizations, indexing strategies, and explain "why" specific Postgres features are used.
•	Project Structure: Follow a standard Go project layout (e.g., /cmd, /internal, /pkg).



To test a high-throughput system like a log aggregator, you need to simulate "synthetic load." You want to move beyond manual curl commands and create a Load Generator—which is actually a second mini-project that looks great on a resume.
Here is how you can mimic a real microservice environment to stress-test your backend.
1. Build a "Log Spammer" (The Load Generator)
Create a separate Go script (or a sub-module in your project) specifically designed to generate traffic. This script should:
•	Use Goroutines: Spin up 10, 50, or 100 concurrent routines.
•	Generate Random Data: Use a library like gofakeit to generate random service names (e.g., "auth-service", "payment-gateway"), log levels (INFO, WARN, ERROR), and messages.
•	Control Velocity: Use a time.Ticker to send a burst of logs every 100ms.
2. The Testing Strategy
To prove your system works for a job switch, you need to test three specific scenarios:
A. The "Happy Path" (Functional Testing)
Send a single log and verify it appears in Postgres.
•	Tool: Postman or curl.
•	What to check: Does the created_at timestamp in the DB match the actual time? Is the JSONB data searchable?
B. The "Burst" (Concurrency Testing)
This is where your Sysadmin/DBA skills shine. Use your Log Spammer to send 10,000 logs as fast as possible.
•	The Goal: Observe how your Go application handles memory.
•	The DBA Angle: Check the Postgres pg_stat_activity. Are there too many open connections? This is where you'll realize you need a Connection Pool and Batch Inserts.
C. The "Graceful Shutdown" (Reliability Testing)
This is a classic senior-level interview topic.
	1.	Start your Load Generator so logs are flowing.
	2.	Send a Ctrl+C (SIGINT) to your Go backend.
	3.	The Test: Your app should stop accepting new logs but finish writing the logs currently in its internal buffer/channel to Postgres before actually turning off.
	4.	Verification: Count the logs in the generator vs. the logs in the DB. They should match perfectly (zero data loss).
3. Recommended Tools for Testing
Tool	Purpose	Why for this project?
k6 (by Grafana)	Load Testing	It uses JavaScript and is excellent for simulating thousands of VUs (Virtual Users) hitting your API.
Docker Compose	Environment	Run your Go App, Postgres, and a "Mock Service" in separate containers to simulate a network environment.
Table Driven Tests	Unit Testing	Use Go’s built-in testing package to test your log-parsing logic with various edge cases (empty strings, huge payloads).




1. The Workflow: A Simplified View
Think of your project like a high-speed funnel.
	1.	The Producers (Microservices): These are just scripts that "talk" to your API. They send a JSON blob every time something happens (e.g., {"level": "error", "message": "Connection failed", "service": "auth"}).
	2.	The Gateway (Your Go API): This is the front door. It receives the JSON. Instead of going straight to the database (which is slow), it puts the message into a Go Channel (a fast, in-memory queue).
	3.	The Worker (Your Logic): A background routine in Go watches that channel. It waits until it has, say, 100 logs or 5 seconds have passed, and then it does one big "Batch Insert" into Postgres.
	4.	The Storage (Postgres): Your database stores the logs. Because you are a DBA, you’ll set up "Partitioning" so that logs from Monday are in one table and logs from Tuesday are in another. This keeps queries lightning fast.