# IoT Implementation Tasks

## Phase 1: Module Scaffold & Domain
- [ ] Create directory structure `backend/backend/internal/modules/iot`.
- [ ] Define `Device` struct in `internal/modules/iot/domain/device.go`.
- [ ] Define `Telemetry` struct in `internal/modules/iot/domain/telemetry.go`.
- [ ] Define `Command` struct in `internal/modules/iot/domain/command.go`.
- [ ] Create Repository interfaces in `internal/modules/iot/domain/repository.go`.

## Phase 2: Persistence (Database)
- [ ] Implement `DeviceRepository` (Postgres/GORM) in `infrastructure/postgres/device_repo.go`.
- [ ] Implement `TelemetryRepository` in `infrastructure/postgres/telemetry_repo.go`.
- [ ] Create GORM migration for `devices`, `telemetries`, `commands` tables.
- [ ] Write unit tests for Repositories.

## Phase 3: Core Logic (Use Cases)
- [ ] Implement `DeviceUseCase` (Register, List, Update Status).
- [ ] Implement `TelemetryUseCase` (Ingest, Query History).
- [ ] Implement `CommandUseCase` (Queue Command, Update Status).
- [ ] Write unit tests for Use Cases.

## Phase 4: HTTP API
- [ ] Create Gin/Chi handlers for Device Management.
- [ ] Create handlers for Telemetry Querying.
- [ ] Register routes in `internal/modules/iot/module.go`.
- [ ] Add integration tests for API endpoints.

## Phase 5: MQTT Integration
- [ ] Research/Select Go MQTT Broker library (e.g., `mochi-mqtt`) or Client library (`paho`).
- [ ] Implement `MQTTService` to listen on `kthulu/devices/+/telemetry`.
- [ ] Wire `MQTTService` to `TelemetryUseCase`.
- [ ] Implement Command dispatching via MQTT publish to `kthulu/devices/{id}/commands`.

## Phase 6: Frontend (React)
- [ ] Generate `Device` entity frontend code.
- [ ] Create `DeviceList` page.
- [ ] Create `DeviceDetail` page with a simple chart for Telemetry (using Recharts).
- [ ] Add "Send Command" button/modal on Detail page.

## Phase 7: Verification & Docs
- [ ] Verify End-to-End data flow (Device -> MQTT -> DB -> API -> Frontend).
- [ ] Update Swagger/OpenAPI documentation.
- [ ] Write `docs/IOT_GUIDE.md`.
