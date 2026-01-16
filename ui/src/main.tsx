import { lazy } from "react";
import ReactDOM from "react-dom/client";
import {
  createBrowserRouter,
  isRouteErrorResponse,
  redirect,
  type RouteObject,
  RouterProvider,
  useRouteError,
} from "react-router";
import "./index.css";
import { ExclamationTriangleIcon } from "@heroicons/react/16/solid";

import { initTestHooks } from "@/test/testHooks";
import { CLOUD_API, CLOUD_ENABLE_VERSIONED_UI, getDeviceAPI } from "@/ui.config";
import api from "@/api";
import Root from "@/root";
import { m } from "@localizations/messages.js";
import Card from "@components/Card";
import EmptyCard from "@components/EmptyCard";
import NotFoundPage from "@components/NotFoundPage";
import DeviceRoute, { LocalDevice } from "@routes/devices.$id";
import WelcomeRoute, { DeviceStatus } from "@routes/welcome-local";
import LoginLocalRoute from "@routes/login-local";
import WelcomeLocalModeRoute from "@routes/welcome-local.mode";
import WelcomeLocalPasswordRoute from "@routes/welcome-local.password";
import SetupRoute from "@routes/devices.$id.setup";
import DeviceIdRename from "@routes/devices.$id.rename";
import DevicesRoute from "@routes/devices";
import SettingsIndexRoute from "@routes/devices.$id.settings._index";
import SettingsAccessIndexRoute from "@routes/devices.$id.settings.access._index";
import Notifications from "@/notifications";
const SignupRoute = lazy(() => import("@routes/signup"));
const LoginRoute = lazy(() => import("@routes/login"));
const OtherSessionRoute = lazy(() => import("@routes/devices.$id.other-session"));
const MountRoute = lazy(() => import("@routes/devices.$id.mount"));
const SettingsRoute = lazy(() => import("@routes/devices.$id.settings"));
const SettingsMouseRoute = lazy(() => import("@routes/devices.$id.settings.mouse"));
const SettingsKeyboardRoute = lazy(() => import("@routes/devices.$id.settings.keyboard"));
const SettingsAdvancedRoute = lazy(() => import("@routes/devices.$id.settings.advanced"));
const SettingsVideoRoute = lazy(() => import("@routes/devices.$id.settings.video"));
const SettingsAppearanceRoute = lazy(() => import("@routes/devices.$id.settings.appearance"));
const SettingsGeneralIndexRoute = lazy(() => import("@routes/devices.$id.settings.general._index"));
// OTA update route disabled
// const SettingsGeneralUpdateRoute = lazy(
//   () => import("@routes/devices.$id.settings.general.update"),
// );
const SettingsNetworkRoute = lazy(() => import("@routes/devices.$id.settings.network"));
const SecurityAccessLocalAuthRoute = lazy(
  () => import("@routes/devices.$id.settings.access.local-auth"),
);
const SettingsMacrosRoute = lazy(() => import("@routes/devices.$id.settings.macros"));
const SettingsMacrosAddRoute = lazy(() => import("@routes/devices.$id.settings.macros.add"));
const SettingsMacrosEditRoute = lazy(() => import("@routes/devices.$id.settings.macros.edit"));
const SettingsConnectionsRoute = lazy(() => import("@routes/settings.connections"));
const NativeSetupRoute = lazy(() => import("@routes/native-setup"));

export const isOnDevice = import.meta.env.MODE === "device";
export const isNative = import.meta.env.MODE === "tauri";
export const isInCloud = !isOnDevice && !isNative;

// Initialize E2E test hooks (safe to call in all environments)
initTestHooks();

// Initialize native config if in Tauri mode
if (isNative) {
  // Load config immediately and wait for it
  const configPromise = import('@/stores/nativeConfigStore').then(async ({ useNativeConfig }) => {
    try {
      await useNativeConfig.getState().loadConfig();
      console.log('Config loaded successfully');
    } catch (error) {
      console.error('Failed to load config on startup:', error);
      // Don't throw - let the app handle it in checkDeviceAuth
    }
  });
  
  // Store the promise so we can wait for it in checkDeviceAuth
  (window as any).__configLoadPromise = configPromise;
}

