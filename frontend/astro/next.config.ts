import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  // Enable standalone output for Docker deployment
  output: "standalone",
  
  // Disable telemetry
  telemetry: false,
  
  // Image optimization for external domains (if needed)
  images: {
    unoptimized: true,
  },
};

export default nextConfig;
