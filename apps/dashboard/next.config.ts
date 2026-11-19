import type { NextConfig } from "next";

import path from "path";

const nextConfig: NextConfig = {
  output: "standalone",
  allowedDevOrigins: process.env.NODE_ENV === "development" ? ["*.local", "*.lan"] : [],
  turbopack: {
    root: path.join(import.meta.dirname, "../.."),
  },
};

export default nextConfig;
