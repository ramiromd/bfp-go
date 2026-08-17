package persistence

import (
	"testing"
	"testing/fstest"

	sharedpersistence "ramiromd/budget/internal/shared/infrastructure/persistence"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// datasourceMapFS arma un fs.FS en memoria a partir de un mapa nombre de
// archivo -> contenido YAML, para no depender de fixtures en disco.
func datasourceMapFS(files map[string]string) fstest.MapFS {

	fsys := fstest.MapFS{}
	for name, content := range files {
		fsys[name] = &fstest.MapFile{Data: []byte(content)}
	}

	return fsys
}

// newYamlCategoryDatasource arma el datasource sobre un origen de archivos en
// memoria, sin abrirlo: cada test decide si llama a Open().
func newYamlCategoryDatasource(t *testing.T, files map[string]string) *YamlCategoryDatasource {
	t.Helper()

	reader, err := sharedpersistence.NewYMLReader(datasourceMapFS(files), ".")
	require.Nil(t, err, "NewYMLReader no debería fallar armando el reader de prueba")

	return NewYamlCategoryDatasource(reader)
}

func TestYamlCategoryDatasource_Open(t *testing.T) {

	t.Run("WithValidFiles_LoadsAllRecordsAcrossLevels", func(t *testing.T) {

		// Given: un datasource sobre dos archivos de distinto nivel.
		datasource := newYamlCategoryDatasource(t, map[string]string{
			"categories.0.yml": "categories:\n  - id: services\n    name: Servicios\n",
			"categories.1.yml": "categories:\n  - id: light\n    name: Electricidad\n    category_id: services\n",
		})

		// When: se abre el datasource y se consultan los records.
		err := datasource.Open()
		records, recordsErr := datasource.Records()

		// Then: no falla y expone los records de ambos archivos.
		require.Nil(t, err, "Open no debería fallar con archivos válidos")
		require.Nil(t, recordsErr, "Records no debería fallar luego de Open")
		assert.Len(t, records, 2, "Records debería incluir los records de los dos archivos")
	})

	t.Run("WithNoMatchingFiles_Fails", func(t *testing.T) {

		// Given: un datasource sobre un origen sin archivos categories.*.yml.
		datasource := newYamlCategoryDatasource(t, map[string]string{
			"beneficiaries.0.yml": "beneficiaries: []\n",
		})

		// When: se intenta abrir.
		err := datasource.Open()

		// Then: falla, propagando el error del YMLReader.
		assert.NotNil(t, err, "Open debería fallar si ningún archivo respeta el patrón")
	})

	t.Run("WithInvalidYaml_FailsAndDoesNotActivate", func(t *testing.T) {

		// Given: un datasource sobre un archivo con YAML mal formado.
		datasource := newYamlCategoryDatasource(t, map[string]string{
			"categories.0.yml": "categories: [this is not valid yaml",
		})

		// When: se intenta abrir y luego se consultan los records.
		openErr := datasource.Open()
		records, recordsErr := datasource.Records()

		// Then: Open falla y el datasource sigue sin estar activo.
		require.NotNil(t, openErr, "Open debería fallar ante un YAML mal formado")
		assert.NotNil(t, recordsErr, "Records debería seguir fallando si Open nunca completó con éxito")
		assert.Nil(t, records, "Records no debería devolver records cuando falla")
	})

	t.Run("CalledTwice_KeepsReturningTheSameRecords", func(t *testing.T) {

		// Given: un datasource ya abierto una vez.
		datasource := newYamlCategoryDatasource(t, map[string]string{
			"categories.0.yml": "categories:\n  - id: services\n    name: Servicios\n",
		})
		require.Nil(t, datasource.Open(), "la primera apertura no debería fallar")

		// When: se abre una segunda vez y se consultan los records.
		secondOpenErr := datasource.Open()
		records, recordsErr := datasource.Records()

		// Then: no falla y sigue exponiendo el mismo record.
		require.Nil(t, secondOpenErr, "una segunda llamada a Open no debería fallar")
		require.Nil(t, recordsErr, "Records no debería fallar luego de Open")
		assert.Len(t, records, 1, "Records debería seguir devolviendo un único record")
	})
}

func TestYamlCategoryDatasource_Records(t *testing.T) {

	t.Run("BeforeOpen_Fails", func(t *testing.T) {

		// Given: un datasource recién construido, nunca abierto.
		datasource := newYamlCategoryDatasource(t, map[string]string{
			"categories.0.yml": "categories:\n  - id: services\n    name: Servicios\n",
		})

		// When: se consultan los records sin haber llamado a Open.
		records, err := datasource.Records()

		// Then: falla en vez de abrir implícitamente.
		assert.NotNil(t, err, "Records debería fallar si el datasource no fue abierto")
		assert.Nil(t, records, "Records no debería devolver records cuando falla")
	})
}

func TestYamlCategoryDatasource_RecordAt(t *testing.T) {

	t.Run("BeforeOpen_Fails", func(t *testing.T) {

		// Given: un datasource recién construido, nunca abierto.
		datasource := newYamlCategoryDatasource(t, map[string]string{
			"categories.0.yml": "categories:\n  - id: services\n    name: Servicios\n",
		})

		// When: se consulta una posición sin haber llamado a Open.
		record, err := datasource.RecordAt(0)

		// Then: falla en vez de abrir implícitamente, devolviendo el zero value.
		assert.NotNil(t, err, "RecordAt debería fallar si el datasource no fue abierto")
		assert.Equal(t, categoryDsRecord{}, record, "RecordAt debería devolver el zero value cuando falla")
	})

	t.Run("WithAValidPosition_ReturnsTheMatchingRecord", func(t *testing.T) {

		// Given: un datasource abierto sobre un único archivo con dos records.
		datasource := newYamlCategoryDatasource(t, map[string]string{
			"categories.0.yml": "categories:\n  - id: services\n    name: Servicios\n  - id: goods\n    name: Bienes\n",
		})
		require.Nil(t, datasource.Open(), "Open no debería fallar con archivos válidos")

		// When: se consulta la segunda posición.
		record, err := datasource.RecordAt(1)

		// Then: se obtiene el record ubicado en esa posición.
		require.Nil(t, err, "RecordAt no debería fallar con una posición válida")
		assert.Equal(t, "goods", record.Id, "RecordAt debería devolver el record ubicado en la posición pedida")
	})

	t.Run("WithANegativePosition_Fails", func(t *testing.T) {

		// Given: un datasource abierto.
		datasource := newYamlCategoryDatasource(t, map[string]string{
			"categories.0.yml": "categories:\n  - id: services\n    name: Servicios\n",
		})
		require.Nil(t, datasource.Open(), "Open no debería fallar con archivos válidos")

		// When: se consulta una posición negativa.
		record, err := datasource.RecordAt(-1)

		// Then: falla en vez de acceder fuera de rango.
		assert.NotNil(t, err, "RecordAt debería fallar ante una posición negativa")
		assert.Equal(t, categoryDsRecord{}, record, "RecordAt debería devolver el zero value cuando falla")
	})

	t.Run("WithAPositionBeyondTheLastRecord_Fails", func(t *testing.T) {

		// Given: un datasource abierto con un único record.
		datasource := newYamlCategoryDatasource(t, map[string]string{
			"categories.0.yml": "categories:\n  - id: services\n    name: Servicios\n",
		})
		require.Nil(t, datasource.Open(), "Open no debería fallar con archivos válidos")

		// When: se consulta una posición fuera de rango.
		record, err := datasource.RecordAt(1)

		// Then: falla en vez de entrar en pánico.
		assert.NotNil(t, err, "RecordAt debería fallar ante una posición fuera de rango")
		assert.Equal(t, categoryDsRecord{}, record, "RecordAt debería devolver el zero value cuando falla")
	})
}