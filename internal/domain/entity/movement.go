package entity

import (
	"time"

	"github.com/google/uuid"
)

type Movement struct {
	// Id es el identificador único del movimiento.
	Id uuid.UUID
	// date es la fecha del movimiento.
	Date time.Time
	// amount es el importe del movimiento.
	Amount float64
	// description es la descripción del movimiento.
	Description string
	// comments son las observaciones asociadas al movimiento.
	Comments string
	// category es la categoría del movimiento.
	Category string
}

// NewMovement crea un movimiento con los datos indicados y le asigna un
// identificador único.
func NewMovement(date time.Time, amount float64, description string, comments string, category string) *Movement {
	return &Movement{
		Id:          uuid.New(),
		Date:        date,
		Amount:      amount,
		Description: description,
		Comments:    comments,
		Category:    category,
	}
}
