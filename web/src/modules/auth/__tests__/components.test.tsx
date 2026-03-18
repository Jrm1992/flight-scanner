import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen, fireEvent, cleanup } from "@testing-library/react";
import LoginForm from "../LoginForm";
import RegisterForm from "../RegisterForm";

afterEach(cleanup);

describe("LoginForm", () => {
  it("renders email and password fields", () => {
    render(
      <LoginForm
        email="" onEmailChange={vi.fn()}
        password="" onPasswordChange={vi.fn()}
        loading={false} error="" onSubmit={vi.fn()}
        onSwitchToRegister={vi.fn()}
      />
    );
    expect(screen.getByPlaceholderText("you@example.com")).toBeDefined();
    expect(screen.getByPlaceholderText("At least 8 characters")).toBeDefined();
  });

  it("shows error message", () => {
    render(
      <LoginForm
        email="" onEmailChange={vi.fn()}
        password="" onPasswordChange={vi.fn()}
        loading={false} error="Invalid credentials" onSubmit={vi.fn()}
        onSwitchToRegister={vi.fn()}
      />
    );
    expect(screen.getByText("Invalid credentials")).toBeDefined();
  });

  it("calls onSwitchToRegister", () => {
    const onSwitch = vi.fn();
    render(
      <LoginForm
        email="" onEmailChange={vi.fn()}
        password="" onPasswordChange={vi.fn()}
        loading={false} error="" onSubmit={vi.fn()}
        onSwitchToRegister={onSwitch}
      />
    );
    fireEvent.click(screen.getByText("Sign Up"));
    expect(onSwitch).toHaveBeenCalled();
  });

  it("calls onEmailChange on input", () => {
    const onEmailChange = vi.fn();
    render(
      <LoginForm
        email="" onEmailChange={onEmailChange}
        password="" onPasswordChange={vi.fn()}
        loading={false} error="" onSubmit={vi.fn()}
        onSwitchToRegister={vi.fn()}
      />
    );
    fireEvent.change(screen.getByPlaceholderText("you@example.com"), {
      target: { value: "test@test.com" },
    });
    expect(onEmailChange).toHaveBeenCalledWith("test@test.com");
  });
});

describe("RegisterForm", () => {
  it("renders name, email and password fields", () => {
    render(
      <RegisterForm
        name="" onNameChange={vi.fn()}
        email="" onEmailChange={vi.fn()}
        password="" onPasswordChange={vi.fn()}
        loading={false} error="" onSubmit={vi.fn()}
        onSwitchToLogin={vi.fn()}
      />
    );
    expect(screen.getByPlaceholderText("Your name")).toBeDefined();
    expect(screen.getByPlaceholderText("you@example.com")).toBeDefined();
    expect(screen.getByPlaceholderText("At least 8 characters")).toBeDefined();
  });

  it("shows error message", () => {
    render(
      <RegisterForm
        name="" onNameChange={vi.fn()}
        email="" onEmailChange={vi.fn()}
        password="" onPasswordChange={vi.fn()}
        loading={false} error="Email taken" onSubmit={vi.fn()}
        onSwitchToLogin={vi.fn()}
      />
    );
    expect(screen.getByText("Email taken")).toBeDefined();
  });

  it("calls onSwitchToLogin", () => {
    const onSwitch = vi.fn();
    render(
      <RegisterForm
        name="" onNameChange={vi.fn()}
        email="" onEmailChange={vi.fn()}
        password="" onPasswordChange={vi.fn()}
        loading={false} error="" onSubmit={vi.fn()}
        onSwitchToLogin={onSwitch}
      />
    );
    fireEvent.click(screen.getByText("Sign In"));
    expect(onSwitch).toHaveBeenCalled();
  });
});
