import { motion } from "framer-motion";
import Card from "@/components/ui/Card";
import LoginForm from "./LoginForm";
import RegisterForm from "./RegisterForm";

interface AuthViewProps {
  mode: "login" | "register";
  // Login
  loginEmail: string;
  onLoginEmailChange: (v: string) => void;
  loginPassword: string;
  onLoginPasswordChange: (v: string) => void;
  loginLoading: boolean;
  loginError: string;
  onLoginSubmit: (e: React.FormEvent) => void;
  // Register
  registerName: string;
  onRegisterNameChange: (v: string) => void;
  registerEmail: string;
  onRegisterEmailChange: (v: string) => void;
  registerPassword: string;
  onRegisterPasswordChange: (v: string) => void;
  registerLoading: boolean;
  registerError: string;
  onRegisterSubmit: (e: React.FormEvent) => void;
  // Switch
  onSwitchToLogin: () => void;
  onSwitchToRegister: () => void;
}

export default function AuthView({
  mode,
  loginEmail,
  onLoginEmailChange,
  loginPassword,
  onLoginPasswordChange,
  loginLoading,
  loginError,
  onLoginSubmit,
  registerName,
  onRegisterNameChange,
  registerEmail,
  onRegisterEmailChange,
  registerPassword,
  onRegisterPasswordChange,
  registerLoading,
  registerError,
  onRegisterSubmit,
  onSwitchToLogin,
  onSwitchToRegister,
}: AuthViewProps) {
  return (
    <div className="min-h-screen bg-[#0a0e1a] flex items-center justify-center px-4 relative overflow-hidden">
      <div className="absolute top-1/4 -left-1/4 w-96 h-96 bg-cyan-500/10 rounded-full blur-3xl" />
      <div className="absolute bottom-1/4 -right-1/4 w-96 h-96 bg-amber-500/10 rounded-full blur-3xl" />

      <motion.div
        className="w-full max-w-sm relative z-10"
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5, ease: "easeOut" }}
      >
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold tracking-tight bg-gradient-to-r from-cyan-400 to-cyan-200 bg-clip-text text-transparent">
            Flight Price Monitor
          </h1>
          <p className="text-sm text-muted mt-2">
            Track flight prices and get alerts when they drop
          </p>
        </div>

        <Card>
          <Card.Header>
            <h2 className="text-lg font-semibold text-foreground">
              {mode === "login" ? "Welcome back" : "Create an account"}
            </h2>
          </Card.Header>
          <Card.Body>
            {mode === "login" ? (
              <LoginForm
                email={loginEmail}
                onEmailChange={onLoginEmailChange}
                password={loginPassword}
                onPasswordChange={onLoginPasswordChange}
                loading={loginLoading}
                error={loginError}
                onSubmit={onLoginSubmit}
                onSwitchToRegister={onSwitchToRegister}
              />
            ) : (
              <RegisterForm
                name={registerName}
                onNameChange={onRegisterNameChange}
                email={registerEmail}
                onEmailChange={onRegisterEmailChange}
                password={registerPassword}
                onPasswordChange={onRegisterPasswordChange}
                loading={registerLoading}
                error={registerError}
                onSubmit={onRegisterSubmit}
                onSwitchToLogin={onSwitchToLogin}
              />
            )}
          </Card.Body>
        </Card>
      </motion.div>
    </div>
  );
}
