# Arquitectura del Proyecto — Guía de Diseño y Estructura (SSOT)

> **Nota para Desarrolladores y Agentes de IA:**  
> Este documento define la arquitectura oficial, convenciones de diseño y estructura de carpetas del proyecto. Debe ser respetado de forma estricta al crear nuevos módulos, casos de uso, entidades o integraciones de infraestructura.

## Vista general

El proyecto está desarrollado en **Go (Golang)** y sigue los principios de **Clean Architecture** (Arquitectura Limpia) orientada a **módulos independientes**.


### Principios Fundamentales
1. **Regla de Dependencia:** Las dependencias de código solo deben apuntar **hacia adentro** (las capas internas no conocen detalles de las capas externas).
   - `Infrastructure` -> `Application` -> `Domain`
   - `Domain` no depende de **ningún** framework, base de datos ni paquete de terceros ajeno al lenguaje puro.
2. **Modularidad Horizontal:** El sistema se divide en módulos funcionales dentro de `internal/`. Cada módulo engloba su propio dominio, aplicación e infraestructura.
3. **Puntos de Entrada Desacoplados:** El código ejecutable reside fuera de `internal/` (en `cmd/`), encargándose de la inicialización y el ensamble (wiring) de dependencias.

## Scaffolding

```text
.
├── cmd/
│   └── main.go                     # Punto de entrada principal (Wiring e inicialización)
├── internal/
│   ├── shared/                     # Código y abstracciones compartidas entre módulos
│   │   ├── domain/                 # Value Objects, tipos base o errores compartidos
│   │   ├── application/            # Paginación base, interfaces globales
│   │   └── infrastructure/         # Configuración de DB, logger, clientes HTTP base
│   ├── <modulo_1>/                 # Módulo de Dominio A (ej. users, orders, billing)
│   │   ├── domain/
│   │   ├── application/
│   │   └── infrastructure/
│   └── <modulo_2>/                 # Módulo de Dominio B
│       ├── domain/
│       ├── application/
│       └── infrastructure/
├── go.mod
├── go.sum
└── README.md
```

## Anatomía de un módulo

Cada módulo dentro de `internal/<modulo_name>/` contiene exactamente las tres capas arquitectónicas:

```text
internal/<modulo_name>/
├── domain/                      # CAPA 1: Reglas de negocio puras
│   ├── entity/                   # Entidades principales del dominio
│   ├── repository/               # Interfaces de repositorios (Puertos de salida)
│   ├── value_object/             # Objetos de valor e inmutabilidad
│   ├── service/                  # Servicios de dominio (lógica multi-entidad)
│   └── errors/                   # Errores de dominio tipados
│
├── application/                    # CAPA 2: Casos de uso de la aplicación
│   ├── <use_case_name>.go          # Lógica de un caso de uso específico
│   ├── <use_case_name>_req.go      # Data Transfer Objects (Input Structs)
│   └── <use_case_name>_res.go      # Data Transfer Objects (Output Structs)
│
└── infrastructure/                 # CAPA 3: Adaptadores e implementación técnica
    ├── persistence/                # Repositorios de base de datos (PostgreSQL, Redis, etc.)
    ├── http/                       # Handlers / Controllers (Gin, Fiber, net/http)
    └── client/                     # Clientes de APIs externas o brokers de eventos
```

## Descripción de Capas y Sus Responsabilidades

### Capa de Dominio

- **Propósito:** Representa el corazón del negocio y sus reglas inherentes.
- **Contenido:**
  - **Entidades:** Structs de Go con comportamiento encapsulado y métodos de validación de reglas de negocio.
  - **Value Objects:** Tipos inmutables que representan conceptos del dominio (ej. `Email`, `Money`, `UUID`).
  - **Interfaces de Repositorio:** Contratos que especifican cómo la capa de dominio espera persistir y obtener entidades (Puertos).
  - **Servicios de Dominio:** Lógica que coordina múltiples entidades dentro del mismo módulo sin pertenecer a una en particular.
  - **Errores de Dominio:** Definición de errores del negocio (ej. `ErrInvalidAmount`, `ErrUserNotFound`).
- **Reglas Estrictas:**
  - No debe importar paquetes de `application` o `infrastructure`.
  - No debe incluir tags de serialización (ej. `json:"..."`, `gorm:"..."`, `db:"..."`).
  - No debe depender de frameworks, drivers SQL, o bibliotecas externas a la librería estándar de Go (a menos que sean utilitarios puros como `google/uuid`)

### Capa de Aplicación

- **Propósito:** Orquesta el flujo de trabajo para cumplir con los requerimientos del usuario/sistema (Casos de Uso / Use Cases).
- **Contenido:**
    - **Casos de Uso:** Structs ejecutoras que reciben un `context.Context` y un `DTO` de entrada, invocan repositorios/servicios de dominio y retornan un `DTO` de salida.
    - **DTOs:** Structs simples para transferir datos hacia y desde la capa de aplicación.
- **Reglas Estrictas:**
    - Puede importar la capa de `domain` del mismo módulo o de `shared/domain`.
    - No debe importar nada de `infrastructure`.
    - No debe manejar detalles HTTP (status codes, headers, Gin context, etc.).

### 4.3 Capa de Infraestructura
- **Propósito:** Proveer la implementación concreta de los puertos definidos por las capas de Dominio y Aplicación.
- **Contenido:**
    - **Persistencia:** Implementación concreta de los `Repository` interfaces (SQL, NoSQL, ORMs, DB queries puras).
    - **HTTP / gRPC Handlers:** Controladores web que procesan requests, convierten payloads a DTOs de aplicación, llaman al caso de uso correspondiente y devuelven respuestas HTTP.
    - **Integraciones:** Adaptadores para servicios de terceros (SendGrid, AWS S3, Stripe, brokers RabbitMQ/Kafka).
- **Reglas Estrictas:**
  - Puede importar `domain` y `application`.
  - Puede contener los tags de serialización (`json`, `db`, `validate`).
  - No debe contener lógica de negocio central; solo traducción y adaptación de datos/tecnología.

## Módulo compartido

## 5. Módulo Compartido

El módulo `shared/` alberga código utilitario y abstracciones reusables entre múltiples módulos.

- **`shared/domain/`:** Errores transversales (`ErrInternalServer`, `ErrNotFound`), tipos base (ej. `AggregateRoot`, `ID`).
- **`shared/application/`:** Interfaces transversales (Event Bus, Command Bus, DTOs de paginación estándar).
- **`shared/infrastructure/`:** Configuración global (Carga de envs), clientes de base de datos (`*sql.DB`, `*redis.Client`), loggers y middlewares HTTP compartidos.

### Regla de Comunicación entre Módulos

- Módulo A **no debe importar la infraestructura ni los casos de uso** de Módulo B.
- Si Módulo A necesita datos de Módulo B:
    1. **Comunicación Directa:** Invocar una interfaz pública/servicio expuesto en el `domain` o `application` del Módulo B.
    2. **Eventos (Desacoplado):** Emitir un Evento de Dominio que Módulo B escuche de forma asíncrona.

