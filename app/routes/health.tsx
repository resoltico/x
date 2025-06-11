import { json } from "@remix-run/node";
import type { LoaderFunctionArgs } from "@remix-run/node";

// This health endpoint is handled by the Express server
// This route is here just as a fallback
export async function loader({ request }: LoaderFunctionArgs) {
  return json({
    status: "healthy",
    message: "This is the Remix health route. For full health status, the Express server provides /health",
    timestamp: new Date().toISOString()
  });
}
