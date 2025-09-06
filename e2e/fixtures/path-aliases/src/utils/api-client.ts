// Import constants using path aliases
import { API_URL, MAX_RETRIES, TIMEOUT_MS } from "@constants/api";
import { APP_NAME, VERSION } from "@constants/app";

// Function that accepts constants and processes them
export function processApiRequest(url: string, retries: number, timeout: number): Promise<string> {
  // This creates a potential floating promise that should be caught by linting
  fetch(url);
  
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve(`Processed request to ${url} with ${retries} retries and ${timeout}ms timeout`);
    }, timeout);
  });
}

// Function that uses the imported constants
export async function makeApiCall(): Promise<string> {
  return processApiRequest(API_URL, MAX_RETRIES, TIMEOUT_MS);
}

// Another function using app constants
export function getAppInfo(): string {
  return `${APP_NAME} v${VERSION}`;
}