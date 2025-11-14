import fp from 'fastify-plugin';
import fastifySwagger from '@fastify/swagger';
import fastifySwaggerUi from '@fastify/swagger-ui';
import path from 'path';
import { fileURLToPath } from 'url';
import YAML from 'yamljs';
import type { FastifyInstance } from 'fastify';

// Helper to get __dirname in ES modules
const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
// Assumes swagger.yaml is in the root directory
const rootDir = path.resolve(__dirname, '..'); 

// Load the YAML file synchronously
const swaggerDocument = YAML.load(path.join(rootDir, 'swagger.yaml'));

const swaggerPlugin = async (fastify: FastifyInstance) => {
    // 1. Register @fastify/swagger in 'static' mode
    await fastify.register(fastifySwagger, {
        mode: 'static',
        specification: {
            document: swaggerDocument,
        },
    });

    // 2. Register @fastify/swagger-ui to serve the documentation
    await fastify.register(fastifySwaggerUi, {
        routePrefix: '/docs', 
        uiConfig: {
            docExpansion: 'full',
            deepLinking: true,
        },
    });
};

export default (fp as any)(swaggerPlugin);
