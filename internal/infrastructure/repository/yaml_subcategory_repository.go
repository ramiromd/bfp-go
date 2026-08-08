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
var _ domain.ISubcategoryRepository = (*YamlSubcategoryRepository)(nil)

// subcategoriesYml es el contenido de subcategories.yml. Se incrusta en el
// binario durante la compilación, de modo que el repositorio no depende del
// directorio de trabajo ni del archivo en disco.
//
//go:embed subcategories.yml
var subcategoriesYml []byte

// subcategoryFile representa la estructura del archivo subcategories.yml.
type subcategoryFile struct {
	Subcategories []subcategoryRecord `yaml:"subcategories"`
}

// subcategoryRecord representa una subcategoría tal como está escrita en el
// archivo.
type subcategoryRecord struct {
	Id          string `yaml:"id"`
	CategoryId  string `yaml:"category_id"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// YamlSubcategoryRepository expone las subcategorías definidas en
// subcategories.yml. El archivo se interpreta una única vez y el resultado
// queda en memoria.
//
// No es seguro para uso concurrente.
type YamlSubcategoryRepository struct {
	subcategories []*entity.Subcategory
}

// NewYamlSubcategoryRepository crea el repositorio de subcategorías. El archivo
// no se interpreta hasta la primera consulta.
func NewYamlSubcategoryRepository() *YamlSubcategoryRepository {
	return &YamlSubcategoryRepository{}
}

// FindAll devuelve todas las subcategorías. Retorna un error si el origen de
// datos no puede leerse.
func (this *YamlSubcategoryRepository) FindAll() ([]*entity.Subcategory, error) {

	if err := this.load(); err != nil {
		return nil, err
	}

	return this.subcategories, nil
}

// FindById devuelve la subcategoría con el identificador indicado, o nil si no
// existe. Retorna un error si el origen de datos no puede leerse.
func (this *YamlSubcategoryRepository) FindById(id string) (*entity.Subcategory, error) {

	if err := this.load(); err != nil {
		return nil, err
	}

	for _, subcategory := range this.subcategories {
		if subcategory.Id == id {
			return subcategory, nil
		}
	}

	return nil, nil
}

// FindByCategoryId devuelve las subcategorías que pertenecen a la categoría
// indicada. Devuelve una lista vacía si la categoría no tiene subcategorías.
// Retorna un error si el origen de datos no puede leerse.
func (this *YamlSubcategoryRepository) FindByCategoryId(categoryId string) ([]*entity.Subcategory, error) {

	if err := this.load(); err != nil {
		return nil, err
	}

	result := make([]*entity.Subcategory, 0)
	for _, subcategory := range this.subcategories {
		if subcategory.CategoryId == categoryId {
			result = append(result, subcategory)
		}
	}

	return result, nil
}

// load interpreta el archivo incrustado y arma las entidades. Las llamadas
// posteriores a la primera reutilizan el resultado.
func (this *YamlSubcategoryRepository) load() error {

	if this.subcategories != nil {
		return nil
	}

	file := subcategoryFile{}
	if err := yaml.Unmarshal(subcategoriesYml, &file); err != nil {
		return fmt.Errorf("el archivo subcategories.yml no tiene un formato válido: %w", err)
	}

	subcategories := make([]*entity.Subcategory, 0, len(file.Subcategories))
	for _, record := range file.Subcategories {
		subcategories = append(subcategories, entity.NewSubcategory(record.Id, record.CategoryId, record.Name, record.Description))
	}

	this.subcategories = subcategories

	return nil
}
