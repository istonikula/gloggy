# config

Implements:
- cavekit-config.md R1 (Config file location and defaults)
- cavekit-config.md R2 (Invalid config handling)
- cavekit-config.md R3 (Forward compatibility)
- cavekit-config.md R4 (Themes)
- cavekit-config.md R5 (Field and display settings, incl. top-level `scrolloff`)

Build tasks: T-006, T-007, T-008, T-009, T-010, T-024, T-025, T-078, T-084, T-085, T-130, T-171, T-175, T-176, T-177, T-178, T-180 (build-site.md)

Theme palette fidelity: per-theme palette struct vars + upstream-source citation constants (`TokyoNightSource`, `CatppuccinMochaSource`, `MaterialDarkSource`, discoverable via `theme.Source(name)`) make canonical-vs-local drift visible at review time. See `context/refs/research-brief-theme-palettes.md` for the canonical palette references these constants point to.
