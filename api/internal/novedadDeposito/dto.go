package novedaddeposito

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// EnviarNovedadDTO es el body para enviar una novedad a UN depósito puntual
// (POST /depositos/:depositoId/novedades).
type EnviarNovedadDTO struct {
	Texto string `json:"texto" binding:"required"`
}

// EnviarNovedadDifusionDTO es el body para que el Auditor (o Administrador)
// envíe la misma novedad a varios depósitos, o a todos (POST
// /novedades/difusion). Si Todos es true, DepositosIDs se ignora.
type EnviarNovedadDifusionDTO struct {
	Texto        string   `json:"texto" binding:"required"`
	DepositosIDs []string `json:"depositos_ids"`
	Todos        bool     `json:"todos"`
}

// ResponderNovedadDTO es el body con el que el operario responde una
// novedad. Responder implica haberla leído: no hace falta un PUT .../leida
// separado antes (ver Service.Responder).
type ResponderNovedadDTO struct {
	Texto string `json:"texto" binding:"required"`
}

// FiltroNovedadesDTO agrupa filtro y paginación del listado de novedades de
// un depósito.
type FiltroNovedadesDTO struct {
	SoloNoLeidas  string `form:"solo_no_leidas"` // "true" para filtrar solo las no leídas
	Pagina        int    `form:"pagina"`
	TamanioPagina int    `form:"tamanioPagina"`
}

// RespuestaNovedadDTO es la vista de salida de una respuesta.
type RespuestaNovedadDTO struct {
	Texto          string `json:"texto"`
	UsuarioID      string `json:"usuario_id"`
	FechaRespuesta string `json:"fecha_respuesta"`
}

// NovedadDTO es la vista de salida de una novedad.
//
//   - Leido: para un operario, si ÉL la leyó; para el resto de los roles
//     (que no marcan lectura), si la leyó al menos un operario.
//   - Respuesta: para un operario, su última respuesta; para el resto, la
//     última respuesta recibida. Se mantiene por compatibilidad; el detalle
//     completo está en Respuestas.
type NovedadDTO struct {
	ID         string                `json:"id"`
	DepositoID string                `json:"deposito_id"`
	Texto      string                `json:"texto"`
	FechaEnvio string                `json:"fecha_envio"`
	Leido      bool                  `json:"leido"`
	LeidoPor   []string              `json:"leido_por"`
	Respuesta  *RespuestaNovedadDTO  `json:"respuesta,omitempty"`
	Respuestas []RespuestaNovedadDTO `json:"respuestas"`
}

// NovedadesPaginadoDTO envuelve el listado con metadata de paginación y el
// contador de no leídas. NoLeidas es independiente del filtro y de la
// página pedidos, para que el frontend pueda mostrar el badge "3 no leídas"
// sin tener que traer todas las novedades para contarlas. Para un operario
// cuenta las que ÉL no leyó; para el resto de los roles, las que todavía no
// leyó ningún operario.
type NovedadesPaginadoDTO struct {
	Items         []NovedadDTO `json:"items"`
	Total         int64        `json:"total"`
	NoLeidas      int64        `json:"no_leidas"`
	Pagina        int          `json:"pagina"`
	TamanioPagina int          `json:"tamanioPagina"`
}

func mapearRespuestaDTO(r RespuestaNovedad) RespuestaNovedadDTO {
	return RespuestaNovedadDTO{
		Texto:          r.Texto,
		UsuarioID:      r.UsuarioID.Hex(),
		FechaRespuesta: r.FechaRespuesta.Format(time.RFC3339),
	}
}

// ConvertirADTO transforma la entidad a su DTO de salida, calculando el
// estado de lectura desde el punto de vista del lector. lector es el
// operario que consulta, o nil si consulta otro rol (gerente, auditor,
// administrador).
func (n NovedadDeposito) ConvertirADTO(lector *bson.ObjectID) NovedadDTO {
	dto := NovedadDTO{
		ID:         n.ID.Hex(),
		DepositoID: n.DepositoID.Hex(),
		Texto:      n.Texto,
		FechaEnvio: n.FechaEnvio.Format(time.RFC3339),
		LeidoPor:   make([]string, 0, len(n.LeidoPor)),
		Respuestas: make([]RespuestaNovedadDTO, 0, len(n.Respuestas)),
	}

	for _, id := range n.LeidoPor {
		dto.LeidoPor = append(dto.LeidoPor, id.Hex())
	}

	if lector != nil {
		dto.Leido = n.LeidaPor(*lector)
	} else {
		dto.Leido = len(n.LeidoPor) > 0
	}

	for _, r := range n.Respuestas {
		rDTO := mapearRespuestaDTO(r)
		dto.Respuestas = append(dto.Respuestas, rDTO)

		if lector == nil || r.UsuarioID == *lector {
			ultima := rDTO
			dto.Respuesta = &ultima
		}
	}

	return dto
}

// ConvertirAModelo arma una NovedadDeposito nueva dirigida a depositoID.
func (dto EnviarNovedadDTO) ConvertirAModelo(depositoID string) (NovedadDeposito, error) {
	depOID, err := bson.ObjectIDFromHex(depositoID)
	if err != nil {
		return NovedadDeposito{}, err
	}

	return NovedadDeposito{
		DepositoID: depOID,
		Texto:      dto.Texto,
		FechaEnvio: time.Now(),
		LeidoPor:   []bson.ObjectID{},
		Respuestas: []RespuestaNovedad{},
	}, nil
}
