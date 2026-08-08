package repository

import "ramiromd/budget/internal/domain/entity"

// ICategoryRepository da acceso a las categorías disponibles para clasificar
// los movimientos.
type ICategoryRepository interface {
	// FindAll devuelve todas las categorías. Retorna un error si el origen de
	// datos no puede leerse.
	FindAll() ([]*entity.Category, error)
	// FindById devuelve la categoría con el identificador indicado, o nil si no
	// existe. Retorna un error si el origen de datos no puede leerse.
	FindById(id string) (*entity.Category, error)
}
