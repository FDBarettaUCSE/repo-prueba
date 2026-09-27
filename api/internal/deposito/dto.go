package deposito

import "go.mongodb.org/mongo-driver/v2/bson"

type DepositoDTO struct {
	ID        string `json:"id"`
	Nombre    string `json:"nombre"`
	Localidad string `json:"localidad"`
	Provincia string `json:"provincia"`
}

type CrearDepositoDTO struct {
	Nombre    string `json:"nombre" binding:"required"`
	Localidad string `json:"localidad" binding:"required"`
	Provincia string `json:"provincia" binding:"required"`
}

func MapeoADeposito(d DepositoDTO) (Deposito, error) {
	oid, err := bson.ObjectIDFromHex(d.ID)
	if err != nil {
		return Deposito{}, err
	}

	return Deposito{
		ID:        oid,
		Nombre:    d.Nombre,
		Localidad: d.Localidad,
		Provincia: d.Provincia,
	}, nil
}

func MapeoCrearDTOADeposito(d CrearDepositoDTO) (Deposito, error) {
	return Deposito{
		Nombre:    d.Nombre,
		Localidad: d.Localidad,
		Provincia: d.Provincia,
	}, nil
}
