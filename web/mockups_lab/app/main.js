import { startFinanceOverview } from "./boot/startFinanceOverview.js";
import { bindSidebarNavigation } from "./shell/sidebar.js";

function resolveAppBasePath(scope = window) {
  const pathname = String((scope.location && scope.location.pathname) || "/");
  if (pathname.endsWith("/") || pathname.endsWith("/index.html")) {
    return "./mockups_lab/";
  }
  return "./";
}

const router = bindSidebarNavigation(document);
router.show("overview");

const appBasePath = resolveAppBasePath(window);
window.__FO_APP_BASE_PATH__ = appBasePath;
if (document.body && document.body.dataset) {
  document.body.dataset.foAppBasePath = appBasePath;
}

startFinanceOverview({
  variant: "mockup1",
  scope: window,
  doc: document,
  bootConfig: {
    basePath: appBasePath,
  },
});
