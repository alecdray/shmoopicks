# Wax Wiki

> Source of truth for product, architecture, and direction.

## Graph

```mermaid
graph TD
    VISION[vision]
    FEATURES[features]
    ROADMAP[roadmap]
    DATA_MODEL[data-model]
    ARCHITECTURE[architecture]
    INTEGRATIONS[integrations]
    FRONTEND[frontend]

    VISION --> FEATURES
    VISION --> ROADMAP
    FEATURES --> DATA_MODEL
    FEATURES --> FRONTEND
    ROADMAP --> FEATURES
    DATA_MODEL --> ARCHITECTURE
    DATA_MODEL --> FEATURES
    ARCHITECTURE --> INTEGRATIONS
    ARCHITECTURE --> FRONTEND
    INTEGRATIONS --> DATA_MODEL
    FRONTEND --> ARCHITECTURE
```

## Nodes

| Node | Description |
|---|---|
| [vision](./pages/vision.md) | What Wax is, why it exists, design philosophy |
| [features](./pages/features.md) | Shipped features and their scope |
| [roadmap](./pages/roadmap.md) | Planned features and future direction |
| [data-model](./pages/data-model.md) | Core domain entities and their relationships |
| [architecture](./pages/architecture.md) | System structure, modules, and key patterns |
| [integrations](./pages/integrations.md) | External services and how they connect |
| [frontend](./pages/frontend.md) | UX approach, rendering model, styling |
