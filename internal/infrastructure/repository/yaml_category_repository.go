package repository

import (
	_ "embed"
	"fmt"

	"ramiromd/budget/internal/domain/entity"
	domain "ramiromd/budget/internal/domain/repository"

	"gopkg.in/yaml.v3"
)

// Verifica en tiempo de compilación que el repositorio cumple el contrato del
// dominio.
var _ domain.ICategoryRepository = (*YamlCategoryRepository)(nil)

// categoriesYml es el contenido de categories.yml. Se incrusta en el binario
// durante la compilación, de modo que el repositorio no depende del directorio
// de trabajo ni del archivo en disco.
//
//go:embed categories.yml
var categoriesYml []byte

// categoryFile representa la estructura del archivo categories.yml.
type categoryFile struct {
	Categories []categoryRecord `yaml:"categories"`
}

// categoryRecord representa una categoría tal como está escrita en el archivo.
type categoryRecord struct {
	Id          string `yaml:"id"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// YamlCategoryRepository expone las categorías definidas en categories.yml. El
// archivo se interpreta una única vez y el resultado queda en memoria.
//
// No es seguro para uso concurrente.
type YamlCategoryRepository struct {
	categories []*entity.Category
}

// NewYamlCategoryRepository crea el repositorio de categorías. El archivo no se
// interpreta hasta la primera consulta.
func NewYamlCategoryRepository() *YamlCategoryRepository {
	return &YamlCategoryRepository{}
}

// FindAll devuelve todas las categorías. Retorna un error si el origen de
// datos no puede leerse.
func (this *YamlCategoryRepository) FindAll() ([]*entity.Category, error) {

	if err := this.load(); err != nil {
		return nil, err
	}

	return this.categories, nil
}

// FindById devuelve la categoría con el identificador indicado, o nil si no
// existe. Retorna un error si el origen de datos no puede leerse.
func (this *YamlCategoryRepository) FindById(id string) (*entity.Category, error) {

	if err := this.load(); err != nil {
		return nil, err
	}

	for _, category := range this.categories {
		if category.Id == id {
			return category, nil
		}
	}

	return nil, nil
}

// load interpreta el archivo incrustado y arma las entidades. Las llamadas
// posteriores a la primera reutilizan el resultado.
func (this *YamlCategoryRepository) load() error {

	if this.categories != nil {
		return nil
	}

	file := categoryFile{}
	if err := yaml.Unmarshal(categoriesYml, &file); err != nil {
		return fmt.Errorf("el archivo categories.yml no tiene un formato válido: %w", err)
	}

	categories := make([]*entity.Category, 0, len(file.Categories))
	for _, record := range file.Categories {
		categories = append(categories, entity.NewCategory(record.Id, record.Name, record.Description))
	}

	this.categories = categories

	return nil
}
