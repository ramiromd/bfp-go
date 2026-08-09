package persistence

import (
	"embed"
	"fmt"
	"io/fs"
	"sort"

	"ramiromd/budget/internal/category/domain/entity"
	"ramiromd/budget/internal/category/domain/repository"

	"gopkg.in/yaml.v3"
)

// Verifica en tiempo de compilación que el repositorio cumple el contrato del
// dominio.
var _ repository.CategoryRepository = (*YamlCategoryRepository)(nil)

// categoriesPattern es el patrón que nombra los archivos de categorías: uno por
// nivel de la jerarquía, numerados de menor a mayor.
const categoriesPattern = "categories.*.yml"

// categoriesFS contiene los archivos categories.<n>.yml, un archivo por nivel
// de la jerarquía. Se incrustan en el binario durante la compilación, de modo
// que el repositorio no depende del directorio de trabajo ni de los archivos en
// disco.
//
// Los archivos se descubren en runtime, así que agregar un nivel es agregar el
// archivo: no hace falta tocar este código.
//
//go:embed categories.*.yml
var categoriesFS embed.FS

// categoryFile representa la estructura de un archivo categories.<n>.yml.
type categoryFile struct {
	Categories []categoryRecord `yaml:"categories"`
}

// categoryRecord representa una categoría tal como está escrita en el archivo.
type categoryRecord struct {
	Id          string `yaml:"id"`
	Name        string `yaml:"name"`
	CategoryId  string `yaml:"category_id"`
	Description string `yaml:"description"`
}

// YamlCategoryRepository expone las categorías definidas en los archivos
// categories.<n>.yml. Los archivos se interpretan una única vez y el resultado
// queda en memoria: categories conserva el orden de carga e indexes mapea cada
// identificador a su posición.
//
// No es seguro para uso concurrente.
type YamlCategoryRepository struct {
	initialized bool
	categories  []*entity.Category
	indexes     map[string]int
}

// NewYamlCategoryRepository crea el repositorio de categorías. El archivo no se
// interpreta hasta la primera consulta.
func NewYamlCategoryRepository() *YamlCategoryRepository {
	return &YamlCategoryRepository{
		initialized: false,
		categories:  nil,
		indexes:     nil,
	}
}

// FindAll devuelve todas las categorías, de todos los niveles. Retorna un error
// si el origen de datos no puede leerse.
func (this *YamlCategoryRepository) FindAll() ([]*entity.Category, error) {

	if err := this.initialize(); err != nil {
		return nil, err
	}

	return this.categories, nil
}

// FindById devuelve la categoría con el identificador indicado, o nil si no
// existe. Retorna un error si el origen de datos no puede leerse.
func (this *YamlCategoryRepository) FindById(id string) (*entity.Category, error) {

	if err := this.initialize(); err != nil {
		return nil, err
	}

	position, exists := this.indexes[id]
	if !exists {
		return nil, nil
	}

	return this.categories[position], nil
}

// Count devuelve la cantidad de categorías definidas. Retorna un error si el
// origen de datos no puede leerse.
func (this *YamlCategoryRepository) Count() (int, error) {

	if err := this.initialize(); err != nil {
		return 0, err
	}

	return len(this.categories), nil
}

// initialize descubre los archivos incrustados y los interpreta en una única
// colección de categorías. Las llamadas posteriores a la primera reutilizan el
// resultado.
func (this *YamlCategoryRepository) initialize() error {

	if this.initialized {
		return nil
	}

	fileNames, err := fs.Glob(categoriesFS, categoriesPattern)
	if err != nil {
		return fmt.Errorf("No se pudieron listar los archivos de categorías: %w", err)
	}

	// El orden importa: una categoría sólo puede referenciar como padre a otra de
	// un nivel anterior, así que al recorrer los archivos de menor a mayor nivel
	// el padre ya está armado cuando se lo necesita. El orden lexicográfico
	// alcanza mientras los niveles sean de un dígito; del décimo en adelante
	// habría que numerarlos con dos ("categories.09.yml").
	sort.Strings(fileNames)

	this.categories = make([]*entity.Category, 0)
	this.indexes = make(map[string]int)

	for _, fileName := range fileNames {

		content, err := categoriesFS.ReadFile(fileName)
		if err != nil {
			return fmt.Errorf("No se pudo leer el archivo %s: %w", fileName, err)
		}

		if err := this.load(content, fileName); err != nil {
			return err
		}
	}

	this.initialized = true

	return nil
}

// load interpreta el contenido de un archivo de categorías y suma sus entidades
// a la colección. Espera que las categorías referenciadas como padre ya estén
// cargadas.
func (this *YamlCategoryRepository) load(content []byte, fileName string) error {

	file := categoryFile{}
	if err := yaml.Unmarshal(content, &file); err != nil {
		return fmt.Errorf("El archivo %s no tiene un formato válido: %w", fileName, err)
	}

	for _, record := range file.Categories {

		if _, exists := this.indexes[record.Id]; exists {
			return fmt.Errorf("El archivo %s define una categoría duplicada: %q", fileName, record.Id)
		}

		var parent *entity.Category

		if record.CategoryId != "" {
			position, exists := this.indexes[record.CategoryId]
			if !exists {
				return fmt.Errorf("La categoría %q del archivo %s referencia una categoría padre inexistente: %q", record.Id, fileName, record.CategoryId)
			}
			parent = this.categories[position]
		}

		category := entity.NewCategory(record.Id, parent, record.Name, record.Description)
		this.categories = append(this.categories, category)
		this.indexes[record.Id] = len(this.categories) - 1
	}

	return nil
}