export async function checkCloudAuth() {
  const res = await fetch(`${CLOUD_API}/me`, {
    mode: "cors",
    credentials: "include",
    headers: { "Content-Type": "application/json" },
  });

  if (res.status === 401) {
    throw redirect(`/login?returnTo=${window.location.href}`);
  }

  return await res.json();
}

export async function checkDeviceAuth() {
  // Wait for config to load in native mode
  if (isNative && (window as any).__configLoadPromise) {
    try {
      await (window as any).__configLoadPromise;
    } catch (error) {
      console.error('Failed to load config:', error);
    }
  }
  
  const deviceAPI = getDeviceAPI();
  
  // If no backend configured yet in native mode, return a special state
  // The app will handle this by showing a setup screen
  if (isNative && !deviceAPI) {
    return { authMode: null, needsSetup: true };
  }
  
  const res = await api
    .GET(`${deviceAPI}/device/status`)
    .then(res => res.json() as Promise<DeviceStatus>);

  if (!res.isSetup) throw redirect("/welcome");

  const deviceRes = await api.GET(`${deviceAPI}/device`);
  if (deviceRes.status === 401) throw redirect("/login-local");
  if (deviceRes.ok) {
    const device = (await deviceRes.json()) as LocalDevice;
    return { authMode: device.authMode };
  }

  throw new Error("Error fetching device");
}

export async function checkAuth() {
  return isOnDevice || isNative ? checkDeviceAuth() : checkCloudAuth();
}

let router: ReturnType<typeof createBrowserRouter>;

const getDeviceRoute = (r: Omit<RouteObject, "children" | "index">): RouteObject => {
  const settingsChildren: RouteObject[] = [
    {
      index: true,
      loader: SettingsIndexRoute.loader,
    },
    {
      path: "general",
      children: [
        {
          index: true,
          element: <SettingsGeneralIndexRoute />,
        },
      ],
    },
    {
      path: "mouse",
      element: <SettingsMouseRoute />,
    },
    {
      path: "keyboard",
      element: <SettingsKeyboardRoute />,
    },
    {
      path: "advanced",
      element: <SettingsAdvancedRoute />,
    },
    {
      path: "network",
      element: <SettingsNetworkRoute />,
    },
    {
      path: "access",
      children: [
        {
          index: true,
          element: <SettingsAccessIndexRoute />,
          loader: SettingsAccessIndexRoute.loader,
        },
        {
          path: "local-auth",
          element: <SecurityAccessLocalAuthRoute />,
        },
      ],
    },
    {
      path: "video",
      element: <SettingsVideoRoute />,
    },
    {
      path: "appearance",
      element: <SettingsAppearanceRoute />,
    },
    {
      path: "macros",
      children: [
        {
          index: true,
          element: <SettingsMacrosRoute />,
        },
        {
          path: "add",
          element: <SettingsMacrosAddRoute />,
        },
        {
          path: ":macroId/edit",
          element: <SettingsMacrosEditRoute />,
        },
      ],
    },
  ];
  
  // Add connections route only for native mode
  if (isNative) {
    settingsChildren.push({
      path: "connections",
      element: <SettingsConnectionsRoute />,
    });
  }
  
  return {
    element: <DeviceRoute />,
    loader: DeviceRoute.loader,
    ...r,
    children: [
      {
        path: "other-session",
        element: <OtherSessionRoute />,
      },
      {
        path: "mount",
        element: <MountRoute />,
      },
      {
        path: "settings",
        element: <SettingsRoute />,
        children: settingsChildren,
      },
    ],
  };
};

