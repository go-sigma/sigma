import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react-swc';

export default defineConfig({
  plugins: [react()],
  build: {
    chunkSizeWarningLimit: 2048,
    minify: "terser",
    target: "esnext",
    rollupOptions: {
      output: {
        manualChunks(id: any): string {
          if (id.includes("node_modules")) {
            let name = id.toString().split("node_modules/")[1].split("/")[0].toString();
            if (name.includes("@monaco-editor")) {
              return "vendor.1";
            }
            if (id.includes("monaco-editor")) {
              if (id.includes("basic-languages")) {
                return name + "-" + "basic-languages"
              } else if (id.includes("browser/ui") || id.includes("editor/common")) {
                return name + "-" + "ui"
              } else if (id.includes("platform/") || id.includes("base/")) {
                return name + "-" + "common"
              } else {
                return name + "-" + "other"
              }
            }
            if (name.includes("flowbit")) {
              return "flowbit";
            }
            if (name.includes("react") || name.includes("tailwind")) {
              return "react+tailwind";
            }
            if (name.includes("parse5") || name.includes("bytemd") || name.includes("axios") || name.includes("codemirror") || name.includes("xterm")) {
              return "vendor-modules";
            }
            return "vendor";
          }
        },
      }
    }
  }
});
