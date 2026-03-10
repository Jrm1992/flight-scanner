"use client";

import { createContext, useContext } from "react";

type CardVariant = "default" | "highlight" | "muted";

const CardContext = createContext<{ variant: CardVariant }>({ variant: "default" });

const variantStyles: Record<CardVariant, string> = {
  default:
    "bg-white/5 backdrop-blur-xl border-white/10",
  highlight:
    "bg-emerald-500/5 backdrop-blur-xl border-emerald-500/20",
  muted:
    "bg-white/[0.02] backdrop-blur-xl border-white/5",
};

interface CardProps {
  variant?: CardVariant;
  children: React.ReactNode;
  className?: string;
}

function Card({ variant = "default", children, className = "" }: CardProps) {
  return (
    <CardContext.Provider value={{ variant }}>
      <div
        className={`rounded-[var(--radius-lg)] border shadow-[var(--shadow-sm)] hover:shadow-[var(--shadow-md)] hover:border-white/15 transition-all duration-[var(--transition-base)] ${variantStyles[variant]} ${className}`}
      >
        {children}
      </div>
    </CardContext.Provider>
  );
}

function CardHeader({
  children,
  action,
  className = "",
}: {
  children: React.ReactNode;
  action?: React.ReactNode;
  className?: string;
}) {
  return (
    <div
      className={`flex items-center justify-between px-5 pt-5 pb-0 ${className}`}
    >
      <div className="flex-1 min-w-0">{children}</div>
      {action && <div className="ml-4 flex-shrink-0">{action}</div>}
    </div>
  );
}

function CardBody({
  children,
  className = "",
}: {
  children: React.ReactNode;
  className?: string;
}) {
  return <div className={`px-5 py-4 ${className}`}>{children}</div>;
}

function CardFooter({
  children,
  className = "",
}: {
  children: React.ReactNode;
  className?: string;
}) {
  return (
    <div
      className={`px-5 pb-4 pt-0 flex items-center gap-2 ${className}`}
    >
      {children}
    </div>
  );
}

Card.Header = CardHeader;
Card.Body = CardBody;
Card.Footer = CardFooter;

export default Card;
export { useContext as useCardContext, CardContext };
