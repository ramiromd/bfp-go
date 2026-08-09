package repository

import "ramiromd/budget/internal/category/domain/entity"

type CategoryRepository interface {
	FindAll() ([]*entity.Category, error)
	FindById(id string) (*entity.Category, error)
	Count() (int, error)
}
