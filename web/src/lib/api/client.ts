import createFetchClient from "openapi-fetch";
import createClient from "openapi-react-query";

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
