package entity

// Beneficiary es el destinatario de un movimiento, por ejemplo "PedidosYa".
// Permite clasificar automáticamente un movimiento: si su descripción coincide
// con Pattern, se le asigna la subcategoría del beneficiario.
type Beneficiary struct {
	// Id es el identificador único del beneficiario, por ejemplo "pedidosya".
	Id string
	// Name es el nombre del beneficiario tal como se muestra al usuario.
	Name string
	// Description describe la actividad del beneficiario.
	Description string
	// Pattern es la expresión que debe encontrarse en la descripción de un
	// movimiento para asociarlo a este beneficiario.
	Pattern string
	// SubcategoryId es el identificador de la subcategoría que se le asigna a
	// los movimientos de este beneficiario.
	SubcategoryId string
}

// NewBeneficiary crea un beneficiario con los datos indicados.
func NewBeneficiary(id string, name string, description string, pattern string, subcategoryId string) *Beneficiary {
	return &Beneficiary{
		Id:            id,
		Name:          name,
		Description:   description,
		Pattern:       pattern,
		SubcategoryId: subcategoryId,
	}
}