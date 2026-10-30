
import { useCallback, useEffect, useState } from "react";
import { useAuth } from "../../auth/AuthContext";
import { obtenerReporteInventario } from "../../api/reportes";
import { listarDepositos } from "../../api/depositos";
import { listarCategorias } from "../../api/categorias";
import { ErrorAPI } from "../../api/respuestas";
import type {
  DepositoDTO,
  ReporteInventarioDTO,
} from "../../api/types";

interface Props {
  activo: boolean;
}

export function ReporteInventario({ activo }: Props) {
  const { token, usuario } = useAuth();

  const [reporte, setReporte] = useState<ReporteInventarioDTO | null>(null);
  const [depositos, setDepositos] = useState<DepositoDTO[]>([]);
  const [categorias, setCategorias] =
    useState<Awaited<ReturnType<typeof listarCategorias>>>([]);

  const [depositoFiltro, setDepositoFiltro] = useState("");
  const [categoriaFiltro, setCategoriaFiltro] = useState("");
  const [pagina, setPagina] = useState(1);
  const [vistaActiva, setVistaActiva] = useState<"deposito" | "categoria">("deposito");

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
      const resultado = await obtenerReporteInventario(token, {
        deposito_id: depositoFiltro || undefined,
        categoria_id: categoriaFiltro || undefined,
        pagina,
      });

      setReporte(resultado);
    } catch (e) {
      setError(
        e instanceof ErrorAPI
          ? e.message
          : "No se pudo cargar el reporte de inventario.",
      );
    } finally {
      setCargando(false);
    }
  }, [token, activo, depositoFiltro, categoriaFiltro, pagina]);

  // Cargar los depósitos y las categorías.
  useEffect(() => {
    if (!activo || !token) return;

    async function cargarCatalogos() {
      try {
        const resultadoCategorias = await listarCategorias(token!);
        setCategorias(resultadoCategorias);

        if (puedeElegirDeposito) {
          const resultadoDepositos = await listarDepositos(token!);
          setDepositos(resultadoDepositos);
        }
      } catch (e) {
        setError(
          e instanceof ErrorAPI
            ? e.message
            : "No se pudieron cargar los filtros.",
        );
      }
    }

    void cargarCatalogos();
  }, [activo, token, puedeElegirDeposito]);

  useEffect(() => {
    void cargarReporte();
  }, [cargarReporte]);

  const totalPaginasDeposito = reporte
    ? Math.max(
        1,
        Math.ceil(
          reporte.paginacion_deposito.total /
            reporte.paginacion_deposito.tamanioPagina,
        ),
      )
    : 1;

  const totalPaginasCategoria = reporte
    ? Math.max(
        1,
        Math.ceil(
          reporte.paginacion_categoria.total /
            reporte.paginacion_categoria.tamanioPagina,
        ),
      )
    : 1;

  return (
    <section className="panel">
      <h2>Reporte de inventario</h2>

      {error && (
        <p className="error" role="alert">
          {error}
        </p>
      )}

      <div className="filtros-barra mb-4">
        <div className="row g-3">
          {puedeElegirDeposito && (
            <div className="col-12 col-md-5">
              <label htmlFor="filtro-deposito-inventario" className="form-label">
                Depósito
              </label>
              <select
                id="filtro-deposito-inventario"
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

          <div className="col-12 col-md-5">
            <label htmlFor="filtro-categoria-inventario" className="form-label">
              Categoría
            </label>
            <select
              id="filtro-categoria-inventario"
              className="form-select"
              value={categoriaFiltro}
              onChange={(e) => {
                setCategoriaFiltro(e.target.value);
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
        </div>
      </div>

      {cargando ? (
        <div className="d-flex align-items-center gap-2 text-muted my-3">
          <span
            className="spinner-border spinner-border-sm"
            aria-hidden="true"
          />
          <p className="mb-0">Cargando reporte de inventario...</p>
        </div>
      ) : reporte ? (
        <>
          <div className="reporte-resumen">
            <div className="reporte-indicador">
              <span>Stock total</span>
              <strong>{reporte.stock_total.toLocaleString("es-AR")}</strong>
            </div>

            <div className="reporte-indicador">
              <span>Valorización total</span>
              <strong>
                {reporte.valorizado_total.toLocaleString("es-AR", {
                  style: "currency",
                  currency: "ARS",
                })}
              </strong>
            </div>
          </div>
          <div className="acciones mb-3">
            <button
              type="button"
              className={`btn ${vistaActiva === "deposito" ? "btn-primary" : "btn-outline-primary"}`}
              onClick={() => {
                setVistaActiva("deposito");
                setPagina(1);
              }}
            >
              Por depósito
            </button>

            <button
              type="button"
              className={`btn ${vistaActiva === "categoria" ? "btn-primary" : "btn-outline-primary"}`}
              onClick={() => {
                setVistaActiva("categoria");
                setPagina(1);
              }}
            >
              Por categoría
            </button>
          </div>

          {vistaActiva === "deposito" ? (
            <section>
              <h2>Inventario por depósito</h2>

              <div className="tabla-wrapper table-responsive">
                <table className="table table-hover align-middle mb-0">
                  <thead>
                    <tr>
                      <th>Depósito</th>
                      <th>Stock</th>
                      <th>Valorización</th>
                    </tr>
                  </thead>

                  <tbody>
                    {reporte.por_deposito.length === 0 ? (
                      <tr>
                        <td colSpan={3} className="text-center text-muted py-4">
                          No hay datos para mostrar.
                        </td>
                      </tr>
                    ) : (
                      reporte.por_deposito.map((item) => (
                        <tr key={item.deposito_id}>
                          <td className="fw-medium">{item.deposito_nombre}</td>
                          <td>{item.stock.toLocaleString("es-AR")}</td>
                          <td>
                            {item.valorizado.toLocaleString("es-AR", {
                              style: "currency",
                              currency: "ARS",
                            })}
                          </td>
                        </tr>
                      ))
                    )}
                  </tbody>
                </table>
              </div>

              <div className="paginacion">
                <span>
                  Página {reporte.paginacion_deposito.pagina} de{" "}
                  {totalPaginasDeposito}
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
                    disabled={pagina >= totalPaginasDeposito || cargando}
                    onClick={() => setPagina((actual) => actual + 1)}
                  >
                    Siguiente
                  </button>
                </div>
              </div>
            </section>
          ) : (
            <section>
              <h2>Inventario por categoría</h2>

              <div className="tabla-wrapper table-responsive">
                <table className="table table-hover align-middle mb-0">
                  <thead>
                    <tr>
                      <th>Categoría</th>
                      <th>Stock</th>
                      <th>Valorización</th>
                    </tr>
                  </thead>

                  <tbody>
                    {reporte.por_categoria.length === 0 ? (
                      <tr>
                        <td colSpan={3} className="text-center text-muted py-4">
                          No hay datos para mostrar.
                        </td>
                      </tr>
                    ) : (
                      reporte.por_categoria.map((item) => (
                        <tr key={item.categoria_id}>
                          <td className="fw-medium">{item.categoria_nombre}</td>
                          <td>{item.stock.toLocaleString("es-AR")}</td>
                          <td>
                            {item.valorizado.toLocaleString("es-AR", {
                              style: "currency",
                              currency: "ARS",
                            })}
                          </td>
                        </tr>
                      ))
                    )}
                  </tbody>
                </table>
              </div>

              <div className="paginacion">
                <span>
                  Página {reporte.paginacion_categoria.pagina} de{" "}
                  {totalPaginasCategoria}
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
                    disabled={pagina >= totalPaginasCategoria || cargando}
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
