# Integration architecture

Each child package owns one upstream protocol and exposes only the operations required by application services. Application calls enter provider-specific queues before reaching these clients. Adapter tests pin methods, paths, authentication, response limits, and secret-redaction behavior to the local upstream references.
