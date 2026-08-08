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
	// Fecha de liquidacion del movimiento. En algunos casos, la fecha de liquidación puede ser diferente a la fecha del movimiento, por ejemplo, en transferencias bancarias que se procesan en días hábiles.
	SettlementDate time.Time
	// installments es la cantidad de cuotas del movimiento. En algunos casos, un movimiento puede estar asociado a un plan de pagos en cuotas, y este campo indica cuántas cuotas se han pagado o se deben pagar.
	// Por ejemplo, si un movimiento corresponde a una compra en 3 cuotas, este campo tendrá el valor 3.
	Installments int
	// installmentNumber es el número de la cuota del movimiento. En algunos casos, un movimiento puede estar asociado a un plan de pagos en cuotas, y este campo indica el número de la cuota que se está pagando.
	InstallmentNumber int
	// comments son las observaciones asociadas al movimiento.
	Comments string
	// category es la categoría del movimiento.
	Category string
}
