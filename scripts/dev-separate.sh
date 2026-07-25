#!/bin/bash
# Start backend and frontend as separate processes
cd "$(dirname "$0")/.."

echo "Starting backend on http://localhost:3001..."
(cd backend && go run .) &
BACKEND_PID=$!

echo "Starting frontend on http://localhost:4321..."
(cd frontend && npm run dev) &
FRONTEND_PID=$!

echo ""
echo "Backend PID:  $BACKEND_PID"
echo "Frontend PID: $FRONTEND_PID"
echo ""
echo "Press Ctrl+C to stop both processes."

wait