if (isOnDevice || isNative) {
  const routes: RouteObject[] = [
    {
      path: "/welcome/mode",
      element: <WelcomeLocalModeRoute />,
      action: WelcomeLocalModeRoute.action,
    },
    {
      path: "/welcome/password",
      element: <WelcomeLocalPasswordRoute />,
      action: WelcomeLocalPasswordRoute.action,
    },
    {
      path: "/welcome",
      element: <WelcomeRoute />,
      loader: WelcomeRoute.loader,
    },
    {
      path: "/login-local",
      element: <LoginLocalRoute />,
      action: LoginLocalRoute.action,
      loader: LoginLocalRoute.loader,
    },
    getDeviceRoute({
      path: "/",
      errorElement: <ErrorBoundary />,
      HydrateFallback: () => <div className="p-4">{m.loading()}</div>,
    }),
  ];
  
  // Add native setup route for first-run
  if (isNative) {
    routes.unshift({
      path: "/setup",
      element: <NativeSetupRoute />,
    });
  }
  
  router = createBrowserRouter(routes);
} else {
  const routeObjects: RouteObject[] = [
    {
      errorElement: <ErrorBoundary />,
      children: [
        { path: "signup", element: <SignupRoute /> },
        { path: "login", element: <LoginRoute /> },
        {
          path: "/",
          element: <Root />,
          children: [
            {
              index: true,
              loader: async () => {
                await checkAuth();
                return redirect(`/devices`);
              },
            },
            {
              path: "devices/:id/setup",
              element: <SetupRoute />,
              action: SetupRoute.action,
              loader: SetupRoute.loader,
            },
            getDeviceRoute({
              path: "devices/:id",
            }),
            {
              path: "devices/:id/rename",
              element: <DeviceIdRename />,
              loader: DeviceIdRename.loader,
              action: DeviceIdRename.action,
            },
            {
              path: "devices",
              element: <DevicesRoute />,
              loader: DevicesRoute.loader,
            },
          ],
        },
      ],
    },
  ];

  // if versioned UI is not enabled, we need to add a route that redirects to the non-versioned route
  if (!CLOUD_ENABLE_VERSIONED_UI) {
    routeObjects.unshift({
      path: "v/:version/*",
      element: <Root />,
      loader: async ({ params }) => {
        throw redirect(`/${params["*"]}`);
      },
    });
  }

  router = createBrowserRouter(routeObjects, { basename: import.meta.env.BASE_URL });
}

document.addEventListener("DOMContentLoaded", () => {
  ReactDOM.createRoot(document.getElementById("root")!).render(
    <>
      <RouterProvider router={router} />
      <Notifications
        toastOptions={{
          className:
            "rounded-sm border-none bg-white text-black shadow-sm outline-1 outline-slate-800/30",
        }}
        max={2}
      />
    </>,
  );
});

// eslint-disable-next-line react-refresh/only-export-components
function ErrorBoundary() {
  const error = useRouteError();
  if (isRouteErrorResponse(error)) {
    if (error.status === 404) return <NotFoundPage />;
  }

  const getErrorMessage = (err: unknown): string | null => {
    // If it's a route error response, try to read a string at err.data.error.message or err.data.error safely
    if (isRouteErrorResponse(err)) {
      const data = (err as { data?: unknown }).data;
      if (data && typeof data === "object") {
        const maybeError = (data as Record<string, unknown>)["error"];
        if (maybeError) {
          if (typeof maybeError === "object") {
            const msg = (maybeError as Record<string, unknown>)["message"];
            if (typeof msg === "string") return msg;
          } else if (typeof maybeError === "string") {
            return maybeError;
          }
        }
      }
    }

    // Fallback: check plain object message property
    if (err && typeof err === "object") {
      const maybeMsg = (err as Record<string, unknown>)["message"];
      if (typeof maybeMsg === "string") return maybeMsg;
    }

    return null;
  };

  const errorMessage = getErrorMessage(error);

  return (
    <div className="h-full w-full">
      <div className="flex h-full items-center justify-center">
        <div className="w-full max-w-2xl">
          <EmptyCard
            IconElm={ExclamationTriangleIcon}
            headline={m.oh_no()}
            description={m.something_went_wrong()}
            BtnElm={
              errorMessage && (
                <Card>
                  <div className="flex items-center font-mono">
                    <div className="flex p-2 text-black dark:text-white">
                      <span className="text-sm">{errorMessage}</span>
                    </div>
                  </div>
                </Card>
              )
            }
          />
        </div>
      </div>
    </div>
  );
}
