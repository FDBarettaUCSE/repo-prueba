import { useCallback, useEffect, useState, type FormEvent } from "react";
import { useAuth } from "../auth/AuthContext";
import {
  actualizarItemOrdenCompra,
  actualizarOrdenCompra,
  agregarItemOrdenCompra,
  confirmarOrdenCompra,
  crearOrdenCompra,
  descartarOrdenCompra,
  eliminarItemOrdenCompra,
  listarOrdenesCompra,
  obtenerOrdenCompra,
  registrarRecepcionOrdenCompra,
  sugerirOrdenesCompra,
} from "../api/ordenesCompra";
import { listarDepositos } from "../api/depositos";
import { listarProductos } from "../api/productos";
import {
  listarProductosSuministrados,
  listarProveedores,
} from "../api/proveedores";
import { ErrorAPI } from "../api/respuestas";
import type {
  DepositoDTO,
  EstadoOrdenCompra,
  OrdenCompraDTO,
  ProductoDTO,
  ProductoSuministradoDTO,
  ProveedorDTO,
} from "../api/types";

const TAMANIO_PAGINA = 10;

const ESTADOS: EstadoOrdenCompra[] = [
  "borrador",
  "confirmada",
  "parcial",
  "completada",
  "descartada",
];

const COLOR_ESTADO: Record<EstadoOrdenCompra, string> = {
  borrador: "secondary",
  confirmada: "primary",
  parcial: "warning",
  completada: "success",
  descartada: "dark",
};

const formatoMoneda = new Intl.NumberFormat("es-AR", {
  style: "currency",
  currency: "ARS",
});

