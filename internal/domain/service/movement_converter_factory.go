package service

import (
	"fmt"
	"strings"
)

type MovementConverterFactory struct {
}

// NewMovementConverterFactory crea una instancia de la factory.
func NewMovementConverterFactory() *MovementConverterFactory {
	return &MovementConverterFactory{}
}

// GetConverter devuelve el conversor correspondiente al banco y al tipo de
// archivo indicados. El tipo de archivo puede ser "account" (extracto de
// cuenta) o "card" (resumen de tarjeta). Retorna un error si la combinación no
// está soportada.
func (this *MovementConverterFactory) GetConverter(bank string, fileType string) (IBankMovementConverter, error) {

	normalizedBank := strings.ToLower(strings.TrimSpace(bank))
	normalizedFileType := strings.ToLower(strings.TrimSpace(fileType))

	switch normalizedBank {
	case "galicia":
		switch normalizedFileType {
		case "account":
			return &GaliciaAccountMovementConverter{}, nil
		}
	}

	return nil, fmt.Errorf("no existe un conversor para el banco %q y el tipo de archivo %q", bank, fileType)
}
