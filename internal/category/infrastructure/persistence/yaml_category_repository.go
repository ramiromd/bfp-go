package persistence

import (
	"fmt"

	"ramiromd/budget/internal/category/domain/entity"
)

// YamlCategoryRepository expone las categorías definidas en los archivos
// categories.<n>.yml. Los archivos se interpretan una única vez y el resultado
// queda en memoria: categories conserva el orden de carga e indexes mapea cada
// identificador a su posición.
//
// No es seguro para uso concurrente.
type YamlCategoryRepository struct {
	indexes    map[string]int
	datasource *YamlCategoryDatasource
}

// NewYamlCategoryRepositoryFromDatasource crea el repositorio a partir de un
// YamlCategoryDatasource ya construido. Si la indexación falla (por ejemplo,
// por un id duplicado), el error queda guardado y toda consulta posterior lo
// devuelve, en vez de operar sobre un índice parcial.
func NewYamlCategoryRepositoryFromDatasource(datasource *YamlCategoryDatasource) (*YamlCategoryRepository, error) {
	self := &YamlCategoryRepository{
		indexes:    make(map[string]int, 0),
		datasource: datasource,
	}

	if err := self.indexRecords(); err != nil {
		return &YamlCategoryRepository{}, err
	}

	return self, nil
}

func (this *YamlCategoryRepository) ParentLookup(parentId string) (*entity.Category, error) {

	if parentId == "" {
		return nil, nil
	}

	position, exists := this.indexes[parentId]
	var parent *entity.Category

	if exists {
		record, _ := this.datasource.RecordAt(position)
		parent = entity.NewCategory(record.Id, nil, record.Name, record.Description)
	} else {
		return nil, fmt.Errorf("La categoría %q no existe.", parentId)
	}

	return parent, nil

}

// indexRecords abre el datasource y mapea cada id de categoría a su posición
// entre los records. Falla si dos records comparten el mismo id.
func (this *YamlCategoryRepository) indexRecords() error {

	if err := this.datasource.Open(); err != nil {
		return err
	}

	records, err := this.datasource.Records()
	if err != nil {
		return err
	}

	for position, record := range records {

		if _, exists := this.indexes[record.Id]; exists {
			return fmt.Errorf("Categoría duplicada: %q", record.Id)
		}

		this.indexes[record.Id] = position
	}

	return nil
}

// FindAll devuelve todas las categorías, de todos los niveles, mapeadas a
// partir de los records del datasource. Retorna un error si el datasource no
// puede leerse o si una categoría referencia un padre inexistente.
func (this *YamlCategoryRepository) FindAll() ([]*entity.Category, error) {

	records, err := this.datasource.Records()
	if err != nil {
		return nil, err
	}

	categories := make([]*entity.Category, len(records))

	for position, record := range records {

		parent, err := this.ParentLookup(record.CategoryId)
		if err != nil {
			return nil, err
		}

		categories[position] = entity.NewCategory(record.Id, parent, record.Name, record.Description)
	}

	return categories, nil
}

// FindById devuelve la categoría con el identificador indicado, mapeada a
// partir de su record en el datasource, o nil si no existe. Resuelve el padre
// buscándolo recursivamente por su id. Retorna un error si el datasource no
// puede leerse o si la categoría referencia un padre inexistente.
func (this *YamlCategoryRepository) FindById(id string) (*entity.Category, error) {

	position, exists := this.indexes[id]
	if !exists {
		return nil, nil
	}

	record, err := this.datasource.RecordAt(position)
	if err != nil {
		return nil, err
	}

	parent, err := this.ParentLookup(record.CategoryId)
	if err != nil {
		return nil, err
	}

	return entity.NewCategory(record.Id, parent, record.Name, record.Description), nil
}

// Count devuelve la cantidad de categorías definidas. Retorna un error si el
// datasource no puede leerse.
func (this *YamlCategoryRepository) Count() (int, error) {

	records, err := this.datasource.Records()
	if err != nil {
		return 0, err
	}

	return len(records), nil
}
