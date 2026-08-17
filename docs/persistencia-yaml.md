# Persistencia YAML: separación de responsabilidades

## Contexto

Los repositorios YAML (`YamlCategoryRepository` y los que vengan para
subcategorías y beneficiarios) mezclaban tres responsabilidades en una sola
struct: descubrir y leer archivos (`fs.Glob` + `fs.ReadFile`), interpretar el
YAML (`yaml.Unmarshal` a un record propio) y construir el grafo de entidades
del dominio (resolución de padres, indexado, caché). El resultado eran
repositorios extensos para lo que en realidad hacen.

Se decidió partir esa responsabilidad en tres componentes.

## Por qué no genéricos

Se evaluó un `YamlDatasource[TRecord, TRoot, TId]` genérico, único para los
tres tipos de YML. Se descartó: la única parte realmente compartible entre
`categories`, `subcategories` y `beneficiaries` es el I/O (glob + lectura de
bytes), que no necesita conocer ningún tipo — no hace falta parametrizarlo. La
parte que sí varía por tipo (la forma del YAML, el `Unmarshal`, la resolución
de relaciones) es distinta en cada dominio y los genéricos no la unifican;
solo agregan la verbosidad de instanciar el tipo en cada uso a cambio de
compartir unas pocas líneas.

## Los tres componentes

### 1. `YMLReader` — I/O y convención de nombres

- Recibe un `fs.FS` y un patrón (ej. `"categories.*.yml"`).
- Descubre los archivos que matchean el patrón y los devuelve ordenados (el
  orden importa: los niveles se cargan de menor a mayor porque una categoría
  solo puede referenciar como padre a una ya cargada).
- Valida la convención de nombres: que haya al menos un archivo, que la
  numeración sea coherente.
- Devuelve bytes crudos por archivo. **No** parsea YAML — un YAML corrupto no
  es un error suyo.
- Es el único componente reusable sin cambios entre los tres tipos de YML,
  porque no conoce ningún tipo de dominio. Se le pasa el patrón desde afuera.

### 2. Datasource — uno por tipo (`YamlCategoryDatasource`, ...)

- Le pide bytes al `YMLReader` con el patrón de su propio tipo.
- Hace `yaml.Unmarshal` a su record propio (`categoryRecord`). Los structs con
  tags `yaml:"..."` viven acá.
- Expone algo como `Records() ([]categoryRecord, error)`, agregando todos los
  archivos del tipo en una sola colección ordenada.
- Cachea en memoria después de la primera carga (de ahí "Memory" en el
  nombre original propuesto) — mismo patrón *lazy* que hoy tiene
  `initialize()`, pero acotado a records, no a entidades.
- **No** sabe nada de `entity.Category`, de padres ni de índices por id. Un
  YAML mal formado es su error; un id duplicado o un padre inexistente no —
  eso requiere mirar la colección completa, no la forma de un registro.

### 3. Repository — uno por tipo (`YamlCategoryRepository`, ...)

- Le pide records a su Datasource y arma las entidades del dominio
  (`entity.Category`), resolviendo `CategoryId` contra las instancias ya
  construidas.
- Acá viven las validaciones de integridad: id duplicado y padre inexistente
  (cubiertas hoy por `TestIntegrityCheck`). No son sobre el YAML sino sobre la
  coherencia del conjunto, así que no se mueven al Datasource.
- Mantiene su propio índice (`map[string]int`) y su propio caché de
  entidades, independiente del caché de records del Datasource — construir el
  grafo de punteros (padre → hijo) tiene su propio costo.
- Es el único que implementa el contrato del dominio
  (`repository.CategoryRepository`), verificado en tiempo de compilación con
  `var _ repository.CategoryRepository = (*YamlCategoryRepository)(nil)`.

## Convención de nombres

`Yaml<Tipo>Datasource` y `Yaml<Tipo>Repository`, siguiendo el prefijo de
tecnología que ya usa `YamlCategoryRepository`. Se descartó nombrar al
Datasource como un `MemoryDatasource` genérico instanciado por tipo: "en
memoria" es un detalle interno del componente, no parte de su identidad.

## Pendiente de decidir

Si el Repository depende del Datasource por su tipo concreto o a través de
una interfaz chica definida en el propio paquete de infraestructura. Con tipo
concreto es más simple y ya hay inyección de dependencias vía `fs.FS`; con
interfaz se gana la posibilidad de testear el Repository con un Datasource
falso, sin pasar por YAML real. Hoy los tests del repositorio ya ejercitan
todo de punta a punta con fixtures reales (`yml_test/`), así que no es obvio
que haga falta la interfaz.

## Estado

`YMLReader` implementado en
`internal/shared/infrastructure/persistence/yml_reader.go`, con sus tests.
Vive en `shared` porque no conoce ningún tipo de dominio y lo van a usar los
tres tipos de YML por igual. Falta migrar `YamlCategoryRepository` a este
esquema (Datasource + Repository usando el Reader) y replicarlo para
subcategorías y beneficiarios cuando corresponda.
