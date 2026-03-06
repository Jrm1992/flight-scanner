export interface User {
  id: string;
  email: string;
  name: string;
  created_at: string;
}

export interface AuthResponse {
  token: string;
  user: User;
}

export interface RegisterRequest {
  email: string;
  password: string;
  name: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface Route {
  id: string;
  origin: string;
  destination: string;
  departure_date: string;
  return_date?: string;
  alert_price: number;
  check_frequency_minutes: number;
  status: "active" | "paused";
  created_at: string;
  updated_at: string;
  current_price?: number;
  last_check_at?: string;
  price_trend?: string;
}

export interface CreateRouteRequest {
  origin: string;
  destination: string;
  departure_date: string;
  return_date?: string;
  alert_price: number;
  check_frequency_minutes: number;
}

export interface UpdateRouteRequest {
  alert_price?: number;
  check_frequency_minutes?: number;
}

export interface FlightResult {
  price: number;
  airline: string;
  flight_number: string;
  departure_code: string;
  arrival_code: string;
  departure: string;
  arrival: string;
  duration_minutes: number;
  stops: number;
}

export interface SearchResponse {
  origin: string;
  destination: string;
  date: string;
  currency: string;
  results: FlightResult[];
  count: number;
}

export interface PriceHistory {
  id: string;
  route_id: string;
  min_price: number;
  max_price: number;
  avg_price: number;
  airline: string;
  checked_at: string;
}

export interface PriceStats {
  min_price: number;
  max_price: number;
  avg_price: number;
  since: string;
}

export interface HistoryResponse {
  route_id: string;
  days: number;
  history: PriceHistory[];
  stats: PriceStats;
  count: number;
}

export interface Alert {
  id: string;
  route_id: string;
  alert_price: number;
  triggered_price: number;
  triggered_at: string;
  notified: boolean;
  notified_at?: string;
}
