package entity

// Category es la entidad que representa una categoría de clasificación

type Category struct {
	// Id es el identificador único de la categoría, por ejemplo "light".
	Id string
	// Parent es la categoría de nivel superior a la que pertenece la categoría actual.
	Parent *Category
	// Name es el nombre de la categoría tal como se muestra al usuario.
	Name string
	// Description explica qué tipo de movimientos agrupa la categoría.
	Description string
}

func NewCategory(id string, parent *Category, name string, description string) *Category {
	return &Category{
		Id:          id,
		Parent:      parent,
		Name:        name,
		Description: description,
	}
}
