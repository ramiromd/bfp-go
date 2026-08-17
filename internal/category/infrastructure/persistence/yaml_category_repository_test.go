package persistence

import (
	"testing"

	"ramiromd/budget/internal/category/domain/entity"
	"ramiromd/budget/internal/category/domain/repository"
	sharedpersistence "ramiromd/budget/internal/shared/infrastructure/persistence"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Verifica en tiempo de compilación que el repositorio cumple el contrato del
// dominio.
var _ repository.CategoryRepository = (*YamlCategoryRepository)(nil)

// twoLevelCategoryFiles arma dos archivos de categorías en memoria: un nivel
// raíz ("services") y un nivel anidado con tres categorías que lo referencian
// como padre. Es el dataset del camino feliz para el resto de los tests.
func twoLevelCategoryFiles() map[string]string {
	return map[string]string{
		"categories.0.yml": `categories:
  - id: "services"
    name: "Servicios básicos"
    category_id: ""
    description: "Servicios esenciales incluyendo agua, electricidad y gas."
`,
		"categories.1.yml": `categories:
  - id: "light"
    category_id: "services"
    name: "Electricidad"
    description: "Servicios de electricidad y energía."
  - id: "water"
    category_id: "services"
    name: "Agua"
    description: "Servicios de suministro de agua potable."
  - id: "gas"
    category_id: "services"
    name: "Gas"
    description: "Servicios de suministro de gas natural y propano."
`,
	}
}

// duplicateCategoryIdFiles arma un único archivo que define el id "services"
// dos veces, para ejercitar la validación de duplicados.
func duplicateCategoryIdFiles() map[string]string {
	return map[string]string{
		"categories.0.yml": `categories:
  - id: "services"
    name: "Servicios básicos"
    category_id: ""
    description: "Servicios esenciales incluyendo agua, electricidad y gas."
  - id: "services"
    name: "Servicios básicos"
    category_id: ""
    description: "Servicios esenciales incluyendo agua, electricidad y gas."
`,
	}
}

// nonexistentParentCategoryFiles arma un único archivo donde una categoría
// referencia como padre un id ("foo") que no está definido en ningún lado.
func nonexistentParentCategoryFiles() map[string]string {
	return map[string]string{
		"categories.1.yml": `categories:
  - id: "light"
    category_id: "services"
    name: "Electricidad"
    description: "Servicios de electricidad y energía."
  - id: "water"
    category_id: "services"
    name: "Agua"
    description: "Servicios de suministro de agua potable."
  - id: "gas"
    category_id: "foo"
    name: "Gas"
    description: "Servicios de suministro de gas natural y propano."
`,
	}
}

// newYamlCategoryRepositoryFromFiles arma el repositorio sobre un origen de
// archivos en memoria, pasando por el mismo camino que un caller de
// producción: YMLReader -> Datasource -> Repository. Propaga el error de
// construcción en vez de asumir que no falla, para que los tests de
// integridad puedan inspeccionarlo.
func newYamlCategoryRepositoryFromFiles(t *testing.T, files map[string]string) (*YamlCategoryRepository, error) {
	t.Helper()

	reader, err := sharedpersistence.NewYMLReader(datasourceMapFS(files), ".")
	require.Nil(t, err, "NewYMLReader no debería fallar armando el reader de prueba")

	datasource := NewYamlCategoryDatasource(reader)

	return NewYamlCategoryRepositoryFromDatasource(datasource)
}

// mustNewYamlCategoryRepositoryFromFiles es newYamlCategoryRepositoryFromFiles
// para los datasets del camino feliz: falla el test si la construcción del
// repositorio devuelve error, para que el resto del test no tenga que
// repetir esa verificación.
func mustNewYamlCategoryRepositoryFromFiles(t *testing.T, files map[string]string) *YamlCategoryRepository {
	t.Helper()

	repository, err := newYamlCategoryRepositoryFromFiles(t, files)
	require.Nil(t, err, "la construcción del repositorio no debería fallar")

	return repository
}

// categoryIds proyecta los identificadores de una colección de categorías.
func categoryIds(categories []*entity.Category) []string {

	ids := make([]string, 0, len(categories))
	for _, category := range categories {
		ids = append(ids, category.Id)
	}

	return ids
}

// categoryById busca una categoría dentro de una colección, o devuelve nil si
// no está. Se resuelve acá y no con FindById para que los tests de FindAll no
// dependan de otro método del repositorio.
func categoryById(categories []*entity.Category, id string) *entity.Category {

	for _, category := range categories {
		if category.Id == id {
			return category
		}
	}

	return nil
}

func TestYamlCategoryRepositoryFromDatasource_FindAll(t *testing.T) {

	t.Run("with a two-level dataset, returns every category defined in it", func(t *testing.T) {

		// Given: el repositorio de prueba armado sobre un dataset de dos niveles.
		repository := mustNewYamlCategoryRepositoryFromFiles(t, twoLevelCategoryFiles())

		// When: se consulta la colección completa.
		categories, err := repository.FindAll()

		// Then: están las categorías de los dos archivos del dataset. El orden de
		// carga no importa acá, solo que estén todas.
		require.Nil(t, err, "FindAll no debería fallar con un dataset válido")
		assert.ElementsMatch(t, []string{"services", "light", "water", "gas"}, categoryIds(categories), "FindAll debería devolver las categorías de todos los archivos del dataset")
	})

	t.Run("with a category that has no parent, returns it detached", func(t *testing.T) {

		// Given: el repositorio de prueba armado sobre un dataset de dos niveles.
		repository := mustNewYamlCategoryRepositoryFromFiles(t, twoLevelCategoryFiles())

		// When: se consulta la colección completa.
		categories, err := repository.FindAll()
		require.Nil(t, err, "FindAll no debería fallar con un dataset válido")

		topLevelCategory := categoryById(categories, "services")

		// Then: la categoría de primer nivel no tiene padre.
		require.NotNil(t, topLevelCategory, "la categoría de primer nivel %q debería estar en la colección", "services")
		assert.Nil(t, topLevelCategory.Parent, "una categoría con category_id vacío no debería tener padre")
	})

	t.Run("with a category that has a parent, resolves it with the parent's data", func(t *testing.T) {

		// Given: el repositorio de prueba armado sobre un dataset de dos niveles.
		repository := mustNewYamlCategoryRepositoryFromFiles(t, twoLevelCategoryFiles())

		// When: se consulta la colección completa.
		categories, err := repository.FindAll()
		require.Nil(t, err, "FindAll no debería fallar con un dataset válido")

		nestedCategory := categoryById(categories, "light")

		// Then: la categoría anidada tiene un padre con el id correcto.
		require.NotNil(t, nestedCategory, "la categoría anidada %q debería estar en la colección", "light")
		require.NotNil(t, nestedCategory.Parent, "la categoría anidada %q debería tener padre", "light")
		assert.Equal(t, "services", nestedCategory.Parent.Id, "el padre resuelto debería tener el id referenciado por category_id")
	})

	t.Run("called twice, returns an equivalent collection", func(t *testing.T) {

		// Given: un repositorio de prueba ya construido.
		repository := mustNewYamlCategoryRepositoryFromFiles(t, twoLevelCategoryFiles())

		firstCall, firstCallErr := repository.FindAll()

		// When: se vuelve a consultar la colección completa.
		secondCall, secondCallErr := repository.FindAll()

		// Then: se obtienen las mismas categorías que la primera vez.
		require.Nil(t, firstCallErr, "la primera llamada a FindAll no debería fallar")
		require.Nil(t, secondCallErr, "la segunda llamada a FindAll no debería fallar")
		assert.Equal(t, firstCall, secondCall, "ambas llamadas a FindAll deberían devolver colecciones equivalentes")
	})
}

func TestYamlCategoryRepositoryFromDatasource_FindById(t *testing.T) {

	t.Run("with an existing id, returns the matching category", func(t *testing.T) {

		// Given: el repositorio de prueba armado sobre un dataset de dos niveles.
		repository := mustNewYamlCategoryRepositoryFromFiles(t, twoLevelCategoryFiles())

		// When: se busca una categoría por id.
		category, err := repository.FindById("light")

		// Then: se obtiene la categoría con ese id.
		require.Nil(t, err, "FindById no debería fallar con un dataset válido")
		require.NotNil(t, category, "FindById debería encontrar la categoría %q", "light")
		assert.Equal(t, "light", category.Id)
		assert.Equal(t, "Electricidad", category.Name)
	})

	t.Run("with a top-level id, returns the category without a parent", func(t *testing.T) {

		// Given: el repositorio de prueba armado sobre un dataset de dos niveles.
		repository := mustNewYamlCategoryRepositoryFromFiles(t, twoLevelCategoryFiles())

		// When: se busca la categoría de primer nivel.
		category, err := repository.FindById("services")

		// Then: se obtiene sin padre.
		require.Nil(t, err, "FindById no debería fallar con un dataset válido")
		require.NotNil(t, category, "FindById debería encontrar la categoría %q", "services")
		assert.Nil(t, category.Parent, "la categoría de primer nivel no debería tener padre")
	})

	t.Run("with a nested id, returns the category pointing to its parent", func(t *testing.T) {

		// Given: el repositorio de prueba armado sobre un dataset de dos niveles.
		repository := mustNewYamlCategoryRepositoryFromFiles(t, twoLevelCategoryFiles())

		// When: se busca una categoría anidada.
		category, err := repository.FindById("gas")

		// Then: se obtiene apuntando a su padre.
		require.Nil(t, err, "FindById no debería fallar con un dataset válido")
		require.NotNil(t, category, "FindById debería encontrar la categoría %q", "gas")
		require.NotNil(t, category.Parent, "la categoría %q debería tener padre", "gas")
		assert.Equal(t, "services", category.Parent.Id)
	})

	t.Run("with an id that does not exist, returns nil without error", func(t *testing.T) {

		// Given: el repositorio de prueba armado sobre un dataset de dos niveles.
		repository := mustNewYamlCategoryRepositoryFromFiles(t, twoLevelCategoryFiles())

		// When: se busca un id que no está en el dataset.
		category, err := repository.FindById("does-not-exist")

		// Then: se obtiene nil, sin error.
		require.Nil(t, err, "FindById no debería fallar con un dataset válido")
		assert.Nil(t, category, "FindById debería devolver nil para un id inexistente")
	})
}

func TestYamlCategoryRepositoryFromDatasource_Count(t *testing.T) {

	t.Run("with a two-level dataset, returns the total number of categories", func(t *testing.T) {

		// Given: el repositorio de prueba armado sobre un dataset de dos niveles.
		repository := mustNewYamlCategoryRepositoryFromFiles(t, twoLevelCategoryFiles())

		// When: se consulta la cantidad de categorías.
		count, err := repository.Count()

		// Then: se obtiene el total definido en el dataset: 1 en categories.0.yml y 3 en categories.1.yml.
		require.Nil(t, err, "Count no debería fallar con un dataset válido")
		assert.Equal(t, 4, count, "Count debería sumar las categorías de todos los archivos del dataset")
	})

	t.Run("returns the same number of categories as FindAll", func(t *testing.T) {

		// Given: el repositorio de prueba armado sobre un dataset de dos niveles.
		repository := mustNewYamlCategoryRepositoryFromFiles(t, twoLevelCategoryFiles())

		// When: se consulta la cantidad de categorías y la colección completa.
		count, countErr := repository.Count()
		categories, findAllErr := repository.FindAll()

		// Then: ambos valores describen la misma colección.
		require.Nil(t, countErr, "Count no debería fallar con un dataset válido")
		require.Nil(t, findAllErr, "FindAll no debería fallar con un dataset válido")
		assert.Equal(t, len(categories), count, "Count y FindAll deberían describir la misma colección")
	})
}

func TestYamlCategoryRepositoryFromDatasource_FindCategories(t *testing.T) {

	t.Run("with a two-level dataset, returns only the top-level categories", func(t *testing.T) {

		// Given: el repositorio de prueba armado sobre un dataset de dos niveles.
		repository := mustNewYamlCategoryRepositoryFromFiles(t, twoLevelCategoryFiles())

		// When: se consultan las categorías de primer nivel.
		categories, err := repository.FindCategories()

		// Then: solo aparece la categoría sin category_id, no las anidadas.
		require.Nil(t, err, "FindCategories no debería fallar con un dataset válido")
		assert.ElementsMatch(t, []string{"services"}, categoryIds(categories), "FindCategories debería devolver solo las categorías de primer nivel")
	})

	t.Run("with a top-level category, returns it without a parent", func(t *testing.T) {

		// Given: el repositorio de prueba armado sobre un dataset de dos niveles.
		repository := mustNewYamlCategoryRepositoryFromFiles(t, twoLevelCategoryFiles())

		// When: se consultan las categorías de primer nivel.
		categories, err := repository.FindCategories()
		require.Nil(t, err, "FindCategories no debería fallar con un dataset válido")

		topLevelCategory := categoryById(categories, "services")

		// Then: no tiene padre, porque su category_id está vacío.
		require.NotNil(t, topLevelCategory, "la categoría de primer nivel %q debería estar en la colección", "services")
		assert.Nil(t, topLevelCategory.Parent, "una categoría de primer nivel no debería tener padre")
	})
}

func TestYamlCategoryRepositoryFromDatasource_FindSubcategories(t *testing.T) {

	t.Run("with a two-level dataset, returns only the nested categories", func(t *testing.T) {

		// Given: el repositorio de prueba armado sobre un dataset de dos niveles.
		repository := mustNewYamlCategoryRepositoryFromFiles(t, twoLevelCategoryFiles())

		// When: se consultan las subcategorías.
		categories, err := repository.FindSubcategories()

		// Then: aparecen las tres categorías anidadas, no la de primer nivel.
		require.Nil(t, err, "FindSubcategories no debería fallar con un dataset válido")
		assert.ElementsMatch(t, []string{"light", "water", "gas"}, categoryIds(categories), "FindSubcategories debería devolver solo las categorías anidadas")
	})

	t.Run("with a nested category, resolves it with the parent's data", func(t *testing.T) {

		// Given: el repositorio de prueba armado sobre un dataset de dos niveles.
		repository := mustNewYamlCategoryRepositoryFromFiles(t, twoLevelCategoryFiles())

		// When: se consultan las subcategorías.
		categories, err := repository.FindSubcategories()
		require.Nil(t, err, "FindSubcategories no debería fallar con un dataset válido")

		nestedCategory := categoryById(categories, "light")

		// Then: tiene un padre con el id referenciado por category_id.
		require.NotNil(t, nestedCategory, "la categoría anidada %q debería estar en la colección", "light")
		require.NotNil(t, nestedCategory.Parent, "la categoría anidada %q debería tener padre", "light")
		assert.Equal(t, "services", nestedCategory.Parent.Id, "el padre resuelto debería tener el id referenciado por category_id")
	})

	t.Run("with a category that references a nonexistent parent, fails", func(t *testing.T) {

		// Given: el repositorio de prueba armado sobre un archivo que referencia
		// una categoría padre que no está definida en ningún lado.
		repository := mustNewYamlCategoryRepositoryFromFiles(t, nonexistentParentCategoryFiles())

		// When: se consultan las subcategorías.
		categories, err := repository.FindSubcategories()

		// Then: falla en vez de devolver una colección parcial o con un padre nulo silencioso.
		require.NotNil(t, err, "FindSubcategories debería fallar ante una categoría que referencia un padre inexistente")
		assert.Nil(t, categories, "FindSubcategories no debería devolver una colección parcial cuando falla")
		assert.Contains(t, err.Error(), "no existe", "el error debería indicar que el padre referenciado no existe")
	})
}

// TestIntegrityCheck agrupa los escenarios donde los archivos de categorías
// violan las reglas de integridad que el repositorio verifica: ids duplicados
// (al construirse, en indexRecords) y padres inexistentes (al consultar, en
// ParentLookup).
func TestIntegrityCheck(t *testing.T) {

	t.Run("with a duplicate category id, construction fails", func(t *testing.T) {

		// Given/When: se construye el repositorio sobre un archivo que define el
		// id "services" dos veces.
		_, err := newYamlCategoryRepositoryFromFiles(t, duplicateCategoryIdFiles())

		// Then: la construcción falla en vez de indexar la primera o la última definición.
		require.NotNil(t, err, "la construcción debería fallar ante una categoría duplicada")
		assert.Contains(t, err.Error(), "duplicada", "el error debería indicar que la categoría está duplicada")
	})

	t.Run("with a category that references a nonexistent parent, FindAll fails", func(t *testing.T) {

		// Given: el repositorio de prueba armado sobre un archivo que referencia
		// una categoría padre que no está definida en ningún lado.
		repository := mustNewYamlCategoryRepositoryFromFiles(t, nonexistentParentCategoryFiles())

		// When: se consulta la colección completa.
		categories, err := repository.FindAll()

		// Then: falla en vez de cargar la categoría sin padre o con un padre nulo silencioso.
		require.NotNil(t, err, "FindAll debería fallar ante una categoría que referencia un padre inexistente")
		assert.Nil(t, categories, "FindAll no debería devolver una colección parcial cuando falla")
		assert.Contains(t, err.Error(), "no existe", "el error debería indicar que el padre referenciado no existe")
	})
}
