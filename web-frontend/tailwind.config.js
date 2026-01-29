/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      minHeight: {
        'dvh': '100dvh',
      },
      height: {
        'dvh': '100dvh',
      },
    },
  },
  plugins: [],
}

