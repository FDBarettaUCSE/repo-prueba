import { useEffect, useState } from "react";
import { useAuth } from "../auth/AuthContext";

import {
  listarTransferencias,
  solicitarTransferencia,
  aprobarTransferencia,
  rechazarTransferencia,
  completarTransferencia,
} from "../api/transferencias";

import { listarProductos } from "../api/productos";
import { listarDepositos } from "../api/depositos";

import type {
  TransferenciaDTO,
  EstadoTransferencia,
  SolicitarTransferenciaDTO,
} from "../api/types";

import type { ProductoDTO } from "../api/types";
import type { DepositoDTO } from "../api/types";
import { formatearFecha } from "../api/fechas";

const TAMANIO_PAGINA = 10;

export function TransferenciasPage() {
  const { token, usuario } = useAuth();

  const rol = usuario?.rol;
  const depositoUsuario = usuario?.deposito_asociado;

  const esAdministrador = rol === "administrador";
  const esAuditor = rol === "auditor";

  const puedeSolicitar = rol === "gerente" || rol === "operario";

  const [transferencias, setTransferencias] = useState<TransferenciaDTO[]>([]);
  const [productos, setProductos] = useState<ProductoDTO[]>([]);
  const [depositos, setDepositos] = useState<DepositoDTO[]>([]);

  const [pagina, setPagina] = useState(1);
  const [total, setTotal] = useState(0);

  const [filtroDeposito, setFiltroDeposito] = useState("");
  const [filtroEstado, setFiltroEstado] = useState("");

  const [cargando, setCargando] = useState(false);
  const [error, setError] = useState("");
  const [mensaje, setMensaje] = useState("");

  const [origenID, setOrigenID] = useState("");
  const [destinoID, setDestinoID] = useState("");
  const [productoID, setProductoID] = useState("");
  const [cantidad, setCantidad] = useState("");
  const [motivo, setMotivo] = useState("");

  const totalPaginas = Math.max(1, Math.ceil(total / TAMANIO_PAGINA));

  useEffect(() => {
    if (depositoUsuario) {
      setOrigenID(depositoUsuario);
    }
  }, [depositoUsuario]);

  useEffect(() => {
    if (!token) return;

    async function cargarCatalogos() {
      try {
        const [productosRespuesta, depositosRespuesta] = await Promise.all([
          listarProductos(token!, { todos: "true" }),
          listarDepositos(token!),
        ]);

        setProductos(productosRespuesta.items);
        setDepositos(depositosRespuesta);
      } catch (err) {
        setError(
          err instanceof Error
            ? err.message
            : "No se pudieron cargar los productos y depósitos."
        );
      }
    }

    cargarCatalogos();
  }, [token]);

  useEffect(() => {
    if (!token) return;

    cargarTransferencias();
  }, [token, pagina, filtroDeposito, filtroEstado]);

  async function cargarTransferencias(
    paginaActual = pagina,
    deposito = filtroDeposito,
    estado = filtroEstado
  ) {
    if (!token) return;

    setCargando(true);
    setError("");

    try {
      const respuesta = await listarTransferencias(token, {
        pagina: paginaActual,
        tamanioPagina: TAMANIO_PAGINA,
        ...(deposito ? { deposito_id: deposito } : {}),
        ...(estado
          ? { estado: estado as EstadoTransferencia }
          : {}),
      });

      setTransferencias(respuesta.items);
      setTotal(respuesta.total);
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : "No se pudieron cargar las transferencias."
      );
    } finally {
      setCargando(false);
    }
  }

  function obtenerNombreDeposito(id: string) {
    return depositos.find((d) => d.id === id)?.nombre ?? id;
  }

  function obtenerNombreProducto(id: string) {
    return productos.find((p) => p.id === id)?.nombre ?? id;
  }


  function obtenerError(err: unknown, mensajePredeterminado: string) {
    return err instanceof Error ? err.message : mensajePredeterminado;
  }

  async function manejarSolicitud(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();

    setError("");
    setMensaje("");

    if (!token) return;

    if (!depositoUsuario || origenID !== depositoUsuario) {
      setError("Solo podés solicitar transferencias desde tu depósito.");
      return;
    }

    if (!destinoID || !productoID || !cantidad) {
      setError("Completá todos los campos obligatorios.");
      return;
    }

    if (origenID === destinoID) {
      setError("El depósito de origen y destino deben ser diferentes.");
      return;
    }

    const cantidadNumerica = Number(cantidad);

    if (!Number.isInteger(cantidadNumerica) || cantidadNumerica <= 0) {
      setError("La cantidad debe ser un número entero mayor a cero.");
      return;
    }

    const dto: SolicitarTransferenciaDTO = {
      deposito_origen_id: origenID,
      deposito_destino_id: destinoID,
      producto_id: productoID,
      cantidad: cantidadNumerica,
      motivo,
    };

    try {
      await solicitarTransferencia(dto, token);

      setMensaje("La transferencia fue solicitada correctamente.");

      setDestinoID("");
      setProductoID("");
      setCantidad("");
      setMotivo("");

      setPagina(1);
      await cargarTransferencias(1);
    } catch (err) {
      setError(obtenerError(err, "No se pudo solicitar la transferencia."));
    }
  }

  async function manejarAprobacion(id: string) {
    if (!token) return;

    setError("");
    setMensaje("");

    try {
      await aprobarTransferencia(id, token);

      setMensaje("La transferencia fue aprobada correctamente.");
      await cargarTransferencias();
    } catch (err) {
      setError(obtenerError(err, "No se pudo aprobar la transferencia."));
    }
  }

  async function manejarRechazo(id: string) {
    if (!token) return;

    const motivoRechazo = window.prompt(
      "Ingresá el motivo del rechazo:"
    );

    if (motivoRechazo === null) return;

    if (!motivoRechazo.trim()) {
      setError("Debés indicar un motivo para rechazar la transferencia.");
      return;
    }

    setError("");
    setMensaje("");

    try {
      await rechazarTransferencia(
        id,
        {
          accion: "rechazar",
          motivo: motivoRechazo,
        },
        token
      );

      setMensaje("La transferencia fue rechazada correctamente.");
      await cargarTransferencias();
    } catch (err) {
      setError(obtenerError(err, "No se pudo rechazar la transferencia."));
    }
  }

  async function manejarCompletado(id: string) {
    if (!token) return;

    setError("");
    setMensaje("");

    try {
      await completarTransferencia(id, {}, token);

      setMensaje("La transferencia fue completada correctamente.");
      await cargarTransferencias();
    } catch (err) {
      setError(obtenerError(err, "No se pudo completar la transferencia."));
    }
  }

  function puedeResolver(t: TransferenciaDTO) {
    if (
      t.estado !== "solicitada" ||
      (rol !== "administrador" && rol !== "gerente")
    ) {
      return false;
    }

    if (esAdministrador) return true;

    return depositoUsuario === t.deposito_destino_id;
  }

  function puedeCompletar(t: TransferenciaDTO) {
    if (
      t.estado !== "aprobada" ||
      (rol !== "administrador" && rol !== "gerente" && rol !== "operario")
    ) {
      return false;
    }

    if (esAdministrador) return true;

    return depositoUsuario === t.deposito_destino_id;
  }

  function limpiarFiltros() {
    setFiltroDeposito("");
    setFiltroEstado("");
    setPagina(1);

    cargarTransferencias(1, "", "");
  }

  return (
    <div className="panel">
      <div className="d-flex flex-wrap align-items-center justify-content-between gap-2 mb-3">
        <h1 className="mb-0">Transferencias de stock</h1>
        <span className="badge text-bg-light border">
          Total: {total} {total === 1 ? "transferencia" : "transferencias"}
        </span>
      </div>

      {error && (
        <p className="error" role="alert">
          {error}
        </p>
      )}
      {mensaje && <p className="mensaje">{mensaje}</p>}

      {puedeSolicitar && (
        <section className="filtros-barra mb-4">
          <h2>Solicitar transferencia</h2>

          <form onSubmit={manejarSolicitud}>
            <div className="row g-3">
              <div className="col-12 col-md-6 col-lg-3">
                <label className="form-label">Depósito de origen</label>

                <select className="form-select" value={origenID} disabled>
                  <option value="">
                    Seleccioná un depósito
                  </option>

                  {depositos
                    .filter((d) => d.id === depositoUsuario)
                    .map((d) => (
                      <option key={d.id} value={d.id}>
                        {d.nombre}
                      </option>
                    ))}
                </select>
              </div>

              <div className="col-12 col-md-6 col-lg-3">
                <label className="form-label">Depósito de destino</label>

                <select
                  className="form-select"
                  value={destinoID}
                  onChange={(e) => setDestinoID(e.target.value)}
                  required
                >
                  <option value="">Seleccioná un depósito</option>

                  {depositos
                    .filter((d) => d.id !== origenID)
                    .map((d) => (
                      <option key={d.id} value={d.id}>
                        {d.nombre}
                      </option>
                    ))}
                </select>
              </div>

              <div className="col-12 col-md-6 col-lg-3">
                <label className="form-label">Producto</label>

                <select
                  className="form-select"
                  value={productoID}
                  onChange={(e) => setProductoID(e.target.value)}
                  required
                >
                  <option value="">Seleccioná un producto</option>

                  {productos.map((p) => (
                    <option key={p.id} value={p.id}>
                      {p.nombre}
                    </option>
                  ))}
                </select>
              </div>

              <div className="col-12 col-md-6 col-lg-3">
                <label className="form-label">Cantidad</label>

                <input
                  type="number"
                  className="form-control"
                  min="1"
                  step="1"
                  value={cantidad}
                  onChange={(e) => setCantidad(e.target.value)}
                  required
                />
              </div>

              <div className="col-12">
                <label className="form-label">Motivo</label>

                <textarea
                  className="form-control"
                  rows={2}
                  value={motivo}
                  onChange={(e) => setMotivo(e.target.value)}
                  placeholder="Motivo de la transferencia"
                />
              </div>
            </div>

            <div className="acciones mt-1">
              <button type="submit" className="btn btn-primary">
                Solicitar transferencia
              </button>
            </div>
          </form>
        </section>
      )}

      <section className="mt-3 pt-0 border-0">
        <h2>Historial de transferencias</h2>

        <div className="filtros-barra">
          <div className="row g-3 align-items-end">
            <div className="col-12 col-md-5">
              <label className="form-label">Filtrar por depósito</label>

              <select
                className="form-select"
                value={filtroDeposito}
                onChange={(e) => {
                  setFiltroDeposito(e.target.value);
                  setPagina(1);
                }}
                disabled={!esAdministrador && !esAuditor}
              >
                <option value="">Todos los depósitos</option>

                {depositos.map((d) => (
                  <option key={d.id} value={d.id}>
                    {d.nombre}
                  </option>
                ))}
              </select>
            </div>

            <div className="col-12 col-md-5">
              <label className="form-label">Filtrar por estado</label>

              <select
                className="form-select"
                value={filtroEstado}
                onChange={(e) => {
                  setFiltroEstado(e.target.value);
                  setPagina(1);
                }}
              >
                <option value="">Todos los estados</option>
                <option value="solicitada">Solicitada</option>
                <option value="aprobada">Aprobada</option>
                <option value="rechazada">Rechazada</option>
                <option value="completada">Completada</option>
              </select>
            </div>

            <div className="col-12 col-md-2">
              {(filtroDeposito || filtroEstado) && (
                <button
                  type="button"
                  className="btn btn-outline-secondary w-100"
                  onClick={limpiarFiltros}
                >
                  Limpiar filtros
                </button>
              )}
            </div>
          </div>
        </div>

        {cargando ? (
          <div className="d-flex align-items-center gap-2 text-muted my-3">
            <span className="spinner-border spinner-border-sm" aria-hidden="true" />
            <p className="mb-0">Cargando transferencias...</p>
          </div>
        ) : transferencias.length === 0 ? (
          <p className="text-muted my-3">No hay transferencias para mostrar.</p>
        ) : (
          <div className="tabla-wrapper table-responsive">
            <table className="table table-hover align-middle mb-0">
              <thead>
                <tr>
                  <th>Fecha</th>
                  <th>Origen</th>
                  <th>Destino</th>
                  <th>Producto</th>
                  <th>Cantidad</th>
                  <th>Estado</th>
                  <th>Motivo</th>
                  <th className="text-end">Acciones</th>
                </tr>
              </thead>

              <tbody>
                {transferencias.map((t) => (
                  <tr key={t.id}>
                    <td className="text-nowrap small">
                      {formatearFecha(t.fecha_solicitud)}
                    </td>

                    <td>
                      {obtenerNombreDeposito(t.deposito_origen_id)}
                    </td>

                    <td>
                      {obtenerNombreDeposito(t.deposito_destino_id)}
                    </td>

                    <td className="fw-medium">
                      {obtenerNombreProducto(t.producto_id)}
                    </td>

                    <td className="fw-semibold">{t.cantidad}</td>

                    <td>
                      <span
                        className={`badge rounded-pill text-capitalize ${
                          t.estado === "completada"
                            ? "text-bg-success"
                            : t.estado === "aprobada"
                              ? "text-bg-info"
                              : t.estado === "rechazada"
                                ? "text-bg-danger"
                                : "text-bg-warning"
                        }`}
                      >
                        {t.estado}
                      </span>
                    </td>

                    <td>{t.motivo || "-"}</td>

                    <td>
                      <div className="acciones justify-content-end">
                        {puedeResolver(t) && (
                          <>
                            <button
                              type="button"
                              className="btn btn-success btn-sm"
                              onClick={() => manejarAprobacion(t.id)}
                            >
                              Aprobar
                            </button>

                            <button
                              type="button"
                              className="btn btn-outline-danger btn-sm"
                              onClick={() => manejarRechazo(t.id)}
                            >
                              Rechazar
                            </button>
                          </>
                        )}

                        {puedeCompletar(t) && (
                          <button
                            type="button"
                            className="btn btn-primary btn-sm"
                            onClick={() => manejarCompletado(t.id)}
                          >
                            Completar
                          </button>
                        )}

                        {!puedeResolver(t) && !puedeCompletar(t) && (
                          <span className="text-muted">-</span>
                        )}
                      </div>
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
            disabled={pagina <= 1}
            onClick={() => setPagina((p) => p - 1)}
          >
            Anterior
          </button>

          <span className="text-muted small">
            Página {pagina} de {totalPaginas}
          </span>

          <button
            type="button"
            className="btn btn-outline-secondary btn-sm"
            disabled={pagina >= totalPaginas}
            onClick={() => setPagina((p) => p + 1)}
          >
            Siguiente
          </button>
        </div>
      </section>
    </div>
  );
}
