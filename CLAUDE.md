# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Sobre el proyecto

Herramienta CLI en Go que lee extractos bancarios y resúmenes de tarjeta (Banco Galicia), los normaliza a un modelo común de movimientos y les asigna una clasificación (categoría/subcategoría). Módulo Go: `ramiromd/budget`. Todo el código y la documentación están en español (comentarios, mensajes de error, nombres de archivos de datos).

## Comandos

```bash
go build ./...
go vet ./...
go test ./...                        # aún no hay archivos _test.go en el repo
go test ./internal/domain/... -run TestNombre -v   # un test puntual

# Ejecutar el CLI
go run ./cmd <archivo> <account|card> <banco> <true|false>
go run ./cmd resources/input/Extracto_00035609454.csv account galicia true
```

Los cuatro argumentos son obligatorios y posicionales: ruta del archivo, tipo (`account` = extracto de cuenta, `card` = resumen de tarjeta), banco, y si la primera línea trae cabeceras.

## Arquitectura

`ARCHITECTURE.md` en la raíz es el **SSOT** de la arquitectura: Clean Architecture con módulos horizontales bajo `internal/<modulo>/{domain,application,infrastructure}`. Léelo antes de crear módulos, casos de uso o entidades. La skill `module-creation` (`.claude/skills/`) apunta a ese mismo documento.

### Estado actual: migración en curso

El repo está a mitad de camino entre la estructura vieja y la que define `ARCHITECTURE.md`:

- `internal/domain/`, `internal/application/`, `internal/infrastructure/` — estructura **plana heredada**, todavía sin capa de módulo. Es donde vive todo el código funcional hoy.
- `internal/classification/` — primer módulo con la forma correcta; por ahora es un scaffold vacío (solo `.gitkeep`).

`Category`, `Subcategory`, `Beneficiary`, sus interfaces de repositorio y sus implementaciones YAML son conceptualmente del módulo `classification`, pero siguen en el árbol plano. **No los migres salvo que se pida explícitamente.**

Violaciones conocidas de la regla de dependencia, presentes en el código actual:

- `internal/domain/service/galicia_account_movement_converter.go` importa `internal/infrastructure/repository` e instancia repositorios concretos dentro del dominio. Al tocar ese archivo, tenerlo en cuenta; corregirlo implica inyectar los repositorios desde `application`/`cmd`.
- `Movement.Classification` es un `*entity.Subcategory`: cuando `Subcategory` se mueva a `classification`, ese campo cruza el límite entre módulos.

### Flujo de procesamiento

`cmd/main.go` parsea argv y arma un `ExportMovementsReq` → `application.ExportMovements` abre el CSV (delimitador `;`, `FieldsPerRecord = -1` porque las columnas varían) y lo lee **registro a registro**, no línea a línea: un registro puede ocupar varias líneas cuando un campo va entrecomillado → `service.MovementConverterFactory` resuelve el conversor por `(banco, tipo)` → cada conversor implementa `IBankMovementConverter.Convert([]string) *entity.Movement`.

Un registro inválido **no aborta el archivo**: el conversor devuelve `nil`, se cuenta en `FailedRecords` y el bucle sigue. Los contadores (`ReadRecords`/`ExportedRecords`/`FailedRecords`) son el resultado del caso de uso.

El `Id` del movimiento es un UUID **determinístico** (`uuid.NewMD5` sobre `fecha|monto|descripción`), no aleatorio: reprocesar el mismo archivo produce los mismos ids.

### Clasificación

`beneficiaries.yml` mapea un `pattern` (regex, aplicado sin distinguir mayúsculas) a un `subcategory_id`. Si la descripción del movimiento matchea, se le asigna esa subcategoría. Jerarquía: `Category` → `Subcategory` → el movimiento apunta a la subcategoría.

Los tres `.yml` (`categories`, `subcategories`, `beneficiaries`) son datos de referencia versionados, incrustados en el binario con `go:embed` — no se leen del disco en runtime, así que el CLI no depende del directorio de trabajo. Editar un `.yml` requiere recompilar.

### Convenciones del código

- Receptores de método se llaman `this`.
- Interfaces de repositorio con prefijo `I` (`ICategoryRepository`); implementaciones con el prefijo de la tecnología (`YamlCategoryRepository`).
- Cada implementación de repositorio afirma el contrato en tiempo de compilación: `var _ domain.IXRepository = (*YamlX)(nil)`.
- Los repositorios YAML cargan de forma perezosa en `load()` y cachean en memoria. **No son seguros para uso concurrente.**
- Entidades con constructor `NewX(...)`, sin tags de serialización (regla del dominio). Los tags `yaml:"..."` viven en structs `xRecord`/`xFile` privados de la capa de infraestructura.
- Un caso de uso son tres archivos: `<nombre>.go`, `<nombre>_req.go`, `<nombre>_res.go`.

## Datos sensibles

`resources/input/` y `resources/output/` están en `.gitignore` porque contienen extractos reales (números de cuenta, CUIT, tarjetas). Se versiona la estructura de carpetas, no el contenido. No commitear archivos de esos directorios ni pegar su contenido en el repo.

## Skills del proyecto

- `gal-visa-credit` — formato de entrada/salida para convertir resúmenes Visa de Galicia (PDF → CSV). Contiene la especificación exacta de columnas, separadores y nombres de archivo por moneda.
- `module-creation` — redirige a `ARCHITECTURE.md` al crear un módulo nuevo.
