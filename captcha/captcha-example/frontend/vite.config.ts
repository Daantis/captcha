import { resolve } from 'node:path'
import { defineConfig } from 'vite'
import { viteSingleFile } from "vite-plugin-singlefile";

export default defineConfig({
    build: {
        outDir: "../internal/captcha/html",
        chunkSizeWarningLimit: 1000,
        cssCodeSplit: false,
        modulePreload: false,
        target: "es2020",
        emptyOutDir: true,
        rollupOptions: {
            input: {
                "odd-grid": resolve(__dirname, "odd-grid.html"),
                "reality-swipe": resolve(__dirname, "reality-swipe.html"),
                "foreign-letter": resolve(__dirname, "foreign-letter.html"),
                "two-baskets": resolve(__dirname, "two-baskets.html"),
                "track-object": resolve(__dirname, "track-object.html"),
            },
        },
    },
    plugins: [
        viteSingleFile({
            useRecommendedBuildConfig: false,
        })
    ]
})
