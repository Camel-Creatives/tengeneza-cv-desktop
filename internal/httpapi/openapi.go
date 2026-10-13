package httpapi

import _ "embed"

//go:embed openapi.yaml
var openAPISpec string

const swaggerHTML = `<!doctype html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Tengeneza CV API Docs</title><link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css"></head><body><div id="swagger-ui"></div><script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script><script>SwaggerUIBundle({url:'/docs/openapi.yaml',dom_id:'#swagger-ui',persistAuthorization:true,deepLinking:true})</script></body></html>`
