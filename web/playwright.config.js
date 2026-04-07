import { defineConfig } from "@playwright/test";

const externalBaseURL = process.env.PLAYWRIGHT_BASE_URL
  ? String(process.env.PLAYWRIGHT_BASE_URL)
  : "";
const skipWebServer = process.env.PLAYWRIGHT_SKIP_WEBSERVER === "1" || Boolean(externalBaseURL);

export default defineConfig({
  testDir: "./e2e",
  timeout: 30_000,
  use: {
    baseURL: externalBaseURL || "http://127.0.0.1:8180",
    headless: true,
  },
  webServer: skipWebServer
    ? undefined
    : {
        command: "python3 -m http.server 8180 -d .",
        port: 8180,
        reuseExistingServer: false,
      },
});
