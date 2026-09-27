package usuario

import (
	"api/internal/auditoria"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type RegistrarDTO struct {
	Correo           string `json:"correo" binding:"required,email"`
	Contraseña       string `json:"contraseña" binding:"required,min=8,max=64"`
	Rol              Rol    `json:"rol" binding:"required"`
	DepositoAsociado string `json:"deposito_asociado,omitempty" binding:"omitempty,len=24,hexadecimal"`
}
type CambiarContraseñaDTO struct {
	Contraseña      string `json:"contraseña" binding:"required"`
	ContraseñaNueva string `json:"contraseña_nueva" binding:"required,min=8,max=64,nefield=Contraseña"`
}

type CambiarRolDTO struct {
	Rol              Rol    `json:"rol" binding:"required"`
	DepositoAsociado string `json:"deposito_asociado,omitempty" binding:"omitempty,len=24,hexadecimal"`
}

type IngresarDTO struct {
	Correo     string `json:"correo" binding:"required,email"`
	Contraseña string `json:"contraseña" binding:"required"`
}

type UsuarioDTO struct {
	ID     string `json:"id"`
	Correo string `json:"correo"`
}

type CambiarEstadoUsuarioDTO struct {
	estado bool
}

// Que rol le damos a los usuarios nuevos? Preguntar
func (*RegistrarDTO) NuevoUsuarioRegistrar(dto RegistrarDTO, hash string, adminId, depositoId bson.ObjectID) *Usuario {
	return &Usuario{
		Correo:             dto.Correo,
		ContraseñaHasheada: hash,
		Rol:                dto.Rol,
		Activado:           false,
		DepositoAsociado:   depositoId,
		Auditoria: auditoria.Auditoria{
			CreadoPor: adminId,
			CreadoEn:  time.Now(),
		},
	}
}

func (*Usuario) ConvertirAUsuarioDTO(usuario Usuario) UsuarioDTO {
	return UsuarioDTO{
		ID:     usuario.ID.Hex(),
		Correo: usuario.Correo,
	}
}
