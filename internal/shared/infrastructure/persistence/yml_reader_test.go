package persistence

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mapFS arma un fs.FS en memoria a partir de un mapa nombre de archivo →
// contenido. El YMLReader no interpreta el contenido, así que alcanza con
// bytes arbitrarios: lo único relevante para sus tests es el nombre.
func mapFS(files map[string]string) fstest.MapFS {

	fsys := fstest.MapFS{}
	for name, content := range files {
		fsys[name] = &fstest.MapFile{Data: []byte(content)}
	}

	return fsys
}

// newYMLReader arma el lector y falla el test si path no es válido, para que
// los casos que solo ejercitan Read no tengan que repetir esa verificación.
func newYMLReader(t *testing.T, files fstest.MapFS, path string) *YMLReader {
	t.Helper()

	reader, err := NewYMLReader(files, path)
	require.Nil(t, err, "NewYMLReader no debería fallar con la ruta %q", path)

	return reader
}

func TestNewYMLReader(t *testing.T) {

	t.Run("with an invalid path, fails", func(t *testing.T) {

		// Given: un origen de archivos válido.
		files := mapFS(map[string]string{
			"categories.0.yml": "content",
		})

		// When: se crea el lector con una ruta que no respeta las reglas de fs.FS.
		reader, err := NewYMLReader(files, "")

		// Then: falla en vez de quedar apuntando a un subárbol indefinido.
		require.NotNil(t, err, "NewYMLReader debería fallar ante una ruta inválida")
		assert.Nil(t, reader, "NewYMLReader no debería devolver un lector cuando falla")
	})

	t.Run("with \".\", reads from the root of files", func(t *testing.T) {

		// Given: un lector apuntado a la raíz del origen con ".".
		reader := newYMLReader(t, mapFS(map[string]string{
			"categories.0.yml": "content",
		}), ".")

		// When: se leen los archivos que matchean el patrón.
		contents, err := reader.Read("categories.*.yml")

		// Then: encuentra el archivo ubicado en la raíz.
		require.Nil(t, err, "Read no debería fallar con archivos válidos")
		assert.Equal(t, [][]byte{[]byte("content")}, contents)
	})

	t.Run("with a subdirectory path, only reads files under it", func(t *testing.T) {

		// Given: dos datasets bajo distintas subcarpetas del mismo origen.
		files := mapFS(map[string]string{
			"01/categories.0.yml": "dataset 01",
			"02/categories.0.yml": "dataset 02",
		})

		// When: se lee un lector acotado a una sola subcarpeta.
		reader := newYMLReader(t, files, "01")
		contents, err := reader.Read("categories.*.yml")

		// Then: solo aparece el archivo de esa subcarpeta.
		require.Nil(t, err, "Read no debería fallar con archivos válidos")
		assert.Equal(t, [][]byte{[]byte("dataset 01")}, contents, "Read no debería ver archivos de otra ruta")
	})
}

func TestYMLReader_Read(t *testing.T) {

	t.Run("with files that match the pattern, returns their content ordered by level", func(t *testing.T) {

		// Given: un lector sobre archivos en orden alfabético inverso al numérico.
		reader := newYMLReader(t, mapFS(map[string]string{
			"categories.1.yml": "level one",
			"categories.0.yml": "level zero",
		}), ".")

		// When: se leen los archivos que matchean el patrón.
		contents, err := reader.Read("categories.*.yml")

		// Then: el contenido queda ordenado de menor a mayor nivel.
		require.Nil(t, err, "Read no debería fallar con archivos válidos")
		assert.Equal(t, [][]byte{[]byte("level zero"), []byte("level one")}, contents, "Read debería devolver el contenido ordenado por nivel numérico")
	})

	t.Run("with a two-digit level, orders it after single-digit levels", func(t *testing.T) {

		// Given: un lector con un nivel de dos dígitos, que el orden lexicográfico
		// ubicaría antes que el de un dígito.
		reader := newYMLReader(t, mapFS(map[string]string{
			"categories.10.yml": "level ten",
			"categories.2.yml":  "level two",
		}), ".")

		// When: se leen los archivos.
		contents, err := reader.Read("categories.*.yml")

		// Then: el nivel 2 queda antes que el 10, aunque "10" ordene antes que "2" como string.
		require.Nil(t, err, "Read no debería fallar con archivos válidos")
		assert.Equal(t, [][]byte{[]byte("level two"), []byte("level ten")}, contents, "Read debería ordenar por el valor numérico del nivel, no por el nombre del archivo")
	})

	t.Run("with no files matching the pattern, fails", func(t *testing.T) {

		// Given: un lector sin archivos que matcheen el patrón.
		reader := newYMLReader(t, mapFS(map[string]string{
			"beneficiaries.0.yml": "content",
		}), ".")

		// When: se leen los archivos con un patrón que no matchea nada.
		contents, err := reader.Read("categories.*.yml")

		// Then: falla en vez de devolver una colección vacía silenciosamente.
		require.NotNil(t, err, "Read debería fallar si ningún archivo respeta el patrón")
		assert.Nil(t, contents, "Read no debería devolver contenido cuando falla")
	})

	t.Run("with a file whose level is not a number, fails", func(t *testing.T) {

		// Given: un archivo que matchea el patrón pero no numera su nivel.
		reader := newYMLReader(t, mapFS(map[string]string{
			"categories.foo.yml": "content",
		}), ".")

		// When: se leen los archivos.
		contents, err := reader.Read("categories.*.yml")

		// Then: falla, en vez de tratarlo como si no tuviera orden.
		require.NotNil(t, err, "Read debería fallar ante un archivo que no numera su nivel")
		assert.Nil(t, contents, "Read no debería devolver contenido cuando falla")
	})

	t.Run("with two files that number the same level, fails", func(t *testing.T) {

		// Given: dos archivos que, numerados de forma distinta, resuelven al mismo nivel.
		reader := newYMLReader(t, mapFS(map[string]string{
			"categories.01.yml": "content a",
			"categories.1.yml":  "content b",
		}), ".")

		// When: se leen los archivos.
		contents, err := reader.Read("categories.*.yml")

		// Then: falla, en vez de descartar uno de los dos en silencio.
		require.NotNil(t, err, "Read debería fallar ante dos archivos que numeran el mismo nivel")
		assert.Nil(t, contents, "Read no debería devolver contenido cuando falla")
	})

	t.Run("with an invalid glob pattern, fails", func(t *testing.T) {

		// Given: un lector sobre cualquier origen de archivos.
		reader := newYMLReader(t, mapFS(map[string]string{
			"categories.0.yml": "content",
		}), ".")

		// When: se lo consulta con un patrón sintácticamente inválido.
		contents, err := reader.Read("categories.[.yml")

		// Then: falla en vez de propagar un resultado vacío indistinguible de "no matchea nada".
		require.NotNil(t, err, "Read debería fallar ante un patrón inválido")
		assert.Nil(t, contents, "Read no debería devolver contenido cuando falla")
	})
}