import { useEffect, useState, type FormEvent } from "react";

import { useAuth } from "../auth/AuthContext";

import { listarDepositos } from "../api/depositos";
import { listarCategorias } from "../api/categorias";
import {
  listarStockDeposito,
  configurarStockMinimoDeposito,
  quitarStockMinimoDeposito,
} from "../api/depositoProductos";

import type {
  DepositoDTO,
  CategoriaDTO,
  StockDepositoDTO,
  StockDepositoPaginadoDTO,
} from "../api/types";

export function StockPage() {
  const { usuario, token } = useAuth();

  const [depositos, setDepositos] = useState<DepositoDTO[]>([]);
  const [categorias, setCategorias] = useState<CategoriaDTO[]>([]);

  const [depositoId, setDepositoId] = useState("");
  const [stock, setStock] = useState<StockDepositoPaginadoDTO | null>(
    null,
  );

  const [categoriaId, setCategoriaId] = useState("");
  const [bajoMinimo, setBajoMinimo] = useState(false);
  const [pagina, setPagina] = useState(1);

  const [cargando, setCargando] = useState(false);
  const [error, setError] = useState("");
  const [recargar, setRecargar] = useState(0);

  const [productoEditando, setProductoEditando] =
    useState<StockDepositoDTO | null>(null);
  const [nuevoMinimo, setNuevoMinimo] = useState("");
  const [guardando, setGuardando] = useState(false);

  const esAdministrador = usuario?.rol === "administrador";
  const esAuditor = usuario?.rol === "auditor";
  const esGerente = usuario?.rol === "gerente";
  const esOperario = usuario?.rol === "operario";

  const puedeSeleccionarDeposito = esAdministrador || esAuditor;
  const puedeEditarMinimo = esAdministrador || esGerente;

  // Cargar los depósitos y las categorías.
  useEffect(() => {
    if (!token || !usuario) return;

    const usuarioActual = usuario;
    const tokenActual = token;

    async function cargarDatosIniciales() {
      try {
        setError("");

        const categoriasObtenidas =
          await listarCategorias(tokenActual);

        setCategorias(categoriasObtenidas);

        if (puedeSeleccionarDeposito) {
          const depositosObtenidos =
            await listarDepositos(tokenActual);

          setDepositos(depositosObtenidos);

          setDepositoId((actual) => {
            if (
              actual &&
              depositosObtenidos.some((d) => d.id === actual)
            ) {
              return actual;
            }

            return depositosObtenidos[0]?.id ?? "";
          });
        } else {
          setDepositoId(
            usuarioActual.deposito_asociado ?? "",
          );
        }
      } catch (err) {
        setError(
          err instanceof Error
            ? err.message
            : "No se pudieron cargar los datos iniciales.",
        );
      }
    }

    void cargarDatosIniciales();
  }, [token, usuario, puedeSeleccionarDeposito]);

  // Consultar el inventario del depósito seleccionado.
  useEffect(() => {
    if (!token || !depositoId) {
      setStock(null);
      return;
    }

    let cancelado = false;

    async function cargarStock() {
      try {
        setCargando(true);
        setError("");

        const resultado = await listarStockDeposito(
          depositoId,
          token!,
          {
            categoriaId: categoriaId || undefined,
            bajoMinimo,
            pagina,
            tamanioPagina: 20,
          },
        );

        if (!cancelado) {
          setStock(resultado);
        }
      } catch (err) {
        if (!cancelado) {
          setStock(null);
          setError(
            err instanceof Error
              ? err.message
              : "No se pudo cargar el inventario.",
          );
        }
      } finally {
        if (!cancelado) {
          setCargando(false);
        }
      }
    }

    void cargarStock();

    return () => {
      cancelado = true;
    };
  }, [
    token,
    depositoId,
    categoriaId,
    bajoMinimo,
    pagina,
    recargar,
  ]);

  function cambiarDeposito(nuevoDepositoId: string) {
    setDepositoId(nuevoDepositoId);
    setPagina(1);
    setStock(null);
  }

  function abrirEdicion(producto: StockDepositoDTO) {
    setProductoEditando(producto);

    // Si existe una excepción, mostrarla.
    // Si no existe, partir del mínimo general efectivo.
    setNuevoMinimo(
      String(producto.stockMinimoOverride ?? producto.stockMinimo),
    );
  }

  function cerrarEdicion() {
    setProductoEditando(null);
    setNuevoMinimo("");
  }

  async function guardarMinimo(evento: FormEvent<HTMLFormElement>) {
    evento.preventDefault();

    if (!token || !productoEditando || !depositoId) return;

    const valor = nuevoMinimo.trim() === "" ? NaN : Number(nuevoMinimo);

    if (!Number.isInteger(valor) || valor < 0) {
      setError("El stock mínimo debe ser un entero mayor o igual a cero.");
      return;
    }

    try {
      setGuardando(true);
      setError("");

      await configurarStockMinimoDeposito(
        depositoId,
        productoEditando.productoId,
        { stockMinimo: valor },
        token,
      );

      cerrarEdicion();
      setRecargar((actual) => actual + 1);
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : "No se pudo configurar el mínimo particular.",
      );
    } finally {
      setGuardando(false);
    }
  }

  async function restablecerMinimo(producto: StockDepositoDTO) {
    if (!token || !depositoId) return;

    const confirmar = window.confirm(
      `¿Quitar la excepción para "${producto.productoNombre}"? ` +
        "El depósito volverá a utilizar el mínimo general del producto.",
    );

    if (!confirmar) return;

    try {
      setError("");

      await quitarStockMinimoDeposito(
        depositoId,
        producto.productoId,
        token,
      );

      if (productoEditando?.productoId === producto.productoId) {
        cerrarEdicion();
      }

      setRecargar((actual) => actual + 1);
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : "No se pudo quitar el mínimo particular.",
      );
    }
  }

  if (!usuario || !token) {
    return (
      <div className="panel">
        <p className="mb-0">Debés iniciar sesión para consultar el stock.</p>
      </div>
    );
  }

  if (
    !puedeSeleccionarDeposito &&
    !usuario.deposito_asociado
  ) {
    return (
      <div className="panel">
        <p className="error mb-0">
          Tu usuario no tiene un depósito asignado. Contactá al
          administrador.
        </p>
      </div>
    );
  }

  return (
    <section className="panel">
      <div className="d-flex flex-wrap align-items-center justify-content-between gap-2 mb-3">
        <h1 className="mb-0">Stock por depósito</h1>
        {stock && (
          <span className="badge text-bg-light border">
            Resultados: {stock.total}
          </span>
        )}
      </div>

      {error && (
        <p className="error" role="alert">
          {error}
        </p>
      )}

      <div className="filtros-barra">
        <div className="row g-3 align-items-end">
          {puedeSeleccionarDeposito && (
            <div className="col-12 col-md-4">
              <label htmlFor="deposito" className="form-label">
                Depósito
              </label>
              <select
                id="deposito"
                className="form-select"
                value={depositoId}
                onChange={(evento) => cambiarDeposito(evento.target.value)}
              >
                <option value="">Seleccioná un depósito</option>

                {depositos.map((deposito) => (
                  <option key={deposito.id} value={deposito.id}>
                    {deposito.nombre}
                  </option>
                ))}
              </select>
            </div>
          )}

          <div className="col-12 col-md-4">
            <label htmlFor="categoria" className="form-label">
              Categoría
            </label>
            <select
              id="categoria"
              className="form-select"
              value={categoriaId}
              onChange={(evento) => {
                setCategoriaId(evento.target.value);
                setPagina(1);
              }}
            >
              <option value="">Todas las categorías</option>

              {categorias.map((categoria) => (
                <option key={categoria.id} value={categoria.id}>
                  {categoria.nombre}
                </option>
              ))}
            </select>
          </div>

          <div className="col-12 col-md-4">
            <div className="form-check mb-2">
              <label className="checkbox form-check-label">
                <input
                  type="checkbox"
                  className="form-check-input mt-0"
                  checked={bajoMinimo}
                  onChange={(evento) => {
                    setBajoMinimo(evento.target.checked);
                    setPagina(1);
                  }}
                />
                Mostrar solo productos bajo el mínimo
              </label>
            </div>
          </div>
        </div>
      </div>

      {productoEditando && puedeEditarMinimo && (
        <form onSubmit={guardarMinimo} className="filtros-barra border-primary mb-4">
          <h2 className="mb-1">
            Mínimo particular: {productoEditando.productoNombre}
          </h2>

          <p className="text-muted small mb-2">
            Mínimo efectivo actual: <strong>{productoEditando.stockMinimo}</strong>
          </p>

          <div className="row g-3 align-items-end">
            <div className="col-12 col-md-4">
              <label htmlFor="nuevoMinimo" className="form-label">
                Nuevo mínimo particular
              </label>

              <input
                id="nuevoMinimo"
                type="number"
                className="form-control"
                min="0"
                step="1"
                required
                value={nuevoMinimo}
                onChange={(evento) => setNuevoMinimo(evento.target.value)}
              />
            </div>

            <div className="col-12 col-md-8">
              <div className="acciones">
                <button
                  type="submit"
                  className="btn btn-primary"
                  disabled={guardando}
                >
                  {guardando ? "Guardando..." : "Guardar"}
                </button>

                <button
                  type="button"
                  className="btn btn-outline-secondary"
                  disabled={guardando}
                  onClick={cerrarEdicion}
                >
                  Cancelar
                </button>
              </div>
            </div>
          </div>
        </form>
      )}

      {cargando && (
        <div className="d-flex align-items-center gap-2 text-muted my-3">
          <span className="spinner-border spinner-border-sm" aria-hidden="true" />
          <p className="mb-0">Cargando inventario...</p>
        </div>
      )}

      {!cargando && depositoId && stock && (
        <>
          {stock.items.length === 0 ? (
            <p className="text-muted my-3">
              No hay productos para los filtros seleccionados.
            </p>
          ) : (
            <div className="tabla-wrapper table-responsive">
              <table className="table table-hover align-middle mb-0">
                <thead>
                  <tr>
                    <th>Producto</th>
                    <th>Categoría</th>
                    <th>Unidad</th>
                    <th>Stock actual</th>
                    <th>Mínimo efectivo</th>
                    {!esOperario && <th>Valorizado</th>}
                    <th>Estado</th>
                    {puedeEditarMinimo && <th className="text-end">Acciones</th>}
                  </tr>
                </thead>

                <tbody>
                  {stock.items.map((producto) => {
                    const tieneExcepcion =
                      producto.stockMinimoOverride !== undefined;

                    return (
                      <tr key={producto.productoId}>
                        <td className="fw-medium">{producto.productoNombre}</td>
                        <td>{producto.categoriaNombre}</td>
                        <td>
                          <span className="badge text-bg-light border text-capitalize">
                            {producto.unidadMedida}
                          </span>
                        </td>
                        <td className="fw-semibold">{producto.stock}</td>
                        <td>
                          {producto.stockMinimo}
                          {tieneExcepcion && (
                            <span className="text-primary fw-bold" title="Mínimo particular">
                              {" *"}
                            </span>
                          )}
                        </td>
                        {!esOperario && (
                          <td>
                            {producto.valorizado !== undefined
                              ? producto.valorizado.toLocaleString("es-AR", {
                                  style: "currency",
                                  currency: "ARS",
                                })
                              : "—"}
                          </td>
                        )}
                        <td>
                          <span
                            className={`badge rounded-pill ${
                              producto.bajoMinimo
                                ? "text-bg-danger"
                                : "text-bg-success"
                            }`}
                          >
                            {producto.bajoMinimo
                              ? "Bajo mínimo"
                              : "Normal"}
                          </span>
                        </td>

                        {puedeEditarMinimo && (
                          <td>
                            <div className="acciones justify-content-end">
                              <button
                                type="button"
                                className="btn btn-outline-primary btn-sm"
                                onClick={() => abrirEdicion(producto)}
                              >
                                Configurar mínimo
                              </button>

                              {tieneExcepcion && (
                                <button
                                  type="button"
                                  className="btn btn-outline-secondary btn-sm"
                                  onClick={() =>
                                    void restablecerMinimo(producto)
                                  }
                                >
                                  Restablecer general
                                </button>
                              )}
                            </div>
                          </td>
                        )}
                      </tr>
                    );
                  })}
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
              Página {stock.pagina} · Total {stock.total}
            </span>

            <button
              type="button"
              className="btn btn-outline-secondary btn-sm"
              disabled={
                cargando ||
                pagina * stock.tamanioPagina >= stock.total
              }
              onClick={() => setPagina((actual) => actual + 1)}
            >
              Siguiente
            </button>
          </div>

          <p className="text-muted small mt-2 mb-0">
            * El producto utiliza un mínimo particular configurado para
            este depósito.
          </p>
        </>
      )}

      {!cargando && !depositoId && puedeSeleccionarDeposito && (
        <p className="text-muted mb-0">
          Seleccioná un depósito para consultar su inventario.
        </p>
      )}
    </section>
  );
}
