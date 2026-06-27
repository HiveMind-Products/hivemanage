import createFetchClient from "openapi-fetch";
import createClient from "openapi-react-query";

/**
 * TODO: The openapi-ts migration is incomplete.
 *
 * Current state:
 *   - The generated types file (`./v1.d.ts`) and its source (`../docs/swagger.json`)
 *     do not exist in the repo, so the `paths` type below is hand-written by hand
 *     for only the two `/dash/system/*` endpoints we currently consume.
 *   - `$api` is used in exactly one place: the version badge in AppSidebar
 *     (`$api.useQuery("get", "/dash/system/version")`). All other data access still
 *     goes through `fetchApi` in `@/utils/http-util`.
 *
 * To finish the migration:
 *   1. Commit the API spec to `docs/swagger.json` (generated from the Go backend).
 *   2. Run `pnpm generate:api` to produce `./src/lib/api/v1.d.ts`.
 *   3. Replace the hand-written `paths` type below with `import type { paths } from "./v1"`.
 *   4. Incrementally route the remaining hooks (tokens, storage, datasets, logs, etc.)
 *      through the generated `$api`/`paths` instead of hand-written `fetchApi` URLs.
 *
 * Do NOT remove the dependency in the meantime — the version badge relies on it.
 */
type paths = {
  "/dash/system/version": {
    get: {
      responses: {
        200: {
          content: {
            "application/json": {
              current: string;
              latest: string;
              update_available: boolean;
              last_checked: string;
            };
          };
        };
      };
    };
  };
  "/dash/system/config": {
    get: {
      responses: {
        200: {
          content: {
            "application/json": {
              bucket_domain: string;
            };
          };
        };
      };
    };
  };
};

export const fetchClient = createFetchClient<paths>({
  baseUrl: "/api",
});

export const $api = createClient(fetchClient);
