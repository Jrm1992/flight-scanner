"use client";

import { useAuth } from "@/modules/auth/AuthContext";
import Auth from "@/modules/auth";
import { useAppModel } from "@/modules/app/model";
import AppView from "@/modules/app/view";

export default function Home() {
  const auth = useAuth();
  const model = useAppModel();

  if (!auth.isAuthenticated) {
    return <Auth onLogin={auth.login} onRegister={auth.register} />;
  }

  return (
    <AppView
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
