#!/bin/bash
# Run Opencoder directly from source using vite-node
cd "$(dirname "$0")/opencoder"
export GOOGLE_GENERATIVE_AI_API_KEY="${GOOGLE_GENERATIVE_AI_API_KEY}"
export OPENCODER_KTHULU=true
npx vite-node src/index.ts "$@"
