package repository

import "ramiromd/budget/internal/domain/entity"

// ISubcategoryRepository da acceso a las subcategorías disponibles para
// clasificar los movimientos.
type ISubcategoryRepository interface {
	// FindAll devuelve todas las subcategorías. Retorna un error si el origen
	// de datos no puede leerse.
	FindAll() ([]*entity.Subcategory, error)
	// FindById devuelve la subcategoría con el identificador indicado, o nil si
	// no existe. Retorna un error si el origen de datos no puede leerse.
	FindById(id string) (*entity.Subcategory, error)
	// FindByCategoryId devuelve las subcategorías que pertenecen a la categoría
	// indicada. Devuelve una lista vacía si la categoría no tiene
	// subcategorías. Retorna un error si el origen de datos no puede leerse.
	FindByCategoryId(categoryId string) ([]*entity.Subcategory, error)
}
