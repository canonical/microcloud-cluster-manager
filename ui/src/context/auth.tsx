import type { FC, ReactNode } from "react";
import { createContext, useContext } from "react";
import { useServer } from "./useServer";
import { hasGroup } from "util/access";

interface ContextProps {
  isAuthenticated: boolean;
  isAuthLoading: boolean;
  isAdmin: boolean;
}

const initialState: ContextProps = {
  isAuthenticated: false,
  isAuthLoading: true,
  isAdmin: false,
};

export const AuthContext = createContext<ContextProps>(initialState);

interface ProviderProps {
  children: ReactNode;
}

export const AuthProvider: FC<ProviderProps> = ({ children }) => {
  const { data: server, isLoading } = useServer();

  const isAuthenticated = !isLoading && !!server?.trusted;
  const isAdmin = hasGroup(server?.groups || [], "admins");

  return (
    <AuthContext.Provider
      value={{
        isAuthenticated,
        isAuthLoading: isLoading,
        isAdmin,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
};

export function useAuth() {
  return useContext(AuthContext);
}
