import type { ReactNode } from "react";
import { Navigate } from "react-router-dom";
import { useAuth } from "../auth/AuthContext";

// Envuelve una página que exige sesión iniciada. "cargando" evita un falso
// negativo mientras AuthProvider todavía está resolviendo el token guardado
// contra GET /usuarios/me (ver Unidad 7, Clase 4).
export function RutaProtegida({ children }: { children: ReactNode }) {
  const { usuario, cargando } = useAuth();

  if (cargando) {
    return (
      <div className="pagina-centrada">
        <div className="d-flex align-items-center gap-2 text-secondary fw-medium">
          <span
            className="spinner-border spinner-border-sm text-primary"
            role="status"
            aria-hidden="true"
          />
          <span>Cargando sesión...</span>
        </div>
      </div>
    );
  }
  if (!usuario) return <Navigate to="/login" replace />;
  return children;
}
