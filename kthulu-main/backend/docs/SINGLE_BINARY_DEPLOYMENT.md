# Kthulu Single Binary Deployment

Kthulu can be deployed as a single binary that serves both the API and the frontend application. This approach simplifies deployment, reduces infrastructure complexity, and provides excellent performance.

## 🚀 **Quick Start**

### Build and Run Locally

```bash
# Build everything (frontend + backend) into a single binary
make build-fullstack

# Run the application
cd backend
./kthulu-app
```

The application will be available at:

- **Frontend**: http://localhost:8080
- **API**: http://localhost:8080/api/\*
- **Health Check**: http://localhost:8080/health
- **API Docs**: http://localhost:8080/docs

## 🏗️ **How It Works**

### Architecture

```
┌─────────────────────────────────────────┐
│           Single Go Binary              │
├─────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────────┐   │
│  │   Static    │  │   API Routes    │   │
│  │   Handler   │  │   /api/*        │   │
│  │   /*        │  │   /health       │   │
│  │             │  │   /docs         │   │
│  └─────────────┘  └─────────────────┘   │
├─────────────────────────────────────────┤
│           Chi Router                    │
├─────────────────────────────────────────┤
│           HTTP Server                   │
└─────────────────────────────────────────┘
```

### Request Flow

1. **API Requests** (`/api/*`, `/health`, `/docs`) → API handlers
2. **Static Assets** (`/assets/*`, `/favicon.ico`) → File server
3. **SPA Routes** (everything else) → `index.html` (React Router)

### Build Process

1. **Frontend Build**: Vite compiles React app to `backend/public/`
2. **Backend Build**: Go compiles binary with embedded static file server
3. **Result**: Single binary + `public/` directory

## 📦 **Deployment Options**

### Option 1: Direct Binary Deployment

```bash
# Build for production
make build-prod

# Deploy files
scp backend/kthulu-app user@server:/opt/kthulu/
scp -r backend/public user@server:/opt/kthulu/

# Run on server
cd /opt/kthulu
./kthulu-app
```

### Option 2: Docker Deployment

```bash
# Build Docker image
docker build -f Dockerfile.fullstack -t kthulu-app .

# Run with docker-compose
docker-compose -f docker-compose.prod.yml up -d
```

### Option 3: Cloud Platform Deployment

#### Heroku

```bash
# Create Heroku app
heroku create my-kthulu-app

# Set buildpacks
heroku buildpacks:add heroku/nodejs
heroku buildpacks:add heroku/go

# Deploy
git push heroku main
```

#### Fly.io

```bash
# Initialize Fly app
fly launch

# Deploy
fly deploy
```

#### Railway

```bash
# Connect to Railway
railway login
railway init

# Deploy
railway up
```

## ⚙️ **Configuration**

### Environment Variables

```bash
# Server
HTTP_ADDR=:8080
ENV=production
RATE_LIMIT_RPS=10
RATE_LIMIT_BURST=20

# Database
DATABASE_URL=postgres://user:pass@host:5432/dbname

# JWT
JWT_SECRET=your-secret-key
JWT_REFRESH_SECRET=your-refresh-secret

# Modules (customize as needed)
MODULES=health,auth,user,access,notifier,organization,contact,product,invoice,static

# SMTP (optional)
SMTP_ENABLED=true
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
```

### Module Configuration

The `static` module is automatically included in core modules. You can disable it by excluding it from the `MODULES` environment variable if you want to serve the frontend separately.

```bash
# Without static module (API only)
MODULES=health,auth,user,access,notifier,organization,contact,product,invoice

# With static module (fullstack)
MODULES=health,auth,user,access,notifier,organization,contact,product,invoice,static
```

## 🔧 **Development Workflow**

### Development Mode

For development, you can run frontend and backend separately:

```bash
# Terminal 1: Backend
cd backend
go run ./cmd/service

# Terminal 2: Frontend (with proxy)
cd frontend
npm run dev
```

The Vite dev server will proxy API requests to the backend.

### Production Build

```bash
# Full production build
make build-fullstack

# Or step by step
make build-frontend  # Build React app
make build-backend   # Build Go binary
```

## 📊 **Performance Benefits**

