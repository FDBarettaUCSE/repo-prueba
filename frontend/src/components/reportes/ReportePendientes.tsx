
import { useCallback, useEffect, useState } from "react";
import { useAuth } from "../../auth/AuthContext";
import { obtenerReportePendientes } from "../../api/reportes";
import { listarDepositos } from "../../api/depositos";
import { ErrorAPI } from "../../api/respuestas";
import type {
  DepositoDTO,
  ReportePendientesDTO,
} from "../../api/types";

interface Props {
  activo: boolean;
}

export function ReportePendientes({ activo }: Props) {
  const { token, usuario } = useAuth();

  const [reporte, setReporte] = useState<ReportePendientesDTO | null>(null);
  const [depositos, setDepositos] = useState<DepositoDTO[]>([]);

  const [depositoFiltro, setDepositoFiltro] = useState("");
  const [pagina, setPagina] = useState(1);
  const [vistaActiva, setVistaActiva] = useState<"ordenes" | "transferencias">(
    "ordenes",
  );

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
      const resultado = await obtenerReportePendientes(token, {
        deposito_id: depositoFiltro || undefined,
        pagina,
      });

      setReporte(resultado);
    } catch (e) {
      setError(
        e instanceof ErrorAPI
          ? e.message
          : "No se pudo cargar el reporte de pendientes.",
      );
    } finally {
      setCargando(false);
    }
  }, [token, activo, depositoFiltro, pagina]);

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

  const totalPaginasOrdenes = reporte
    ? Math.max(
        1,
        Math.ceil(
          reporte.paginacion_ordenes.total /
            reporte.paginacion_ordenes.tamanioPagina,
        ),
      )
    : 1;

  const totalPaginasTransferencias = reporte
    ? Math.max(
        1,
        Math.ceil(
          reporte.paginacion_transferencias.total /
            reporte.paginacion_transferencias.tamanioPagina,
        ),
      )
    : 1;

  return (
    <section className="panel">
      <h2>Reporte de pendientes</h2>

      {error && (
        <p className="error" role="alert">
          {error}
        </p>
      )}

      {puedeElegirDeposito && (
        <div className="filtros-barra mb-4">
          <div className="row g-3">
            <div className="col-12 col-md-6">
              <label htmlFor="filtro-deposito-pendientes" className="form-label">
                Depósito
              </label>

              <select
                id="filtro-deposito-pendientes"
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
          </div>
        </div>
      )}

      <div className="d-flex flex-wrap gap-2 mb-4">
        <button
          type="button"
          className={`btn ${
            vistaActiva === "ordenes"
              ? "btn-primary"
              : "btn-outline-primary"
          }`}
          onClick={() => {
            setVistaActiva("ordenes");
            setPagina(1);
          }}
        >
          Órdenes de compra
        </button>

        <button
          type="button"
          className={`btn ${
            vistaActiva === "transferencias"
              ? "btn-primary"
              : "btn-outline-primary"
          }`}
          onClick={() => {
            setVistaActiva("transferencias");
            setPagina(1);
          }}
        >
          Transferencias
        </button>
      </div>

      {cargando ? (
        <div className="d-flex align-items-center gap-2 text-muted my-3">
          <span
            className="spinner-border spinner-border-sm"
            aria-hidden="true"
          />
          <p className="mb-0">Cargando pendientes...</p>
        </div>
      ) : reporte ? (
        <>
          {vistaActiva === "ordenes" ? (
            <section>
              <h2>Estado de órdenes de compra</h2>

              <div className="tabla-wrapper table-responsive">
                <table className="table table-hover align-middle mb-0">
                  <thead>
                    <tr>
                      <th>Depósito</th>
                      <th>Borrador</th>
                      <th>Confirmada</th>
                      <th>Parcial</th>
                      <th>Completada</th>
                    </tr>
                  </thead>

                  <tbody>
                    {reporte.ordenes.length === 0 ? (
                      <tr>
                        <td colSpan={5} className="text-center text-muted py-4">
                          No hay órdenes para mostrar.
                        </td>
                      </tr>
                    ) : (
                      reporte.ordenes.map((item) => (
                        <tr key={item.deposito_id}>
                          <td className="fw-medium">{item.deposito_nombre}</td>
                          <td>{item.borrador}</td>
                          <td>{item.confirmada}</td>
                          <td>{item.parcial}</td>
                          <td>{item.completada}</td>
                        </tr>
                      ))
                    )}
                  </tbody>
                </table>
              </div>

              <div className="paginacion">
                <span>
                  Página {reporte.paginacion_ordenes.pagina} de{" "}
                  {totalPaginasOrdenes}
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
                    disabled={pagina >= totalPaginasOrdenes || cargando}
                    onClick={() => setPagina((actual) => actual + 1)}
                  >
                    Siguiente
                  </button>
                </div>
              </div>
            </section>
          ) : (
            <section>
              <h2>Estado de transferencias</h2>

              <div className="tabla-wrapper table-responsive">
                <table className="table table-hover align-middle mb-0">
                  <thead>
                    <tr>
                      <th>Depósito</th>
                      <th>Solicitadas como origen</th>
                      <th>Pendientes de aprobar</th>
                      <th>Aprobadas por recibir</th>
                    </tr>
                  </thead>

                  <tbody>
                    {reporte.transferencias.length === 0 ? (
                      <tr>
                        <td colSpan={4} className="text-center text-muted py-4">
                          No hay transferencias para mostrar.
                        </td>
                      </tr>
                    ) : (
                      reporte.transferencias.map((item) => (
                        <tr key={item.deposito_id}>
                          <td className="fw-medium">{item.deposito_nombre}</td>
                          <td>{item.solicitadas_como_origen}</td>
                          <td>{item.pendientes_de_aprobar}</td>
                          <td>{item.aprobadas_por_recibir}</td>
                        </tr>
                      ))
                    )}
                  </tbody>
                </table>
              </div>

              <div className="paginacion">
                <span>
                  Página {reporte.paginacion_transferencias.pagina} de{" "}
                  {totalPaginasTransferencias}
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
                    disabled={pagina >= totalPaginasTransferencias || cargando}
                    onClick={() => setPagina((actual) => actual + 1)}
                  >
                    Siguiente
                  </button>
                </div>
              </div>
            </section>
          )}
        </>
      ) : null}
    </section>
  );
}
