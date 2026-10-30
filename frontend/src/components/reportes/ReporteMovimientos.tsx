import { useCallback, useEffect, useState } from "react";
import { useAuth } from "../../auth/AuthContext";
import { obtenerReporteMovimientos } from "../../api/reportes";
import { listarDepositos } from "../../api/depositos";
import { ErrorAPI } from "../../api/respuestas";
import type {
  DepositoDTO,
  ReporteMovimientosDTO,
  TipoMovimientoReporte,
} from "../../api/types";

interface Props {
  activo: boolean;
}

export function ReporteMovimientos({ activo }: Props) {
  const { token, usuario } = useAuth();

  const [reporte, setReporte] = useState<ReporteMovimientosDTO | null>(null);
  const [depositos, setDepositos] = useState<DepositoDTO[]>([]);

  const [depositoFiltro, setDepositoFiltro] = useState("");
  const [tipoFiltro, setTipoFiltro] = useState<TipoMovimientoReporte | "">("");
  const [fechaDesde, setFechaDesde] = useState("");
  const [fechaHasta, setFechaHasta] = useState("");
  const [pagina, setPagina] = useState(1);

  const [cargando, setCargando] = useState(false);
  const [error, setError] = useState("");

  const puedeElegirDeposito =
    usuario?.rol === "administrador" || usuario?.rol === "auditor";

  // Cargar los datos del reporte.
  const cargarReporte = useCallback(async () => {
    if (!token || !activo) return;

    setCargando(true);
    setError("");

    try {
      const resultado = await obtenerReporteMovimientos(token, {
        deposito_id: depositoFiltro || undefined,
        tipo: tipoFiltro || undefined,
        fecha_desde: fechaDesde || undefined,
        fecha_hasta: fechaHasta || undefined,
        pagina,
      });

      setReporte(resultado);
    } catch (e) {
      setError(
        e instanceof ErrorAPI
          ? e.message
          : "No se pudo cargar el reporte de movimientos.",
      );
    } finally {
      setCargando(false);
    }
  }, [
    token,
    activo,
    depositoFiltro,
    tipoFiltro,
    fechaDesde,
    fechaHasta,
    pagina,
  ]);

  // Cargar los depósitos disponibles para el filtro.
  useEffect(() => {
    if (!activo || !token || !puedeElegirDeposito) return;

    async function cargarCatalogos() {
      try {
        const resultado = await listarDepositos(token!);
        setDepositos(resultado);
      } catch (e) {
        setError(
          e instanceof ErrorAPI
            ? e.message
            : "No se pudieron cargar los depósitos.",
        );
      }
    }

    void cargarCatalogos();
  }, [activo, token, puedeElegirDeposito]);

  useEffect(() => {
    void cargarReporte();
  }, [cargarReporte]);

  const totalPaginas = reporte
    ? Math.max(
        1,
        Math.ceil(
          reporte.paginacion.total / reporte.paginacion.tamanioPagina,
        ),
      )
    : 1;

  return (
    <section className="panel">
      <h2>Reporte de movimientos</h2>

      {error && (
        <p className="error" role="alert">
          {error}
        </p>
      )}

      <div className="filtros-barra mb-4">
        <div className="row g-3">
          {puedeElegirDeposito && (
            <div className="col-12 col-md-6">
              <label htmlFor="filtro-deposito-movimientos" className="form-label">
                Depósito
              </label>
              <select
                id="filtro-deposito-movimientos"
                className="form-select"
                value={depositoFiltro}
                onChange={(e) => {
                  setDepositoFiltro(e.target.value);
                  setPagina(1);
                }}
              >
                <option value="">Todos los depósitos</option>
                {depositos.map((deposito) => (
                  <option key={deposito.id} value={deposito.id}>
                    {deposito.nombre}
                  </option>
                ))}
              </select>
            </div>
          )}

          <div className="col-12 col-md-6">
            <label htmlFor="filtro-tipo-movimiento" className="form-label">
              Tipo de movimiento
            </label>
            <select
              id="filtro-tipo-movimiento"
              className="form-select"
              value={tipoFiltro}
              onChange={(e) => {
                setTipoFiltro(e.target.value as TipoMovimientoReporte | "");
                setPagina(1);
              }}
            >
              <option value="">Todos los tipos</option>
              <option value="ingreso">Ingreso</option>
              <option value="egreso">Egreso</option>
              <option value="ajuste">Ajuste</option>
              <option value="transferencia">Transferencia</option>
            </select>
          </div>

          <div className="col-12 col-md-6">
            <label htmlFor="fecha-desde-movimientos" className="form-label">
              Fecha desde
            </label>
            <input
              id="fecha-desde-movimientos"
              type="date"
              className="form-control"
              value={fechaDesde}
              max={fechaHasta || undefined}
              onChange={(e) => {
                setFechaDesde(e.target.value);
                setPagina(1);
              }}
            />
          </div>

          <div className="col-12 col-md-6">
            <label htmlFor="fecha-hasta-movimientos" className="form-label">
              Fecha hasta
            </label>
            <input
              id="fecha-hasta-movimientos"
              type="date"
              className="form-control"
              value={fechaHasta}
              min={fechaDesde || undefined}
              onChange={(e) => {
                setFechaHasta(e.target.value);
                setPagina(1);
              }}
            />
          </div>
        </div>
      </div>

      {cargando ? (
        <div className="d-flex align-items-center gap-2 text-muted my-3">
          <span
            className="spinner-border spinner-border-sm"
            aria-hidden="true"
          />
          <p className="mb-0">Cargando movimientos...</p>
        </div>
      ) : reporte ? (
        <>
          <div className="reporte-resumen">
            <div className="reporte-indicador">
              <span>Total de movimientos</span>
              <strong>
                {reporte.total_movimientos.toLocaleString("es-AR")}
              </strong>
            </div>
          </div>

          <div className="tabla-wrapper table-responsive">
            <table className="table table-hover align-middle mb-0">
              <thead>
                <tr>
                  <th>Fecha</th>
                  <th>Depósito</th>
                  <th>Tipo</th>
                  <th>Cantidad de eventos</th>
                  <th>Unidades netas</th>
                </tr>
              </thead>

              <tbody>
                {reporte.resumen.length === 0 ? (
                  <tr>
                    <td colSpan={5} className="text-center text-muted py-4">
                      No hay movimientos para mostrar.
                    </td>
                  </tr>
                ) : (
                  reporte.resumen.map((item, indice) => (
                    <tr
                      key={`${item.deposito_id}-${item.fecha}-${item.tipo}-${indice}`}
                    >
                      <td>
                        {new Date(`${item.fecha}T00:00:00`).toLocaleDateString(
                          "es-AR",
                        )}
                      </td>
                      <td className="fw-medium">{item.deposito_nombre}</td>
                      <td className="text-capitalize">{item.tipo}</td>
                      <td>{item.cantidad_eventos}</td>
                      <td>{item.unidades_netas.toLocaleString("es-AR")}</td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>

          <div className="paginacion">
            <span>
              Página {reporte.paginacion.pagina} de {totalPaginas}
            </span>

            <div className="acciones">
              <button
                type="button"
                className="btn btn-outline-primary btn-sm"
                disabled={pagina <= 1 || cargando}
                onClick={() => setPagina((actual) => actual - 1)}
              >
                Anterior
              </button>

              <button
                type="button"
                className="btn btn-outline-primary btn-sm"
                disabled={pagina >= totalPaginas || cargando}
                onClick={() => setPagina((actual) => actual + 1)}
              >
                Siguiente
              </button>
            </div>
          </div>
        </>
      ) : null}
    </section>
  );
}