export function OrdenesCompraPage() {
  const { token, usuario } = useAuth();

  const [ordenes, setOrdenes] = useState<OrdenCompraDTO[]>([]);
  const [depositos, setDepositos] = useState<DepositoDTO[]>([]);
  const [proveedores, setProveedores] = useState<ProveedorDTO[]>([]);
  const [productos, setProductos] = useState<ProductoDTO[]>([]);
  const [suministrados, setSuministrados] = useState<ProductoSuministradoDTO[]>([]);

  const [detalle, setDetalle] = useState<OrdenCompraDTO | null>(null);
  const [cantidadesEdicion, setCantidadesEdicion] = useState<Record<string, string>>({});
  const [recepcion, setRecepcion] = useState<Record<string, string>>({});

  const [pagina, setPagina] = useState(1);
  const [total, setTotal] = useState(0);
  const [filtroDeposito, setFiltroDeposito] = useState("");
  const [filtroEstado, setFiltroEstado] = useState("");
  const [filtroProveedor, setFiltroProveedor] = useState("");

  const [mostrarAlta, setMostrarAlta] = useState(false);
  const [altaDeposito, setAltaDeposito] = useState("");
  const [altaProveedor, setAltaProveedor] = useState("");
  const [proveedorEdicion, setProveedorEdicion] = useState("");
  const [nuevoProductoID, setNuevoProductoID] = useState("");
  const [nuevaCantidad, setNuevaCantidad] = useState("");

  const [cargando, setCargando] = useState(false);
  const [enviando, setEnviando] = useState(false);
  const [error, setError] = useState("");
  const [mensaje, setMensaje] = useState("");

  const esAdministrador = usuario?.rol === "administrador";
  const puedeGestionar = esAdministrador || usuario?.rol === "gerente";
  const veTodosLosDepositos = esAdministrador || usuario?.rol === "auditor";
  const depositoUsuario = usuario?.deposito_asociado;

  const totalPaginas = Math.max(1, Math.ceil(total / TAMANIO_PAGINA));
  const esBorrador = detalle?.estado === "borrador";
  const admiteRecepcion =
    detalle?.estado === "confirmada" || detalle?.estado === "parcial";

  // Carga los datos necesarios para los selectores.
  useEffect(() => {
    if (!token) return;

    listarDepositos(token)
      .then(setDepositos)
      .catch(() => setDepositos([]));

    listarProveedores(token)
      .then(setProveedores)
      .catch(() => setProveedores([]));

    listarProductos(token)
      .then((resultado) => setProductos(resultado.items))
      .catch(() => setProductos([]));
  }, [token]);

  // Carga las órdenes según los filtros y la página seleccionada.
  const cargarOrdenes = useCallback(async () => {
    if (!token) return;

    setCargando(true);
    setError("");

    try {
      const resultado = await listarOrdenesCompra(token, {
        deposito_id: veTodosLosDepositos
          ? filtroDeposito || undefined
          : depositoUsuario,
        estado: (filtroEstado || undefined) as EstadoOrdenCompra | undefined,
        proveedor_id: filtroProveedor || undefined,
        pagina,
        tamanioPagina: TAMANIO_PAGINA,
      });

      setOrdenes(resultado.items ?? []);
      setTotal(resultado.total);
    } catch (e) {
      setError(
        e instanceof ErrorAPI
          ? e.message
          : "No se pudieron cargar las órdenes de compra.",
      );
    } finally {
      setCargando(false);
    }
  }, [
    token,
    veTodosLosDepositos,
    depositoUsuario,
    filtroDeposito,
    filtroEstado,
    filtroProveedor,
    pagina,
  ]);

  useEffect(() => {
    void cargarOrdenes();
  }, [cargarOrdenes]);

  // Carga los productos que suministra el proveedor de la orden seleccionada.
  useEffect(() => {
    if (!token || !detalle?.proveedorId) {
      setSuministrados([]);
      return;
    }

    listarProductosSuministrados(detalle.proveedorId, token)
      .then(setSuministrados)
      .catch(() => setSuministrados([]));
  }, [token, detalle?.proveedorId]);

  function nombreDeposito(id: string) {
    return depositos.find((deposito) => deposito.id === id)?.nombre ?? id;
  }

  function nombreProveedor(id: string) {
    return proveedores.find((proveedor) => proveedor.id === id)?.nombre ?? id;
  }

  function nombreProducto(id: string) {
    const producto = productos.find((p) => p.id === id);
    if (producto) return producto.nombre;

    const suministrado = suministrados.find((p) => p.productoId === id);
    return suministrado?.productoNombre ?? id;
  }

  function totalOrden(orden: OrdenCompraDTO) {
    return (orden.items ?? []).reduce(
      (total, item) => total + item.cantidadSolicitada * item.costoUnitario,
      0,
    );
  }

  function mostrarDetalle(orden: OrdenCompraDTO | null) {
    setDetalle(orden ? { ...orden, items: orden.items ?? [] } : null);
    setCantidadesEdicion({});
    setRecepcion({});
    setNuevoProductoID("");
    setNuevaCantidad("");
    setProveedorEdicion(orden?.proveedorId ?? "");
  }

  async function abrirDetalle(id: string) {
    if (!token) return;

    setError("");
    setMensaje("");

    try {
      const resultado = await obtenerOrdenCompra(token, id);
      mostrarDetalle(resultado);
    } catch (e) {
      setError(
        e instanceof ErrorAPI ? e.message : "No se pudo cargar la orden.",
      );
    }
  }

  // Ejecuta una operación y actualiza el listado y el detalle.
  async function ejecutar(
    accion: () => Promise<unknown>,
    textoOk: string,
    cerrarDetalle = false,
  ) {
    if (!token) return;

    setEnviando(true);
    setError("");
    setMensaje("");

    try {
      await accion();
      setMensaje(textoOk);

      if (cerrarDetalle) {
        mostrarDetalle(null);
      } else if (detalle) {
        const resultado = await obtenerOrdenCompra(token, detalle.id);
        mostrarDetalle(resultado);
      }

      await cargarOrdenes();
    } catch (e) {
      setError(
        e instanceof ErrorAPI ? e.message : "La operación no pudo completarse.",
      );
    } finally {
      setEnviando(false);
    }
  }

  async function manejarAlta(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (!token) return;

    const depositoId = esAdministrador ? altaDeposito : depositoUsuario;

    if (!depositoId || !altaProveedor) {
      setError("Elegí el depósito y el proveedor.");
      return;
    }

    setEnviando(true);
    setError("");
    setMensaje("");

    try {
      const nuevaOrden = await crearOrdenCompra(token, {
        proveedorId: altaProveedor,
        depositoId,
      });

      setMensaje("Orden creada en estado borrador. Agregá los productos.");
      setMostrarAlta(false);
      setAltaProveedor("");
      mostrarDetalle(nuevaOrden);
      await cargarOrdenes();
    } catch (e) {
      setError(
        e instanceof ErrorAPI ? e.message : "No se pudo crear la orden.",
      );
    } finally {
      setEnviando(false);
    }
  }

  async function manejarSugerencias() {
    if (!token) return;

    setEnviando(true);
    setError("");
    setMensaje("");

    try {
      const depositoId = esAdministrador
        ? filtroDeposito || undefined
        : depositoUsuario;
      const resultado = await sugerirOrdenesCompra(token, depositoId);

      setMensaje(
        resultado.cantidad > 0
          ? `${resultado.mensaje} Revisá los borradores generados.`
          : resultado.mensaje || "No hay productos por debajo del mínimo.",
      );
      setFiltroEstado("borrador");
      setPagina(1);
    } catch (e) {
      setError(
        e instanceof ErrorAPI
          ? e.message
          : "No se pudieron generar las sugerencias.",
      );
    } finally {
      setEnviando(false);
    }
  }

  function agregarItem() {
    if (!detalle || !token) return;

    const cantidad = Number(nuevaCantidad);
    if (!nuevoProductoID || !Number.isInteger(cantidad) || cantidad <= 0) {
      setError("Elegí un producto y una cantidad entera mayor a cero.");
      return;
    }

    void ejecutar(
      () =>
        agregarItemOrdenCompra(token, detalle.id, {
          productoId: nuevoProductoID,
          cantidadSolicitada: cantidad,
        }),
      "Producto agregado a la orden.",
    );
  }

  function guardarCantidad(itemId: string, productoId: string) {
    if (!detalle || !token) return;

    const cantidad = Number(cantidadesEdicion[itemId]);
    if (!Number.isInteger(cantidad) || cantidad <= 0) {
      setError("La cantidad debe ser un entero mayor a cero.");
      return;
    }

    void ejecutar(
      () =>
        actualizarItemOrdenCompra(token, detalle.id, itemId, {
          productoId,
          cantidadSolicitada: cantidad,
        }),
      "Cantidad actualizada.",
    );
  }

  function quitarItem(itemId: string) {
    if (!detalle || !token) return;

    void ejecutar(
      () => eliminarItemOrdenCompra(token, detalle.id, itemId),
      "Producto quitado de la orden.",
    );
  }

  function cambiarProveedor() {
    if (!detalle || !token || !proveedorEdicion) return;

    void ejecutar(
      () =>
        actualizarOrdenCompra(token, detalle.id, {
          proveedorId: proveedorEdicion,
          depositoId: detalle.depositoId,
        }),
      "Proveedor actualizado.",
    );
  }

  function confirmar() {
    if (!detalle || !token) return;

    const continuar = window.confirm(
      "Al confirmar, la orden queda cerrada en productos y cantidades y se considera enviada al proveedor. ¿Continuar?",
    );
    if (!continuar) return;

    void ejecutar(
      () => confirmarOrdenCompra(token, detalle.id),
      "Orden confirmada.",
    );
  }

  function descartar() {
    if (!detalle || !token) return;
    if (!window.confirm("¿Descartar este borrador?")) return;

    void ejecutar(
      () => descartarOrdenCompra(token, detalle.id),
      "Borrador descartado.",
      true,
    );
  }

  function registrarRecepcion() {
    if (!detalle || !token) return;

    const items: { itemId: string; cantidad: number }[] = [];

    for (const item of detalle.items) {
      const texto = recepcion[item.id];
      if (!texto) continue;

      const cantidad = Number(texto);
      const pendiente = item.cantidadSolicitada - item.cantidadRecibida;

      if (!Number.isInteger(cantidad) || cantidad < 0) {
        setError("Las cantidades recibidas deben ser enteros positivos.");
        return;
      }

      if (cantidad > pendiente) {
        setError(
          `No se puede recibir más de lo pendiente (${pendiente}) de ${nombreProducto(item.productoId)}.`,
        );
        return;
      }

      if (cantidad > 0) items.push({ itemId: item.id, cantidad });
    }

    if (items.length === 0) {
      setError("Indicá la cantidad recibida de al menos un producto.");
      return;
    }

    void ejecutar(
      () => registrarRecepcionOrdenCompra(token, detalle.id, { items }),
      "Recepción registrada: el stock del depósito fue actualizado.",
    );
  }

  function limpiarFiltros() {
    setFiltroDeposito("");
    setFiltroEstado("");
    setFiltroProveedor("");
    setPagina(1);
  }

  return (
    <main>
      <div className="d-flex flex-wrap align-items-center justify-content-between gap-2 mb-3">
        <div>
          <h1 className="mb-0">Órdenes de compra</h1>
          <p className="text-muted mb-0">
            Flujo: borrador → confirmada → (parcial) → completada
            {!puedeGestionar && " · Solo lectura"}
          </p>
        </div>

        {puedeGestionar && (
          <div className="acciones">
            <button
              type="button"
              className="btn btn-outline-primary"
              disabled={enviando}
              onClick={() => void manejarSugerencias()}
            >
              Generar sugerencias por stock bajo
            </button>
            <button
              type="button"
              className="btn btn-primary"
              onClick={() => {
                setMostrarAlta(!mostrarAlta);
                setError("");
                setMensaje("");
              }}
            >
              {mostrarAlta ? "Cancelar" : "+ Nueva orden"}
            </button>
          </div>
        )}
      </div>

      {error && <p className="error" role="alert">{error}</p>}
      {mensaje && <p className="mensaje" role="status">{mensaje}</p>}

      {puedeGestionar && mostrarAlta && (
        <section className="panel">
          <h2>Nueva orden de compra</h2>
          <form onSubmit={manejarAlta}>
            <div className="row g-3 align-items-end">
              {esAdministrador && (
                <div className="col-12 col-md-4">
                  <label className="form-label mb-0">
                    Depósito destino
                    <select
                      className="form-select"
                      value={altaDeposito}
                      onChange={(e) => setAltaDeposito(e.target.value)}
                      required
                    >
                      <option value="">Seleccionar depósito</option>
                      {depositos.map((deposito) => (
                        <option key={deposito.id} value={deposito.id}>
                          {deposito.nombre}
                        </option>
                      ))}
                    </select>
                  </label>
                </div>
              )}

              <div className="col-12 col-md-4">
                <label className="form-label mb-0">
                  Proveedor
                  <select
                    className="form-select"
                    value={altaProveedor}
                    onChange={(e) => setAltaProveedor(e.target.value)}
                    required
                  >
                    <option value="">Seleccionar proveedor</option>
                    {proveedores.map((proveedor) => (
                      <option key={proveedor.id} value={proveedor.id}>
                        {proveedor.nombre}
                      </option>
                    ))}
                  </select>
                </label>
              </div>

              <div className="col-12 col-md-4">
                <div className="acciones">
                  <button type="submit" className="btn btn-primary" disabled={enviando}>
                    Crear borrador
                  </button>
                  <button
                    type="button"
                    className="btn btn-outline-secondary"
                    onClick={() => setMostrarAlta(false)}
                  >
                    Cancelar
                  </button>
                </div>
              </div>
            </div>
          </form>
        </section>
      )}

      <section className="panel">
        <h2>Filtros</h2>
        <div className="row g-3 align-items-end">
          {veTodosLosDepositos && (
            <div className="col-12 col-md-3">
              <label className="form-label mb-0">
                Depósito
                <select
                  className="form-select"
                  value={filtroDeposito}
                  onChange={(e) => {
                    setFiltroDeposito(e.target.value);
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
              </label>
            </div>
          )}

          <div className="col-12 col-md-3">
            <label className="form-label mb-0">
              Estado
              <select
                className="form-select"
                value={filtroEstado}
                onChange={(e) => {
                  setFiltroEstado(e.target.value);
                  setPagina(1);
                }}
              >
                <option value="">Todos los estados</option>
                {ESTADOS.map((estado) => (
                  <option key={estado} value={estado}>
                    {estado}
                  </option>
                ))}
              </select>
            </label>
          </div>

          <div className="col-12 col-md-3">
            <label className="form-label mb-0">
              Proveedor
              <select
                className="form-select"
                value={filtroProveedor}
                onChange={(e) => {
                  setFiltroProveedor(e.target.value);
                  setPagina(1);
                }}
              >
                <option value="">Todos los proveedores</option>
                {proveedores.map((proveedor) => (
                  <option key={proveedor.id} value={proveedor.id}>
                    {proveedor.nombre}
                  </option>
                ))}
              </select>
            </label>
          </div>

          <div className="col-12 col-md-3">
            <button
              type="button"
              className="btn btn-outline-secondary"
              onClick={limpiarFiltros}
            >
              Limpiar filtros
            </button>
          </div>
        </div>
      </section>

      <section className="panel">
        <div className="d-flex flex-wrap align-items-center justify-content-between gap-2 mb-3">
          <h2 className="mb-0">Listado de órdenes</h2>
          <span className="text-muted small">
            {total} orden{total === 1 ? "" : "es"}
          </span>
        </div>

        {cargando ? (
          <div className="d-flex align-items-center gap-2 text-muted my-3">
            <span className="spinner-border spinner-border-sm" aria-hidden="true" />
            <p className="mb-0">Cargando órdenes...</p>
          </div>
        ) : (
          <>
            <div className="tabla-wrapper table-responsive">
              <table className="table table-hover align-middle mb-0">
                <thead>
                  <tr>
                    <th>Depósito</th>
                    <th>Proveedor</th>
                    <th>Estado</th>
                    <th className="text-end">Productos</th>
                    <th className="text-end">Total</th>
                    <th className="text-end">Acciones</th>
                  </tr>
                </thead>
                <tbody>
                  {ordenes.length === 0 ? (
                    <tr>
                      <td colSpan={6} className="text-center text-muted py-4">
                        No hay órdenes para mostrar.
                      </td>
                    </tr>
                  ) : (
                    ordenes.map((orden) => (
                      <tr key={orden.id}>
                        <td>{nombreDeposito(orden.depositoId)}</td>
                        <td>{nombreProveedor(orden.proveedorId)}</td>
                        <td>
                          <span className={`badge text-bg-${COLOR_ESTADO[orden.estado]}`}>
                            {orden.estado}
                          </span>
                        </td>
                        <td className="text-end">{orden.items?.length ?? 0}</td>
                        <td className="text-end">{formatoMoneda.format(totalOrden(orden))}</td>
                        <td>
                          <div className="acciones justify-content-end">
                            <button
                              type="button"
                              className="btn btn-outline-primary btn-sm"
                              onClick={() => void abrirDetalle(orden.id)}
                            >
                              Ver detalle
                            </button>
                          </div>
                        </td>
                      </tr>
                    ))
                  )}
                </tbody>
              </table>
            </div>

            <div className="paginacion">
              <button
                type="button"
                className="btn btn-outline-secondary btn-sm"
                disabled={pagina <= 1 || cargando}
                onClick={() => setPagina(pagina - 1)}
              >
                Anterior
              </button>
              <span className="text-muted small">
                Página {pagina} de {totalPaginas} — {total} órdenes
              </span>
              <button
                type="button"
                className="btn btn-outline-secondary btn-sm"
                disabled={pagina >= totalPaginas || cargando}
                onClick={() => setPagina(pagina + 1)}
              >
                Siguiente
              </button>
            </div>
          </>
        )}
      </section>

      {detalle && (
        <section className="panel">
          <div className="d-flex flex-wrap align-items-center justify-content-between gap-2 mb-3">
            <div>
              <h2 className="mb-1">Detalle de la orden</h2>
              <p className="text-muted mb-0">
                {nombreProveedor(detalle.proveedorId)} → {nombreDeposito(detalle.depositoId)}{" "}
                <span className={`badge text-bg-${COLOR_ESTADO[detalle.estado]}`}>
                  {detalle.estado}
                </span>
              </p>
            </div>
            <button
              type="button"
              className="btn btn-outline-secondary btn-sm"
              onClick={() => mostrarDetalle(null)}
            >
              Cerrar
            </button>
          </div>

          {puedeGestionar && esBorrador && (
            <div className="row g-2 align-items-end mb-3">
              <div className="col-12 col-md-6">
                <label className="form-label mb-0">
                  Proveedor
                  <select
                    className="form-select"
                    value={proveedorEdicion}
                    onChange={(e) => setProveedorEdicion(e.target.value)}
                  >
                    {proveedores.map((proveedor) => (
                      <option key={proveedor.id} value={proveedor.id}>
                        {proveedor.nombre}
                      </option>
                    ))}
                  </select>
                </label>
              </div>
              <div className="col-12 col-md-3">
                <button
                  type="button"
                  className="btn btn-outline-primary"
                  disabled={enviando || proveedorEdicion === detalle.proveedorId}
                  onClick={cambiarProveedor}
                >
                  Cambiar proveedor
                </button>
              </div>
            </div>
          )}

          <div className="tabla-wrapper table-responsive">
            <table className="table table-hover align-middle mb-0">
              <thead>
                <tr>
                  <th>Producto</th>
                  <th className="text-end">Solicitada</th>
                  <th className="text-end">Recibida</th>
                  <th className="text-end">Costo unitario</th>
                  <th className="text-end">Subtotal</th>
                  {puedeGestionar && esBorrador && <th>Editar</th>}
                  {puedeGestionar && admiteRecepcion && <th>Recibir ahora</th>}
                </tr>
              </thead>
              <tbody>
                {(detalle.items ?? []).length === 0 ? (
                  <tr>
                    <td
                      colSpan={
                        5 +
                        (puedeGestionar && esBorrador ? 1 : 0) +
                        (puedeGestionar && admiteRecepcion ? 1 : 0)
                      }
                      className="text-center text-muted py-3"
                    >
                      La orden todavía no tiene productos.
                    </td>
                  </tr>
                ) : (
                  (detalle.items ?? []).map((item) => {
                    const pendiente = item.cantidadSolicitada - item.cantidadRecibida;

                    return (
                      <tr key={item.id}>
                        <td>{nombreProducto(item.productoId)}</td>
                        <td className="text-end">{item.cantidadSolicitada}</td>
                        <td className="text-end">{item.cantidadRecibida}</td>
                        <td className="text-end">{formatoMoneda.format(item.costoUnitario)}</td>
                        <td className="text-end">
                          {formatoMoneda.format(item.cantidadSolicitada * item.costoUnitario)}
                        </td>

                        {puedeGestionar && esBorrador && (
                          <td>
                            <div className="acciones">
                              <input
                                type="number"
                                min={1}
                                className="form-control form-control-sm"
                                style={{ width: 90 }}
                                value={cantidadesEdicion[item.id] ?? String(item.cantidadSolicitada)}
                                onChange={(e) =>
                                  setCantidadesEdicion({
                                    ...cantidadesEdicion,
                                    [item.id]: e.target.value,
                                  })
                                }
                              />
                              <button
                                type="button"
                                className="btn btn-outline-primary btn-sm"
                                disabled={enviando || cantidadesEdicion[item.id] === undefined}
                                onClick={() => guardarCantidad(item.id, item.productoId)}
                              >
                                Guardar
                              </button>
                              <button
                                type="button"
                                className="btn btn-outline-danger btn-sm"
                                disabled={enviando}
                                onClick={() => quitarItem(item.id)}
                              >
                                Quitar
                              </button>
                            </div>
                          </td>
                        )}

                        {puedeGestionar && admiteRecepcion && (
                          <td>
                            {pendiente > 0 ? (
                              <input
                                type="number"
                                min={0}
                                max={pendiente}
                                placeholder={`Pendiente: ${pendiente}`}
                                className="form-control form-control-sm"
                                value={recepcion[item.id] ?? ""}
                                onChange={(e) =>
                                  setRecepcion({
                                    ...recepcion,
                                    [item.id]: e.target.value,
                                  })
                                }
                              />
                            ) : (
                              <span className="badge text-bg-success">Completo</span>
                            )}
                          </td>
                        )}
                      </tr>
                    );
                  })
                )}
              </tbody>
              <tfoot>
                <tr>
                  <th colSpan={4} className="text-end">Total</th>
                  <th className="text-end">{formatoMoneda.format(totalOrden(detalle))}</th>
                  {(puedeGestionar && (esBorrador || admiteRecepcion)) && (
                    <th colSpan={(esBorrador ? 1 : 0) + (admiteRecepcion ? 1 : 0)} />
                  )}
                </tr>
              </tfoot>
            </table>
          </div>

          {puedeGestionar && esBorrador && (
            <div className="row g-2 align-items-end mt-3">
              <div className="col-12 col-md-5">
                <label className="form-label mb-0">
                  Agregar producto del proveedor
                  <select
                    className="form-select"
                    value={nuevoProductoID}
                    onChange={(e) => setNuevoProductoID(e.target.value)}
                  >
                    <option value="">Seleccionar producto</option>
                    {suministrados.map((producto) => (
                      <option key={producto.productoId} value={producto.productoId}>
                        {producto.productoNombre} · {formatoMoneda.format(producto.costo)} ·{" "}
                        {producto.tiempoEntregaDias} días
                      </option>
                    ))}
                  </select>
                </label>
              </div>
              <div className="col-6 col-md-2">
                <label className="form-label mb-0">
                  Cantidad
                  <input
                    type="number"
                    min={1}
                    className="form-control"
                    value={nuevaCantidad}
                    onChange={(e) => setNuevaCantidad(e.target.value)}
                  />
                </label>
              </div>
              <div className="col-6 col-md-3">
                <button
                  type="button"
                  className="btn btn-outline-primary"
                  disabled={enviando}
                  onClick={agregarItem}
                >
                  Agregar producto
                </button>
              </div>
            </div>
          )}

          {puedeGestionar && (esBorrador || admiteRecepcion) && (
            <div className="acciones mt-3">
              {esBorrador && (
                <>
                  <button
                    type="button"
                    className="btn btn-success"
                    disabled={enviando || (detalle.items ?? []).length === 0}
                    onClick={confirmar}
                  >
                    Confirmar orden
                  </button>
                  <button
                    type="button"
                    className="btn btn-outline-danger"
                    disabled={enviando}
                    onClick={descartar}
                  >
                    Descartar borrador
                  </button>
                </>
              )}

              {admiteRecepcion && (
                <button
                  type="button"
                  className="btn btn-success"
                  disabled={enviando}
                  onClick={registrarRecepcion}
                >
                  Registrar recepción
                </button>
              )}
            </div>
          )}

          {detalle.estado === "completada" && (
            <p className="text-muted mt-3 mb-0">
              Orden completada: no admite nuevas recepciones ni modificaciones.
            </p>
          )}
        </section>
      )}
    </main>
  );
}
