import { defineConfig, loadEnv } from 'vite';
import react from '@vitejs/plugin-react';
// https://vitejs.dev/config/
export default defineConfig(function (_a) {
    var mode = _a.mode;
    // Load env file based on mode
    var env = loadEnv(mode, process.cwd(), '');
    return {
        plugins: [react()],
        define: {
            'import.meta.env.VITE_TAILSCALE_AUTH_KEY': JSON.stringify(env.VITE_TAILSCALE_AUTH_KEY || ''),
            'import.meta.env.VITE_TAILA2A_URL': JSON.stringify(env.VITE_TAILA2A_URL || 'http://localhost:8080'),
            'import.meta.env.VITE_TAILFS_URL': JSON.stringify(env.VITE_TAILFS_URL || 'http://localhost:8081'),
            'import.meta.env.VITE_DEBUG': JSON.stringify(env.VITE_DEBUG || 'false'),
            'import.meta.env.VITE_REFRESH_INTERVAL': JSON.stringify(env.VITE_REFRESH_INTERVAL || '30'),
        },
        server: {
            port: 3000,
            proxy: {
                '/api/taila2a': {
                    target: env.VITE_TAILA2A_URL || 'http://localhost:8080',
                    changeOrigin: true,
                    rewrite: function (path) { return path.replace(/^\/api\/taila2a/, ''); },
                },
                '/api/tailfs': {
                    target: env.VITE_TAILFS_URL || 'http://localhost:8081',
                    changeOrigin: true,
                    rewrite: function (path) { return path.replace(/^\/api\/tailfs/, ''); },
                },
            },
        },
        test: {
            globals: true,
            environment: 'jsdom',
            setupFiles: './src/test/setup.ts',
        },
    };
});
