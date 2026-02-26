import { useState, useCallback, useEffect } from "react";
import {
  getRoutes,
  createRoute,
  updateRoute,
  deleteRoute,
  pauseRoute,
  resumeRoute,
} from "@/lib/api";
import type { Route, CreateRouteRequest, UpdateRouteRequest } from "@/lib/types";

export function useRoutesModel() {
  const [routes, setRoutes] = useState<Route[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const loadRoutes = useCallback(async () => {
    try {
      const data = await getRoutes();
      setRoutes(data);
      setError("");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load routes");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadRoutes();
  }, [loadRoutes]);

  const create = useCallback(
    async (req: CreateRouteRequest) => {
      await createRoute(req);
      await loadRoutes();
    },
    [loadRoutes]
  );

  const update = useCallback(
    async (id: string, req: UpdateRouteRequest) => {
      await updateRoute(id, req);
      await loadRoutes();
    },
    [loadRoutes]
  );

  const remove = useCallback(
    async (id: string) => {
      await deleteRoute(id);
      await loadRoutes();
    },
    [loadRoutes]
  );

  const pause = useCallback(
    async (id: string) => {
      await pauseRoute(id);
      await loadRoutes();
    },
    [loadRoutes]
  );

  const resume = useCallback(
    async (id: string) => {
      await resumeRoute(id);
      await loadRoutes();
    },
    [loadRoutes]
  );

  return { routes, loading, error, setError, loadRoutes, create, update, remove, pause, resume };
}
