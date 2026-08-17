package persistence

import (
	"fmt"

	sharedpersistence "ramiromd/budget/internal/shared/infrastructure/persistence"

	"gopkg.in/yaml.v3"
)

// categoryFile representa la estructura de un archivo categories.<n>.yml.
type categoryDsRoot struct {
	Categories []categoryDsRecord `yaml:"categories"`
}

// categoryRecord representa una categoría tal como está escrita en el archivo.
type categoryDsRecord struct {
	Id          string `yaml:"id"`
	Name        string `yaml:"name"`
	CategoryId  string `yaml:"category_id"`
	Description string `yaml:"description"`
}

type YamlCategoryDatasource struct {
	reader  *sharedpersistence.YMLReader
	records []categoryDsRecord
	active  bool
}

// NewYamlCategoryDatasource crea el datasource de categorías a partir de un
// YMLReader ya construido. Los records no se interpretan hasta llamar a
// Open().
func NewYamlCategoryDatasource(reader *sharedpersistence.YMLReader) *YamlCategoryDatasource {
	return &YamlCategoryDatasource{
		reader:  reader,
		records: make([]categoryDsRecord, 0),
		active:  false,
	}
}

// Open descubre y lee los archivos categories.*.yml y acumula sus records en
// memoria. reader.Read devuelve un contenido por archivo, así que cada uno se
// interpreta por separado y sus categorías se agregan a una única colección.
func (this *YamlCategoryDatasource) Open() error {

	contents, err := this.reader.Read("categories.*.yml")
	if err != nil {
		return err
	}

	records := make([]categoryDsRecord, 0)

	for _, content := range contents {
		root := categoryDsRoot{}
		if err := yaml.Unmarshal(content, &root); err != nil {
			return err
		}
		records = append(records, root.Categories...)
	}

	this.records = records
	this.active = true
	return nil
}

// RecordAt devuelve el record ubicado en position. Retorna un error si el
// datasource no fue abierto con Open() o si position está fuera de rango.
func (this *YamlCategoryDatasource) RecordAt(position int) (categoryDsRecord, error) {

	if !this.active {
		return categoryDsRecord{}, fmt.Errorf("El datasource no está abierto: llamar a Open() antes de consultarlo")
	}

	if position < 0 || position >= len(this.records) {
		return categoryDsRecord{}, fmt.Errorf("La posición %d está fuera de rango: hay %d records", position, len(this.records))
	}

	return this.records[position], nil
}

// Records devuelve todos los records cargados. Retorna un error si el
// datasource no fue abierto con Open().
func (this *YamlCategoryDatasource) Records() ([]categoryDsRecord, error) {

	if !this.active {
		return nil, fmt.Errorf("El datasource no está abierto: llamar a Open() antes de consultarlo")
	}

	return this.records, nil
}
