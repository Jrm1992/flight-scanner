"use client";

import { useAuthFormModel } from "./model";
import type { LoginRequest, RegisterRequest } from "@/lib/types";
import AuthView from "./view";

interface AuthProps {
  onLogin: (req: LoginRequest) => Promise<void>;
  onRegister: (req: RegisterRequest) => Promise<void>;
}

export default function Auth({ onLogin, onRegister }: AuthProps) {
  const model = useAuthFormModel(onLogin, onRegister);

  return (
    <AuthView
      mode={model.mode}
      loginEmail={model.loginEmail}
      onLoginEmailChange={model.setLoginEmail}
      loginPassword={model.loginPassword}
      onLoginPasswordChange={model.setLoginPassword}
      loginLoading={model.loginLoading}
      loginError={model.loginError}
      onLoginSubmit={model.handleLogin}
      registerName={model.registerName}
      onRegisterNameChange={model.setRegisterName}
      registerEmail={model.registerEmail}
      onRegisterEmailChange={model.setRegisterEmail}
      registerPassword={model.registerPassword}
      onRegisterPasswordChange={model.setRegisterPassword}
      registerLoading={model.registerLoading}
      registerError={model.registerError}
      onRegisterSubmit={model.handleRegister}
      onSwitchToLogin={model.switchToLogin}
      onSwitchToRegister={model.switchToRegister}
    />
  );
}
