
import { useEffect, useState } from "react";
import { useAuth } from "../auth/AuthContext";
import { ReporteInventario } from "../components/reportes/ReporteInventario";
import { ReporteMovimientos } from "../components/reportes/ReporteMovimientos";
import { ReporteAlertas } from "../components/reportes/ReporteAlertas";
import { ReportePendientes } from "../components/reportes/ReportePendientes";

type TipoReporte =
  | "inventario"
  | "movimientos"
  | "alertas"
  | "pendientes";

export function ReportesPage() {
  const { usuario } = useAuth();

  const [reporteActivo, setReporteActivo] =
    useState<TipoReporte>("inventario");

  const esOperario = usuario?.rol === "operario";

  const puedeVerReportes =
    usuario?.rol === "administrador" ||
    usuario?.rol === "gerente" ||
    usuario?.rol === "auditor" ||
    esOperario;

  // El operario solamente puede consultar el reporte de alertas.
  useEffect(() => {
    if (esOperario) {
      setReporteActivo("alertas");
    }
  }, [esOperario]);

  if (!puedeVerReportes) {
    return (
      <div className="panel">
        <p className="text-muted mb-0">
          No tenés permisos para consultar los reportes.
        </p>
      </div>
    );
  }

  return (
    <div className="panel">
      <h1>Reportes</h1>

      <div className="reportes-nav">
        {!esOperario && (
          <>
            <button
              type="button"
              className={`btn ${
                reporteActivo === "inventario"
                  ? "btn-primary"
                  : "btn-outline-primary"
              }`}
              onClick={() => setReporteActivo("inventario")}
            >
              Inventario
            </button>

            <button
              type="button"
              className={`btn ${
                reporteActivo === "movimientos"
                  ? "btn-primary"
                  : "btn-outline-primary"
              }`}
              onClick={() => setReporteActivo("movimientos")}
            >
              Movimientos
            </button>
          </>
        )}

        <button
          type="button"
          className={`btn ${
            reporteActivo === "alertas"
              ? "btn-primary"
              : "btn-outline-primary"
          }`}
          onClick={() => setReporteActivo("alertas")}
        >
          Alertas
        </button>

        {!esOperario && (
          <button
            type="button"
            className={`btn ${
              reporteActivo === "pendientes"
                ? "btn-primary"
                : "btn-outline-primary"
            }`}
            onClick={() => setReporteActivo("pendientes")}
          >
            Pendientes
          </button>
        )}
      </div>

      {!esOperario && (
        <>
          <div
            style={{
              display: reporteActivo === "inventario" ? "block" : "none",
            }}
          >
            <ReporteInventario activo={reporteActivo === "inventario"} />
          </div>

          <div
            style={{
              display: reporteActivo === "movimientos" ? "block" : "none",
            }}
          >
            <ReporteMovimientos activo={reporteActivo === "movimientos"} />
          </div>

          <div
            style={{
              display: reporteActivo === "pendientes" ? "block" : "none",
            }}
          >
            <ReportePendientes activo={reporteActivo === "pendientes"} />
          </div>
        </>
      )}

      <div
        style={{
          display: reporteActivo === "alertas" ? "block" : "none",
        }}
      >
        <ReporteAlertas activo={reporteActivo === "alertas"} />
      </div>
    </div>
  );
}
