
import { useEffect, useState } from "react";
import { useAuth } from "../auth/AuthContext";

import {
  obtenerReporteInventario,
  obtenerReporteAlertas,
  obtenerReportePendientes,
} from "../api/reportes";

import { listarMovimientos } from "../api/movimientosStock";
import { listarTransferencias } from "../api/transferencias";
import { listarNovedadesDeposito } from "../api/novedadesDepositos";

import type {
  EstadoTransferencia,
  ReporteInventarioDTO,
  ReporteAlertasDTO,
  ReportePendientesDTO,
  MovimientosPaginadosDTO,
  NovedadesPaginadoDTO,
} from "../api/types";

import { ErrorAPI } from "../api/respuestas";

const ESTADOS_TRANSFERENCIA: EstadoTransferencia[] = [
  "solicitada",
  "aprobada",
  "rechazada",
  "completada",
];

function formatearNumero(valor: number): string {
  return valor.toLocaleString("es-AR");
}

function formatearMoneda(valor: number): string {
  return valor.toLocaleString("es-AR", {
    style: "currency",
    currency: "ARS",
  });
}

function formatearFecha(fecha: string): string {
  const fechaConvertida = new Date(fecha);

  if (Number.isNaN(fechaConvertida.getTime())) {
    return fecha;
  }

  return fechaConvertida.toLocaleString("es-AR");
}

function obtenerNombreEstado(estado: EstadoTransferencia): string {
  const nombres: Record<EstadoTransferencia, string> = {
    solicitada: "Solicitadas",
    aprobada: "Aprobadas",
    rechazada: "Rechazadas",
    completada: "Completadas",
  };

  return nombres[estado];
}

