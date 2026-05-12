/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        'warm-bg': '#f6f5f4',
        'warm-text': 'rgba(0,0,0,0.95)',
        'warm-secondary': '#615d59',
        'warm-muted': '#a39e98',
        'warm-blue': '#0075de',
        'warm-blue-bg': '#f2f9ff',
        'warm-blue-text': '#097fe8',
      },
      fontFamily: {
        sans: ['Inter', '-apple-system', 'BlinkMacSystemFont', 'Segoe UI', 'sans-serif'],
      },
    },
  },
  plugins: [],
}
