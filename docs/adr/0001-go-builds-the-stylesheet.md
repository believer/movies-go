# Go builds the stylesheet

CSS lives next to what it styles: tokens and base styles in `styles/`, component CSS beside its component, view CSS beside its view. The app embeds these files and joins them into one stylesheet at startup, putting each file in the cascade layer its directory implies. It serves the result at a URL versioned by the content's hash. We chose this over a build step because it adds no dependencies, leaves no generated file to commit or let go stale, and Go already owned the stylesheet's hash. A CSS change in dev reaches the browser through the watcher's normal restart.

## Considered Options

- **Keep the make step and rerun it in the dev watcher.** This fixes the dev loop, but the generated stylesheet stays committed and the production image ships whatever was last generated.
- **Native CSS `@import ... layer()`.** This needs no build at all, but the browser fetches the files in a request waterfall, and imported files can't carry the cache-busting version under a one-year cache.
- **esbuild's Go API as a bundler.** It would add minification, but it's a new dependency and its `@scope` support was unverified. It remains an option if minification is wanted later.
- **One stylesheet file.** This is the simplest option, but it gives up keeping a component's styles with its markup.
