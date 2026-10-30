package usuario

import (
	"api/internal/auditoria"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

const MaxTamanioPagina = 100

type RegistrarDTO struct {
	Correo           string `json:"correo" binding:"required,email"`
	Contrasenia      string `json:"contrasenia" binding:"required,min=8,max=64"`
	Rol              Rol    `json:"rol" binding:"required"`
	DepositoAsociado string `json:"deposito_asociado,omitempty" binding:"omitempty,len=24,hexadecimal"`
}

type CambiarContraseniaDTO struct {
	Contrasenia      string `json:"contrasenia" binding:"required"`
	ContraseniaNueva string `json:"contrasenia_nueva" binding:"required,min=8,max=64,nefield=Contrasenia"`
}

type CambiarRolDTO struct {
	Rol              Rol    `json:"rol" binding:"required"`
	DepositoAsociado string `json:"deposito_asociado,omitempty" binding:"omitempty,len=24,hexadecimal"`
}

type IngresarDTO struct {
	Correo      string `json:"correo" binding:"required,email"`
	Contrasenia string `json:"contrasenia" binding:"required"`
}

type EventoAuditoriaUsuarioDTO struct {
	Accion       string `json:"accion"`
	Detalle      string `json:"detalle,omitempty"`
	RealizadoPor string `json:"realizado_por"`
	Fecha        string `json:"fecha"`
}

type UsuarioDTO struct {
	ID                 string                      `json:"id"`
	Correo             string                      `json:"correo"`
	Rol                Rol                         `json:"rol"`
	Activado           bool                        `json:"activado"`
	DepositoAsociado   string                      `json:"deposito_asociado,omitempty"`
	HistorialAuditoria []EventoAuditoriaUsuarioDTO `json:"historial_auditoria,omitempty"`
}

type UsuariosPaginadosDTO struct {
	Items         []UsuarioDTO `json:"items"`
	Total         int64        `json:"total"`
	Pagina        int          `json:"pagina"`
	TamanioPagina int          `json:"tamanioPagina"`
}

type CambiarEstadoUsuarioDTO struct {
	Estado bool `json:"estado"`
}

func (*RegistrarDTO) NuevoUsuarioRegistrar(dto RegistrarDTO, hash string, adminId, depositoId bson.ObjectID) *Usuario {
	ahora := time.Now()
	detalle := "rol=" + string(dto.Rol)
	if !depositoId.IsZero() {
		detalle += ", deposito=" + depositoId.Hex()
	}
	return &Usuario{
		Correo:              dto.Correo,
		ContraseniaHasheada: hash,
		Rol:                 dto.Rol,
		Activado:            true,
		DepositoAsociado:    depositoId,
		Auditoria: auditoria.Auditoria{
			CreadoPor: adminId,
			CreadoEn:  ahora,
		},
		HistorialAuditoria: []EventoAuditoriaUsuario{
			{
				Accion:       "alta_usuario",
				Detalle:      detalle,
				RealizadoPor: adminId,
				Fecha:        ahora,
			},
		},
	}
}

func (*Usuario) ConvertirAUsuarioDTO(usuario Usuario) UsuarioDTO {
	depo := ""
	if !usuario.DepositoAsociado.IsZero() {
		depo = usuario.DepositoAsociado.Hex()
	}
	var historial []EventoAuditoriaUsuarioDTO
	if len(usuario.HistorialAuditoria) > 0 {
		historial = make([]EventoAuditoriaUsuarioDTO, 0, len(usuario.HistorialAuditoria))
		for _, ev := range usuario.HistorialAuditoria {
			historial = append(historial, EventoAuditoriaUsuarioDTO{
				Accion:       ev.Accion,
				Detalle:      ev.Detalle,
				RealizadoPor: ev.RealizadoPor.Hex(),
				Fecha:        ev.Fecha.Format(time.RFC3339),
			})
		}
	}
	return UsuarioDTO{
		ID:                 usuario.ID.Hex(),
		Correo:             usuario.Correo,
		Rol:                usuario.Rol,
		Activado:           usuario.Activado,
		DepositoAsociado:   depo,
		HistorialAuditoria: historial,
	}
}
