import { defineConfig } from "@playwright/test";

if (!process.env.XKVM_URL) {
  throw new Error("XKVM_URL environment variable is required");
}

export default defineConfig({
  testDir: "./e2e",
  timeout: 60000,
  workers: 1,
  reporter: "list",
  use: {
    baseURL: process.env.XKVM_URL,
    trace: "retain-on-failure",
    video: "retain-on-failure",
    screenshot: "only-on-failure",
  },
});
