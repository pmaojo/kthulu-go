#!/bin/bash

# Start the server in background
cd backend
go run cmd/service/main.go &
SERVER_PID=$!

# Wait for server to start
sleep 3

# Test the register endpoint
echo "Testing register endpoint..."
curl -v -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'

# Kill the server
kill $SERVER_PID