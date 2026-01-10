package deployment

const K8sDeploymentTemplate = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{.AppName}}
  namespace: {{.Namespace}}
  labels:
    app: {{.AppName}}
spec:
  replicas: {{.Replicas}}
  selector:
    matchLabels:
      app: {{.AppName}}
  template:
    metadata:
      labels:
        app: {{.AppName}}
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8080"
    spec:
      containers:
      - name: {{.AppName}}
        image: {{.Image}}
        ports:
        - containerPort: 8080
        env:
        # - name: DB_HOST
        #   valueFrom:
        #     secretKeyRef:
        #       name: {{.AppName}}-secrets
        #       key: db-host
        resources:
          requests:
            cpu: "100m"
            memory: "128Mi"
          limits:
            cpu: "500m"
            memory: "512Mi"
`

const K8sServiceTemplate = `apiVersion: v1
kind: Service
metadata:
  name: {{.AppName}}
  namespace: {{.Namespace}}
spec:
  selector:
    app: {{.AppName}}
  ports:
    - protocol: TCP
      port: 80
      targetPort: 8080
  type: LoadBalancer
`

const K8sHPATemplate = `apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: {{.AppName}}-hpa
  namespace: {{.Namespace}}
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: {{.AppName}}
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
`

const DockerfileTemplate = `FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main %s

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]
`

const WasmerConfigTemplate = `[package]
name = "{{.AppName}}"
version = "0.1.0"
description = "Kthulu App on Wasmer Edge"

[[module]]
name = "server"
source = "build/app.wasm"
abi = "wasi"

[[command]]
name = "server"
module = "server"
runner = "https://webc.org/runner/wasi"
annotations = { "wasi.entrypoint" = ["server"] }

[app]
name = "{{.AppName}}"

[[app.service]]
name = "server"
type = "http"
port = 8080
command = "server"
`
