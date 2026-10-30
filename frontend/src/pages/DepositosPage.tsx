import { useCallback, useEffect, useState, type FormEvent } from "react";
import { useAuth } from "../auth/AuthContext";
import {
  listarDepositos,
  crearDeposito,
  actualizarDeposito,
  eliminarDeposito,
} from "../api/depositos";
import { ErrorAPI } from "../api/respuestas";
import type { DepositoDTO } from "../api/types";

interface FormularioDeposito {
  nombre: string;
  localidad: string;
  provincia: string;
}

const FORMULARIO_VACIO: FormularioDeposito = {
  nombre: "",
  localidad: "",
  provincia: "",
};

export function DepositosPage() {
  const { token, usuario } = useAuth();

  const [depositos, setDepositos] = useState<DepositoDTO[]>([]);
  const [cargando, setCargando] = useState(true);
  const [errorLista, setErrorLista] = useState<string | null>(null);

  const [formulario, setFormulario] =
    useState<FormularioDeposito>(FORMULARIO_VACIO);
  const [editandoId, setEditandoId] = useState<string | null>(null);

  const [enviando, setEnviando] = useState(false);
  const [errorFormulario, setErrorFormulario] = useState<string | null>(null);

  const esAdministrador = usuario?.rol === "administrador";

  // Cargar los depósitos desde la API.
  const cargarDepositos = useCallback(async () => {
    if (!token) {
      setCargando(false);
      return;
    }

    setCargando(true);
    setErrorLista(null);

    try {
      const datos = await listarDepositos(token);
      setDepositos(datos);
    } catch (e) {
      setErrorLista(
        e instanceof ErrorAPI
          ? e.message
          : "No se pudieron cargar los depósitos",
      );
    } finally {
      setCargando(false);
    }
  }, [token]);

  useEffect(() => {
    cargarDepositos();
  }, [cargarDepositos]);

  // Preparar el formulario para crear un depósito.
  function empezarCreacion() {
    setFormulario(FORMULARIO_VACIO);
    setEditandoId(null);
    setErrorFormulario(null);
  }

  // Cargar los datos del depósito seleccionado.
  function empezarEdicion(deposito: DepositoDTO) {
    setFormulario({
      nombre: deposito.nombre,
      localidad: deposito.localidad,
      provincia: deposito.provincia,
    });
    setEditandoId(deposito.id);
    setErrorFormulario(null);
  }

  // Limpiar el formulario.
  function cancelarEdicion() {
    setFormulario(FORMULARIO_VACIO);
    setEditandoId(null);
    setErrorFormulario(null);
  }

  // Crear o actualizar el depósito.
  async function manejarEnvio(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();

    if (!token) return;

    setErrorFormulario(null);
    setEnviando(true);

    try {
      const dto = {
        nombre: formulario.nombre.trim(),
        localidad: formulario.localidad.trim(),
        provincia: formulario.provincia.trim(),
      };

      if (editandoId) {
        await actualizarDeposito(editandoId, dto, token);
      } else {
        await crearDeposito(dto, token);
      }

      cancelarEdicion();
      await cargarDepositos();
    } catch (e) {
      setErrorFormulario(
        e instanceof ErrorAPI
          ? e.message
          : "No se pudo guardar el depósito",
      );
    } finally {
      setEnviando(false);
    }
  }

  // Eliminar el depósito después de confirmar la operación.
  async function manejarEliminar(id: string, nombre: string) {
    if (!token) return;

    if (!window.confirm(`¿Querés eliminar el depósito "${nombre}"?`)) {
      return;
    }

    setErrorLista(null);

    try {
      await eliminarDeposito(id, token);

      if (editandoId === id) {
        cancelarEdicion();
      }

      await cargarDepositos();
    } catch (e) {
      setErrorLista(
        e instanceof ErrorAPI
          ? e.message
          : "No se pudo eliminar el depósito",
      );
    }
  }

  return (
    <div className="panel">
      <div className="d-flex flex-wrap align-items-center justify-content-between gap-2 mb-3">
        <h1 className="mb-0">Depósitos</h1>
        {esAdministrador && editandoId && (
          <div className="acciones">
            <button
              type="button"
              className="btn btn-outline-primary btn-sm"
              onClick={empezarCreacion}
            >
              + Nuevo depósito
            </button>
          </div>
        )}
      </div>

      {esAdministrador && (
        <form onSubmit={manejarEnvio} className="filtros-barra mb-4">
          <h2 className="mb-2">
            {editandoId ? "Editar depósito" : "Registrar depósito"}
          </h2>

          <div className="row g-3">
            <div className="col-12 col-md-4">
              <label htmlFor="nombre" className="form-label">
                Nombre
              </label>
              <input
                id="nombre"
                type="text"
                className="form-control"
                placeholder="Ej.: Depósito Central"
                value={formulario.nombre}
                onChange={(e) =>
                  setFormulario({
                    ...formulario,
                    nombre: e.target.value,
                  })
                }
                required
              />
            </div>

            <div className="col-12 col-md-4">
              <label htmlFor="localidad" className="form-label">
                Localidad
              </label>
              <input
                id="localidad"
                type="text"
                className="form-control"
                placeholder="Ej.: Rafaela"
                value={formulario.localidad}
                onChange={(e) =>
                  setFormulario({
                    ...formulario,
                    localidad: e.target.value,
                  })
                }
                required
              />
            </div>

            <div className="col-12 col-md-4">
              <label htmlFor="provincia" className="form-label">
                Provincia
              </label>
              <input
                id="provincia"
                type="text"
                className="form-control"
                placeholder="Ej.: Santa Fe"
                value={formulario.provincia}
                onChange={(e) =>
                  setFormulario({
                    ...formulario,
                    provincia: e.target.value,
                  })
                }
                required
              />
            </div>
          </div>

          {errorFormulario && (
            <p className="error mb-0" role="alert">
              {errorFormulario}
            </p>
          )}

          <div className="acciones mt-1">
            <button
              type="submit"
              className="btn btn-primary"
              disabled={enviando}
            >
              {enviando ? "Guardando..." : "Guardar"}
            </button>

            {editandoId && (
              <button
                type="button"
                className="btn btn-outline-secondary"
                onClick={cancelarEdicion}
                disabled={enviando}
              >
                Cancelar edición
              </button>
            )}
          </div>
        </form>
      )}

      {cargando && (
        <div className="d-flex align-items-center gap-2 text-muted my-3">
          <span className="spinner-border spinner-border-sm" aria-hidden="true" />
          <p className="mb-0">Cargando depósitos...</p>
        </div>
      )}

      {errorLista && (
        <p className="error" role="alert">
          {errorLista}
        </p>
      )}

      {!cargando && !errorLista && depositos.length === 0 && (
        <p className="text-muted mb-0">No hay depósitos registrados.</p>
      )}

      {!errorLista && depositos.length > 0 && (
        <div className="tabla-wrapper table-responsive">
          <table className="table table-hover align-middle mb-0">
            <thead>
              <tr>
                <th>Nombre</th>
                <th>Localidad</th>
                <th>Provincia</th>
                {esAdministrador && <th className="text-end">Acciones</th>}
              </tr>
            </thead>

            <tbody>
              {depositos.map((deposito) => (
                <tr key={deposito.id}>
                  <td className="fw-medium">{deposito.nombre}</td>
                  <td>{deposito.localidad}</td>
                  <td>{deposito.provincia}</td>

                  {esAdministrador && (
                    <td>
                      <div className="acciones justify-content-end">
                        <button
                          type="button"
                          className="btn btn-outline-primary btn-sm"
                          onClick={() => empezarEdicion(deposito)}
                        >
                          Editar
                        </button>

                        <button
                          type="button"
                          className="btn btn-outline-danger btn-sm"
                          onClick={() =>
                            manejarEliminar(deposito.id, deposito.nombre)
                          }
                        >
                          Eliminar
                        </button>
                      </div>
                    </td>
                  )}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
