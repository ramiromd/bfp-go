package persistence

import (
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"
)

// YMLReader descubre y lee los archivos que componen un YML particionado por
// niveles (por ejemplo categories.0.yml, categories.1.yml...), sin conocer su
// contenido. No interpreta YAML: un archivo con contenido corrupto no es un
// error de este componente. Solo valida la convención de nombres: que haya al
// menos un archivo y que la numeración de niveles sea coherente.
type YMLReader struct {
	files fs.FS
}

// NewYMLReader crea un lector acotado a path dentro del origen de archivos
// indicado. path es obligatorio: para leer desde la raíz de files hay que
// pasar explícitamente ".". Falla si path no es una ruta válida dentro de
// files.
func NewYMLReader(files fs.FS, path string) (*YMLReader, error) {

	scoped, err := fs.Sub(files, path)
	if err != nil {
		return nil, fmt.Errorf("La ruta %q no es válida: %w", path, err)
	}

	return &YMLReader{files: scoped}, nil
}

// Read descubre los archivos que matchean pattern y devuelve su contenido
// crudo, ordenado de menor a mayor nivel. pattern debe tener exactamente un
// '*', que es donde se espera el número de nivel (ej. "categories.*.yml"): el
// orden se resuelve por ese número, no por orden lexicográfico del nombre de
// archivo, así que un nivel de dos dígitos ordena correctamente respecto de
// uno de un dígito.
func (this *YMLReader) Read(pattern string) ([][]byte, error) {

	fileNames, err := fs.Glob(this.files, pattern)
	if err != nil {
		return nil, fmt.Errorf("El patrón %q no es válido: %w", pattern, err)
	}

	if len(fileNames) == 0 {
		return nil, fmt.Errorf("No se encontró ningún archivo que respete el patrón %q", pattern)
	}

	levels, err := this.levelsOf(pattern, fileNames)
	if err != nil {
		return nil, err
	}

	sort.Slice(fileNames, func(i, j int) bool {
		return levels[fileNames[i]] < levels[fileNames[j]]
	})

	contents := make([][]byte, 0, len(fileNames))
	for _, fileName := range fileNames {

		content, err := fs.ReadFile(this.files, fileName)
		if err != nil {
			return nil, fmt.Errorf("No se pudo leer el archivo %s: %w", fileName, err)
		}

		contents = append(contents, content)
	}

	return contents, nil
}

// levelsOf extrae el nivel numérico de cada archivo a partir de la posición
// del '*' en pattern, y valida que ningún nivel se repita entre archivos.
func (this *YMLReader) levelsOf(pattern string, fileNames []string) (map[string]int, error) {

	starIndex := strings.IndexByte(pattern, '*')
	if starIndex == -1 {
		return nil, fmt.Errorf("El patrón %q no tiene un '*' que indique dónde va el nivel", pattern)
	}
	prefix, suffix := pattern[:starIndex], pattern[starIndex+1:]

	levels := make(map[string]int, len(fileNames))
	seenLevels := make(map[int]string, len(fileNames))

	for _, fileName := range fileNames {

		token := strings.TrimSuffix(strings.TrimPrefix(fileName, prefix), suffix)

		level, err := strconv.Atoi(token)
		if err != nil {
			return nil, fmt.Errorf("El archivo %s no respeta la convención de numeración esperada por %q: %w", fileName, pattern, err)
		}

		if other, exists := seenLevels[level]; exists {
			return nil, fmt.Errorf("Los archivos %s y %s numeran el mismo nivel (%d)", other, fileName, level)
		}

		levels[fileName] = level
		seenLevels[level] = fileName
	}

	return levels, nil
}