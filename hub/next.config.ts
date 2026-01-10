import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: "export",
  basePath: "/hub",
  images: {
    unoptimized: true,
  },
  reactCompiler: true,
};

export default nextConfig;
