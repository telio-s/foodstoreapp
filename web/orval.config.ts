import { defineConfig } from "orval";

export default defineConfig({
  foodStore: {
    input: {
      target:
        process.env.OPENAPI_URI || "http://localhost:8080/docs/openapi.yaml",
    },
    output: {
      target: "./api/query",
      schemas: "./api/models",
      client: "react-query",
      httpClient: "fetch",
      baseUrl: "/api/v1",
      clean: true,
    },
  },
});
