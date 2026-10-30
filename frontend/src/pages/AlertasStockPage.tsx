
import { useCallback, useEffect, useState } from "react";
import { useAuth } from "../auth/AuthContext";
import { listarDepositos } from "../api/depositos";
import { listarAlertasStock } from "../api/alertasStock";
import { ErrorAPI } from "../api/respuestas";
import type {
  AlertaStockDTO,
  AlertasPaginadasDTO,
  DepositoDTO,
} from "../api/types";
import { formatearFecha } from "../api/fechas";

export function AlertasStockPage() {
  const { token, usuario } = useAuth();

  const [alertas, setAlertas] = useState<AlertaStockDTO[]>([]);
  const [depositos, setDepositos] = useState<DepositoDTO[]>([]);
  const [depositoId, setDepositoId] = useState("");
  const [soloPendientes, setSoloPendientes] = useState(true);
  const [pagina, setPagina] = useState(1);
  const [total, setTotal] = useState(0);
  const [pendientes, setPendientes] = useState(0);
  const [tamanioPagina, setTamanioPagina] = useState(10);

  const [cargando, setCargando] = useState(false);
  const [error, setError] = useState("");

  const rol = usuario?.rol;
  const esAdministrador = rol === "administrador";
  const esAuditor = rol === "auditor";
  const puedeFiltrarDeposito = esAdministrador || esAuditor;

  const cargarDepositos = useCallback(async () => {
    if (!token || !puedeFiltrarDeposito) return;

    try {
      const resultado = await listarDepositos(token);
      setDepositos(resultado);
    } catch (e) {
      setError(
        e instanceof ErrorAPI
          ? e.message
          : "No se pudieron cargar los depósitos."
      );
    }
  }, [token, puedeFiltrarDeposito]);

  const cargarAlertas = useCallback(async () => {
    if (!token) return;

    setCargando(true);
    setError("");

    try {
      const filtros = {
        deposito_id: depositoId || undefined,
        solo_pendientes: soloPendientes,
        pagina,
        tamanioPagina,
      };

      const resultado: AlertasPaginadasDTO =
        await listarAlertasStock(token, filtros);

      setAlertas(resultado.items);
      setTotal(resultado.total);
      setPendientes(resultado.pendientes);
      setTamanioPagina(resultado.tamanioPagina);
    } catch (e) {
      setError(
        e instanceof ErrorAPI
          ? e.message
          : "No se pudieron cargar las alertas."
      );
    } finally {
      setCargando(false);
    }
  }, [token, depositoId, soloPendientes, pagina, tamanioPagina]);

  useEffect(() => {
    cargarDepositos();
  }, [cargarDepositos]);

  useEffect(() => {
    cargarAlertas();
  }, [cargarAlertas]);

  function cambiarDeposito(valor: string) {
    setDepositoId(valor);
    setPagina(1);
  }

  function cambiarSoloPendientes(valor: boolean) {
    setSoloPendientes(valor);
    setPagina(1);
  }

  const totalPaginas = Math.max(1, Math.ceil(total / tamanioPagina));

  
  return (
    <div className="panel">
      <h1 className="mb-3">Alertas de stock</h1>

      <div className="row g-3 mb-4 resumen-alertas">
        <div className="col-12 col-sm-6 col-md-4">
          <div className="card border-warning bg-warning bg-opacity-10 h-100">
            <div className="card-body py-3">
              <div className="text-muted small fw-semibold text-uppercase">
                Alertas pendientes
              </div>
              <div className="fs-3 fw-bold text-dark mt-1">{pendientes}</div>
            </div>
          </div>
        </div>

        <div className="col-12 col-sm-6 col-md-4">
          <div className="card border-secondary bg-light h-100">
            <div className="card-body py-3">
              <div className="text-muted small fw-semibold text-uppercase">
                Alertas encontradas
              </div>
              <div className="fs-3 fw-bold text-dark mt-1">{total}</div>
            </div>
          </div>
        </div>
      </div>

      <section className="mt-3 pt-0 border-0 seccion-alertas">
        <h2>Listado de alertas</h2>

        <div className="filtros-barra filtros-alertas">
          <div className="row g-3 align-items-end">
            {puedeFiltrarDeposito && (
              <div className="col-12 col-md-5">
                <label className="form-label mb-0">
                  Filtrar por depósito:
                  <select
                    className="form-select"
                    value={depositoId}
                    onChange={(e) => cambiarDeposito(e.target.value)}
                  >
                    <option value="">Todos los depósitos</option>
                    {depositos.map((deposito) => (
                      <option key={deposito.id} value={deposito.id}>
                        {deposito.nombre}
                      </option>
                    ))}
                  </select>
                </label>
              </div>
            )}

            <div className="col-12 col-md-5">
              <div className="form-check mb-2">
                <label className="checkbox form-check-label">
                  <input
                    type="checkbox"
                    className="form-check-input mt-0"
                    checked={soloPendientes}
                    onChange={(e) => cambiarSoloPendientes(e.target.checked)}
                  />
                  Mostrar solamente pendientes
                </label>
              </div>
            </div>
          </div>
        </div>

        {error && (
          <p className="error" role="alert">
            {error}
          </p>
        )}

        {cargando ? (
          <div className="d-flex align-items-center gap-2 text-muted my-3">
            <span
              className="spinner-border spinner-border-sm"
              aria-hidden="true"
            />
            <p className="mb-0">Cargando alertas...</p>
          </div>
        ) : alertas.length === 0 ? (
          <p className="text-muted my-3">
            No se encontraron alertas con los filtros seleccionados.
          </p>
        ) : (
          <div className="tabla-wrapper table-responsive">
            <table className="table table-hover align-middle mb-0">
              <thead>
                <tr>
                  <th>Tipo</th>
                  <th>Producto</th>
                  <th>Depósito</th>
                  <th>Información</th>
                  <th>Fecha de creación</th>
                  <th>Estado</th>
                </tr>
              </thead>

              <tbody>
                {alertas.map((alerta) => (
                  <tr key={alerta.id}>
                    <td>
                      <span
                        className={`badge rounded-pill ${
                          alerta.tipo === "stock_bajo"
                            ? "text-bg-danger"
                            : "text-bg-warning"
                        }`}
                      >
                        {alerta.tipo === "stock_bajo"
                          ? "Stock bajo"
                          : "Vencimiento"}
                      </span>
                    </td>

                    <td className="fw-medium">{alerta.producto_nombre}</td>

                    <td>{alerta.deposito_nombre}</td>

                    <td className="small">
                      {alerta.tipo === "stock_bajo" ? (
                        <div>
                          <div>
                            Stock actual: <strong>{alerta.stock_actual}</strong>
                          </div>
                          <div className="text-muted">
                            Stock mínimo: {alerta.stock_minimo}
                          </div>
                        </div>
                      ) : alerta.fecha_vencimiento ? (
                        <div>
                          Vencimiento:{" "}
                          <strong>
                            {new Date(
                              alerta.fecha_vencimiento
                            ).toLocaleDateString("es-AR")}
                          </strong>
                        </div>
                      ) : (
                        <span className="text-muted">Sin fecha informada</span>
                      )}
                    </td>

                    <td className="text-nowrap small">
                      {formatearFecha(alerta.fecha_creacion)}
                    </td>

                    <td>
                      <span
                        className={`badge rounded-pill ${
                          alerta.resuelta
                            ? "text-bg-success"
                            : "text-bg-secondary"
                        }`}
                      >
                        {alerta.resuelta ? "Resuelta" : "Pendiente"}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        <div className="paginacion">
          <button
            type="button"
            className="btn btn-outline-secondary btn-sm"
            disabled={pagina <= 1 || cargando}
            onClick={() => setPagina((actual) => actual - 1)}
          >
            Anterior
          </button>

          <span className="text-muted small">
            Página {pagina} de {totalPaginas}
          </span>

          <button
            type="button"
            className="btn btn-outline-secondary btn-sm"
            disabled={pagina >= totalPaginas || cargando}
            onClick={() => setPagina((actual) => actual + 1)}
          >
            Siguiente
          </button>
        </div>
      </section>
    </div>
  );
}
