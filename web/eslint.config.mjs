import nextCoreWebVitals from "eslint-config-next/core-web-vitals";
import nextTypescript from "eslint-config-next/typescript";

const eslintConfig = [
  ...nextCoreWebVitals,
  ...nextTypescript,
  { settings: { react: { version: "19.2.7" } } },
  { ignores: [".next/**", "node_modules/**", "pkg/**", "next-env.d.ts"] },
];

export default eslintConfig;
