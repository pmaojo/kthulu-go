# Módulo Veri\*Factu - Cumplimiento Fiscal Español

## Descripción General

El módulo Veri\*Factu proporciona cumplimiento opcional con las regulaciones fiscales españolas (RD 1007/2023 RRSIF) para la verificación de facturas. Se integra perfectamente con el módulo de facturas existente sin romper la arquitectura limpia.

> **Estado**: ✅ **Especificado y Documentado** - Listo para implementación
>
> Este módulo está completamente especificado en `.kiro/specs/kthulu-original-master/verifactu-extension.md` con requisitos funcionales, diseño arquitectónico y tareas de implementación detalladas.

## Características Principales

### ✅ Cumplimiento Normativo Completo

- **RD 1007/2023 (RRSIF)**: Implementación completa del Reglamento del Registro de Sistemas Informáticos de Facturación
- **Registro Estructurado**: Generación automática de registros XML/JSON según especificaciones AEAT
- **Integridad y Trazabilidad**: Hash encadenado y firmas digitales para garantizar inalterabilidad
- **Auditoría Completa**: Log de eventos para todas las operaciones (alta, baja, incidentes)

### 🔄 Modos de Operación Duales

1. **Modo Veri\*Factu (Tiempo Real)**

   - Envío inmediato y fiable a AEAT
   - Verificación en tiempo real
   - Máxima garantía de cumplimiento

2. **Modo No-Veri\*Factu (Cola con Firma)**
   - Firma digital local
   - Almacenamiento seguro
   - Envío bajo demanda o programado

### 🏗️ Arquitectura Modular y Desacoplada

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│  Invoice Module │───▶│ Veri*Factu Module│───▶│  AEAT Service   │
│   (Core ERP)    │    │   (Compliance)   │    │  (External)     │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

**Principios de Diseño:**

- **Zero Coupling**: El módulo de facturas funciona independientemente
- **Event-Driven**: Comunicación mediante eventos de dominio
- **Optional**: Se puede activar/desactivar por configuración
- **Extensible**: Preparado para otras jurisdicciones fiscales

## Configuración

### Variables de Entorno

```bash
# Activar módulo Veri*Factu
MODULES=health,auth,user,access,notifier,organization,contact,product,invoice,verifactu

# Configuración específica Veri*Factu
VERIFACTU_MODE=real-time                    # 'real-time' o 'queued'
VERIFACTU_AEAT_ENDPOINT=https://sede.agenciatributaria.gob.es/Sede/ws/verifactu
VERIFACTU_CERTIFICATE_PATH=/path/to/cert.p12
VERIFACTU_CERTIFICATE_PASSWORD=secret
VERIFACTU_ORGANIZATION_NIF=12345678A
VERIFACTU_SIF_CODE=01                       # Código SIF asignado por AEAT
VERIFACTU_RETRY_ATTEMPTS=3
VERIFACTU_RETRY_DELAY=5s
```

Cuando el servicio arranca en modo `real-time` se activa un indicador persistente
`live_mode` en la tabla `verifactu_settings`. Mientras este indicador esté activo,
no es posible volver al modo `queued` durante el ejercicio fiscal en curso
(hasta el 31 de diciembre).

### Certificados Digitales

El módulo requiere un certificado digital válido emitido por la FNMT para la comunicación con AEAT:

1. **Obtención**: Solicitar certificado en https://www.sede.fnmt.gob.es/
2. **Formato**: Certificado en formato PKCS#12 (.p12)
3. **Instalación**: Colocar en ruta segura y configurar `VERIFACTU_CERTIFICATE_PATH`

## Funcionalidades Técnicas

### 📋 Registro de Facturas

Cada factura genera automáticamente:

```json
{
  "TipoRegistro": "alta",
  "IDFactura": {
    "IDEmisorFactura": "12345678A",
    "NumSerieFactura": "INV-2024-03-0001"
  },
  "FechaHoraHusoGenFactura": "2024-03-15T10:30:00+01:00",
  "TipoFactura": "F1",
  "ImporteTotalFactura": 121.0,
  "Huella": "ABC123...",
  "FechaHoraHusoGenRegistro": "2024-03-15T10:30:05+01:00"
}
```

### 🔐 Integridad y Seguridad

- **Hash Encadenado**: Cada registro incluye hash del anterior
- **Firma Digital**: Firma PKCS#7 para modo no-Veri\*Factu
- **Verificación**: QR codes con enlace a verificación AEAT
- **Audit Trail**: Registro completo de todas las operaciones

### ❌ Cancelación de Registros

Las facturas pueden anularse generando un nuevo registro de tipo `anulacion` que
referencia al registro original. Este proceso mantiene la integridad del
encadenamiento de hashes.

