import type { FC, ReactNode } from "react";
import { createContext, useContext } from "react";
import { useServer } from "./useServer";

interface ContextProps {
  isAuthenticated: boolean;
  isAuthLoading: boolean;
}

const initialState: ContextProps = {
  isAuthenticated: false,
  isAuthLoading: true,
};

export const AuthContext = createContext<ContextProps>(initialState);

interface ProviderProps {
  children: ReactNode;
}

export const AuthProvider: FC<ProviderProps> = ({ children }) => {
  const { data: server, isLoading } = useServer();

  const isAuthenticated = !isLoading && !!server?.trusted;

  return (
    <AuthContext.Provider
      value={{
        isAuthenticated,
        isAuthLoading: isLoading,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
};

export function useAuth() {
  return useContext(AuthContext);
}
