import swaggerAutogen from "swagger-autogen"

const doc = {
  info: {
    version: "1.0.0",
    title: "i9rfs HTTP (REST) API",
    description: "A Remote File System API Server",
  },
  servers: [
    {
      url: `http://localhost:${process.env.PORT || 8000}`,
      description: "Development server",
    },
  ],

  components: {},
}

const outputFile = "./swagger.json"
const routes = ["../src/index.ts"]

/* NOTE: If you are using the express Router, you must pass in the 'routes' only the 
root file where the route starts, such as index.js, app.js, routes.js, etc ... */

swaggerAutogen({ openapi: "3.0.0" })(outputFile, routes, doc)
