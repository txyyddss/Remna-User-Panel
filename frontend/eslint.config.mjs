import js from "@eslint/js";
import eslintConfigPrettier from "eslint-config-prettier";
import svelte from "eslint-plugin-svelte";
import globals from "globals";

/** @type {import("eslint").Linter.Config[]} */
export default [
  {
    ignores: ["**/node_modules/**", "../internal/webassets/templates/**"],
  },
  js.configs.recommended,
  ...svelte.configs["flat/base"],
  {
    files: ["src/**/*.{js,svelte}", "scripts/**/*.mjs", "tests/**/*.js"],
    rules: {
      "no-unused-vars": [
        "warn",
        {
          argsIgnorePattern: "^_",
          varsIgnorePattern: "^_",
          caughtErrorsIgnorePattern: "^_",
        },
      ],
      "no-empty": "warn",
    },
  },
  {
    files: ["src/**/*.svelte"],
    rules: {
      "no-useless-assignment": "off",
    },
  },
  {
    files: ["src/**/*.{js,svelte}"],
    languageOptions: {
      globals: {
        ...globals.browser,
        ...globals.es2021,
      },
      sourceType: "module",
    },
  },
  {
    files: ["scripts/**/*.mjs"],
    languageOptions: {
      globals: {
        ...globals.node,
        ...globals.es2021,
      },
      sourceType: "module",
    },
  },
  {
    files: ["tests/**/*.js"],
    languageOptions: {
      globals: {
        ...globals.browser,
        ...globals.es2021,
      },
      sourceType: "module",
    },
  },
  eslintConfigPrettier,
];
