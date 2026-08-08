package service

import (
	"fmt"
	"ramiromd/budget/internal/domain/entity"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// galiciaAccountFieldCount es la cantidad de columnas que tiene un extracto de
// cuenta del Banco Galicia: Fecha, Movimiento, Débito, Crédito, Saldo Parcial y
// Comentarios.
const galiciaAccountFieldCount = 6

type GaliciaAccountMovementConverter struct{}

func (this *GaliciaAccountMovementConverter) Convert(record []string) *entity.Movement {

	if len(record) < galiciaAccountFieldCount {
		fmt.Println("Wrong field count")
		return nil
	}

	date, err := time.Parse("02/01/2006", strings.TrimSpace(record[0]))
	if err != nil {
		fmt.Println(err)
		return nil
	}

	// 1. Reemplazamos los saltos de línea estilo Windows (\r\n)
	description := strings.ReplaceAll(record[1], "\r\n", " ")

	// 2. Reemplazamos los saltos de línea estilo Unix (\n)
	description = strings.ReplaceAll(description, "\n", " ")

	description = strings.TrimSpace(description)

	// 1. Reemplazamos los saltos de línea estilo Windows (\r\n)
	comments := strings.ReplaceAll(record[5], "\r\n", " ")

	// 2. Reemplazamos los saltos de línea estilo Unix (\n)
	comments = strings.ReplaceAll(comments, "\n", " ")

	comments = strings.TrimSpace(comments)

	// 1. Quitar los puntos de miles -> queda "-6600,00"
	debitText := strings.ReplaceAll(record[2], ".", "")

	// 2. Cambiar la coma decimal por un punto -> queda "-6600.00"
	debitText = strings.ReplaceAll(debitText, ",", ".")

	debit, err := strconv.ParseFloat(debitText, 64)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	creditText := strings.ReplaceAll(record[3], ".", "")

	// 2. Cambiar la coma decimal por un punto -> queda "-6600.00"
	creditText = strings.ReplaceAll(creditText, ",", ".")
	credit, err := strconv.ParseFloat(creditText, 64)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	amount := credit - debit
	reference := fmt.Sprintf("%s|%.2f|%s", date.Format("2006-01-02"), amount, description)
	movementId := uuid.NewMD5(uuid.Nil, []byte(reference))

	movement := &entity.Movement{}
	movement.Id = movementId
	movement.Date = date
	movement.Description = description
	movement.Amount = credit - debit
	movement.Comments = comments
	movement.SettlementDate = date
	movement.Installments = 1
	movement.InstallmentNumber = 1
	movement.Category = "Sin categoría"

	return movement
}

func (this *GaliciaAccountMovementConverter) GetDescription() string {
	return "Movimiento de cuenta del Banco Galicia."
}
