package evaluacionstock

import (
	"context"
	"errors"
	"time"

	alertastock "api/internal/alertaStock"
	"api/internal/auditoria"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type EvaluadorStock struct {
	alertaRepositorio alertastock.AlertaStockRepositorio
}

func NuevoEvaluadorStock(repo alertastock.AlertaStockRepositorio) *EvaluadorStock {
	return &EvaluadorStock{
		alertaRepositorio: repo,
	}
}

func (e *EvaluadorStock) EvaluarStock(
	ctx context.Context,
	depositoID, productoID bson.ObjectID,
	stockActual, stockMinimo int32,
	usuarioID bson.ObjectID,
) error {

	alerta, err := e.alertaRepositorio.BuscarActiva(
		ctx,
		depositoID,
		productoID,
		alertastock.AlertaStockBajo,
	)

	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return err
	}

	if stockActual < stockMinimo {
		if errors.Is(err, mongo.ErrNoDocuments) {
			_, err = e.alertaRepositorio.Crear(ctx, alertastock.AlertaStock{
				Tipo:        alertastock.AlertaStockBajo,
				ProductoID:  productoID,
				DepositoID:  depositoID,
				StockMinimo: stockMinimo,
				StockActual: stockActual,
				Resuelta:    false,
				Auditoria: auditoria.Auditoria{
					CreadoPor: usuarioID,
					CreadoEn:  time.Now(),
				},
			})

			return err
		}

		return e.alertaRepositorio.ActualizarStock(
			ctx,
			alerta.ID.Hex(),
			stockActual,
			stockMinimo,
			usuarioID,
		)
	}

	if err == nil {
		_, err = e.alertaRepositorio.MarcarResuelta(
			ctx,
			alerta.ID.Hex(),
			usuarioID,
		)
		return err
	}

	return nil
}
