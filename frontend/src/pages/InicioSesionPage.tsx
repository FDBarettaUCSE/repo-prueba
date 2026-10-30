import { useState, type FormEvent } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "../auth/AuthContext";
import { ErrorAPI } from "../api/respuestas";

export function IniciarSesionPage() {
  const { login } = useAuth();
  const navigate = useNavigate();

  const [correo, setCorreo] = useState("");
  const [contrasenia, setContrasenia] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [enviando, setEnviando] = useState(false);

  async function manejarEnvio(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setError(null);
    setEnviando(true);

    try {
      await login(correo, contrasenia);
      navigate("/");
    } catch (err) {
      setError(
        err instanceof ErrorAPI
          ? err.message
          : "No se pudo iniciar sesión",
      );
    } finally {
      setEnviando(false);
    }
  }

  return (
    <div className="pagina-centrada">
      <div className="tarjeta">
        <div className="text-center mb-4">
          <div className="fs-1 mb-2" aria-hidden="true">📦</div>
          <h1 className="mb-1">Iniciar sesión en GeStock</h1>
          <p className="text-muted small mb-0">
            Sistema de Gestión de Stock — Delta Insumos S.A.
          </p>
        </div>

        <form onSubmit={manejarEnvio}>
          <label className="form-label mb-0">
            Correo electrónico
            <input
              type="email"
              className="form-control"
              placeholder="usuario@deltainsumos.com"
              value={correo}
              onChange={(e) => setCorreo(e.target.value)}
              autoComplete="username"
              required
            />
          </label>

          <label className="form-label mb-0">
            Contraseña
            <input
              type="password"
              className="form-control"
              placeholder="••••••••"
              value={contrasenia}
              onChange={(e) => setContrasenia(e.target.value)}
              autoComplete="current-password"
              required
            />
          </label>

          {error && (
            <p className="error alert alert-danger py-2 mb-0" role="alert">
              {error}
            </p>
          )}

          <button
            type="submit"
            className="btn btn-primary w-100 py-2 fw-semibold mt-2"
            disabled={enviando}
          >
            {enviando ? "Ingresando..." : "Ingresar"}
          </button>
        </form>
      </div>
    </div>
  );
}
