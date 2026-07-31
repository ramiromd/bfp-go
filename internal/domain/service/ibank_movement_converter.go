package service

import "ramiromd/budget/internal/domain/entity"

type IBankMovementConverter interface {
	GetDescription() string
	// Convert transforma un registro ya delimitado del archivo de entrada en un
	// movimiento. Devuelve nil si el registro no puede interpretarse.
	Convert(record []string) *entity.Movement
}