export function InicioPage() {
  const { usuario, token } = useAuth();

  const rol = usuario?.rol ?? "";
  const depositoUsuario = usuario?.deposito_asociado ?? "";

  const esAdministrador = rol === "administrador";
  const esGerente = rol === "gerente";
  const esOperario = rol === "operario";
  const esAuditor = rol === "auditor";

  // Reportes
  const [inventario, setInventario] =
    useState<ReporteInventarioDTO | null>(null);

  const [alertas, setAlertas] =
    useState<ReporteAlertasDTO | null>(null);

  const [pendientes, setPendientes] =
    useState<ReportePendientesDTO | null>(null);

  // Actividad
  const [movimientos, setMovimientos] =
    useState<MovimientosPaginadosDTO | null>(null);

  const [transferenciasPorEstado, setTransferenciasPorEstado] =
    useState<Record<EstadoTransferencia, number>>({
      solicitada: 0,
      aprobada: 0,
      rechazada: 0,
      completada: 0,
    });

  const [novedades, setNovedades] =
    useState<NovedadesPaginadoDTO | null>(null);

  // Estado de la página
  const [cargando, setCargando] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    async function cargarDashboard() {
      if (!token || !usuario) return;

      if ((esGerente || esOperario) && !depositoUsuario) {
        setError("El usuario no tiene un depósito asociado.");
        return;
      }

      setCargando(true);
      setError("");

      // Limpiar datos anteriores
      setInventario(null);
      setAlertas(null);
      setPendientes(null);
      setMovimientos(null);
      setNovedades(null);

      setTransferenciasPorEstado({
        solicitada: 0,
        aprobada: 0,
        rechazada: 0,
        completada: 0,
      });

      try {
        const filtros = esGerente || esOperario
          ? { deposito_id: depositoUsuario }
          : {};

        // Información común a todos los roles
        const [
          respuestaAlertas,
          respuestaMovimientos,
          respuestaInventario,
          respuestaPendientes,
        ] = await Promise.all([
          obtenerReporteAlertas(token, filtros),

          listarMovimientos(token, {
            ...filtros,
            pagina: 1,
            tamanioPagina: 5,
          }),

          esAdministrador || esGerente || esAuditor
            ? obtenerReporteInventario(token, filtros)
            : Promise.resolve(null),

          esAdministrador || esGerente || esAuditor
            ? obtenerReportePendientes(token, filtros)
            : Promise.resolve(null),
        ]);

        setAlertas(respuestaAlertas);
        setMovimientos(respuestaMovimientos);
        setInventario(respuestaInventario);
        setPendientes(respuestaPendientes);

        // Información adicional del operario
        if (esOperario) {
          const [respuestasTransferencias, respuestaNovedades] =
            await Promise.all([
              Promise.all(
                ESTADOS_TRANSFERENCIA.map((estado) =>
                  listarTransferencias(token, {
                    deposito_id: depositoUsuario,
                    estado,
                    pagina: 1,
                    tamanioPagina: 1,
                  })
                )
              ),

              listarNovedadesDeposito(
                token,
                depositoUsuario,
                {
                  solo_no_leidas: true,
                  pagina: 1,
                  tamanioPagina: 1,
                }
              ),
            ]);

          const conteos = {} as Record<EstadoTransferencia, number>;

          ESTADOS_TRANSFERENCIA.forEach((estado, indice) => {
            conteos[estado] = respuestasTransferencias[indice].total;
          });

          setTransferenciasPorEstado(conteos);
          setNovedades(respuestaNovedades);
        }
      } catch (err) {
        if (err instanceof ErrorAPI) {
          setError(err.message);
        } else if (err instanceof Error) {
          setError(err.message);
        } else {
          setError("Ocurrió un error al cargar el dashboard.");
        }
      } finally {
        setCargando(false);
      }
    }

    cargarDashboard();
  }, [token, usuario, rol, depositoUsuario]);

  if (!usuario) return null;

  const totalOrdenesPorRevisar =
    pendientes?.ordenes.reduce(
      (total, deposito) => total + deposito.borrador,
      0
    ) ?? 0;

  const totalTransferenciasPendientes =
    pendientes?.transferencias.reduce(
      (total, deposito) => total + deposito.pendientes_de_aprobar,
      0
    ) ?? 0;

  return (
    <div>
      <h2 className="mb-4">Inicio</h2>

      {cargando && (
        <p className="text-muted">Cargando información...</p>
      )}

      {error && (
        <div className="alert alert-danger">
          {error}
        </div>
      )}

      {!cargando && !error && (
        <>
          {/* INVENTARIO: administrador, gerente y auditor */}
          {inventario && (
            <section className="mb-5">
              <h4 className="mb-3">Inventario</h4>

              <div className="row g-3 mb-4">
                <div className="col-md-6">
                  <div className="card h-100">
                    <div className="card-body">
                      <h6 className="text-muted">Stock total</h6>
                      <h3 className="mb-0">
                        {formatearNumero(inventario.stock_total)}
                      </h3>
                    </div>
                  </div>
                </div>

                <div className="col-md-6">
                  <div className="card h-100">
                    <div className="card-body">
                      <h6 className="text-muted">Valorización total</h6>
                      <h3 className="mb-0">
                        {formatearMoneda(inventario.valorizado_total)}
                      </h3>
                    </div>
                  </div>
                </div>
              </div>

              <div className="row g-3">
                <div className="col-lg-6">
                  <div className="card h-100">
                    <div className="card-body">
                      <h5 className="card-title">
                        Stock por depósito
                      </h5>

                      {inventario.por_deposito.length === 0 ? (
                        <p className="text-muted mb-0">
                          No hay información disponible.
                        </p>
                      ) : (
                        <ul className="list-group list-group-flush">
                          {inventario.por_deposito.map((deposito) => (
                            <li
                              key={deposito.deposito_id}
                              className="list-group-item px-0"
                            >
                              <div className="d-flex justify-content-between gap-3">
                                <strong>{deposito.deposito_nombre}</strong>
                                <span>
                                  {formatearNumero(deposito.stock)} unidades
                                </span>
                              </div>

                              <small className="text-muted">
                                Valorización: {formatearMoneda(deposito.valorizado)}
                              </small>
                            </li>
                          ))}
                        </ul>
                      )}
                    </div>
                  </div>
                </div>

                <div className="col-lg-6">
                  <div className="card h-100">
                    <div className="card-body">
                      <h5 className="card-title">
                        Stock por categoría
                      </h5>

                      {inventario.por_categoria.length === 0 ? (
                        <p className="text-muted mb-0">
                          No hay información disponible.
                        </p>
                      ) : (
                        <ul className="list-group list-group-flush">
                          {inventario.por_categoria.map((categoria) => (
                            <li
                              key={categoria.categoria_id}
                              className="list-group-item px-0"
                            >
                              <div className="d-flex justify-content-between gap-3">
                                <strong>{categoria.categoria_nombre}</strong>
                                <span>
                                  {formatearNumero(categoria.stock)} unidades
                                </span>
                              </div>

                              <small className="text-muted">
                                Valorización: {formatearMoneda(categoria.valorizado)}
                              </small>
                            </li>
                          ))}
                        </ul>
                      )}
                    </div>
                  </div>
                </div>
              </div>
            </section>
          )}

          {/* ALERTAS: todos los roles */}
          {alertas && (
            <section className="mb-5">
              <h4 className="mb-3">Alertas</h4>

              <div className="row g-3 mb-3">
                <div className="col-md-6">
                  <div className="card h-100">
                    <div className="card-body">
                      <h6 className="text-muted">
                        Productos bajo mínimo
                      </h6>
                      <h3 className="mb-0">
                        {formatearNumero(alertas.paginacion_minimo.total)}
                      </h3>
                    </div>
                  </div>
                </div>

                <div className="col-md-6">
                  <div className="card h-100">
                    <div className="card-body">
                      <h6 className="text-muted">
                        Productos por vencer
                      </h6>
                      <h3 className="mb-0">
                        {formatearNumero(alertas.paginacion_vencimiento.total)}
                      </h3>
                    </div>
                  </div>
                </div>
              </div>

              <div className="row g-3">
                <div className="col-lg-6">
                  <div className="card h-100">
                    <div className="card-body">
                      <h5 className="card-title">
                        Bajo mínimo
                      </h5>

                      {alertas.bajo_minimo.length === 0 ? (
                        <p className="text-muted mb-0">
                          No hay productos bajo mínimo.
                        </p>
                      ) : (
                        <ul className="list-group list-group-flush">
                          {alertas.bajo_minimo.slice(0, 5).map((producto) => (
                            <li
                              key={`${producto.deposito_id}-${producto.producto_id}`}
                              className="list-group-item px-0"
                            >
                              <strong>{producto.producto_nombre}</strong>
                              <div className="small text-muted">
                                {producto.deposito_nombre} · Stock:{" "}
                                {producto.stock} / Mínimo: {producto.stock_minimo}
                              </div>
                            </li>
                          ))}
                        </ul>
                      )}
                    </div>
                  </div>
                </div>

                <div className="col-lg-6">
                  <div className="card h-100">
                    <div className="card-body">
                      <h5 className="card-title">
                        Próximos vencimientos
                      </h5>

                      {alertas.por_vencer.length === 0 ? (
                        <p className="text-muted mb-0">
                          No hay productos próximos a vencer.
                        </p>
                      ) : (
                        <ul className="list-group list-group-flush">
                          {alertas.por_vencer.slice(0, 5).map((producto) => (
                            <li
                              key={`${producto.deposito_id}-${producto.producto_id}`}
                              className="list-group-item px-0"
                            >
                              <strong>{producto.producto_nombre}</strong>
                              <div className="small text-muted">
                                {producto.deposito_nombre} · Vence:{" "}
                                {formatearFecha(producto.fecha_vencimiento)}
                              </div>
                            </li>
                          ))}
                        </ul>
                      )}
                    </div>
                  </div>
                </div>
              </div>
            </section>
          )}

          {/* MOVIMIENTOS: todos los roles */}
          {movimientos && (
            <section className="mb-5">
              <h4 className="mb-3">Últimos movimientos</h4>

              <div className="card">
                <div className="card-body">
                  {movimientos.items.length === 0 ? (
                    <p className="text-muted mb-0">
                      Todavía no hay movimientos registrados.
                    </p>
                  ) : (
                    <ul className="list-group list-group-flush">
                      {movimientos.items.map((movimiento) => (
                        <li
                          key={movimiento.id}
                          className="list-group-item px-0"
                        >
                          <div className="d-flex justify-content-between flex-wrap gap-2">
                            <strong className="text-capitalize">
                              {movimiento.tipo}
                            </strong>

                            <span>
                              {formatearNumero(movimiento.cantidad)} unidades
                            </span>
                          </div>

                          <div className="small text-muted">
                            Producto: {movimiento.producto_nombre}
                          </div>

                          <div className="small text-muted">
                            {formatearFecha(movimiento.fecha_hora)}
                          </div>
                        </li>
                      ))}
                    </ul>
                  )}
                </div>
              </div>
            </section>
          )}

          {/* PENDIENTES: administrador, gerente y auditor */}
          {pendientes && (
            <section className="mb-5">
              <h4 className="mb-3">Pendientes</h4>

              <div className="row g-3">
                <div className="col-md-6">
                  <div className="card h-100">
                    <div className="card-body">
                      <h6 className="text-muted">
                        Órdenes para revisar
                      </h6>
                      <h3 className="mb-0">
                        {formatearNumero(totalOrdenesPorRevisar)}
                      </h3>
                      <small className="text-muted">
                        Órdenes en estado borrador.
                      </small>
                    </div>
                  </div>
                </div>

                <div className="col-md-6">
                  <div className="card h-100">
                    <div className="card-body">
                      <h6 className="text-muted">
                        Transferencias pendientes de aprobación
                      </h6>
                      <h3 className="mb-0">
                        {formatearNumero(totalTransferenciasPendientes)}
                      </h3>
                    </div>
                  </div>
                </div>
              </div>
            </section>
          )}

          {/* ACTIVIDAD DEL OPERARIO */}
          {esOperario && (
            <section className="mb-5">
              <h4 className="mb-3">Mi actividad</h4>

              <div className="row g-3 mb-4">
                <div className="col-md-6">
                  <div className="card h-100">
                    <div className="card-body">
                      <h6 className="text-muted">
                        Novedades sin leer
                      </h6>
                      <h3 className="mb-0">
                        {formatearNumero(novedades?.no_leidas ?? 0)}
                      </h3>
                    </div>
                  </div>
                </div>
              </div>

              <div className="card">
                <div className="card-body">
                  <h5 className="card-title">
                    Mis transferencias
                  </h5>

                  <div className="row g-3">
                    {ESTADOS_TRANSFERENCIA.map((estado) => (
                      <div className="col-6 col-md-3" key={estado}>
                        <div className="border rounded p-3 h-100">
                          <div className="text-muted small">
                            {obtenerNombreEstado(estado)}
                          </div>
                          <h4 className="mb-0">
                            {formatearNumero(transferenciasPorEstado[estado])}
                          </h4>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            </section>
          )}
        </>
      )}
    </div>
  );
}
