import { useCallback, useEffect, useState } from "react";
import type { FormEvent } from "react";

import { useAuth } from "../auth/AuthContext";
import { ErrorAPI } from "../api/respuestas";
import { listarDepositos } from "../api/depositos";
import {
  enviarNovedadADeposito,
  enviarNovedadDifusion,
  listarNovedadesDeposito,
  obtenerNovedad,
  marcarNovedadLeida,
  responderNovedad,
} from "../api/novedadesDepositos";

import type {
  DepositoDTO,
  NovedadDTO,
  NovedadesPaginadoDTO,
} from "../api/types";

export function NovedadesPage() {
  const { token, usuario } = useAuth();

  const rol = usuario?.rol;
  const depositoAsociado = usuario?.deposito_asociado;

  const esAdministrador = rol === "administrador";
  const esGerente = rol === "gerente";
  const esAuditor = rol === "auditor";
  const esOperario = rol === "operario";
  const puedeVerTodosDepositos = esAdministrador || esAuditor;

  const [depositos, setDepositos] = useState<DepositoDTO[]>([]);
  const [depositoSeleccionado, setDepositoSeleccionado] = useState("");

  const [novedades, setNovedades] = useState<NovedadesPaginadoDTO | null>(
    null,
  );
  const [novedadSeleccionada, setNovedadSeleccionada] =
    useState<NovedadDTO | null>(null);

  const [texto, setTexto] = useState("");
  const [respuesta, setRespuesta] = useState("");

  const [depositosSeleccionados, setDepositosSeleccionados] = useState<
    string[]
  >([]);
  const [difundirATodos, setDifundirATodos] = useState(false);

  const [soloNoLeidas, setSoloNoLeidas] = useState(false);
  const [pagina, setPagina] = useState(1);

  const [cargando, setCargando] = useState(false);
  const [error, setError] = useState("");
  const [mensaje, setMensaje] = useState("");

  const tamanioPagina = 10;

  // Depósito utilizado para consultar las novedades.
  const depositoId = puedeVerTodosDepositos
    ? depositoSeleccionado
    : depositoAsociado ?? "";

  // Cargar depósitos disponibles para el administrador y el auditor.
  useEffect(() => {
    if (!token || !puedeVerTodosDepositos) return;

    async function cargarDepositos() {
      try {
        const resultado = await listarDepositos(token!);
        setDepositos(resultado);

        if (resultado.length > 0) {
          setDepositoSeleccionado(resultado[0].id);
        }
      } catch (error) {
        setError(
          error instanceof ErrorAPI
            ? error.message
            : "Error al cargar los depósitos.",
        );
      }
    }

    cargarDepositos();
  }, [token, puedeVerTodosDepositos]);

  // Consultar novedades del depósito seleccionado.
  const cargarNovedades = useCallback(async () => {
    if (!token || !depositoId) return;

    try {
      setCargando(true);
      setError("");

      const resultado = await listarNovedadesDeposito(token, depositoId, {
        solo_no_leidas: soloNoLeidas,
        pagina,
        tamanioPagina,
      });

      setNovedades(resultado);
    } catch (error) {
      setError(
        error instanceof ErrorAPI
          ? error.message
          : "Error al cargar las novedades.",
      );
    } finally {
      setCargando(false);
    }
  }, [token, depositoId, soloNoLeidas, pagina]);

  useEffect(() => {
    cargarNovedades();
  }, [cargarNovedades]);

  // Enviar una novedad a un depósito (gerente).
  async function handleEnviarADeposito(evento: FormEvent<HTMLFormElement>) {
    evento.preventDefault();

    if (!token || !depositoAsociado || !texto.trim()) return;

    try {
      setError("");
      setMensaje("");

      await enviarNovedadADeposito(token, depositoAsociado, {
        texto: texto.trim(),
      });

      setTexto("");
      setMensaje("La novedad fue enviada correctamente.");

      await cargarNovedades();
    } catch (error) {
      setError(
        error instanceof ErrorAPI
          ? error.message
          : "Error al enviar la novedad.",
      );
    }
  }

  // Difundir una novedad (administrador / auditor).
  async function handleEnviarDifusion(evento: FormEvent<HTMLFormElement>) {
    evento.preventDefault();

    if (!token || !texto.trim()) return;

    if (!difundirATodos && depositosSeleccionados.length === 0) {
      setError("Seleccioná al menos un depósito.");
      return;
    }

    try {
      setError("");
      setMensaje("");

      await enviarNovedadDifusion(token, {
        texto: texto.trim(),
        depositos_ids: difundirATodos ? [] : depositosSeleccionados,
        todos: difundirATodos,
      });

      setTexto("");
      setDepositosSeleccionados([]);
      setDifundirATodos(false);

      setMensaje("La difusión fue enviada correctamente.");
      await cargarNovedades();
    } catch (error) {
      setError(
        error instanceof ErrorAPI
          ? error.message
          : "Error al enviar la difusión.",
      );
    }
  }

  // Abrir el detalle de una novedad.
  async function handleVerDetalle(id: string) {
    if (!token || !depositoId) return;

    try {
      setError("");

      const resultado = await obtenerNovedad(token, depositoId, id);
      setNovedadSeleccionada(resultado);
    } catch (error) {
      setError(
        error instanceof ErrorAPI
          ? error.message
          : "Error al obtener la novedad.",
      );
    }
  }

  // Marcar una novedad como leída (operario).
  async function handleMarcarLeida() {
    if (!token || !depositoId || !novedadSeleccionada) return;

    try {
      setError("");

      const resultado = await marcarNovedadLeida(
        token,
        depositoId,
        novedadSeleccionada.id,
      );

      setNovedadSeleccionada(resultado);
      await cargarNovedades();
    } catch (error) {
      setError(
        error instanceof ErrorAPI
          ? error.message
          : "Error al marcar la novedad como leída.",
      );
    }
  }

  // Responder una novedad (operario).
  async function handleResponder(evento: FormEvent<HTMLFormElement>) {
    evento.preventDefault();

    if (!token || !depositoId || !novedadSeleccionada || !respuesta.trim()) {
      return;
    }

    try {
      setError("");

      const resultado = await responderNovedad(
        token,
        depositoId,
        novedadSeleccionada.id,
        { texto: respuesta.trim() },
      );

      setNovedadSeleccionada(resultado);
      setRespuesta("");
      setMensaje("La respuesta fue enviada correctamente.");

      await cargarNovedades();
    } catch (error) {
      setError(
        error instanceof ErrorAPI
          ? error.message
          : "Error al enviar la respuesta.",
      );
    }
  }

  function handleCambiarDeposito(id: string) {
    setDepositoSeleccionado(id);
    setNovedadSeleccionada(null);
    setPagina(1);
    setNovedades(null);
  }

  function handleCambiarFiltro(soloNoLeidas: boolean) {
    setSoloNoLeidas(soloNoLeidas);
    setPagina(1);
    setNovedadSeleccionada(null);
  }

  function handleSeleccionarDeposito(id: string, seleccionado: boolean) {
    setDepositosSeleccionados((actuales) =>
      seleccionado
        ? [...actuales, id]
        : actuales.filter((depositoId) => depositoId !== id),
    );
  }

  function formatearFecha(fecha: string) {
    return new Date(fecha).toLocaleString("es-AR");
  }

  if (!esAdministrador && !esGerente && !esAuditor && !esOperario) {
    return (
      <div className="panel">
        <h1>Novedades</h1>
        <p className="error mb-0">Tu rol no tiene acceso a este módulo.</p>
      </div>
    );
  }

  return (
    <div className="panel">
      <div className="d-flex flex-wrap align-items-center justify-content-between gap-2 mb-3">
        <h1 className="mb-0">Novedades a depósitos</h1>
        {novedades && (
          <div className="d-flex align-items-center gap-2">
            <span className="badge text-bg-light border">
              Total: {novedades.total}
            </span>
            <span
              className={`badge rounded-pill ${
                novedades.no_leidas > 0 ? "text-bg-warning" : "text-bg-success"
              }`}
            >
              No leídas: {novedades.no_leidas}
            </span>
          </div>
        )}
      </div>

      {error && (
        <p className="error" role="alert">
          {error}
        </p>
      )}
      {mensaje && <p className="mensaje">{mensaje}</p>}

      {/* Envío de novedades por parte del gerente. */}
      {esGerente && (
        <section className="filtros-barra mb-4">
          <h3>Enviar novedad al depósito</h3>

          {!depositoAsociado ? (
            <p className="text-muted mb-0">No tenés un depósito asociado.</p>
          ) : (
            <form onSubmit={handleEnviarADeposito}>
              <div>
                <label htmlFor="texto-novedad" className="form-label">
                  Mensaje
                </label>
                <textarea
                  id="texto-novedad"
                  className="form-control"
                  placeholder="Escribí una novedad para los operarios de tu depósito..."
                  value={texto}
                  onChange={(evento) => setTexto(evento.target.value)}
                  required
                  rows={3}
                />
              </div>

              <div className="acciones">
                <button
                  type="submit"
                  className="btn btn-primary"
                  disabled={!texto.trim()}
                >
                  Enviar novedad
                </button>
              </div>
            </form>
          )}
        </section>
      )}

      {/* Difusión de novedades por parte del administrador o auditor. */}
      {puedeVerTodosDepositos && (
        <section className="filtros-barra mb-4">
          <h3>Difundir novedad</h3>

          <form onSubmit={handleEnviarDifusion}>
            <div>
              <label htmlFor="texto-difusion" className="form-label">
                Mensaje
              </label>
              <textarea
                id="texto-difusion"
                className="form-control"
                placeholder="Escribí el mensaje a difundir..."
                value={texto}
                onChange={(evento) => setTexto(evento.target.value)}
                required
                rows={3}
              />
            </div>

            <div className="form-check">
              <label className="checkbox form-check-label">
                <input
                  type="checkbox"
                  className="form-check-input mt-0"
                  checked={difundirATodos}
                  onChange={(evento) => {
                    setDifundirATodos(evento.target.checked);
                    setDepositosSeleccionados([]);
                  }}
                />
                Enviar a todos los depósitos
              </label>
            </div>

            {!difundirATodos && (
              <fieldset className="border rounded p-3 bg-white">
                <legend className="fs-6 fw-semibold mb-2">
                  Seleccionar depósitos
                </legend>

                {depositos.length === 0 ? (
                  <p className="text-muted mb-0">No hay depósitos disponibles.</p>
                ) : (
                  <div className="row g-2">
                    {depositos.map((deposito) => (
                      <div key={deposito.id} className="col-12 col-sm-6 col-md-4">
                        <label className="checkbox form-check-label">
                          <input
                            type="checkbox"
                            className="form-check-input mt-0"
                            checked={depositosSeleccionados.includes(
                              deposito.id,
                            )}
                            onChange={(evento) =>
                              handleSeleccionarDeposito(
                                deposito.id,
                                evento.target.checked,
                              )
                            }
                          />
                          {deposito.nombre} - {deposito.localidad}
                        </label>
                      </div>
                    ))}
                  </div>
                )}
              </fieldset>
            )}

            <div className="acciones">
              <button
                type="submit"
                className="btn btn-primary"
                disabled={
                  !texto.trim() ||
                  (!difundirATodos && depositosSeleccionados.length === 0)
                }
              >
                Enviar difusión
              </button>
            </div>
          </form>
        </section>
      )}

      {/* Consulta de novedades del depósito. */}
      <section className="mt-3 pt-0 border-0">
        <h3>
          {esOperario
            ? "Novedades recibidas"
            : esGerente
              ? "Novedades del depósito"
              : "Consultar novedades"}
        </h3>

        {(puedeVerTodosDepositos || esOperario) && (
          <div className="filtros-barra">
            <div className="row g-3 align-items-end">
              {puedeVerTodosDepositos && (
                <div className="col-12 col-md-6">
                  <label htmlFor="deposito-consulta" className="form-label">
                    Depósito
                  </label>
                  <select
                    id="deposito-consulta"
                    className="form-select"
                    value={depositoSeleccionado}
                    onChange={(evento) =>
                      handleCambiarDeposito(evento.target.value)
                    }
                  >
                    {depositos.map((deposito) => (
                      <option key={deposito.id} value={deposito.id}>
                        {deposito.nombre} - {deposito.localidad}
                      </option>
                    ))}
                  </select>
                </div>
              )}

              {esOperario && (
                <div className="col-12 col-md-6">
                  <div className="form-check">
                    <label className="checkbox form-check-label">
                      <input
                        type="checkbox"
                        className="form-check-input mt-0"
                        checked={soloNoLeidas}
                        onChange={(evento) =>
                          handleCambiarFiltro(evento.target.checked)
                        }
                      />
                      Mostrar solamente las no leídas
                    </label>
                  </div>
                </div>
              )}
            </div>
          </div>
        )}

        {cargando ? (
          <div className="d-flex align-items-center gap-2 text-muted my-3">
            <span className="spinner-border spinner-border-sm" aria-hidden="true" />
            <p className="mb-0">Cargando novedades...</p>
          </div>
        ) : !depositoId ? (
          <p className="text-muted my-3">
            Seleccioná un depósito para consultar sus novedades.
          </p>
        ) : novedades && novedades.items.length === 0 ? (
          <p className="text-muted my-3">No hay novedades para mostrar.</p>
        ) : (
          novedades && (
            <>
              <div className="tabla-wrapper table-responsive">
                <table className="table table-hover align-middle mb-0">
                  <thead>
                    <tr>
                      <th>Fecha</th>
                      <th>Mensaje</th>
                      <th>Estado</th>
                      <th className="text-end">Acciones</th>
                    </tr>
                  </thead>

                  <tbody>
                    {novedades.items.map((novedad) => (
                      <tr key={novedad.id}>
                        <td className="text-nowrap small">
                          {formatearFecha(novedad.fecha_envio)}
                        </td>
                        <td className={novedad.leido ? "" : "fw-semibold"}>
                          {novedad.texto}
                        </td>
                        <td>
                          <span
                            className={`badge rounded-pill ${
                              novedad.leido
                                ? "text-bg-success"
                                : "text-bg-warning"
                            }`}
                          >
                            {novedad.leido ? "Leída" : "No leída"}
                          </span>
                        </td>
                        <td>
                          <div className="acciones justify-content-end">
                            <button
                              type="button"
                              className="btn btn-outline-primary btn-sm"
                              onClick={() => handleVerDetalle(novedad.id)}
                            >
                              Ver detalle
                            </button>
                          </div>
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
                  onClick={() => setPagina((actual) => actual - 1)}
                >
                  Anterior
                </button>

                <span className="text-muted small">
                  Página {novedades.pagina} de{" "}
                  {Math.max(1, Math.ceil(novedades.total / tamanioPagina))}
                </span>

                <button
                  type="button"
                  className="btn btn-outline-secondary btn-sm"
                  disabled={
                    pagina >= Math.ceil(novedades.total / tamanioPagina)
                  }
                  onClick={() => setPagina((actual) => actual + 1)}
                >
                  Siguiente
                </button>
              </div>
            </>
          )
        )}
      </section>

      {/* Detalle de la novedad seleccionada. */}
      {novedadSeleccionada && (
        <section className="card border-primary mt-4 p-3 p-md-4 bg-light bg-opacity-50">
          <div className="d-flex flex-wrap align-items-center justify-content-between gap-2 mb-3">
            <h3 className="mb-0">Detalle de la novedad</h3>
            <span
              className={`badge rounded-pill ${
                novedadSeleccionada.leido
                  ? "text-bg-success"
                  : "text-bg-warning"
              }`}
            >
              {novedadSeleccionada.leido ? "Leída" : "No leída"}
            </span>
          </div>

          <p className="text-muted small mb-2">
            <strong>Fecha de envío:</strong>{" "}
            {formatearFecha(novedadSeleccionada.fecha_envio)}
          </p>

          <div className="p-3 bg-white border rounded mb-3">
            <p className="mb-0">{novedadSeleccionada.texto}</p>
          </div>

          {novedadSeleccionada.respuestas &&
            novedadSeleccionada.respuestas.length > 0 && (
              <div className="mb-3">
                <h4 className="fs-6 fw-semibold mb-2">Respuestas</h4>

                <div className="d-flex flex-column gap-2">
                  {novedadSeleccionada.respuestas.map((item, indice) => (
                    <div
                      key={`${item.usuario_id}-${indice}`}
                      className="p-2 px-3 bg-white border rounded"
                    >
                      <p className="mb-1">{item.texto}</p>
                      <small className="text-muted">
                        {formatearFecha(item.fecha_respuesta)}
                      </small>
                    </div>
                  ))}
                </div>
              </div>
            )}

          {esOperario && (
            <div className="mb-3">
              {!novedadSeleccionada.leido && (
                <div className="mb-3">
                  <button
                    type="button"
                    className="btn btn-success btn-sm"
                    onClick={handleMarcarLeida}
                  >
                    ✓ Marcar como leída
                  </button>
                </div>
              )}

              <form onSubmit={handleResponder}>
                <div>
                  <label htmlFor="respuesta-novedad" className="form-label">
                    Responder
                  </label>
                  <textarea
                    id="respuesta-novedad"
                    className="form-control"
                    placeholder="Escribí tu respuesta..."
                    value={respuesta}
                    onChange={(evento) => setRespuesta(evento.target.value)}
                    required
                    rows={3}
                  />
                </div>

                <div className="acciones">
                  <button
                    type="submit"
                    className="btn btn-primary btn-sm"
                    disabled={!respuesta.trim()}
                  >
                    Enviar respuesta
                  </button>
                </div>
              </form>
            </div>
          )}

          <div className="acciones">
            <button
              type="button"
              className="btn btn-outline-secondary btn-sm"
              onClick={() => setNovedadSeleccionada(null)}
            >
              Cerrar detalle
            </button>
          </div>
        </section>
      )}
    </div>
  );
}
