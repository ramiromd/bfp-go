package repository

import "ramiromd/budget/internal/domain/entity"

// IBeneficiaryRepository da acceso a los beneficiarios conocidos, que son los
// que permiten asignarle automáticamente una subcategoría a un movimiento.
type IBeneficiaryRepository interface {
	// FindAll devuelve todos los beneficiarios. Retorna un error si el origen
	// de datos no puede leerse.
	FindAll() ([]*entity.Beneficiary, error)
	// FindById devuelve el beneficiario con el identificador indicado, o nil si
	// no existe. Retorna un error si el origen de datos no puede leerse.
	FindById(id string) (*entity.Beneficiary, error)
	// FindBySubcategoryId devuelve los beneficiarios asociados a la subcategoría
	// indicada. Devuelve una lista vacía si no hay ninguno. Retorna un error si
	// el origen de datos no puede leerse.
	FindBySubcategoryId(subcategoryId string) ([]*entity.Beneficiary, error)
}