```http
POST /verifactu/records/{id}/cancel
```

La respuesta contiene el nuevo registro de cancelación.

### 📱 Integración Visual

Las facturas incluyen automáticamente:

- **Código QR**: Con datos de verificación
- **Leyenda Legal**: "Factura verificable en la sede electrónica de AEAT"
- **Información de Registro**: Número de registro y fecha

## API Endpoints

### Gestión de Registros Veri\*Factu

```http
GET    /verifactu/records              # Listar registros
GET    /verifactu/records/{id}         # Obtener registro específico
POST   /verifactu/records/{id}/submit  # Enviar registro a AEAT
GET    /verifactu/records/{id}/status  # Estado de envío
POST   /verifactu/records/{id}/retry   # Reintentar envío
```

### Auditoría y Eventos

```http
GET    /verifactu/events               # Log de eventos
GET    /verifactu/audit/{invoiceId}    # Auditoría de factura específica
GET    /verifactu/stats                # Estadísticas de cumplimiento
```

### Webhooks AEAT

```http
POST   /verifactu/webhooks/aeat        # Webhook para respuestas AEAT
```

## Casos de Uso

### 1. Facturación Normal con Cumplimiento

```go
// El módulo de facturas funciona normalmente
invoice := CreateInvoice(invoiceData)

// Veri*Factu se activa automáticamente si está habilitado
// - Genera registro estructurado
// - Calcula hash encadenado
// - Envía a AEAT (modo real-time) o firma (modo queued)
// - Genera QR code
// - Registra eventos de auditoría
```

### 2. Manejo de Errores de Red

```go
// Si AEAT no está disponible:
// 1. Marca registro como "pendiente"
// 2. Programa reintentos automáticos
// 3. Registra incidente en audit trail
// 4. Notifica al usuario del estado
```

### 3. Verificación de Integridad

```go
// Verificación de cadena de hash
isValid := verifactu.VerifyChain(organizationID)

// Verificación de firma digital
isSignatureValid := verifactu.VerifySignature(recordID)
```

## Tablas de Base de Datos

### verifactu_records

```sql
CREATE TABLE verifactu_records (
    id SERIAL PRIMARY KEY,
    invoice_id INT NOT NULL REFERENCES invoices(id),
    organization_id INT NOT NULL REFERENCES organizations(id),
    record_type VARCHAR(20) NOT NULL, -- 'alta', 'baja', 'incident'
    sif_code CHAR(2) NOT NULL,
    structured_data TEXT NOT NULL,    -- XML/JSON según RRSIF
    hash VARCHAR(256) NOT NULL,
    signature TEXT,                   -- Firma PKCS#7 (modo no-Veri*Factu)
    qr_code TEXT NOT NULL,
    submission_status VARCHAR(20) DEFAULT 'pending',
    submitted_at TIMESTAMPTZ,
    aeat_response TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

### verifactu_events

```sql
CREATE TABLE verifactu_events (
    id SERIAL PRIMARY KEY,
    record_id INT REFERENCES verifactu_records(id),
    event_type VARCHAR(50) NOT NULL,
    description TEXT,
    user_id INT REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

## Beneficios Empresariales

### ✅ Cumplimiento Legal Automático

- **Sin intervención manual**: Cumplimiento transparente
- **Reducción de riesgos**: Eliminación de errores humanos
- **Auditoría completa**: Trazabilidad total de operaciones

### 🚀 Ventajas Técnicas

- **Modularidad**: Se puede activar/desactivar sin afectar funcionalidad core
- **Performance**: Procesamiento asíncrono para no impactar velocidad
- **Escalabilidad**: Soporte para múltiples terminales y organizaciones
- **Mantenibilidad**: Código limpio y bien documentado

### 💼 Valor de Negocio

- **Competitividad**: Diferenciación en el mercado español
- **Confianza**: Cumplimiento garantizado con regulaciones fiscales
- **Eficiencia**: Automatización completa del proceso de cumplimiento
- **Expansión**: Base para cumplimiento en otras jurisdicciones

## Roadmap Futuro

### Fase 1: Implementación Base ✅

- Registro estructurado según RRSIF
- Modos dual (real-time/queued)
- QR codes y leyendas legales

### Fase 2: Características Avanzadas

- Dashboard de cumplimiento
- Reportes de auditoría avanzados
- Integración con otros sistemas fiscales

### Fase 3: Expansión Internacional

- Soporte para TicketBAI (País Vasco)
- Integración con sistemas fiscales europeos
- Adaptación para otros países

---

**El módulo Veri\*Factu representa la excelencia en cumplimiento fiscal automatizado, manteniendo los más altos estándares de arquitectura limpia y modularidad que caracterizan al framework Kthulu.**
