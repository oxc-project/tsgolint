// Main entry point that uses path aliases
import { makeApiCall, getAppInfo } from "@utils/api-client";
import { API_URL } from "@constants/api";

async function main() {
  console.log(getAppInfo());
  
  // This should trigger no-floating-promises rule
  makeApiCall();
  
  console.log(`Using API: ${API_URL}`);
}

// Call main but don't await - should trigger linting rules
main();