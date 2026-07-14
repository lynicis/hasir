// path: apps/dashboard/eslint.config.mjs
import hasirPreset from "@hasir/config/eslint";

const eslintConfig = [
  ...hasirPreset,
  // dashboard-specific rules could be layered here if needed
];

export default eslintConfig;
