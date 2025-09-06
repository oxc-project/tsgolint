// Import app name from constants file using path alias
import { APP_NAME } from "@constants/app";

// Function called formatAppInfo, which takes two arguments, the app name, and app version and returns a string
function formatAppInfo(appName: string, appVersion: string): string {
  return `${appName} v${appVersion}`;
}

// Export a new function called getAppInfo which has no args, but calls formatAppInfo with the app name/app version from the constant file and returns the value
export function getAppInfo(): string {
  // Since VERSION was deleted per feedback, using a default version
  return formatAppInfo(APP_NAME, "1.0.0");
}