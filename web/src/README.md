# Frontend source

- `App.vue` composes the application shell and mounts the active route immediately, so cached content is not delayed by exit or entry transitions. Route keys retain the existing reset behavior.
- `env.d.ts` declares Vite and Vue build-time types.
- `main.ts` installs browser compatibility fallbacks, application plugins, and mounts Vue.
