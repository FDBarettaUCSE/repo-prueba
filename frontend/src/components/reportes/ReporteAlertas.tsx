
import { useCallback, useEffect, useState } from "react";
import { useAuth } from "../../auth/AuthContext";
import { obtenerReporteAlertas } from "../../api/reportes";
import { listarDepositos } from "../../api/depositos";
import { ErrorAPI } from "../../api/respuestas";
import type {
  DepositoDTO,
  ReporteAlertasDTO,
} from "../../api/types";

interface Props {
  activo: boolean;
}

export function ReporteAlertas({ activo }: Props) {
  const { token, usuario } = useAuth();

  const [reporte, setReporte] = useState<ReporteAlertasDTO | null>(null);
  const [depositos, setDepositos] = useState<DepositoDTO[]>([]);

  const [depositoFiltro, setDepositoFiltro] = useState("");
  const [diasVencimiento, setDiasVencimiento] = useState(30);
  const [pagina, setPagina] = useState(1);
  const [vistaActiva, setVistaActiva] = useState<"minimo" | "vencimiento">("minimo");

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
      const resultado = await obtenerReporteAlertas(token, {
        deposito_id: depositoFiltro || undefined,
        dias_vencimiento: diasVencimiento,
        pagina,
      });

      setReporte(resultado);
    } catch (e) {
      setError(
        e instanceof ErrorAPI
          ? e.message
          : "No se pudo cargar el reporte de alertas.",
      );
    } finally {
      setCargando(false);
    }
  }, [token, activo, depositoFiltro, diasVencimiento, pagina]);

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

  const totalPaginasMinimo = reporte
    ? Math.max(
        1,
        Math.ceil(
          reporte.paginacion_minimo.total /
            reporte.paginacion_minimo.tamanioPagina,
        ),
      )
    : 1;

  const totalPaginasVencimiento = reporte
    ? Math.max(
        1,
        Math.ceil(
          reporte.paginacion_vencimiento.total /
            reporte.paginacion_vencimiento.tamanioPagina,
        ),
      )
    : 1;

  return (
    <section className="panel">
      <h2>Reporte de alertas</h2>

      {error && (
        <p className="error" role="alert">
          {error}
        </p>
      )}

      <div className="filtros-barra mb-4">
        <div className="row g-3">
          {puedeElegirDeposito && (
            <div className="col-12 col-md-6">
              <label htmlFor="filtro-deposito-alertas" className="form-label">
                Depósito
              </label>
              <select
                id="filtro-deposito-alertas"
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

          <div className="col-12 col-md-6">
            <label htmlFor="filtro-dias-vencimiento" className="form-label">
              Próximos a vencer dentro de
            </label>
            <select
              id="filtro-dias-vencimiento"
              className="form-select"
              value={diasVencimiento}
              onChange={(e) => {
                setDiasVencimiento(Number(e.target.value));
                setPagina(1);
              }}
            >
              <option value={7}>7 días</option>
              <option value={15}>15 días</option>
              <option value={30}>30 días</option>
              <option value={60}>60 días</option>
              <option value={90}>90 días</option>
            </select>
          </div>
        </div>
      </div>

      <div className="acciones mb-3">
        <button
          type="button"
          className={`btn ${vistaActiva === "minimo" ? "btn-primary" : "btn-outline-primary"}`}
          onClick={() => {
            setVistaActiva("minimo");
            setPagina(1);
          }}
        >
          Stock bajo el mínimo
        </button>

        <button
          type="button"
          className={`btn ${vistaActiva === "vencimiento" ? "btn-primary" : "btn-outline-primary"}`}
          onClick={() => {
            setVistaActiva("vencimiento");
            setPagina(1);
          }}
        >
          Próximos a vencer
        </button>
      </div>

      {cargando ? (
        <div className="d-flex align-items-center gap-2 text-muted my-3">
          <span
            className="spinner-border spinner-border-sm"
            aria-hidden="true"
          />
          <p className="mb-0">Cargando alertas...</p>
        </div>
      ) : reporte ? (
        <>
          {vistaActiva === "minimo" ? (
            <section>
              <h2>Productos con stock bajo el mínimo</h2>

              <div className="tabla-wrapper table-responsive">
                <table className="table table-hover align-middle mb-0">
                  <thead>
                    <tr>
                      <th>Depósito</th>
                      <th>Producto</th>
                      <th>Stock actual</th>
                      <th>Stock mínimo</th>
                    </tr>
                  </thead>

                  <tbody>
                    {reporte.bajo_minimo.length === 0 ? (
                      <tr>
                        <td colSpan={4} className="text-center text-muted py-4">
                          No hay productos con stock bajo el mínimo.
                        </td>
                      </tr>
                    ) : (
                      reporte.bajo_minimo.map((item) => (
                        <tr key={`${item.deposito_id}-${item.producto_id}`}>
                          <td>{item.deposito_nombre}</td>
                          <td className="fw-medium">{item.producto_nombre}</td>
                          <td>{item.stock.toLocaleString("es-AR")}</td>
                          <td>{item.stock_minimo.toLocaleString("es-AR")}</td>
                        </tr>
                      ))
                    )}
                  </tbody>
                </table>
              </div>

              <div className="paginacion">
                <span>
                  Página {reporte.paginacion_minimo.pagina} de{" "}
                  {totalPaginasMinimo}
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
                    disabled={pagina >= totalPaginasMinimo || cargando}
                    onClick={() => setPagina((actual) => actual + 1)}
                  >
                    Siguiente
                  </button>
                </div>
              </div>
            </section>
          ) : (
            <section>
              <h2>Productos próximos a vencer</h2>

              <div className="tabla-wrapper table-responsive">
                <table className="table table-hover align-middle mb-0">
                  <thead>
                    <tr>
                      <th>Depósito</th>
                      <th>Producto</th>
                      <th>Stock</th>
                      <th>Fecha de vencimiento</th>
                      <th>Días restantes</th>
                    </tr>
                  </thead>

                  <tbody>
                    {reporte.por_vencer.length === 0 ? (
                      <tr>
                        <td colSpan={5} className="text-center text-muted py-4">
                          No hay productos próximos a vencer.
                        </td>
                      </tr>
                    ) : (
                      reporte.por_vencer.map((item) => (
                        <tr
                          key={`${item.deposito_id}-${item.producto_id}-${item.fecha_vencimiento}`}
                        >
                          <td>{item.deposito_nombre}</td>
                          <td className="fw-medium">{item.producto_nombre}</td>
                          <td>{item.stock.toLocaleString("es-AR")}</td>
                          <td>
                            {new Date(
                              `${item.fecha_vencimiento}T00:00:00`,
                            ).toLocaleDateString("es-AR")}
                          </td>
                          <td>{item.dias_para_vencer}</td>
                        </tr>
                      ))
                    )}
                  </tbody>
                </table>
              </div>

              <div className="paginacion">
                <span>
                  Página {reporte.paginacion_vencimiento.pagina} de{" "}
                  {totalPaginasVencimiento}
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
                    disabled={pagina >= totalPaginasVencimiento || cargando}
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
