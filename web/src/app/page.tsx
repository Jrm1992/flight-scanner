"use client";

import { useAuth } from "@/modules/auth/AuthContext";
import Auth from "@/modules/auth";
import { useDashboardModel } from "@/modules/dashboard/model";
import DashboardView from "@/modules/dashboard/view";

export default function Home() {
  const auth = useAuth();
  const model = useDashboardModel();

  if (!auth.isAuthenticated) {
    return <Auth onLogin={auth.login} onRegister={auth.register} />;
  }

  return (
    <DashboardView
      userName={auth.user?.name ?? ""}
      onLogout={auth.logout}
      tab={model.tab}
      onTabChange={model.setTab}
      chartRoute={model.chartRoute}
      onCloseHistory={() => model.setChartRoute(null)}
      onViewHistory={(r) => model.setChartRoute(r)}
      onMonitor={model.handleMonitor}
      monitorRequest={model.monitorRequest}
      onMonitorRequestHandled={model.clearMonitorRequest}
    />
  );
}
