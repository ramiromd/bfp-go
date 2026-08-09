package persistence

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// expectedCategoryCount es la cantidad de categorías definidas entre todos los
// archivos categories.<n>.yml incrustados: 4 en categories.0.yml y 8 en
// categories.1.yml. Hay que actualizarlo al agregar o quitar categorías.
const expectedCategoryCount = 12

func TestCount(t *testing.T) {

	t.Run("with embedded category files, returns the total number of categories", func(t *testing.T) {

		// Given: un repositorio recién creado, que todavía no interpretó los archivos.
		repository := NewYamlCategoryRepository()

		// When: se consulta la cantidad de categorías.
		count, err := repository.Count()

		// Then: se obtiene el total definido en los archivos, sin error.
		require.Nil(t, err, "Count should not fail while reading the embedded category files")
		assert.Equal(t, expectedCategoryCount, count, "Count should add up the categories of every category file, not only the first one")
	})

	t.Run("returns the same number of categories as FindAll", func(t *testing.T) {

		// Given: un repositorio recién creado.
		repository := NewYamlCategoryRepository()

		// When: se consulta la cantidad de categorías y la colección completa.
		count, countError := repository.Count()

		categories, findAllError := repository.FindAll()

		// Then: ambos valores describen la misma colección.
		require.Nil(t, countError, "Count should not fail while reading the embedded category files")
		require.Nil(t, findAllError, "FindAll should not fail while reading the embedded category files")
		assert.Equal(t, len(categories), count, "Count and FindAll should describe the same collection")
	})

	t.Run("called twice, returns the same value", func(t *testing.T) {

		// Given: un repositorio que ya interpretó los archivos en una primera consulta.
		repository := NewYamlCategoryRepository()

		firstCall, firstCallError := repository.Count()

		// When: se vuelve a consultar la cantidad de categorías.
		secondCall, secondCallError := repository.Count()

		// Then: se obtiene el mismo valor que la primera vez, sin error.
		require.Nil(t, firstCallError, "the first call to Count should not fail")
		require.Nil(t, secondCallError, "the second call to Count should not fail")
		assert.Equal(t, firstCall, secondCall, "Count should reuse the already loaded collection instead of appending the categories again")
	})
}
