
import {
  createContext,
  useContext,
  useEffect,
  useState,
  type ReactNode,
} from "react";

import * as api from "../api/usuarios";
import type { UsuarioDTO, Rol } from "../api/types";

const CLAVE_TOKEN = "gestock_token";

interface AuthContextValue {
  usuario: UsuarioDTO | null;
  token: string | null;
  cargando: boolean;
  login: (correo: string, contrasenia: string) => Promise<void>;
  registrar: (
    correo: string,
    contrasenia: string,
    rol: Rol,
    deposito_asociado?: string,
  ) => Promise<UsuarioDTO>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({
  children,
}: {
  children: ReactNode;
}) {
  const [usuario, setUsuario] = useState<UsuarioDTO | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [cargando, setCargando] = useState(true);

  // Recuperar la sesión guardada al iniciar la aplicación
  useEffect(() => {
    const tokenGuardado = localStorage.getItem(CLAVE_TOKEN);

    if (!tokenGuardado) {
      setCargando(false);
      return;
    }

    api
      .obtenerPerfil(tokenGuardado)
      .then((usuarioObtenido) => {
        setUsuario(usuarioObtenido);
        setToken(tokenGuardado);
      })
      .catch(() => {
        localStorage.removeItem(CLAVE_TOKEN);
        setUsuario(null);
        setToken(null);
      })
      .finally(() => {
        setCargando(false);
      });
  }, []);

  // Iniciar sesión
  async function login(
    correo: string,
    contrasenia: string,
  ): Promise<void> {
    const respuesta = await api.ingresar({
      correo,
      contrasenia,
    });

    const nuevoToken = respuesta.token;

    localStorage.setItem(CLAVE_TOKEN, nuevoToken);
    setToken(nuevoToken);

    try {
      const usuarioObtenido = await api.obtenerPerfil(nuevoToken);
      setUsuario(usuarioObtenido);
    } catch (error) {
      localStorage.removeItem(CLAVE_TOKEN);
      setToken(null);
      setUsuario(null);
      throw error;
    }
  }

  // Registrar un usuario (requiere token de administrador)
  async function registrar(
    correo: string,
    contrasenia: string,
    rol: Rol,
    deposito_asociado?: string,
  ): Promise<UsuarioDTO> {
    if (!token) {
      throw new Error(
        "Se necesita iniciar sesión para registrar un usuario",
      );
    }

    const dto = {
      correo,
      contrasenia,
      rol,
      ...(deposito_asociado ? { deposito_asociado } : {}),
    };

    return api.registrar(dto, token);
  }

  // Cerrar sesión
  function logout(): void {
    localStorage.removeItem(CLAVE_TOKEN);
    setToken(null);
    setUsuario(null);
  }

  return (
    <AuthContext.Provider
      value={{
        usuario,
        token,
        cargando,
        login,
        registrar,
        logout,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const contexto = useContext(AuthContext);

  if (!contexto) {
    throw new Error(
      "useAuth debe usarse dentro de un AuthProvider",
    );
  }

  return contexto;
}
