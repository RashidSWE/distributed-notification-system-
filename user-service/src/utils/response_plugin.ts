import type { FastifyPluginAsync } from "fastify";
import { successResponse, errorResponse } from "./response.js";

const responsePlugin: FastifyPluginAsync = async (fastify) => {
  fastify.decorate("response", {
    success: successResponse,
    error: errorResponse,
  });
};

export default responsePlugin;

// Type augment for Fastify instance
declare module "fastify" {
  interface FastifyInstance {
    response: {
      success: typeof successResponse;
      error: typeof errorResponse;
    };
  }
}
