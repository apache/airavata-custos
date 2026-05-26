// Ambient declaration so `import "./globals.css"` typechecks under
// `tsc --noEmit` without next-env handling the side-effect import.
declare module "*.css";
