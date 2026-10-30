import { useCallback, useEffect, useState, type FormEvent } from "react";
import { useAuth } from "../auth/AuthContext";
import {
  listarProductos,
  crearProducto,
  actualizarProducto,
  configurarStockMinimo,
  eliminarProducto,
} from "../api/productos";
import { listarCategorias } from "../api/categorias";
import { ErrorAPI } from "../api/respuestas";
import {
  UnidadMedida,
  type ProductoDTO,
  type CrearProductoDTO,
  type ActualizarProductoDTO,
} from "../api/types";

type FormularioProducto = {
  nombre: string;
  descripcion: string;
  categoriaId: string;
  unidadMedida: (typeof UnidadMedida)[keyof typeof UnidadMedida];
  costoUnitario: string;
  stockMinimo: string;
  fechaVencimiento: string;
};

const formularioVacio: FormularioProducto = {
  nombre: "",
  descripcion: "",
  categoriaId: "",
  unidadMedida: UnidadMedida.Unidad,
  costoUnitario: "0",
  stockMinimo: "0",
  fechaVencimiento: "",
};

const TAMANIO_PAGINA = 10;

export function ProductosPage() {
  const { token, usuario } = useAuth();

  const [productos, setProductos] = useState<ProductoDTO[]>([]);
  const [categorias, setCategorias] = useState<
    Awaited<ReturnType<typeof listarCategorias>>
  >([]);

  const [formulario, setFormulario] =
    useState<FormularioProducto>(formularioVacio);

  const [productoEditando, setProductoEditando] =
    useState<string | null>(null);

  const [categoriaFiltro, setCategoriaFiltro] = useState("");
  const [pagina, setPagina] = useState(1);
  const [total, setTotal] = useState(0);

  const [productoStockEditando, setProductoStockEditando] =
    useState<string | null>(null);
  const [stockMinimo, setStockMinimo] = useState("0");

  const [mostrarFormulario, setMostrarFormulario] = useState(false);
  const [cargando, setCargando] = useState(false);
  const [error, setError] = useState("");
  const [mensaje, setMensaje] = useState("");

  const esAdministrador = usuario?.rol === "administrador";
  const esOperario = usuario?.rol === "operario";
  const puedeEditar =
    esAdministrador || usuario?.rol === "gerente";

  const totalPaginas = Math.max(1, Math.ceil(total / TAMANIO_PAGINA));

  // Carga los productos de la página y categoría seleccionadas.
  const cargarProductos = useCallback(async () => {
    if (!token) return;

    setCargando(true);
    setError("");

    try {
      const resultado = await listarProductos(token, {
        pagina,
        categoria_id: categoriaFiltro || undefined,
      });

      setProductos(resultado.items);
      setTotal(resultado.total);
    } catch (e) {
      setError(
        e instanceof ErrorAPI
          ? e.message
          : "No se pudieron cargar los productos.",
      );
    } finally {
      setCargando(false);
    }
  }, [token, pagina, categoriaFiltro]);

  // Carga las categorías para los selectores.
  const cargarCategorias = useCallback(async () => {
    if (!token) return;

    try {
      const resultado = await listarCategorias(token);
      setCategorias(resultado);
    } catch (e) {
      setError(
        e instanceof ErrorAPI
          ? e.message
          : "No se pudieron cargar las categorías.",
      );
    }
  }, [token]);

  useEffect(() => {
    void cargarProductos();
  }, [cargarProductos]);

  useEffect(() => {
    void cargarCategorias();
  }, [cargarCategorias]);

  function limpiarFormulario() {
    setFormulario(formularioVacio);
    setProductoEditando(null);
    setMostrarFormulario(false);
  }

  function comenzarEdicion(producto: ProductoDTO) {
    setFormulario({
      nombre: producto.nombre,
      descripcion: producto.descripcion,
      categoriaId: producto.categoriaId,
      unidadMedida: producto.unidadMedida,
      costoUnitario: String(producto.costoUnitario),
      stockMinimo: String(producto.stockMinimo),
      fechaVencimiento: producto.fechaVencimiento
        ? producto.fechaVencimiento.slice(0, 10)
        : "",
    });

    setProductoEditando(producto.id);
    setMostrarFormulario(true);
    setMensaje("");
    setError("");
  }

  async function manejarEnvioFormulario(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();

    if (!token) return;

    setError("");
    setMensaje("");

    const fechaVencimiento = formulario.fechaVencimiento
      ? `${formulario.fechaVencimiento}T00:00:00Z`
      : undefined;

    const datos = {
      nombre: formulario.nombre.trim(),
      descripcion: formulario.descripcion.trim(),
      categoriaId: formulario.categoriaId,
      unidadMedida: formulario.unidadMedida,
      costoUnitario: Number(formulario.costoUnitario),
      fechaVencimiento,
    };

    try {
      if (productoEditando) {
        const dto: ActualizarProductoDTO = datos;

        await actualizarProducto(productoEditando, dto, token);
        setMensaje("Producto actualizado correctamente.");
      } else {
        const dto: CrearProductoDTO = {
          ...datos,
          stockMinimo: Number(formulario.stockMinimo),
        };

        await crearProducto(dto, token);
        setPagina(1);
        setCategoriaFiltro("");
        setMensaje("Producto creado correctamente.");
      }

      limpiarFormulario();

      // Recarga el listado después de guardar.
      if (!productoEditando) {
        const resultado = await listarProductos(token, { pagina: 1 });
        setProductos(resultado.items);
        setTotal(resultado.total);
      } else {
        await cargarProductos();
      }
    } catch (e) {
      setError(
        e instanceof ErrorAPI
          ? e.message
          : "No se pudo guardar el producto.",
      );
    }
  }

  async function guardarStockMinimo(
    e: FormEvent<HTMLFormElement>,
  ) {
    e.preventDefault();

    if (!token || !productoStockEditando) return;

    setError("");
    setMensaje("");

    const valor = Number(stockMinimo);

    if (!Number.isInteger(valor) || valor < 0) {
      setError("El stock mínimo debe ser un entero mayor o igual a cero.");
      return;
    }

    try {
      await configurarStockMinimo(
        productoStockEditando,
        { stockMinimo: valor },
        token,
      );

      setProductoStockEditando(null);
      setMensaje("Stock mínimo actualizado correctamente.");
      await cargarProductos();
    } catch (e) {
      setError(
        e instanceof ErrorAPI
          ? e.message
          : "No se pudo actualizar el stock mínimo.",
      );
    }
  }

  async function borrarProducto(producto: ProductoDTO) {
    if (!token) return;

    const confirmar = window.confirm(
      `¿Seguro que querés eliminar el producto "${producto.nombre}"?`,
    );

    if (!confirmar) return;

    setError("");
    setMensaje("");

    try {
      await eliminarProducto(producto.id, token);

      setMensaje("Producto eliminado correctamente.");

      // Si borramos el último de una página, volvemos a la anterior.
      if (productos.length === 1 && pagina > 1) {
        setPagina(pagina - 1);
      } else {
        await cargarProductos();
      }
    } catch (e) {
      setError(
        e instanceof ErrorAPI
          ? e.message
          : "No se pudo eliminar el producto.",
      );
    }
  }

  function cambiarCategoria(categoriaId: string) {
    setCategoriaFiltro(categoriaId);
    setPagina(1);
    setMensaje("");
    setError("");
  }

  function formatearFecha(fecha?: string) {
    if (!fecha) return "—";

    const valor = fecha.slice(0, 10).split("-");
    if (valor.length !== 3) return fecha;

    return `${valor[2]}/${valor[1]}/${valor[0]}`;
  }

  return (
    <main>
      <div className="d-flex flex-wrap align-items-center justify-content-between gap-2 mb-3">
        <h1 className="mb-0">Productos</h1>
        {puedeEditar && (
          <button
            type="button"
            className="btn btn-primary"
            onClick={() => {
              limpiarFormulario();
              setMostrarFormulario(true);
              setError("");
              setMensaje("");
            }}
          >
            + Nuevo producto
          </button>
        )}
      </div>

      {error && (
        <p className="error" role="alert">
          {error}
        </p>
      )}
      {mensaje && <p className="mensaje">{mensaje}</p>}

      {mostrarFormulario && puedeEditar && (
        <section className="panel">
          <h2>
            {productoEditando ? "Editar producto" : "Nuevo producto"}
          </h2>

          <form onSubmit={manejarEnvioFormulario}>
            <div className="row g-3">
              <div className="col-12 col-md-6">
                <label className="form-label mb-0">
                  Nombre
                  <input
                    className="form-control"
                    value={formulario.nombre}
                    onChange={(e) =>
                      setFormulario({ ...formulario, nombre: e.target.value })
                    }
                    required
                  />
                </label>
              </div>

              <div className="col-12 col-md-6">
                <label className="form-label mb-0">
                  Categoría
                  <select
                    className="form-select"
                    value={formulario.categoriaId}
                    onChange={(e) =>
                      setFormulario({
                        ...formulario,
                        categoriaId: e.target.value,
                      })
                    }
                    required
                  >
                    <option value="">Seleccionar categoría</option>
                    {categorias.map((categoria) => (
                      <option key={categoria.id} value={categoria.id}>
                        {categoria.nombre}
                      </option>
                    ))}
                  </select>
                </label>
              </div>

              <div className="col-12">
                <label className="form-label mb-0">
                  Descripción
                  <textarea
                    className="form-control"
                    rows={2}
                    value={formulario.descripcion}
                    onChange={(e) =>
                      setFormulario({
                        ...formulario,
                        descripcion: e.target.value,
                      })
                    }
                  />
                </label>
              </div>

              <div className="col-12 col-sm-6 col-md-3">
                <label className="form-label mb-0">
                  Unidad de medida
                  <select
                    className="form-select"
                    value={formulario.unidadMedida}
                    onChange={(e) =>
                      setFormulario({
                        ...formulario,
                        unidadMedida: e.target.value as FormularioProducto["unidadMedida"],
                      })
                    }
                    required
                  >
                    <option value={UnidadMedida.Unidad}>Unidad</option>
                    <option value={UnidadMedida.Kilogramo}>Kilogramo</option>
                    <option value={UnidadMedida.Litro}>Litro</option>
                    <option value={UnidadMedida.Metro}>Metro</option>
                  </select>
                </label>
              </div>

              <div className="col-12 col-sm-6 col-md-3">
                <label className="form-label mb-0">
                  Costo unitario
                  <input
                    type="number"
                    className="form-control"
                    min="0"
                    step="0.01"
                    value={formulario.costoUnitario}
                    onChange={(e) =>
                      setFormulario({
                        ...formulario,
                        costoUnitario: e.target.value,
                      })
                    }
                    required
                  />
                </label>
              </div>

              {!productoEditando && (
                <div className="col-12 col-sm-6 col-md-3">
                  <label className="form-label mb-0">
                    Stock mínimo
                    <input
                      type="number"
                      className="form-control"
                      min="0"
                      step="1"
                      value={formulario.stockMinimo}
                      onChange={(e) =>
                        setFormulario({
                          ...formulario,
                          stockMinimo: e.target.value,
                        })
                      }
                      required
                    />
                  </label>
                </div>
              )}

              <div className="col-12 col-sm-6 col-md-3">
                <label className="form-label mb-0">
                  Fecha de vencimiento (opcional)
                  <input
                    type="date"
                    className="form-control"
                    value={formulario.fechaVencimiento}
                    onChange={(e) =>
                      setFormulario({
                        ...formulario,
                        fechaVencimiento: e.target.value,
                      })
                    }
                  />
                </label>
              </div>
            </div>

            <div className="acciones mt-1">
              <button type="submit" className="btn btn-primary">
                {productoEditando ? "Guardar cambios" : "Crear producto"}
              </button>
              <button
                type="button"
                className="btn btn-outline-secondary"
                onClick={limpiarFormulario}
              >
                Cancelar
              </button>
            </div>
          </form>
        </section>
      )}

      {esAdministrador && productoStockEditando && (
        <section className="panel border-primary">
          <h2>Modificar stock mínimo</h2>

          <form onSubmit={guardarStockMinimo}>
            <div className="row g-3 align-items-end">
              <div className="col-12 col-md-4">
                <label className="form-label mb-0">
                  Nuevo stock mínimo
                  <input
                    type="number"
                    className="form-control"
                    min="0"
                    step="1"
                    value={stockMinimo}
                    onChange={(e) => setStockMinimo(e.target.value)}
                    required
                  />
                </label>
              </div>

              <div className="col-12 col-md-8">
                <div className="acciones">
                  <button type="submit" className="btn btn-primary">
                    Guardar stock mínimo
                  </button>
                  <button
                    type="button"
                    className="btn btn-outline-secondary"
                    onClick={() => setProductoStockEditando(null)}
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
        <div className="d-flex flex-wrap align-items-end justify-content-between gap-3 mb-3">
          <h2 className="mb-0">Listado de productos</h2>

          <div className="acciones">
            <label className="form-label mb-0" style={{ minWidth: "220px" }}>
              Filtrar por categoría
              <select
                className="form-select form-select-sm"
                value={categoriaFiltro}
                onChange={(e) => cambiarCategoria(e.target.value)}
              >
                <option value="">Todas las categorías</option>
                {categorias.map((categoria) => (
                  <option key={categoria.id} value={categoria.id}>
                    {categoria.nombre}
                  </option>
                ))}
              </select>
            </label>
          </div>
        </div>

        {cargando ? (
          <div className="d-flex align-items-center gap-2 text-muted my-3">
            <span className="spinner-border spinner-border-sm" aria-hidden="true" />
            <p className="mb-0">Cargando productos...</p>
          </div>
        ) : (
          <>
            <div className="tabla-wrapper table-responsive">
              <table className="table table-hover align-middle mb-0">
                <thead>
                  <tr>
                    <th>Nombre</th>
                    <th>Categoría</th>
                    <th>Unidad</th>
                    {!esOperario && <th>Costo unitario</th>}
                    <th>Stock mínimo</th>
                    <th>Vencimiento</th>
                    {(puedeEditar || esAdministrador) && (
                      <th className="text-end">Acciones</th>
                    )}
                  </tr>
                </thead>

                <tbody>
                  {productos.length === 0 ? (
                    <tr>
                      <td colSpan={7} className="text-center text-muted py-4">
                        No hay productos para mostrar.
                      </td>
                    </tr>
                  ) : (
                    productos.map((producto) => (
                      <tr key={producto.id}>
                        <td className="fw-medium">{producto.nombre}</td>
                        <td>{producto.categoriaNombre}</td>
                        <td>
                          <span className="badge text-bg-light border text-capitalize">
                            {producto.unidadMedida}
                          </span>
                        </td>
                        {!esOperario && (
                          <td>
                            {producto.costoUnitario.toLocaleString("es-AR", {
                              style: "currency",
                              currency: "ARS",
                            })}
                          </td>
                        )}
                        <td>{producto.stockMinimo}</td>
                        <td>{formatearFecha(producto.fechaVencimiento)}</td>

                        {(puedeEditar || esAdministrador) && (
                          <td>
                            <div className="acciones justify-content-end">
                              {puedeEditar && (
                                <button
                                  type="button"
                                  className="btn btn-outline-primary btn-sm"
                                  onClick={() => comenzarEdicion(producto)}
                                >
                                  Editar
                                </button>
                              )}

                              {esAdministrador && (
                                <>
                                  <button
                                    type="button"
                                    className="btn btn-outline-secondary btn-sm"
                                    onClick={() => {
                                      setProductoStockEditando(producto.id);
                                      setStockMinimo(
                                        String(producto.stockMinimo),
                                      );
                                      setError("");
                                      setMensaje("");
                                    }}
                                  >
                                    Stock mínimo
                                  </button>

                                  <button
                                    type="button"
                                    className="btn btn-outline-danger btn-sm"
                                    onClick={() => void borrarProducto(producto)}
                                  >
                                    Eliminar
                                  </button>
                                </>
                              )}
                            </div>
                          </td>
                        )}
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
                Página {pagina} de {totalPaginas} — {total} productos
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
    </main>
  );
}
