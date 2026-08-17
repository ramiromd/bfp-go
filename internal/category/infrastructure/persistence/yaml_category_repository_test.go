package persistence

import (
	"embed"
	"io/fs"
	"path"
	"testing"

	"ramiromd/budget/internal/category/domain/entity"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testCategoriesFS contiene los datasets de prueba bajo yml_test/<dataset>, uno
// por escenario. Se incrustan igual que los archivos productivos, así que los
// tests usan el mismo mecanismo de carga (embed.FS + fs.Glob) en vez de leer
// del disco.
//
//go:embed yml_test
var testCategoriesFS embed.FS

// newYamlCategoryRepositoryFromDataset arma el repositorio con el constructor
// parametrizado, sobre el dataset de prueba indicado (por ejemplo "00",
// ubicado en yml_test/00).
func newYamlCategoryRepositoryFromDataset(t *testing.T, dataset string) *YamlCategoryRepository {
	t.Helper()

	files, err := fs.Sub(testCategoriesFS, path.Join("yml_test", dataset))
	require.Nil(t, err, "el dataset de prueba %q debería existir bajo yml_test", dataset)

	return NewYamlCategoryRepositoryFrom(files)
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

// expectedProductionCategoryCount es la cantidad de categorías definidas en los
// archivos productivos incrustados en yml/: 4 en categories.0.yml y 8 en
// categories.1.yml. Hay que actualizarlo al agregar o quitar categorías.
const expectedProductionCategoryCount = 12

// TestNewYamlCategoryRepository_Count verifica que el repositorio productivo --
// el que arma NewYamlCategoryRepository sobre los archivos incrustados en yml/
// -- cargue la cantidad de categorías esperada.
func TestNewYamlCategoryRepository_Count(t *testing.T) {

	t.Run("with the production embedded files, returns the total number of categories", func(t *testing.T) {

		// Given: el repositorio productivo, construido sobre los archivos incrustados en yml/.
		repository := NewYamlCategoryRepository()

		// When: se consulta la cantidad de categorías.
		count, err := repository.Count()

		// Then: se obtiene el total definido en los archivos productivos.
		require.Nil(t, err, "Count no debería fallar al leer los archivos productivos")
		assert.Equal(t, expectedProductionCategoryCount, count, "Count debería sumar las categorías de todos los archivos productivos")
	})
}

func TestYamlCategoryRepositoryFrom_FindAll(t *testing.T) {

	t.Run("with the yml_test/00 dataset, returns every category defined in it", func(t *testing.T) {

		// Given: el repositorio de prueba armado sobre yml_test/00.
		repository := newYamlCategoryRepositoryFromDataset(t, "00")

		// When: se consulta la colección completa.
		categories, err := repository.FindAll()

		// Then: están las categorías de los dos archivos del dataset. El orden de
		// carga no importa acá, solo que estén todas.
		require.Nil(t, err, "FindAll no debería fallar al leer el dataset yml_test/00")
		assert.ElementsMatch(t, []string{"services", "light", "water", "gas"}, categoryIds(categories), "FindAll debería devolver las categorías de todos los archivos del dataset")
	})

	t.Run("with a category that has no parent, returns it detached", func(t *testing.T) {

		// Given: el repositorio de prueba armado sobre yml_test/00.
		repository := newYamlCategoryRepositoryFromDataset(t, "00")

		// When: se consulta la colección completa.
		categories, err := repository.FindAll()
		require.Nil(t, err, "FindAll no debería fallar al leer el dataset yml_test/00")

		topLevelCategory := categoryById(categories, "services")

		// Then: la categoría de primer nivel no tiene padre.
		require.NotNil(t, topLevelCategory, "la categoría de primer nivel %q debería estar en la colección", "services")
		assert.Nil(t, topLevelCategory.Parent, "una categoría con category_id vacío no debería tener padre")
	})

	t.Run("with a category that has a parent, resolves it to the already loaded instance", func(t *testing.T) {

		// Given: el repositorio de prueba armado sobre yml_test/00.
		repository := newYamlCategoryRepositoryFromDataset(t, "00")

		// When: se consulta la colección completa.
		categories, err := repository.FindAll()
		require.Nil(t, err, "FindAll no debería fallar al leer el dataset yml_test/00")

		nestedCategory := categoryById(categories, "light")
		parentCategory := categoryById(categories, "services")

		// Then: la categoría anidada apunta a la instancia de su padre, no a una copia.
		require.NotNil(t, nestedCategory, "la categoría anidada %q debería estar en la colección", "light")
		require.NotNil(t, parentCategory, "la categoría padre %q debería estar en la colección", "services")
		assert.Same(t, parentCategory, nestedCategory.Parent, "una categoría anidada debería apuntar a la instancia del padre ya cargada")
	})

	t.Run("called twice, returns the same collection", func(t *testing.T) {

		// Given: un repositorio de prueba que ya interpretó los archivos en una primera consulta.
		repository := newYamlCategoryRepositoryFromDataset(t, "00")

		firstCall, firstCallErr := repository.FindAll()

		// When: se vuelve a consultar la colección completa.
		secondCall, secondCallErr := repository.FindAll()

		// Then: se obtienen las mismas categorías que la primera vez.
		require.Nil(t, firstCallErr, "la primera llamada a FindAll no debería fallar")
		require.Nil(t, secondCallErr, "la segunda llamada a FindAll no debería fallar")
		assert.Equal(t, firstCall, secondCall, "ambas llamadas a FindAll deberían devolver la misma colección")
	})
}

func TestYamlCategoryRepositoryFrom_FindById(t *testing.T) {

	t.Run("with an existing id, returns the matching category", func(t *testing.T) {

		// Given: el repositorio de prueba armado sobre yml_test/00.
		repository := newYamlCategoryRepositoryFromDataset(t, "00")

		// When: se busca una categoría por id.
		category, err := repository.FindById("light")

		// Then: se obtiene la categoría con ese id.
		require.Nil(t, err, "FindById no debería fallar al leer el dataset yml_test/00")
		require.NotNil(t, category, "FindById debería encontrar la categoría %q", "light")
		assert.Equal(t, "light", category.Id)
		assert.Equal(t, "Electricidad", category.Name)
	})

	t.Run("with a top-level id, returns the category without a parent", func(t *testing.T) {

		// Given: el repositorio de prueba armado sobre yml_test/00.
		repository := newYamlCategoryRepositoryFromDataset(t, "00")

		// When: se busca la categoría de primer nivel.
		category, err := repository.FindById("services")

		// Then: se obtiene sin padre.
		require.Nil(t, err, "FindById no debería fallar al leer el dataset yml_test/00")
		require.NotNil(t, category, "FindById debería encontrar la categoría %q", "services")
		assert.Nil(t, category.Parent, "la categoría de primer nivel no debería tener padre")
	})

	t.Run("with a nested id, returns the category pointing to its parent", func(t *testing.T) {

		// Given: el repositorio de prueba armado sobre yml_test/00.
		repository := newYamlCategoryRepositoryFromDataset(t, "00")

		// When: se busca una categoría anidada.
		category, err := repository.FindById("gas")

		// Then: se obtiene apuntando a su padre.
		require.Nil(t, err, "FindById no debería fallar al leer el dataset yml_test/00")
		require.NotNil(t, category, "FindById debería encontrar la categoría %q", "gas")
		require.NotNil(t, category.Parent, "la categoría %q debería tener padre", "gas")
		assert.Equal(t, "services", category.Parent.Id)
	})

	t.Run("with an id that does not exist, returns nil without error", func(t *testing.T) {

		// Given: el repositorio de prueba armado sobre yml_test/00.
		repository := newYamlCategoryRepositoryFromDataset(t, "00")

		// When: se busca un id que no está en el dataset.
		category, err := repository.FindById("does-not-exist")

		// Then: se obtiene nil, sin error.
		require.Nil(t, err, "FindById no debería fallar al leer el dataset yml_test/00")
		assert.Nil(t, category, "FindById debería devolver nil para un id inexistente")
	})
}

func TestYamlCategoryRepositoryFrom_Count(t *testing.T) {

	t.Run("with the yml_test/00 dataset, returns the total number of categories", func(t *testing.T) {

		// Given: el repositorio de prueba armado sobre yml_test/00.
		repository := newYamlCategoryRepositoryFromDataset(t, "00")

		// When: se consulta la cantidad de categorías.
		count, err := repository.Count()

		// Then: se obtiene el total definido en el dataset: 1 en categories.0.yml y 3 en categories.1.yml.
		require.Nil(t, err, "Count no debería fallar al leer el dataset yml_test/00")
		assert.Equal(t, 4, count, "Count debería sumar las categorías de todos los archivos del dataset")
	})

	t.Run("returns the same number of categories as FindAll", func(t *testing.T) {

		// Given: el repositorio de prueba armado sobre yml_test/00.
		repository := newYamlCategoryRepositoryFromDataset(t, "00")

		// When: se consulta la cantidad de categorías y la colección completa.
		count, countErr := repository.Count()
		categories, findAllErr := repository.FindAll()

		// Then: ambos valores describen la misma colección.
		require.Nil(t, countErr, "Count no debería fallar al leer el dataset yml_test/00")
		require.Nil(t, findAllErr, "FindAll no debería fallar al leer el dataset yml_test/00")
		assert.Equal(t, len(categories), count, "Count y FindAll deberían describir la misma colección")
	})
}

// TestIntegrityCheck agrupa los escenarios donde los archivos de categorías
// violan las reglas de integridad que el repositorio verifica al cargar: ids
// duplicados y padres inexistentes. Cada escenario usa un dataset con un único
// archivo, para que el error no dependa de qué otro archivo se cargó antes.
func TestIntegrityCheck(t *testing.T) {

	t.Run("with a duplicate category id, FindAll fails", func(t *testing.T) {

		// Given: el repositorio de prueba armado sobre yml_test/01, cuyo único
		// archivo define el id "services" dos veces.
		repository := newYamlCategoryRepositoryFromDataset(t, "01")

		// When: se consulta la colección completa.
		categories, err := repository.FindAll()

		// Then: falla en vez de quedarse con la primera o la última definición.
		require.NotNil(t, err, "FindAll debería fallar ante una categoría duplicada")
		assert.Nil(t, categories, "FindAll no debería devolver una colección parcial cuando falla")
		assert.Contains(t, err.Error(), "duplicada", "el error debería indicar que la categoría está duplicada")
	})

	t.Run("with a category that references a nonexistent parent, FindAll fails", func(t *testing.T) {

		// Given: el repositorio de prueba armado sobre yml_test/02, cuyo único
		// archivo referencia categorías padre que no están definidas en ningún lado.
		repository := newYamlCategoryRepositoryFromDataset(t, "02")

		// When: se consulta la colección completa.
		categories, err := repository.FindAll()

		// Then: falla en vez de cargar la categoría sin padre o con un padre nulo silencioso.
		require.NotNil(t, err, "FindAll debería fallar ante una categoría que referencia un padre inexistente")
		assert.Nil(t, categories, "FindAll no debería devolver una colección parcial cuando falla")
		assert.Contains(t, err.Error(), "inexistente", "el error debería indicar que el padre referenciado no existe")
	})
}
