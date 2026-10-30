import {
  useCallback,
  useEffect,
  useState,
  type FormEvent,
} from "react";
import { useAuth } from "../auth/AuthContext";
import {
  listarProveedores,
  listarProductosSuministrados,
  agregarProductoSuministrado,
  actualizarProductoSuministrado,
  eliminarProductoSuministrado,
} from "../api/proveedores";
import { listarProductos } from "../api/productos";
import { ErrorAPI } from "../api/respuestas";
import type {
  ProveedorDTO,
  ProductoDTO,
  ProductoSuministradoDTO,
  AgregarProductoSuministradoDTO,
  ActualizarProductoSuministradoDTO,
} from "../api/types";

interface RelacionProductoProveedor extends ProductoSuministradoDTO {
  proveedorId: string;
  proveedorNombre: string;
}

interface FormularioSuministro {
  proveedorId: string;
  productoId: string;
  costo: string;
  tiempoEntregaDias: string;
}

const FORMULARIO_VACIO: FormularioSuministro = {
  proveedorId: "",
  productoId: "",
  costo: "",
  tiempoEntregaDias: "",
};

export function ProductosSuministradosPage() {
  const { token, usuario } = useAuth();

  const puedeEditar =
    usuario?.rol === "administrador" ||
    usuario?.rol === "gerente";

  const [proveedores, setProveedores] = useState<ProveedorDTO[]>([]);
  const [productos, setProductos] = useState<ProductoDTO[]>([]);
  const [relaciones, setRelaciones] = useState<
    RelacionProductoProveedor[]
  >([]);

  const [formulario, setFormulario] =
    useState<FormularioSuministro>(FORMULARIO_VACIO);
  const [relacionEditando, setRelacionEditando] =
    useState<RelacionProductoProveedor | null>(null);

  const [filtroProveedor, setFiltroProveedor] = useState("");
  const [filtroProducto, setFiltroProducto] = useState("");

  const [cargando, setCargando] = useState(true);
  const [cargandoCatalogos, setCargandoCatalogos] = useState(true);
  const [enviando, setEnviando] = useState(false);

  const [error, setError] = useState<string | null>(null);
  const [mensaje, setMensaje] = useState<string | null>(null);

  const cargarCatalogos = useCallback(async () => {
    if (!token) return;

    setCargandoCatalogos(true);
    setError(null);

    try {
      const [datosProveedores, datosProductos] = await Promise.all([
        listarProveedores(token),
        listarProductos(token, { todos: "true" }),
      ]);

      setProveedores(datosProveedores);
      setProductos(datosProductos.items);
    } catch (e) {
      setError(
        e instanceof ErrorAPI
          ? e.message
          : "No se pudieron cargar los proveedores y productos.",
      );
    } finally {
      setCargandoCatalogos(false);
    }
  }, [token]);

  const cargarRelaciones = useCallback(async () => {
    if (!token) return;

    setCargando(true);
    setError(null);

    try {
      const datosProveedores = await listarProveedores(token);

      // Cada proveedor tiene su propia lista de productos suministrados.
      const listas = await Promise.all(
        datosProveedores.map(async (proveedor) => {
          const suministros = await listarProductosSuministrados(
            proveedor.id,
            token,
          );

          return suministros.map((suministro) => ({
            ...suministro,
            proveedorId: proveedor.id,
            proveedorNombre: proveedor.nombre,
          }));
        }),
      );

      setRelaciones(listas.flat());
    } catch (e) {
      setError(
        e instanceof ErrorAPI
          ? e.message
          : "No se pudieron cargar los productos suministrados.",
      );
    } finally {
      setCargando(false);
    }
  }, [token]);

  useEffect(() => {
    void cargarCatalogos();
  }, [cargarCatalogos]);

  useEffect(() => {
    void cargarRelaciones();
  }, [cargarRelaciones]);

  function cancelarEdicion() {
    setFormulario(FORMULARIO_VACIO);
    setRelacionEditando(null);
    setError(null);
  }

  function editarRelacion(relacion: RelacionProductoProveedor) {
    setFormulario({
      proveedorId: relacion.proveedorId,
      productoId: relacion.productoId,
      costo: String(relacion.costo),
      tiempoEntregaDias: String(relacion.tiempoEntregaDias),
    });

    setRelacionEditando(relacion);
    setError(null);
    setMensaje(null);
  }

  async function manejarEnvio(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();

    if (!token || !puedeEditar) return;

    setError(null);
    setMensaje(null);

    if (!relacionEditando && !formulario.proveedorId) {
      setError("Seleccioná un proveedor.");
      return;
    }

    if (!relacionEditando && !formulario.productoId) {
      setError("Seleccioná un producto.");
      return;
    }

    const costo = Number(formulario.costo);
    const tiempoEntregaDias = Number(formulario.tiempoEntregaDias);

    if (!Number.isFinite(costo) || costo < 0) {
      setError("El costo debe ser un número mayor o igual a cero.");
      return;
    }

    if (
      !Number.isInteger(tiempoEntregaDias) ||
      tiempoEntregaDias < 0
    ) {
      setError(
        "El tiempo de entrega debe ser un entero mayor o igual a cero.",
      );
      return;
    }

    setEnviando(true);

    try {
      if (relacionEditando) {
        const dto: ActualizarProductoSuministradoDTO = {
          costo,
          tiempoEntregaDias,
        };

        await actualizarProductoSuministrado(
          relacionEditando.proveedorId,
          relacionEditando.productoId,
          dto,
          token,
        );

        setMensaje("Relación actualizada correctamente.");
      } else {
        const dto: AgregarProductoSuministradoDTO = {
          productoId: formulario.productoId,
          costo,
          tiempoEntregaDias,
        };

        await agregarProductoSuministrado(
          formulario.proveedorId,
          dto,
          token,
        );

        setMensaje("Producto suministrado agregado correctamente.");
      }

      cancelarEdicion();
      await cargarRelaciones();
    } catch (e) {
      setError(
        e instanceof ErrorAPI
          ? e.message
          : "No se pudo guardar la relación.",
      );
    } finally {
      setEnviando(false);
    }
  }

  async function manejarEliminar(
    relacion: RelacionProductoProveedor,
  ) {
    if (!token || !puedeEditar) return;

    if (
      !window.confirm(
        `¿Quitar "${relacion.productoNombre}" de "${relacion.proveedorNombre}"?`,
      )
    ) {
      return;
    }

    setError(null);
    setMensaje(null);

    try {
      await eliminarProductoSuministrado(
        relacion.proveedorId,
        relacion.productoId,
        token,
      );

      if (
        relacionEditando?.proveedorId === relacion.proveedorId &&
        relacionEditando?.productoId === relacion.productoId
      ) {
        cancelarEdicion();
      }

      setMensaje("Relación eliminada correctamente.");
      await cargarRelaciones();
    } catch (e) {
      setError(
        e instanceof ErrorAPI
          ? e.message
          : "No se pudo eliminar la relación.",
      );
    }
  }

  const relacionesFiltradas = relaciones.filter((relacion) => {
    const coincideProveedor =
      !filtroProveedor || relacion.proveedorId === filtroProveedor;

    const coincideProducto =
      !filtroProducto || relacion.productoId === filtroProducto;

    return coincideProveedor && coincideProducto;
  });

  // Evitar ofrecer una relación que ya existe para el mismo proveedor.
  const productosDisponibles = productos.filter(
    (producto) =>
      !relaciones.some(
        (relacion) =>
          relacion.proveedorId === formulario.proveedorId &&
          relacion.productoId === producto.id,
      ),
  );

  function formatearCosto(costo: number) {
    return costo.toLocaleString("es-AR", {
      style: "currency",
      currency: "ARS",
    });
  }

  return (
    <div className="panel">
      <div className="d-flex flex-wrap align-items-center justify-content-between gap-2 mb-3">
        <h1 className="mb-0">Productos suministrados</h1>
        <span className="badge text-bg-light border">
          {relacionesFiltradas.length}{" "}
          {relacionesFiltradas.length === 1 ? "relación" : "relaciones"}
        </span>
      </div>

      {mensaje && <p className="mensaje">{mensaje}</p>}
      {error && (
        <p className="error" role="alert">
          {error}
        </p>
      )}

      {puedeEditar && (
        <section className="filtros-barra mb-4">
          <h2>
            {relacionEditando
              ? "Editar suministro"
              : "Agregar producto suministrado"}
          </h2>

          <form onSubmit={manejarEnvio}>
            {relacionEditando && (
              <div className="alert alert-secondary py-2 mb-1">
                <strong>Proveedor:</strong> {relacionEditando.proveedorNombre}
                {" — "}
                <strong>Producto:</strong> {relacionEditando.productoNombre}
              </div>
            )}

            <div className="row g-3">
              {!relacionEditando && (
                <>
                  <div className="col-12 col-md-6 col-lg-3">
                    <label className="form-label mb-0">
                      Proveedor
                      <select
                        className="form-select"
                        value={formulario.proveedorId}
                        onChange={(e) =>
                          setFormulario({
                            ...formulario,
                            proveedorId: e.target.value,
                            productoId: "",
                          })
                        }
                        required
                        disabled={cargandoCatalogos}
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

                  <div className="col-12 col-md-6 col-lg-3">
                    <label className="form-label mb-0">
                      Producto
                      <select
                        className="form-select"
                        value={formulario.productoId}
                        onChange={(e) =>
                          setFormulario({
                            ...formulario,
                            productoId: e.target.value,
                          })
                        }
                        required
                        disabled={
                          cargandoCatalogos || !formulario.proveedorId
                        }
                      >
                        <option value="">Seleccionar producto</option>
                        {productosDisponibles.map((producto) => (
                          <option key={producto.id} value={producto.id}>
                            {producto.nombre}
                          </option>
                        ))}
                      </select>
                    </label>
                  </div>
                </>
              )}

              <div className="col-12 col-md-6 col-lg-3">
                <label className="form-label mb-0">
                  Costo de suministro
                  <input
                    type="number"
                    className="form-control"
                    min="0"
                    step="0.01"
                    value={formulario.costo}
                    onChange={(e) =>
                      setFormulario({
                        ...formulario,
                        costo: e.target.value,
                      })
                    }
                    required
                  />
                </label>
              </div>

              <div className="col-12 col-md-6 col-lg-3">
                <label className="form-label mb-0">
                  Tiempo de entrega (días)
                  <input
                    type="number"
                    className="form-control"
                    min="0"
                    step="1"
                    value={formulario.tiempoEntregaDias}
                    onChange={(e) =>
                      setFormulario({
                        ...formulario,
                        tiempoEntregaDias: e.target.value,
                      })
                    }
                    required
                  />
                </label>
              </div>
            </div>

            <div className="acciones mt-1">
              <button
                type="submit"
                className="btn btn-primary"
                disabled={
                  enviando ||
                  cargandoCatalogos ||
                  (!relacionEditando &&
                    (!formulario.proveedorId ||
                      !formulario.productoId))
                }
              >
                {enviando
                  ? "Guardando..."
                  : relacionEditando
                    ? "Guardar cambios"
                    : "Agregar relación"}
              </button>

              {relacionEditando && (
                <button
                  type="button"
                  className="btn btn-outline-secondary"
                  onClick={cancelarEdicion}
                  disabled={enviando}
                >
                  Cancelar
                </button>
              )}
            </div>
          </form>
        </section>
      )}

      <section className="mt-3 pt-0 border-0">
        <h2>Listado de productos suministrados</h2>

        <div className="filtros-barra">
          <div className="row g-3 align-items-end">
            <div className="col-12 col-md-5">
              <label className="form-label mb-0">
                Filtrar por proveedor
                <select
                  className="form-select"
                  value={filtroProveedor}
                  onChange={(e) => setFiltroProveedor(e.target.value)}
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

            <div className="col-12 col-md-5">
              <label className="form-label mb-0">
                Filtrar por producto
                <select
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
              </label>
            </div>

            <div className="col-12 col-md-2">
              <button
                type="button"
                className="btn btn-outline-secondary w-100"
                onClick={() => {
                  setFiltroProveedor("");
                  setFiltroProducto("");
                }}
              >
                Limpiar filtros
              </button>
            </div>
          </div>
        </div>

        {(cargando || cargandoCatalogos) && (
          <div className="d-flex align-items-center gap-2 text-muted my-3">
            <span className="spinner-border spinner-border-sm" aria-hidden="true" />
            <p className="mb-0">Cargando información...</p>
          </div>
        )}

        {!cargando && !cargandoCatalogos && (
          <div className="tabla-wrapper table-responsive">
            <table className="table table-hover align-middle mb-0">
              <thead>
                <tr>
                  <th>Proveedor</th>
                  <th>Producto</th>
                  <th>Costo</th>
                  <th>Entrega (días)</th>
                  {puedeEditar && <th className="text-end">Acciones</th>}
                </tr>
              </thead>

              <tbody>
                {relacionesFiltradas.length === 0 ? (
                  <tr>
                    <td
                      colSpan={puedeEditar ? 5 : 4}
                      className="text-center text-muted py-4"
                    >
                      No hay relaciones para mostrar.
                    </td>
                  </tr>
                ) : (
                  relacionesFiltradas.map((relacion) => (
                    <tr
                      key={`${relacion.proveedorId}-${relacion.productoId}`}
                    >
                      <td className="fw-medium">{relacion.proveedorNombre}</td>
                      <td>{relacion.productoNombre}</td>
                      <td>{formatearCosto(relacion.costo)}</td>
                      <td>
                        <span className="badge text-bg-light border">
                          {relacion.tiempoEntregaDias} días
                        </span>
                      </td>

                      {puedeEditar && (
                        <td>
                          <div className="acciones justify-content-end">
                            <button
                              type="button"
                              className="btn btn-outline-primary btn-sm"
                              onClick={() => editarRelacion(relacion)}
                            >
                              Editar
                            </button>

                            <button
                              type="button"
                              className="btn btn-outline-danger btn-sm"
                              onClick={() =>
                                void manejarEliminar(relacion)
                              }
                            >
                              Quitar
                            </button>
                          </div>
                        </td>
                      )}
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        )}
      </section>
    </div>
  );
}
