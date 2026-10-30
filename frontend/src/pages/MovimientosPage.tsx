import { useEffect, useState } from "react";

import { useAuth } from "../auth/AuthContext";

import {
  registrarMovimiento,
  listarMovimientos,
} from "../api/movimientosStock";

import { listarProductos } from "../api/productos";
import { listarDepositos } from "../api/depositos";

import type {
  TipoMovimiento,
  MovimientoStockDTO,
  FiltroMovimientosDTO,
  RegistrarMovimientoDTO,
} from "../api/types";

import type { ProductoDTO } from "../api/types";
import type { DepositoDTO } from "../api/types";

import { ErrorAPI } from "../api/respuestas";

const TAMANIO_PAGINA = 10;

export function MovimientosPage() {
  const { token, usuario } = useAuth();

  const rol = usuario?.rol ?? "";
  const depositoUsuario = usuario?.deposito_asociado ?? "";

  const esAdministrador = rol === "administrador";
  const esAuditor = rol === "auditor";

  const puedeRegistrar =
    rol === "administrador" ||
    rol === "gerente" ||
    rol === "operario";

  // Catálogos
  const [productos, setProductos] = useState<ProductoDTO[]>([]);
  const [depositos, setDepositos] = useState<DepositoDTO[]>([]);

  // Formulario de registro
  const [productoID, setProductoID] = useState("");
  const [depositoID, setDepositoID] = useState("");
  const [tipo, setTipo] = useState<TipoMovimiento>("ingreso");
  const [cantidad, setCantidad] = useState("");
  const [motivo, setMotivo] = useState("");

  // Historial
  const [movimientos, setMovimientos] = useState<
    MovimientoStockDTO[]
  >([]);

  const [pagina, setPagina] = useState(1);
  const [total, setTotal] = useState(0);

  // Filtros
  const [filtroProducto, setFiltroProducto] = useState("");
  const [filtroDeposito, setFiltroDeposito] = useState("");
  const [filtroTipo, setFiltroTipo] = useState("");
  const [fechaDesde, setFechaDesde] = useState("");
  const [fechaHasta, setFechaHasta] = useState("");

  // Estados generales
  const [cargando, setCargando] = useState(false);
  const [cargandoCatalogos, setCargandoCatalogos] = useState(false);
  const [error, setError] = useState("");
  const [mensaje, setMensaje] = useState("");

  // Cargar productos y depósitos
  useEffect(() => {
    async function cargarCatalogos() {
      if (!token) return;

      setCargandoCatalogos(true);

      try {
        const respuestaProductos = await listarProductos(token, {
          todos: "true",
        });

        setProductos(respuestaProductos.items);

        if (esAdministrador || puedeRegistrar || esAuditor) {
          const respuestaDepositos = await listarDepositos(token);
          setDepositos(respuestaDepositos);
        }
      } catch (err) {
        if (err instanceof ErrorAPI) {
          setError(err.message);
        } else {
          setError("No se pudieron cargar los catálogos.");
        }
      } finally {
        setCargandoCatalogos(false);
      }
    }

    cargarCatalogos();
  }, [token, esAdministrador, puedeRegistrar, esAuditor]);

  // Asignar automáticamente el depósito del usuario
  useEffect(() => {
    if (!esAdministrador) {
      setDepositoID(depositoUsuario);
    }
  }, [esAdministrador, depositoUsuario]);

  // Cargar historial
  async function cargarMovimientos(
    paginaSolicitada: number = pagina,
    filtrosPersonalizados?: FiltroMovimientosDTO,
  ) {
    if (!token) return;

    setCargando(true);
    setError("");

    try {
      const filtros: FiltroMovimientosDTO = filtrosPersonalizados ?? {
        pagina: paginaSolicitada,
        tamanioPagina: TAMANIO_PAGINA,
      };

      if (!filtrosPersonalizados) {
        if (filtroProducto) {
          filtros.producto_id = filtroProducto;
        }

        if (esAdministrador || esAuditor) {
          if (filtroDeposito) {
            filtros.deposito_id = filtroDeposito;
          }
        }

        if (filtroTipo) {
          filtros.tipo = filtroTipo as TipoMovimiento;
        }

        if (fechaDesde) {
          filtros.fecha_desde = fechaDesde;
        }

        if (fechaHasta) {
          filtros.fecha_hasta = fechaHasta;
        }
      }

      const respuesta = await listarMovimientos(token, filtros);

      setMovimientos(respuesta.items);
      setTotal(respuesta.total);
      setPagina(respuesta.pagina);
    } catch (err) {
      if (err instanceof ErrorAPI) {
        setError(err.message);
      } else {
        setError("No se pudieron cargar los movimientos.");
      }
    } finally {
      setCargando(false);
    }
  }

  useEffect(() => {
    cargarMovimientos(1);
  }, [token, esAdministrador, esAuditor]);

  // Registrar movimiento
  async function handleRegistrar(e: React.FormEvent) {
    e.preventDefault();

    setError("");
    setMensaje("");

    if (!token) return;

    if (!productoID || !depositoID || !cantidad || !motivo.trim()) {
      setError("Completá todos los campos.");
      return;
    }

    const cantidadNumerica = Number(cantidad);

    if (!Number.isInteger(cantidadNumerica) || cantidadNumerica === 0) {
      setError("La cantidad debe ser un número entero distinto de cero.");
      return;
    }

    if (
      (tipo === "ingreso" || tipo === "egreso") &&
      cantidadNumerica < 0
    ) {
      setError("Los ingresos y egresos requieren una cantidad positiva.");
      return;
    }

    const dto: RegistrarMovimientoDTO = {
      producto_id: productoID,
      deposito_id: depositoID,
      cantidad: cantidadNumerica,
      tipo,
      motivo: motivo.trim(),
    };

    try {
      await registrarMovimiento(dto, token);

      setMensaje("Movimiento registrado correctamente.");

      setProductoID("");
      setCantidad("");
      setMotivo("");
      setTipo("ingreso");

      await cargarMovimientos(1);
    } catch (err) {
      if (err instanceof ErrorAPI) {
        setError(err.message);
      } else {
        setError("No se pudo registrar el movimiento.");
      }
    }
  }

  // Aplicar filtros
  function aplicarFiltros() {
    setPagina(1);
    cargarMovimientos(1);
  }

  // Limpiar filtros
  function limpiarFiltros() {
    setFiltroProducto("");
    setFiltroDeposito("");
    setFiltroTipo("");
    setFechaDesde("");
    setFechaHasta("");
    setPagina(1);

    cargarMovimientos(1, {
      pagina: 1,
      tamanioPagina: TAMANIO_PAGINA,
    });
  }

  const totalPaginas = Math.max(
    1,
    Math.ceil(total / TAMANIO_PAGINA),
  );

  function nombreProducto(id: string) {
    return (
      productos.find((producto) => producto.id === id)?.nombre ??
      id
    );
  }

  function nombreDeposito(id: string) {
    return (
      depositos.find((deposito) => deposito.id === id)?.nombre ??
      id
    );
  }

  function formatearFecha(fecha: string): string {
    const fechaConvertida = new Date(fecha);

    if (Number.isNaN(fechaConvertida.getTime())) {
      return fecha;
    }

    return fechaConvertida.toLocaleString("es-AR", {
      timeZone: "America/Argentina/Buenos_Aires",
      hourCycle: "h23",
    });
  }  
  
  return (
    <div className="panel">
      <div className="d-flex flex-wrap align-items-center justify-content-between gap-2 mb-3">
        <h1 className="mb-0">Movimientos de stock</h1>
        <span className="badge text-bg-light border">
          Total: {total} {total === 1 ? "movimiento" : "movimientos"}
        </span>
      </div>

      {error && (
        <p className="error" role="alert">
          Error: {error}
        </p>
      )}

      {mensaje && <p className="mensaje">{mensaje}</p>}

      {/* REGISTRO DE MOVIMIENTOS */}
      {puedeRegistrar && (
        <section className="filtros-barra mb-4">
          <h2>Registrar movimiento</h2>

          {cargandoCatalogos ? (
            <div className="d-flex align-items-center gap-2 text-muted my-2">
              <span className="spinner-border spinner-border-sm" aria-hidden="true" />
              <p className="mb-0">Cargando productos y depósitos...</p>
            </div>
          ) : (
            <form onSubmit={handleRegistrar}>
              <div className="row g-3">
                <div className="col-12 col-md-6 col-lg-3">
                  <label htmlFor="producto" className="form-label">
                    Producto
                  </label>

                  <select
                    id="producto"
                    className="form-select"
                    value={productoID}
                    onChange={(e) => setProductoID(e.target.value)}
                    required
                  >
                    <option value="">Seleccionar producto</option>

                    {productos.map((producto) => (
                      <option key={producto.id} value={producto.id}>
                        {producto.nombre}
                      </option>
                    ))}
                  </select>
                </div>

                <div className="col-12 col-md-6 col-lg-3">
                  <label htmlFor="deposito" className="form-label">
                    Depósito
                  </label>

                  {esAdministrador ? (
                    <select
                      id="deposito"
                      className="form-select"
                      value={depositoID}
                      onChange={(e) => setDepositoID(e.target.value)}
                      required
                    >
                      <option value="">Seleccionar depósito</option>

                      {depositos.map((deposito) => (
                        <option key={deposito.id} value={deposito.id}>
                          {deposito.nombre}
                        </option>
                      ))}
                    </select>
                  ) : (
                    <input
                      id="deposito"
                      type="text"
                      className="form-control"
                      value={
                        depositos.find(
                          (deposito) => deposito.id === depositoUsuario,
                        )?.nombre ?? "Depósito asignado"
                      }
                      disabled
                    />
                  )}
                </div>

                <div className="col-12 col-md-6 col-lg-3">
                  <label htmlFor="tipo" className="form-label">
                    Tipo de movimiento
                  </label>

                  <select
                    id="tipo"
                    className="form-select"
                    value={tipo}
                    onChange={(e) =>
                      setTipo(e.target.value as TipoMovimiento)
                    }
                  >
                    <option value="ingreso">Ingreso</option>
                    <option value="egreso">Egreso</option>
                    <option value="ajuste">Ajuste</option>
                  </select>
                </div>

                <div className="col-12 col-md-6 col-lg-3">
                  <label htmlFor="cantidad" className="form-label">
                    {tipo === "ajuste"
                      ? "Diferencia de stock (+/-)"
                      : "Cantidad"}
                  </label>

                  <input
                    id="cantidad"
                    type="number"
                    className="form-control"
                    step="1"
                    min={tipo === "ajuste" ? undefined : 1}
                    value={cantidad}
                    onChange={(e) => setCantidad(e.target.value)}
                    placeholder={
                      tipo === "ajuste"
                        ? "Positivo o negativo"
                        : "Ingresá una cantidad positiva"
                    }
                    required
                  />
                </div>

                <div className="col-12">
                  <div className="alert alert-light border py-2 px-3 mb-0 small text-secondary">
                    <div>
                      {tipo === "ingreso" &&
                        "Ingresá una cantidad positiva. Esta cantidad se sumará al stock."}

                      {tipo === "egreso" &&
                        "Ingresá una cantidad positiva. Esta cantidad se restará del stock automáticamente."}

                      {tipo === "ajuste" &&
                        "Ingresá una diferencia positiva para aumentar el stock o negativa para disminuirlo."}
                    </div>
                    <div className="fw-medium mt-1">
                      {tipo === "ingreso" &&
                        "Ejemplo: 10 → se agregan 10 unidades."}

                      {tipo === "egreso" &&
                        "Ejemplo: 5 → se descuentan 5 unidades."}

                      {tipo === "ajuste" &&
                        "Ejemplos: +5 → aumenta 5 unidades; -5 → disminuye 5 unidades."}
                    </div>
                  </div>
                </div>

                <div className="col-12">
                  <label htmlFor="motivo" className="form-label">
                    Motivo
                  </label>

                  <textarea
                    id="motivo"
                    className="form-control"
                    rows={2}
                    placeholder="Describí el motivo del movimiento..."
                    value={motivo}
                    onChange={(e) => setMotivo(e.target.value)}
                    required
                  />
                </div>
              </div>

              <div className="acciones mt-1">
                <button type="submit" className="btn btn-primary">
                  Registrar movimiento
                </button>
              </div>
            </form>
          )}
        </section>
      )}

      {/* HISTORIAL */}
      <section className="mt-3 pt-0 border-0">
        <h2>Historial de movimientos</h2>

        <div className="filtros-barra">
          <div className="row g-3 align-items-end">
            <div className="col-12 col-md-6 col-lg-3">
              <label htmlFor="filtroProducto" className="form-label">
                Producto
              </label>

              <select
                id="filtroProducto"
                className="form-select"
                value={filtroProducto}
                onChange={(e) => setFiltroProducto(e.target.value)}
              >
                <option value="">Todos los productos</option>

                {productos.map((producto) => (
                  <option key={producto.id} value={producto.id}>
                    {producto.nombre}
                  </option>
                ))}
              </select>
            </div>

            {(esAdministrador || esAuditor) && (
              <div className="col-12 col-md-6 col-lg-3">
                <label htmlFor="filtroDeposito" className="form-label">
                  Depósito
                </label>

                <select
                  id="filtroDeposito"
                  className="form-select"
                  value={filtroDeposito}
                  onChange={(e) => setFiltroDeposito(e.target.value)}
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

            <div className="col-12 col-sm-6 col-lg-2">
              <label htmlFor="filtroTipo" className="form-label">
                Tipo
              </label>

              <select
                id="filtroTipo"
                className="form-select"
                value={filtroTipo}
                onChange={(e) => setFiltroTipo(e.target.value)}
              >
                <option value="">Todos</option>
                <option value="ingreso">Ingreso</option>
                <option value="egreso">Egreso</option>
                <option value="ajuste">Ajuste</option>
                <option value="transferencia">Transferencia</option>
              </select>
            </div>

            <div className="col-12 col-sm-6 col-lg-2">
              <label htmlFor="fechaDesde" className="form-label">
                Fecha desde
              </label>

              <input
                id="fechaDesde"
                type="date"
                className="form-control"
                value={fechaDesde}
                onChange={(e) => setFechaDesde(e.target.value)}
              />
            </div>

            <div className="col-12 col-sm-6 col-lg-2">
              <label htmlFor="fechaHasta" className="form-label">
                Fecha hasta
              </label>

              <input
                id="fechaHasta"
                type="date"
                className="form-control"
                value={fechaHasta}
                onChange={(e) => setFechaHasta(e.target.value)}
              />
            </div>
          </div>

          <div className="acciones mt-3">
            <button
              type="button"
              className="btn btn-primary btn-sm"
              onClick={aplicarFiltros}
            >
              Aplicar filtros
            </button>

            <button
              type="button"
              className="btn btn-outline-secondary btn-sm"
              onClick={limpiarFiltros}
            >
              Limpiar filtros
            </button>
          </div>
        </div>

        {cargando ? (
          <div className="d-flex align-items-center gap-2 text-muted my-3">
            <span className="spinner-border spinner-border-sm" aria-hidden="true" />
            <p className="mb-0">Cargando movimientos...</p>
          </div>
        ) : movimientos.length === 0 ? (
          <p className="text-muted my-3">No se encontraron movimientos.</p>
        ) : (
          <>
            <div className="tabla-wrapper table-responsive">
              <table className="table table-hover align-middle mb-0">
                <thead>
                  <tr>
                    <th>Fecha</th>
                    <th>Producto</th>
                    <th>Depósito</th>
                    <th>Tipo</th>
                    <th>Cantidad</th>
                    <th>Motivo</th>
                  </tr>
                </thead>

                <tbody>
                  {movimientos.map((movimiento) => (
                    <tr key={movimiento.id}>
                      <td className="text-nowrap small">
                        {formatearFecha(movimiento.fecha_hora)}
                      </td>

                      <td className="fw-medium">
                        {nombreProducto(movimiento.producto_id)}
                      </td>

                      <td>
                        {nombreDeposito(movimiento.deposito_id)}
                      </td>

                      <td>
                        <span
                          className={`badge rounded-pill text-capitalize ${
                            movimiento.tipo === "ingreso"
                              ? "text-bg-success"
                              : movimiento.tipo === "egreso"
                                ? "text-bg-danger"
                                : "text-bg-warning"
                          }`}
                        >
                          {movimiento.tipo}
                        </span>
                      </td>

                      <td className="fw-semibold">
                        {movimiento.cantidad}
                      </td>

                      <td>
                        {movimiento.motivo}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            <div className="paginacion">
              <button
                type="button"
                className="btn btn-outline-secondary btn-sm"
                disabled={pagina <= 1}
                onClick={() => cargarMovimientos(pagina - 1)}
              >
                Anterior
              </button>

              <span className="text-muted small">
                Página {pagina} de {totalPaginas} — Total de movimientos: {total}
              </span>

              <button
                type="button"
                className="btn btn-outline-secondary btn-sm"
                disabled={pagina >= totalPaginas}
                onClick={() => cargarMovimientos(pagina + 1)}
              >
                Siguiente
              </button>
            </div>
          </>
        )}
      </section>
    </div>
  );
}
