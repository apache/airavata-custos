// PostCSS pipeline for the portal. Tailwind v4 is CSS-first and ships its
// own PostCSS plugin, so this is the only entry needed (no separate
// autoprefixer/tailwindcss plugins as in v3).
export default {
  plugins: {
    "@tailwindcss/postcss": {},
  },
};
