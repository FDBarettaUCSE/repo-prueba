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
	SoloNoLeidas string `form:"solo_no_leidas"` // "true" para filtrar solo las no leídas
	Pagina       int    `form:"pagina"`
	TamañoPagina int    `form:"tamaño_pagina"`
}

// RespuestaNovedadDTO es la vista de salida de una respuesta.
type RespuestaNovedadDTO struct {
	Texto          string `json:"texto"`
	UsuarioID      string `json:"usuario_id"`
	FechaRespuesta string `json:"fecha_respuesta"`
}

// NovedadDTO es la vista de salida de una novedad.
type NovedadDTO struct {
	ID         string               `json:"id"`
	DepositoID string               `json:"deposito_id"`
	Texto      string               `json:"texto"`
	FechaEnvio string               `json:"fecha_envio"`
	Leido      bool                 `json:"leido"`
	Respuesta  *RespuestaNovedadDTO `json:"respuesta,omitempty"`
}

// NovedadesPaginadoDTO envuelve el listado con metadata de paginación y el
// contador de no leídas del depósito. NoLeidas es independiente del filtro y
// de la página pedidos, para que el frontend pueda mostrar el badge "3 no
// leídas" sin tener que traer todas las novedades para contarlas.
type NovedadesPaginadoDTO struct {
	Items        []NovedadDTO `json:"items"`
	Total        int64        `json:"total"`
	NoLeidas     int64        `json:"no_leidas"`
	Pagina       int          `json:"pagina"`
	TamañoPagina int          `json:"tamaño_pagina"`
}

// ConvertirADTO transforma la entidad a su DTO de salida.
func (n NovedadDeposito) ConvertirADTO() NovedadDTO {
	dto := NovedadDTO{
		ID:         n.ID.Hex(),
		DepositoID: n.DepositoID.Hex(),
		Texto:      n.Texto,
		FechaEnvio: n.FechaEnvio.Format(time.RFC3339),
		Leido:      n.Leido,
	}

	if n.Respuesta != nil {
		dto.Respuesta = &RespuestaNovedadDTO{
			Texto:          n.Respuesta.Texto,
			UsuarioID:      n.Respuesta.UsuarioID.Hex(),
			FechaRespuesta: n.Respuesta.FechaRespuesta.Format(time.RFC3339),
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
		Leido:      false,
	}, nil
}