### Single Binary Advantages

- **Reduced Latency**: No network calls between frontend and backend
- **Simplified Caching**: Static assets served with optimal cache headers
- **Lower Resource Usage**: Single process vs multiple containers
- **Easier Scaling**: Scale entire application as one unit

### Benchmarks

Typical performance characteristics:

- **Binary Size**: ~15-25MB (including frontend assets)
- **Memory Usage**: ~50-100MB at startup
- **Cold Start**: <100ms
- **Static Asset Serving**: <1ms response time

## 🛡️ **Security Considerations**

### Static File Security

The static handler includes several security features:

- **Directory Traversal Protection**: Prevents `../` attacks
- **File Existence Validation**: Returns 404 for missing files
- **Appropriate Cache Headers**: Optimizes performance while maintaining security
- **Content Type Detection**: Proper MIME types for all assets

### API Route Protection

- API routes are protected by authentication middleware
- Static routes bypass authentication (as expected for public assets)
- Health and docs endpoints have appropriate access controls

## 🚀 **Deployment Examples**

### Systemd Service

```ini
# /etc/systemd/system/kthulu.service
[Unit]
Description=Kthulu Application
After=network.target

[Service]
Type=simple
User=kthulu
WorkingDirectory=/opt/kthulu
ExecStart=/opt/kthulu/kthulu-app
Restart=always
RestartSec=5
Environment=DATABASE_URL=postgres://...
Environment=JWT_SECRET=...

[Install]
WantedBy=multi-user.target
```

### Nginx Reverse Proxy (Optional)

If you need SSL termination or load balancing:

```nginx
server {
    listen 80;
    server_name yourdomain.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### Docker Compose with SSL

```yaml
version: "3.8"
services:
  kthulu-app:
    build:
      context: .
      dockerfile: Dockerfile.fullstack
    environment:
      HTTP_ADDR: :8080
      # ... other env vars

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
      - ./ssl:/etc/ssl/certs
    depends_on:
      - kthulu-app
```

## 🔍 **Troubleshooting**

### Common Issues

#### Frontend Not Loading

- **Symptom**: API works but frontend shows "Frontend Not Built"
- **Solution**: Run `make build-frontend` or `make build-fullstack`

#### Static Assets 404

- **Symptom**: CSS/JS files return 404
- **Solution**: Ensure `public/` directory exists and contains built assets

#### API Routes Conflicting with Frontend Routes

- **Symptom**: API endpoints return HTML instead of JSON
- **Solution**: Ensure API routes are registered before static handler

### Debug Mode

Set environment variable for detailed logging:

```bash
ENV=development ./kthulu-app
```

This will show detailed logs including static file serving and route matching.

## 📈 **Monitoring**

### Health Checks

The application includes comprehensive health checks:

```bash
# Check application health
curl http://localhost:8080/health

# Response
{
  "status": "healthy",
  "timestamp": "2024-03-15T10:30:00Z",
  "version": "1.0.0",
  "database": "connected",
  "modules": ["health", "auth", "user", "static"]
}
```

### Metrics

Access Prometheus metrics at `/metrics` (if enabled):

```bash
curl http://localhost:8080/metrics
```

## 🎯 **Best Practices**

### Production Deployment

1. **Use Environment Variables**: Never hardcode secrets
2. **Enable HTTPS**: Use reverse proxy or cloud load balancer
3. **Set Up Monitoring**: Health checks, logs, and metrics
4. **Database Migrations**: Run migrations before deployment
5. **Graceful Shutdown**: Application handles SIGTERM properly

### Performance Optimization

1. **Static Asset Caching**: Assets are cached for 1 year
2. **Compression**: Enable gzip compression in reverse proxy
3. **Database Connection Pooling**: Configured automatically
4. **Resource Limits**: Set appropriate memory/CPU limits

### Security

1. **Regular Updates**: Keep dependencies updated
2. **Secure Headers**: Add security headers in reverse proxy
3. **Rate Limiting**: Implement rate limiting for API endpoints
4. **Input Validation**: All inputs are validated and sanitized

---

**The single binary deployment approach makes Kthulu incredibly easy to deploy while maintaining enterprise-grade performance and security.** 🚀
