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
var _ domain.IBeneficiaryRepository = (*YamlBeneficiaryRepository)(nil)

// beneficiariesYml es el contenido de beneficiaries.yml. Se incrusta en el
// binario durante la compilación, de modo que el repositorio no depende del
// directorio de trabajo ni del archivo en disco.
//
//go:embed beneficiaries.yml
var beneficiariesYml []byte

// beneficiaryFile representa la estructura del archivo beneficiaries.yml.
type beneficiaryFile struct {
	Beneficiaries []beneficiaryRecord `yaml:"beneficiaries"`
}

// beneficiaryRecord representa un beneficiario tal como está escrito en el
// archivo.
type beneficiaryRecord struct {
	Id            string `yaml:"id"`
	Name          string `yaml:"name"`
	Description   string `yaml:"description"`
	Pattern       string `yaml:"pattern"`
	SubcategoryId string `yaml:"subcategory_id"`
}

// YamlBeneficiaryRepository expone los beneficiarios definidos en
// beneficiaries.yml. El archivo se interpreta una única vez y el resultado
// queda en memoria.
//
// No es seguro para uso concurrente.
type YamlBeneficiaryRepository struct {
	beneficiaries []*entity.Beneficiary
}

// NewYamlBeneficiaryRepository crea el repositorio de beneficiarios. El archivo
// no se interpreta hasta la primera consulta.
func NewYamlBeneficiaryRepository() *YamlBeneficiaryRepository {
	return &YamlBeneficiaryRepository{}
}

// FindAll devuelve todos los beneficiarios. Retorna un error si el origen de
// datos no puede leerse.
func (this *YamlBeneficiaryRepository) FindAll() ([]*entity.Beneficiary, error) {

	if err := this.load(); err != nil {
		return nil, err
	}

	return this.beneficiaries, nil
}

// FindById devuelve el beneficiario con el identificador indicado, o nil si no
// existe. Retorna un error si el origen de datos no puede leerse.
func (this *YamlBeneficiaryRepository) FindById(id string) (*entity.Beneficiary, error) {

	if err := this.load(); err != nil {
		return nil, err
	}

	for _, beneficiary := range this.beneficiaries {
		if beneficiary.Id == id {
			return beneficiary, nil
		}
	}

	return nil, nil
}

// FindBySubcategoryId devuelve los beneficiarios asociados a la subcategoría
// indicada. Devuelve una lista vacía si no hay ninguno. Retorna un error si el
// origen de datos no puede leerse.
func (this *YamlBeneficiaryRepository) FindBySubcategoryId(subcategoryId string) ([]*entity.Beneficiary, error) {

	if err := this.load(); err != nil {
		return nil, err
	}

	result := make([]*entity.Beneficiary, 0)
	for _, beneficiary := range this.beneficiaries {
		if beneficiary.SubcategoryId == subcategoryId {
			result = append(result, beneficiary)
		}
	}

	return result, nil
}

// load interpreta el archivo incrustado y arma las entidades. Las llamadas
// posteriores a la primera reutilizan el resultado.
func (this *YamlBeneficiaryRepository) load() error {

	if this.beneficiaries != nil {
		return nil
	}

	file := beneficiaryFile{}
	if err := yaml.Unmarshal(beneficiariesYml, &file); err != nil {
		return fmt.Errorf("el archivo beneficiaries.yml no tiene un formato válido: %w", err)
	}

	beneficiaries := make([]*entity.Beneficiary, 0, len(file.Beneficiaries))
	for _, record := range file.Beneficiaries {
		beneficiaries = append(beneficiaries, entity.NewBeneficiary(record.Id, record.Name, record.Description, record.Pattern, record.SubcategoryId))
	}

	this.beneficiaries = beneficiaries

	return nil
}
